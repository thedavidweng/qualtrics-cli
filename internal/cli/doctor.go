package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/thedavidweng/qualtrics-cli/internal/config"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local configuration and credentials",
	Run: func(cmd *cobra.Command, args []string) {
		runLocal("doctor", func() any {
			cfgPath := cfgFile
			if cfgPath == "" {
				cfgPath = config.DefaultConfigPath()
			}
			cfg, err := config.Load(cfgFile)

			checks := []map[string]any{}
			add := func(name string, ok bool, detail string) {
				checks = append(checks, map[string]any{"check": name, "ok": ok, "detail": detail})
			}

			add("config_file", err == nil, cfgPath)
			if err != nil {
				add("config_parse", false, err.Error())
			}

			profileName := cfg.ProfileName
			if profileName == "" {
				profileName = "default"
			}
			add("profile", cfg.Active != nil, profileName)

			dc := ""
			if cfg.Active != nil {
				dc = cfg.Active.Datacenter
				if u, uerr := cfg.BaseURL(); uerr == nil {
					_ = u
				} else {
					add("base_url", false, uerr.Error())
				}
			}
			if dc != "" {
				add("datacenter", true, dc)
			} else if cfg.Active != nil && cfg.Active.BaseURL == "" {
				add("datacenter", false, "not set; run `qualtrics auth set-datacenter <id>`")
			}

			token := cfg.Token()
			tokenSource := "none"
			if cfg.Active != nil {
				switch {
				case cfg.Active.Token == "":
					tokenSource = "none"
				case len(cfg.Active.Token) > 4 && cfg.Active.Token[:4] == "env:":
					tokenSource = "env:" + cfg.Active.Token[4:]
				default:
					tokenSource = "config file"
				}
			}
			add("api_token", token != "", tokenSource)

			return map[string]any{
				"checks":  checks,
				"profile": profileName,
			}
		}, func(data any) {
			m, _ := data.(map[string]any)
			fmt.Printf("profile: %s\n", m["profile"])
			checks, _ := m["checks"].([]map[string]any)
			for _, c := range checks {
				status := "ok  "
				if ok, _ := c["ok"].(bool); !ok {
					status = "FAIL"
				}
				fmt.Printf("[%s] %-14s %s\n", status, c["check"], c["detail"])
			}
		})
	},
}
