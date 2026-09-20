package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of qualtrics",
	Run: func(cmd *cobra.Command, args []string) {
		if err := writeVersion(cmd.OutOrStdout(), profile, jsonMode, pretty); err != nil {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)
		}
	},
}

type versionPayload struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	BuiltBy string `json:"built_by"`
}

func writeVersion(out io.Writer, profileName string, jsonOut, prettyOut bool) error {
	if jsonOut {
		renderer := output.NewRenderer(out, nil, true, prettyOut)
		env := output.NewEnvelope("version", profileName, output.SchemaVersion, requestID, versionPayload{
			Version: version.GetVersion(),
			Commit:  version.GetCommit(),
			Date:    version.GetDate(),
			BuiltBy: version.GetBuiltBy(),
		}, time.Duration(0))
		renderer.RenderSuccess(env)
		return nil
	}
	_, err := fmt.Fprintf(out, "qualtrics version %s (commit: %s, date: %s, built by: %s)\n",
		version.GetVersion(), version.GetCommit(), version.GetDate(), version.GetBuiltBy())
	return err
}
