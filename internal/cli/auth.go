package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/config"
	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
)

var setTokenValue string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage API credentials",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show credential status and probe the API",
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context(), "auth.status", "failed to check credentials",
			func(ctx context.Context, client *qualtrics.Client) (any, error) {
				cfg, err := config.Load(cfgFile)
				if err != nil {
					return nil, err
				}
				status := map[string]any{
					"profile":   cfg.ProfileName,
					"token_set": cfg.Token() != "",
					"state":     "unknown",
				}
				if cfg.Active != nil {
					status["datacenter"] = cfg.Active.Datacenter
				}
				if u, uerr := cfg.BaseURL(); uerr == nil {
					status["base_url"] = u
				}

				var result any
				err = client.Do(ctx, "GET", "/surveys", nil, &result)
				if err != nil {
					if e, ok := err.(*errors.Error); ok {
						switch e.Code {
						case errors.AuthTokenInvalid:
							status["state"] = "invalid_token"
							status["detail"] = e.Message
							return status, nil
						case errors.APIAccessForbidden:
							status["state"] = "no_api_access"
							status["detail"] = e.Message
							return status, nil
						}
					}
					return nil, err
				}
				status["state"] = "ok"
				status["detail"] = "token accepted by Qualtrics"
				return status, nil
			},
			func(data any) {
				m, _ := data.(map[string]any)
				fmt.Printf("profile:    %v\n", m["profile"])
				fmt.Printf("datacenter: %v\n", m["datacenter"])
				fmt.Printf("base url:   %v\n", m["base_url"])
				fmt.Printf("state:      %v\n", m["state"])
				if d, ok := m["detail"].(string); ok && d != "" {
					fmt.Printf("detail:     %s\n", d)
				}
			})
	},
}

var authSetTokenCmd = &cobra.Command{
	Use:   "set-token",
	Short: "Store an API token in the config file",
	Run: func(cmd *cobra.Command, args []string) {
		token := setTokenValue
		if token == "" {
			token = os.Getenv("QUALTRICS_TOKEN")
		}
		if token == "" {
			fmt.Print("API token: ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			token = strings.TrimSpace(line)
		}
		if token == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "no token provided")
			os.Exit(2)
		}

		profileName := resolveProfileName()
		store := config.NewStore(cfgFile)
		if err := store.SetProfileField(profileName, "token", token); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "failed to write config: %v\n", err)
			os.Exit(1)
		}
		if err := store.SetProfileField(profileName, "auth_method", "token"); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("token stored for profile %q\n", profileName)
	},
}

var authSetDatacenterCmd = &cobra.Command{
	Use:   "set-datacenter <id>",
	Short: "Store the Qualtrics datacenter ID (e.g. pdx1, iad1, fra1)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := resolveProfileName()
		store := config.NewStore(cfgFile)
		if err := store.SetProfileField(profileName, "datacenter", args[0]); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("datacenter %q stored for profile %q\n", args[0], profileName)
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the stored API token",
	Run: func(cmd *cobra.Command, args []string) {
		profileName := resolveProfileName()
		store := config.NewStore(cfgFile)
		if err := store.UnsetProfileField(profileName, "token"); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("token removed from profile %q\n", profileName)
	},
}

func resolveProfileName() string {
	if profile != "" && profile != "default" {
		return profile
	}
	cfg, err := config.Load(cfgFile)
	if err == nil && cfg.ProfileName != "" {
		return cfg.ProfileName
	}
	return "default"
}

func init() {
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authSetTokenCmd)
	authCmd.AddCommand(authSetDatacenterCmd)
	authCmd.AddCommand(authLogoutCmd)

	authSetTokenCmd.Flags().StringVar(&setTokenValue, "token", "", "API token (or QUALTRICS_TOKEN env)")
}
