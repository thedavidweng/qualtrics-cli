package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFileOverridesDefaults(t *testing.T) {
	path := writeConfig(t, `
default_profile: work
profiles:
  work:
    datacenter: fra1
    auth_method: token
    token: tok_123
    timeout: 45s
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ProfileName != "work" {
		t.Fatalf("profile = %q, want work", cfg.ProfileName)
	}
	if cfg.Active.Datacenter != "fra1" {
		t.Fatalf("datacenter = %q, want fra1", cfg.Active.Datacenter)
	}
	if cfg.Active.Timeout != 45*time.Second {
		t.Fatalf("timeout = %v, want 45s", cfg.Active.Timeout)
	}
	if got := cfg.Token(); got != "tok_123" {
		t.Fatalf("token = %q, want tok_123", got)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	path := writeConfig(t, `
profiles:
  default:
    datacenter: fra1
    token: file_token
`)
	t.Setenv("QUALTRICS_DATACENTER", "iad1")
	t.Setenv("QUALTRICS_TOKEN", "env_token")
	t.Setenv("QUALTRICS_PROFILE", "default")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Active.Datacenter != "iad1" {
		t.Fatalf("datacenter = %q, want iad1 (env wins)", cfg.Active.Datacenter)
	}
	if cfg.Token() != "env_token" {
		t.Fatalf("token = %q, want env_token", cfg.Token())
	}
}

func TestEnvSecretIndirection(t *testing.T) {
	path := writeConfig(t, `
profiles:
  default:
    datacenter: pdx1
    token: env:MY_QUALTRICS_TOKEN
`)
	t.Setenv("MY_QUALTRICS_TOKEN", "resolved_secret")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Token(); got != "resolved_secret" {
		t.Fatalf("token = %q, want resolved_secret", got)
	}
}

func TestBaseURLFromDatacenter(t *testing.T) {
	path := writeConfig(t, `
profiles:
  default:
    datacenter: sjc1
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := cfg.BaseURL()
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://sjc1.qualtrics.com/API/v3"; got != want {
		t.Fatalf("base url = %q, want %q", got, want)
	}
}

func TestBaseURLOverride(t *testing.T) {
	path := writeConfig(t, `
profiles:
  default:
    base_url: https://proxy.example.com/API/v3/
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := cfg.BaseURL()
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://proxy.example.com/API/v3"; got != want {
		t.Fatalf("base url = %q, want %q (trailing slash trimmed)", got, want)
	}
}

func TestMissingDatacenterErrors(t *testing.T) {
	path := writeConfig(t, `
profiles:
  default:
    token: tok
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.BaseURL(); err == nil {
		t.Fatal("expected error when neither datacenter nor base_url is set")
	}
}

func TestMissingConfigFileIsTolerated(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatalf("missing config should not error: %v", err)
	}
	if cfg.ProfileName != "default" {
		t.Fatalf("profile = %q, want default", cfg.ProfileName)
	}
}
