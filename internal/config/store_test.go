package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	store := NewStore(path)

	if err := store.SetProfileField("default", "datacenter", "pdx1"); err != nil {
		t.Fatal(err)
	}
	if err := store.SetProfileField("default", "token", "tok_abc"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("config perm = %v, want 0600", perm)
	}

	if err := store.UnsetProfileField("default", "token"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Active.Datacenter != "pdx1" {
		t.Fatalf("datacenter = %q, want pdx1", cfg.Active.Datacenter)
	}
	if cfg.Active.Token != "" {
		t.Fatalf("token = %q, want empty after unset", cfg.Active.Token)
	}
}
