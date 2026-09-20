package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
)

var (
	rawMethod string
	rawData   string
)

var rawCmd = &cobra.Command{
	Use:   "raw <METHOD> <path>",
	Short: "Make an authenticated API request directly",
	Long: `Escape hatch for endpoints without a dedicated command. The path is
relative to the API root, e.g. "GET /surveys" or "POST /responseexports".`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		method := rawMethod
		if method == "" {
			method = args[0]
		}
		path := args[1]
		if path == "" || path[0] != '/' {
			path = "/" + path
		}

		run(cmd.Context(), "raw", "request failed",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				var body any
				if rawData != "" {
					if err := json.Unmarshal([]byte(rawData), &body); err != nil {
						return nil, fmt.Errorf("invalid --data JSON: %w", err)
					}
				}
				var result any
				if err := client.Do(ctx, method, path, body, &result); err != nil {
					return nil, err
				}
				return result, nil
			},
			func(data any) {
				out, _ := json.MarshalIndent(data, "", "  ")
				fmt.Println(string(out))
			})
	},
}

func init() {
	rawCmd.Flags().StringVar(&rawMethod, "method", "", "HTTP method (default: first argument)")
	rawCmd.Flags().StringVar(&rawData, "data", "", "JSON request body")
}
