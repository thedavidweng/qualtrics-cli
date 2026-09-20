package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
)

var (
	exportFormat    string
	exportUseLabels bool
	exportTimeZone  string
	exportCompress  bool
	exportBreakouts []string
	exportSeen      string
	exportOutput    string
	exportExtract   bool
	exportWait      bool
	exportInterval  time.Duration
	importFile      string
)

var responsesCmd = &cobra.Command{
	Use:   "responses",
	Short: "Export and import survey responses",
}

var responsesExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Manage response export jobs",
}

var responsesExportStartCmd = &cobra.Command{
	Use:   "start <surveyId>",
	Short: "Start a response export job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		req := qualtrics.ExportRequest{
			SurveyID:             args[0],
			Format:               exportFormat,
			TimeZone:             exportTimeZone,
			BreakoutSets:         exportBreakouts,
			SeenUnansweredRecode: exportSeen,
		}
		if cmd.Flags().Changed("use-labels") {
			v := exportUseLabels
			req.UseLabels = &v
		}
		if cmd.Flags().Changed("compress") {
			v := exportCompress
			req.Compress = &v
		}

		var started qualtrics.ExportStart
		runMutation(cmd, "responses.export.start", "failed to start export", safety.TierRemoteAction,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[0],
					planAfter:  req,
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						s, err := client.StartExport(ctx, &req)
						if err != nil {
							return nil, err
						}
						started = s
						return s, nil
					},
					human: func() {
						fmt.Printf("export started: %s\n", started.ProgressID)
					},
				}, nil
			})

		if !exportWait || dryRun || started.ProgressID == "" {
			return
		}
		waitExport(cmd, args[0], started.ProgressID)
	},
}

func waitExport(cmd *cobra.Command, surveyID, progressID string) {
	start := time.Now()
	renderer := output.NewRenderer(nil, nil, jsonMode, pretty)
	deps, ok := newDeps(renderer, "responses.export.wait", start)
	if !ok {
		return
	}

	ctx := cmd.Context()
	for {
		status, err := deps.Client.GetExport(ctx, progressID)
		if err != nil {
			handleError(renderer, "responses.export.wait", wrapError(err, "export failed"), start)
			return
		}
		if !jsonMode {
			fmt.Printf("\rprogress: %.0f%% (%s)   ", status.PercentComplete, status.Status)
		}
		switch status.Status {
		case qualtrics.JobComplete:
			if !jsonMode {
				fmt.Println()
			}
			downloadExport(ctx, renderer, deps, surveyID, progressID, start)
			return
		case qualtrics.JobFailed:
			handleError(renderer, "responses.export.wait", errors.New(errors.APIError, fmt.Sprintf("export %s failed", progressID), errors.CatAPI, false, nil), start)
			return
		}
		select {
		case <-ctx.Done():
			handleError(renderer, "responses.export.wait", errors.New(errors.NetworkTimeout, "canceled while waiting for export", errors.CatNetwork, false, ctx.Err()), start)
			return
		case <-time.After(exportInterval):
		}
	}
}

func downloadExport(ctx context.Context, renderer *output.Renderer, deps CommandDeps, surveyID, progressID string, start time.Time) {
	data, filename, err := deps.Client.DownloadExport(ctx, progressID)
	if err != nil {
		handleError(renderer, "responses.export.download", wrapError(err, "download failed"), start)
		return
	}
	outPath := exportOutput
	if outPath == "" {
		base := strings.TrimSuffix(filename, ".zip")
		if base == "" {
			base = surveyID + "_" + progressID
		}
		outPath = base + ".zip"
	}
	if err := writeFile(outPath, data); err != nil {
		handleError(renderer, "responses.export.download", wrapError(err, "failed to write file"), start)
		return
	}
	if exportExtract {
		if err := extractZip(outPath); err != nil {
			handleError(renderer, "responses.export.download", wrapError(err, "failed to extract"), start)
			return
		}
	}
	payload := map[string]any{
		"progress_id": progressID,
		"bytes":       len(data),
		"path":        outPath,
		"extracted":   exportExtract,
	}
	if jsonMode {
		renderer.RenderSuccess(output.NewEnvelope("responses.export.download", profile, output.SchemaVersion, requestID, payload, time.Since(start)))
		return
	}
	fmt.Printf("wrote %d bytes to %s\n", len(data), outPath)
}

var responsesExportStatusCmd = &cobra.Command{
	Use:   "status <progressId>",
	Short: "Check an export job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "responses.export.status", "failed to get export status",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetExport(ctx, args[0])
			},
			func(data any) {
				s, _ := data.(qualtrics.ExportStatus)
				fmt.Printf("status: %s (%.0f%% complete)\n", s.Status, s.PercentComplete)
			})
	},
}

