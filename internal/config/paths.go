package config

import (
	"os"
	"path/filepath"
)

func DefaultDir() string {
	return defaultDir
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultDir(), "config.yaml")
}

func defaultDirFor(goos, home, xdgConfigHome string) string {
	if goos != "linux" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "qualtrics-cli")
		}
		return filepath.Join(home, ".qualtrics-cli")
	}
	if xdgConfigHome != "" && filepath.IsAbs(xdgConfigHome) {
		return filepath.Join(xdgConfigHome, "qualtrics-cli")
	}
	return filepath.Join(home, ".config", "qualtrics-cli")
}
