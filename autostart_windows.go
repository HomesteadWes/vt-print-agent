//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runKeyName = "VulcanTunesPrintAgent"

// setAutostart adds/removes an HKCU Run entry for the currently-running exe.
func setAutostart(enable bool) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return setAutostartExe(enable, exe)
}

// setAutostartExe registers a specific exe path to launch at login (used by the
// installer so autostart points at the installed copy, not the downloaded one).
func setAutostartExe(enable bool, exe string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !enable {
		return k.DeleteValue(runKeyName)
	}
	return k.SetStringValue(runKeyName, `"`+exe+`"`)
}

func autostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runKeyName)
	return err == nil
}
