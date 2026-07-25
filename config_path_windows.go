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

// installDir is where the exe lives — the target install location (config +
// SumatraPDF go here).
func installDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}
