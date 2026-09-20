//go:build !windows

package config

import (
	"os"
	"runtime"
)

var defaultDir = func() string {
	home, _ := os.UserHomeDir()
	return defaultDirFor(runtime.GOOS, home, os.Getenv("XDG_CONFIG_HOME"))
}()
