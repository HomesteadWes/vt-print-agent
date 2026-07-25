//go:build windows

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// firstRun (Windows): a self-contained first-launch installer. Prompts for the
// server + agent key, writes config.json into the install folder (next to the
// exe), downloads SumatraPDF into that same folder, and registers autostart at
// login. Returns the resulting config.
func firstRun(cfgPath string) *Config {
	cfg := defaultConfig()

	if server := strings.TrimSpace(psInputBox("VulcanTunes Print Agent — setup", "Back-office URL:", prodURL)); server != "" {
		cfg.BaseURL = strings.TrimRight(server, "/")
	}
	cfg.AgentKey = strings.TrimSpace(psInputBox("VulcanTunes Print Agent — setup",
		"Paste the agent key (Back office → System → Print Agents → New agent):", ""))

	// Everything lives in the install folder (next to the exe).
	dir := installDir()
	cfg.SumatraPath = filepath.Join(dir, "SumatraPDF.exe")
	if err := saveConfig(cfgPath, cfg); err != nil {
		log.Printf("save config: %v", err)
	}

	// Fetch SumatraPDF into the install folder only (skip if already present).
	if _, err := os.Stat(cfg.SumatraPath); err != nil {
		if derr := downloadFile(cfg.BaseURL+"/downloads/SumatraPDF.exe", cfg.SumatraPath); derr != nil {
			log.Printf("download SumatraPDF: %v", derr)
		} else {
			log.Printf("downloaded SumatraPDF to %s", cfg.SumatraPath)
		}
	}

	// Launch at login.
	if err := setAutostart(true); err != nil {
		log.Printf("autostart: %v", err)
	}
	return cfg
}

// psInputBox shows a Windows input dialog via PowerShell and returns the entry
// (empty if the user cancels). No GUI dependencies to bundle.
func psInputBox(title, prompt, def string) string {
	esc := func(s string) string { return strings.ReplaceAll(s, "'", "''") }
	script := fmt.Sprintf(
		"[void][Reflection.Assembly]::LoadWithPartialName('Microsoft.VisualBasic');"+
			"[Microsoft.VisualBasic.Interaction]::InputBox('%s','%s','%s')",
		esc(prompt), esc(title), esc(def),
	)
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\r\n")
}

func downloadFile(url, dest string) error {
	c := &http.Client{Timeout: 90 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
