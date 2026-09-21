package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/qsf"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
)

var (
	defPayloadFile string
	defBlockID     string
	defDescription string
	qsfOutputPath  string
)

var definitionsCmd = &cobra.Command{
	Use:   "definitions",
	Short: "Read and edit survey definitions",
}

var definitionsShowCmd = &cobra.Command{
	Use:   "show <surveyId>",
	Short: "Show a survey definition summary",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.show", "failed to get definition",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				def, err := client.GetDefinition(ctx, args[0])
				if err != nil {
					return nil, err
				}
				if fullOutput {
					return def, nil
				}
				return def.Summary(), nil
			},
			func(data any) {
				if def, ok := data.(qualtrics.Definition); ok {
					printJSON(def)
					return
				}
				m, _ := data.(map[string]any)
				fmt.Printf("survey: %v (%v)\n", m["survey_name"], m["survey_id"])
				fmt.Printf("blocks: %v  questions: %v\n\n", m["blocks"], m["questions"])
				blocks, _ := m["per_block"].([]map[string]any)
				fmt.Printf("%-20s %-30s %s\n", "BLOCK", "DESCRIPTION", "QUESTIONS")
				for _, b := range blocks {
					fmt.Printf("%-20v %-30v %v\n", b["id"], b["description"], b["questions"])
				}
			})
	},
}

var definitionsExportCmd = &cobra.Command{
	Use:   "export <surveyId> -o <file>",
	Short: "Write the full survey definition JSON to a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.export", "failed to export definition",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				def, err := client.GetDefinition(ctx, args[0])
				if err != nil {
					return nil, err
				}
				data, err := json.MarshalIndent(def, "", "  ")
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"survey_id": args[0],
					"bytes":     len(data),
					"path":      defExportPath,
				}, writeFile(defExportPath, data)
			},
			func(data any) {
				if fullOutput {
					printJSON(data)
					return
				}
				m, _ := data.(map[string]any)
				fmt.Printf("wrote %v bytes to %v\n", m["bytes"], m["path"])
			})
	},
}

var defExportPath string

var definitionsImportCmd = &cobra.Command{
	Use:   "import -f <file>",
	Short: "Create a new survey from a definition JSON file",
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.import", "failed to import definition", safety.TierMutation,
			func() (mutation, *errors.Error) {
				if defPayloadFile == "" {
					return mutation{}, errors.New(errors.InvalidArguments, "-f <file> is required (use - for stdin)", errors.CatValidation, false, nil)
				}
				payload, err := readPayload(defPayloadFile)
				if err != nil {
					return mutation{}, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
				}
				if !json.Valid(payload) {
					return mutation{}, errors.New(errors.InvalidArguments, "payload is not valid JSON", errors.CatValidation, false, nil)
				}
				return mutation{
					resourceID: "",
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateFromDefinition(ctx, payload)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("survey imported")
					},
				}, nil
			})
	},
}

var questionsCmd = &cobra.Command{
	Use:   "questions",
	Short: "Manage questions in a survey definition",
}

var questionsListCmd = &cobra.Command{
	Use:   "list <surveyId>",
	Short: "List questions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.questions.list", "failed to list questions",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.ListQuestions(ctx, args[0])
			},
			func(data any) {
				if fullOutput {
					printJSON(data)
					return
				}
				questions, _ := data.([]qualtrics.Question)
				fmt.Printf("%-12s %-14s %-18s %s\n", "QID", "TYPE", "SELECTOR", "TEXT")
				for _, q := range questions {
					fmt.Printf("%-12s %-14s %-18s %s\n", q.QuestionID, q.QuestionType, q.Selector, truncate(q.QuestionText, 50))
				}
				fmt.Printf("\n%d questions\n", len(questions))
			})
	},
}

var questionsShowCmd = &cobra.Command{
	Use:   "show <surveyId> <questionId>",
	Short: "Show one question",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.questions.show", "failed to get question",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetQuestion(ctx, args[0], args[1])
			},
			func(data any) {
				q, _ := data.(qualtrics.Question)
				printJSON(q)
			})
	},
}

var questionsCreateCmd = &cobra.Command{
	Use:   "create <surveyId> -f <question.json>",
	Short: "Create a question from a Qualtrics question JSON payload",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.questions.create", "failed to create question", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requirePayload()
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[0],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateQuestion(ctx, args[0], defBlockID, payload)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("question created")
					},
				}, nil
			})
	},
}

