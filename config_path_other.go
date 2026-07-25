//go:build !windows

package main

import (
	"os"
	"path/filepath"
)

// defaultConfigPath returns the OS config dir (macOS/Linux).
func defaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "vt-print-agent", "config.json")
}

func installDir() string { return filepath.Dir(defaultConfigPath()) }