var responsesExportDownloadCmd = &cobra.Command{
	Use:   "download <progressId> -o <file.zip>",
	Short: "Download a completed export",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "responses.export.download", "failed to download export",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				data, filename, err := client.DownloadExport(ctx, args[0])
				if err != nil {
					return nil, err
				}
				outPath := exportOutput
				if outPath == "" {
					base := strings.TrimSuffix(filename, ".zip")
					if base == "" {
						base = args[0]
					}
					outPath = base + ".zip"
				}
				if err := writeFile(outPath, data); err != nil {
					return nil, err
				}
				if exportExtract {
					if err := extractZip(outPath); err != nil {
						return nil, err
					}
				}
				return map[string]any{
					"progress_id": args[0],
					"bytes":       len(data),
					"path":        outPath,
					"extracted":   exportExtract,
				}, nil
			},
			func(data any) {
				m, _ := data.(map[string]any)
				fmt.Printf("wrote %v bytes to %v\n", m["bytes"], m["path"])
			})
	},
}

var responsesImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Manage response import jobs",
}

var responsesImportStartCmd = &cobra.Command{
	Use:   "start <surveyId>",
	Short: "Start a response import job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "responses.import.start", "failed to start import", safety.TierRemoteAction,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[0],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.StartImport(ctx, args[0])
					},
					human: func() { fmt.Println("import started") },
				}, nil
			})
	},
}

var responsesImportStatusCmd = &cobra.Command{
	Use:   "status <progressId>",
	Short: "Check an import job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "responses.import.status", "failed to get import status",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetImport(ctx, args[0])
			},
			func(data any) {
				s, _ := data.(qualtrics.ImportStatus)
				fmt.Printf("status: %s (%.0f%% complete)\n", s.Status, s.PercentComplete)
			})
	},
}

var responsesImportUploadCmd = &cobra.Command{
	Use:   "upload <progressId> -f <responses.csv>",
	Short: "Upload a response file for an import job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "responses.import.upload", "failed to upload responses", safety.TierRemoteAction,
			func() (mutation, *errors.Error) {
				if importFile == "" {
					return mutation{}, errors.New(errors.InvalidArguments, "-f <file> is required", errors.CatValidation, false, nil)
				}
				content, err := readPayload(importFile)
				if err != nil {
					return mutation{}, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
				}
				name := filepath.Base(importFile)
				return mutation{
					resourceID: args[0],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UploadImport(ctx, args[0], content, name)
					},
					human: func() { fmt.Println("uploaded") },
				}, nil
			})
	},
}

func extractZip(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	dir := strings.TrimSuffix(path, ".zip") + "_extracted"
	for _, f := range reader.File {
		target := filepath.Join(dir, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dir)+string(filepath.Separator)) {
			return fmt.Errorf("zip entry escapes output dir: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	responsesCmd.AddCommand(responsesExportCmd)
	responsesCmd.AddCommand(responsesImportCmd)

	responsesExportCmd.AddCommand(responsesExportStartCmd)
	responsesExportCmd.AddCommand(responsesExportStatusCmd)
	responsesExportCmd.AddCommand(responsesExportDownloadCmd)

	responsesImportCmd.AddCommand(responsesImportStartCmd)
	responsesImportCmd.AddCommand(responsesImportStatusCmd)
	responsesImportCmd.AddCommand(responsesImportUploadCmd)

	responsesExportStartCmd.Flags().StringVar(&exportFormat, "format", "csv", "export format: csv, json, spss, tsv")
	responsesExportStartCmd.Flags().BoolVar(&exportUseLabels, "use-labels", false, "use choice text instead of coded values")
	responsesExportStartCmd.Flags().StringVar(&exportTimeZone, "timezone", "", "time zone for dates")
	responsesExportStartCmd.Flags().BoolVar(&exportCompress, "compress", false, "request compression")
	responsesExportStartCmd.Flags().StringArrayVar(&exportBreakouts, "breakout-set", nil, "breakout set ID (repeatable)")
	responsesExportStartCmd.Flags().StringVar(&exportSeen, "seen-unanswered-recode", "", "recoding value for seen but unanswered questions")
	responsesExportStartCmd.Flags().BoolVar(&exportWait, "wait", false, "poll until complete, then download")
	responsesExportStartCmd.Flags().DurationVar(&exportInterval, "interval", 5*time.Second, "poll interval for --wait")
	responsesExportStartCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file for --wait")
	responsesExportStartCmd.Flags().BoolVar(&exportExtract, "extract", false, "unzip after download")

	responsesExportDownloadCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file")
	responsesExportDownloadCmd.Flags().BoolVar(&exportExtract, "extract", false, "unzip after download")

	responsesImportUploadCmd.Flags().StringVarP(&importFile, "file", "f", "", "response file to upload")
}