var questionsUpdateCmd = &cobra.Command{
	Use:   "update <surveyId> <questionId> -f <question.json>",
	Short: "Replace a question",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.questions.update", "failed to update question", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requirePayload()
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[1],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateQuestion(ctx, args[0], args[1], payload)
					},
					human: func(data any) { fmt.Printf("updated %s\n", args[1]) },
				}, nil
			})
	},
}

var questionsDeleteCmd = &cobra.Command{
	Use:   "delete <surveyId> <questionId>",
	Short: "Delete a question",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.questions.delete", "failed to delete question", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[1],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteQuestion(ctx, args[0], args[1])
					},
					human: func(data any) { fmt.Printf("deleted %s\n", args[1]) },
				}, nil
			})
	},
}

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage blocks in a survey definition",
}

var blocksListCmd = &cobra.Command{
	Use:   "list <surveyId>",
	Short: "List blocks (extracted from the definition)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.blocks.list", "failed to list blocks",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				def, err := client.GetDefinition(ctx, args[0])
				if err != nil {
					return nil, err
				}
				return def.BlockList(), nil
			},
			func(data any) {
				if fullOutput {
					printJSON(data)
					return
				}
				blocks, _ := data.([]qualtrics.Block)
				fmt.Printf("%-20s %-12s %-10s %s\n", "BLOCK", "TYPE", "QUESTIONS", "DESCRIPTION")
				for _, b := range blocks {
					n := 0
					for _, el := range b.BlockElements {
						if el.Type == "Question" {
							n++
						}
					}
					fmt.Printf("%-20s %-12s %-10d %s\n", b.ID, b.Type, n, b.Description)
				}
				fmt.Printf("\n%d blocks\n", len(blocks))
			})
	},
}

var blocksShowCmd = &cobra.Command{
	Use:   "show <surveyId> <blockId>",
	Short: "Show one block",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.blocks.show", "failed to get block",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetBlock(ctx, args[0], args[1])
			},
			func(data any) {
				b, _ := data.(qualtrics.Block)
				printJSON(b)
			})
	},
}

var blocksCreateCmd = &cobra.Command{
	Use:   "create <surveyId>",
	Short: "Create a standard block",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.blocks.create", "failed to create block", safety.TierMutation,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[0],
					planAfter:  qualtrics.CreateBlockRequest{Description: defDescription, Type: "Standard"},
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return client.CreateBlock(ctx, args[0], defDescription)
					},
					human: func(data any) {
						if fullOutput {
							printJSON(data)
							return
						}
						fmt.Println("block created")
					},
				}, nil
			})
	},
}

var blocksUpdateCmd = &cobra.Command{
	Use:   "update <surveyId> <blockId> -f <block.json>",
	Short: "Replace a block",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.blocks.update", "failed to update block", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requirePayload()
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[1],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateBlock(ctx, args[0], args[1], payload)
					},
					human: func(data any) { fmt.Printf("updated %s\n", args[1]) },
				}, nil
			})
	},
}

var blocksDeleteCmd = &cobra.Command{
	Use:   "delete <surveyId> <blockId>",
	Short: "Delete a block",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.blocks.delete", "failed to delete block", safety.TierDestructive,
			func() (mutation, *errors.Error) {
				return mutation{
					resourceID: args[1],
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.DeleteBlock(ctx, args[0], args[1])
					},
					human: func(data any) { fmt.Printf("deleted %s\n", args[1]) },
				}, nil
			})
	},
}

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Show or replace the survey flow",
}

var flowShowCmd = &cobra.Command{
	Use:   "show <surveyId>",
	Short: "Show the survey flow",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.flow.show", "failed to get flow",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetFlow(ctx, args[0])
			},
			func(data any) { printJSON(data) })
	},
}

var flowUpdateCmd = &cobra.Command{
	Use:   "update <surveyId> -f <flow.json>",
	Short: "Replace the survey flow",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.flow.update", "failed to update flow", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requirePayload()
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[0],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateFlow(ctx, args[0], payload)
					},
					human: func(data any) { fmt.Println("flow updated") },
				}, nil
			})
	},
}

