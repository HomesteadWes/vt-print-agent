//go:build !windows

package main

import "log"

// firstRun (macOS/Linux): write a template config the operator edits (or pastes a
// key into via the tray).
func firstRun(cfgPath string) *Config {
	cfg := defaultConfig()
	if err := saveConfig(cfgPath, cfg); err == nil {
		log.Printf("wrote template config to %s — set base_url + agent_key (or use the tray)", cfgPath)
	}
	return cfg
}
