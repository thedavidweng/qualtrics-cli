package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/config"
	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/version"
)

var (
	cfgFile    string
	jsonMode   bool
	pretty     bool
	fullOutput bool
	readOnly   bool
	dryRun     bool
	confirm    bool
	timeout    time.Duration
	profile    string
	requestID  string
)

var RootCmd = &cobra.Command{
	Use:     "qualtrics",
	Short:   "A local, agent-friendly CLI for the Qualtrics API",
	Version: version.GetVersion(),
	Long: `qualtrics-cli is a single-binary command line tool for working with
Qualtrics surveys, responses, distributions, and contacts from your terminal,
scripts, and local agents.`,
	Example: `  qualtrics surveys list --json
  qualtrics definitions show SV_0123456789123
  qualtrics responses export start SV_0123456789123 --format json --wait
  qualtrics raw GET /surveys`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		requestID = uuid.NewString()

		if cfgFile == "" {
			cfgFile = os.Getenv("QUALTRICS_CONFIG")
		}

		cfg, _ := config.Load(cfgFile)

		jsonMode = jsonMode || envBool("QUALTRICS_JSON")
		pretty = pretty || envBool("QUALTRICS_PRETTY")
		fullOutput = fullOutput || envBool("QUALTRICS_FULL")
		readOnly = readOnly || envBool("QUALTRICS_READ_ONLY")
		dryRun = dryRun || envBool("QUALTRICS_DRY_RUN")
		confirm = confirm || envBool("QUALTRICS_CONFIRM")

		if !persistentFlagChanged(cmd, "profile") {
			profile = cfg.ProfileName
		}
		if !persistentFlagChanged(cmd, "timeout") && cfg.Active != nil {
			timeout = cfg.Active.Timeout
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		e, ok := err.(*errors.Error)
		if !ok {
			e = errors.New(errors.InvalidArguments, err.Error(), errors.CatValidation, false, err)
		}
		if jsonMode {
			renderer := output.NewRenderer(nil, nil, true, pretty)
			renderer.RenderError(output.NewErrorEnvelope("qualtrics", profile, output.SchemaVersion, requestID, e, 0))
			os.Exit(e.ExitCode())
		}
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(e.ExitCode())
	}
}

func envBool(key string) bool {
	return config.ParseBool(os.Getenv(key))
}

func persistentFlagChanged(cmd *cobra.Command, name string) bool {
	f := cmd.Root().PersistentFlags().Lookup(name)
	return f != nil && f.Changed
}

func init() {
	RootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Core Commands"})
	RootCmd.AddGroup(&cobra.Group{ID: "surveys", Title: "Survey Platform"})
	RootCmd.AddGroup(&cobra.Group{ID: "data", Title: "Responses & Contacts"})
	RootCmd.AddGroup(&cobra.Group{ID: "utility", Title: "Utilities"})

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/qualtrics-cli/config.yaml)")
	RootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "emit machine-readable JSON")
	RootCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty-print JSON output")
	RootCmd.PersistentFlags().BoolVar(&fullOutput, "full", false, "print full payloads instead of summaries")
	RootCmd.PersistentFlags().BoolVar(&readOnly, "read-only", false, "block remote writes")
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview a remote write without executing it")
	RootCmd.PersistentFlags().BoolVar(&confirm, "confirm", false, "explicitly execute a remote write")
	RootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "set command timeout")
	RootCmd.PersistentFlags().StringVar(&profile, "profile", "default", "use a named profile")

	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(doctorCmd)
	RootCmd.AddCommand(authCmd)
	RootCmd.AddCommand(rawCmd)
	RootCmd.AddCommand(completionCmd)
	RootCmd.AddCommand(surveysCmd)
	RootCmd.AddCommand(responsesCmd)
	RootCmd.AddCommand(definitionsCmd)
	RootCmd.AddCommand(distributionsCmd)
	RootCmd.AddCommand(directoriesCmd)
	RootCmd.AddCommand(eventsCmd)
}