var optionsCmd = &cobra.Command{
	Use:   "options",
	Short: "Show or replace survey options",
}

var optionsShowCmd = &cobra.Command{
	Use:   "show <surveyId>",
	Short: "Show survey options",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "definitions.options.show", "failed to get options",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				return client.GetOptions(ctx, args[0])
			},
			func(data any) { printJSON(data) })
	},
}

var optionsUpdateCmd = &cobra.Command{
	Use:   "update <surveyId> -f <options.json>",
	Short: "Replace survey options",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runMutation(cmd, "definitions.options.update", "failed to update options", safety.TierMutation,
			func() (mutation, *errors.Error) {
				payload, verr := requirePayload()
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: args[0],
					planAfter:  json.RawMessage(payload),
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						return nil, client.UpdateOptions(ctx, args[0], payload)
					},
					human: func(data any) { fmt.Println("options updated") },
				}, nil
			})
	},
}

func requirePayload() ([]byte, *errors.Error) {
	if defPayloadFile == "" {
		return nil, errors.New(errors.InvalidArguments, "-f <file> is required (use - for stdin)", errors.CatValidation, false, nil)
	}
	payload, err := readPayload(defPayloadFile)
	if err != nil {
		return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
	}
	if !json.Valid(payload) {
		return nil, errors.New(errors.InvalidArguments, "payload is not valid JSON", errors.CatValidation, false, nil)
	}
	return payload, nil
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}
	fmt.Println(string(data))
}

