package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
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
		method := strings.ToUpper(rawMethod)
		if method == "" {
			method = strings.ToUpper(args[0])
		}
		path := args[1]
		if path == "" || path[0] != '/' {
			path = "/" + path
		}

		var tier safety.OperationTier
		switch method {
		case "GET", "HEAD", "OPTIONS":
			tier = safety.TierRead
		case "POST", "PUT", "PATCH":
			tier = safety.TierMutation
		case "DELETE":
			tier = safety.TierDestructive
		default:
			renderer := output.NewRenderer(nil, nil, jsonMode, pretty)
			handleError(renderer, "raw", errors.New(errors.InvalidArguments, fmt.Sprintf("unsupported method %q: use GET, POST, PUT, PATCH, or DELETE", method), errors.CatValidation, false, nil), time.Now())
			return
		}

		var result any
		if tier == safety.TierRead {
			run(cmd.Context(), "raw", "request failed", rawCall(method, path, &result), rawHuman(&result))
			return
		}
		runMutation(cmd, "raw", "request failed", tier,
			func() (mutation, *errors.Error) {
				body, verr := rawBody()
				if verr != nil {
					return mutation{}, verr
				}
				return mutation{
					resourceID: path,
					planAfter:  body,
					do: func(ctx context.Context, client *qualtrics.Client) (any, error) {
						if err := client.Do(ctx, method, path, body, &result); err != nil {
							return nil, err
						}
						return result, nil
					},
					human: func(data any) { printJSON(result) },
				}, nil
			})
	},
}

func rawBody() (any, *errors.Error) {
	if rawData == "" {
		return nil, nil
	}
	var body any
	if err := json.Unmarshal([]byte(rawData), &body); err != nil {
		return nil, errors.New(errors.InvalidArguments, fmt.Sprintf("invalid --data JSON: %v", err), errors.CatValidation, false, err)
	}
	return body, nil
}

func rawCall(method, path string, result *any) func(context.Context, *qualtrics.Client) (any, error) {
	return func(ctx context.Context, client *qualtrics.Client) (any, error) {
		body, verr := rawBody()
		if verr != nil {
			return nil, verr
		}
		if err := client.Do(ctx, method, path, body, result); err != nil {
			return nil, err
		}
		return *result, nil
	}
}

func rawHuman(result *any) func(any) {
	return func(any) { printJSON(*result) }
}

func init() {
	rawCmd.Flags().StringVar(&rawMethod, "method", "", "HTTP method (default: first argument)")
	rawCmd.Flags().StringVar(&rawData, "data", "", "JSON request body")
}
