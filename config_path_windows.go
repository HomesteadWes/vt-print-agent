//go:build windows

package main

import (
	"os"
	"path/filepath"
)

// defaultConfigPath keeps config.json next to the .exe — the agent is installed
// and run from a per-user folder, so it's self-contained/portable there.
func defaultConfigPath() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "config.json")
	}
	return "config.json"
}

// appDataInstallDir is the per-user install location: %AppData%\VulcanTunes\PrintAgent.
// The agent copies itself here on first run and keeps config + SumatraPDF alongside.
func appDataInstallDir() string {
	base := os.Getenv("APPDATA")
	if base == "" {
		if d, err := os.UserConfigDir(); err == nil {
			base = d
		}
	}
	if base == "" {
		base = "."
	}
	return filepath.Join(base, "VulcanTunes", "PrintAgent")
}