var definitionsBuildCmd = &cobra.Command{
	Use:   "build <spec.md> [-o output.qsf]",
	Short: "Compile a Markdown survey spec to Qualtrics Survey Format (QSF)",
	Long: `Compile a plain-text markdown survey specification into a .qsf file
ready to import into Qualtrics via Create Project → Import a QSF File.
No API access required.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runLocal("definitions.build", func() (any, *errors.Error) {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			spec, err := qsf.ParseSurvey(string(content))
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			data, err := qsf.BuildQSF(spec)
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			outPath := qsfOutputPath
			if outPath == "" {
				base := strings.TrimSuffix(args[0], ".md")
				outPath = base + ".qsf"
			}
			jsonBytes, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return nil, errors.New(errors.InternalError, err.Error(), errors.CatInternal, false, err)
			}
			if err := os.WriteFile(outPath, jsonBytes, 0o600); err != nil {
				return nil, errors.New(errors.InternalError, err.Error(), errors.CatInternal, false, err)
			}
			totalQ := 0
			for _, b := range spec.Blocks {
				for _, q := range b.Questions {
					if !q.IsPageBreak {
						totalQ++
					}
				}
			}
			return map[string]any{
				"output":    outPath,
				"title":     spec.Title,
				"language":  spec.Language,
				"blocks":    len(spec.Blocks),
				"questions": totalQ,
			}, nil
		}, func(data any) {
			if fullOutput {
				printJSON(data)
				return
			}
			m, _ := data.(map[string]any)
			fmt.Printf("Compiled %s\n", m["output"])
			fmt.Printf("  Title:     %s\n", m["title"])
			fmt.Printf("  Language:  %s\n", m["language"])
			fmt.Printf("  Blocks:    %v\n", m["blocks"])
			fmt.Printf("  Questions: %v\n", m["questions"])
			fmt.Println("\nImport into Qualtrics: Create Project → Import a QSF File")
		})
	},
}

var qsfCmd = &cobra.Command{
	Use:   "qsf",
	Short: "Work with Qualtrics Survey Format (QSF) files offline",
}

var qsfSummaryCmd = &cobra.Command{
	Use:   "summary <file.qsf>",
	Short: "Display structural summary of a QSF file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runLocal("definitions.qsf.summary", func() (any, *errors.Error) {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			summary, err := qsf.SummarizeQSF(data)
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			return summary, nil
		}, func(data any) {
			if fullOutput {
				printJSON(data)
				return
			}
			s, _ := data.(*qsf.SummaryInfo)
			fmt.Printf("survey:    %s (%s)\n", s.SurveyName, s.SurveyID)
			fmt.Printf("language:  %s\n", s.Language)
			fmt.Printf("blocks:    %d\n", s.BlockCount)
			fmt.Printf("questions: %d\n\n", s.TotalQCount)
			fmt.Printf("%-20s %-12s %-10s %s\n", "BLOCK_ID", "TYPE", "QUESTIONS", "DESCRIPTION")
			for _, b := range s.Blocks {
				fmt.Printf("%-20s %-12s %-10d %s\n", b.ID, b.Type, b.Questions, b.Description)
			}
		})
	},
}

var qsfConvertCmd = &cobra.Command{
	Use:   "convert <file.qsf> [-o output.json]",
	Short: "Convert a QSF file into a survey definition JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runLocal("definitions.qsf.convert", func() (any, *errors.Error) {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			defBytes, err := qsf.ConvertQSFToDefinition(data)
			if err != nil {
				return nil, errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
			}
			outPath := qsfOutputPath
			if outPath == "" {
				base := strings.TrimSuffix(args[0], ".qsf")
				outPath = base + "_def.json"
			}
			if err := os.WriteFile(outPath, defBytes, 0o600); err != nil {
				return nil, errors.New(errors.InternalError, err.Error(), errors.CatInternal, false, err)
			}
			return map[string]any{
				"output": outPath,
				"bytes":  len(defBytes),
			}, nil
		}, func(data any) {
			if fullOutput {
				printJSON(data)
				return
			}
			m, _ := data.(map[string]any)
			fmt.Printf("Converted %s (%v bytes)\n", m["output"], m["bytes"])
		})
	},
}

func init() {
	definitionsCmd.AddCommand(definitionsShowCmd)
	definitionsCmd.AddCommand(definitionsExportCmd)
	definitionsCmd.AddCommand(definitionsImportCmd)
	definitionsCmd.AddCommand(definitionsBuildCmd)
	definitionsCmd.AddCommand(qsfCmd)
	definitionsCmd.AddCommand(questionsCmd)
	definitionsCmd.AddCommand(blocksCmd)
	definitionsCmd.AddCommand(flowCmd)
	definitionsCmd.AddCommand(optionsCmd)

	qsfCmd.AddCommand(qsfSummaryCmd)
	qsfCmd.AddCommand(qsfConvertCmd)

	definitionsBuildCmd.Flags().StringVarP(&qsfOutputPath, "output", "o", "", "output .qsf file")
	qsfConvertCmd.Flags().StringVarP(&qsfOutputPath, "output", "o", "", "output definition .json file")

	questionsCmd.AddCommand(questionsListCmd)
	questionsCmd.AddCommand(questionsShowCmd)
	questionsCmd.AddCommand(questionsCreateCmd)
	questionsCmd.AddCommand(questionsUpdateCmd)
	questionsCmd.AddCommand(questionsDeleteCmd)

	blocksCmd.AddCommand(blocksListCmd)
	blocksCmd.AddCommand(blocksShowCmd)
	blocksCmd.AddCommand(blocksCreateCmd)
	blocksCmd.AddCommand(blocksUpdateCmd)
	blocksCmd.AddCommand(blocksDeleteCmd)

	flowCmd.AddCommand(flowShowCmd)
	flowCmd.AddCommand(flowUpdateCmd)

	optionsCmd.AddCommand(optionsShowCmd)
	optionsCmd.AddCommand(optionsUpdateCmd)

	definitionsExportCmd.Flags().StringVarP(&defExportPath, "output", "o", "", "output file")
	definitionsImportCmd.Flags().StringVarP(&defPayloadFile, "file", "f", "", "definition JSON file (- for stdin)")
	questionsCreateCmd.Flags().StringVarP(&defPayloadFile, "file", "f", "", "question JSON file (- for stdin)")
	questionsCreateCmd.Flags().StringVar(&defBlockID, "block-id", "", "target block ID")
	questionsUpdateCmd.Flags().StringVarP(&defPayloadFile, "file", "f", "", "question JSON file (- for stdin)")
	blocksCreateCmd.Flags().StringVar(&defDescription, "description", "", "block description")
	blocksUpdateCmd.Flags().StringVarP(&defPayloadFile, "file", "f", "", "block JSON file (- for stdin)")
	flowUpdateCmd.Flags().StringVarP(&defPayloadFile, "file", "f", "", "flow JSON file (- for stdin)")
	optionsUpdateCmd.Flags().StringVarP(&defPayloadFile, "file", "f", "", "options JSON file (- for stdin)")
}
