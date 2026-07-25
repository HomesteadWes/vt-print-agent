//go:build !windows

package main

// Autostart is Windows-only for now; no-ops elsewhere.
func setAutostart(enable bool) error { return nil }
func autostartEnabled() bool         { return false }
