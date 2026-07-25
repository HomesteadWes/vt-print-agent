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

// firstRun (Windows): a self-installing first launch. If we're not already running
// from the per-user install dir (%AppData%\VulcanTunes\PrintAgent), copy ourselves
// there, prompt for server + key, write config + download SumatraPDF into that
// folder, register autostart pointing at the installed copy, relaunch from there,
// and exit — so nothing is left scattered in the folder it was downloaded to.
func firstRun(cfgPath string) *Config {
	installed := filepath.Join(appDataInstallDir(), "vt-print-agent.exe")
	cur, _ := os.Executable()
	if cur != "" && !strings.EqualFold(cur, installed) {
		return installSelf(installed, cur) // exits on success; returns a config only if install failed
	}
	return promptWriteInPlace(cfgPath) // already in the install dir (or unknown) → in-place
}

// installSelf performs the copy-to-AppData install. On success it relaunches the
// installed exe and exits; on failure it returns a config to run in place.
func installSelf(installed, cur string) *Config {
	dir := filepath.Dir(installed)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("install: mkdir %s: %v", dir, err)
		return promptWriteInPlace(defaultConfigPath())
	}
	cfgPath := filepath.Join(dir, "config.json")

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		// Fresh install → ask for server + key.
		cfg = defaultConfig()
		if s := strings.TrimSpace(psInputBox("VulcanTunes Print Agent — setup", "Back-office URL:", prodURL)); s != "" {
			cfg.BaseURL = strings.TrimRight(s, "/")
		}
		cfg.AgentKey = strings.TrimSpace(psInputBox("VulcanTunes Print Agent — setup",
			"Paste the agent key (Back office → System → Print Agents → New agent):", ""))
		cfg.SumatraPath = filepath.Join(dir, "SumatraPDF.exe")
		if err := saveConfig(cfgPath, cfg); err != nil {
			log.Printf("install: save config: %v", err)
		}
	}

	// Copy ourselves into the install dir (best-effort — may be locked if an
	// installed instance is already running).
	if err := copyFile(cur, installed); err != nil {
		log.Printf("install: copy exe: %v", err)
	}

	// SumatraPDF into the install dir (only if missing).
	if cfg.SumatraPath == "" {
		cfg.SumatraPath = filepath.Join(dir, "SumatraPDF.exe")
	}
	if _, err := os.Stat(cfg.SumatraPath); err != nil {
		if derr := downloadFile(strings.TrimRight(cfg.BaseURL, "/")+"/downloads/SumatraPDF.exe", cfg.SumatraPath); derr != nil {
			log.Printf("install: download SumatraPDF: %v", derr)
		}
	}

	// Autostart the installed copy.
	if err := setAutostartExe(true, installed); err != nil {
		log.Printf("install: autostart: %v", err)
	}

	// Relaunch from the install dir and hand off.
	if err := exec.Command(installed).Start(); err != nil {
		log.Printf("install: relaunch: %v", err)
		return cfg // couldn't relaunch — keep running in place
	}
	psMsg("VulcanTunes Print Agent",
		"Installed to:\n"+dir+"\n\nIt runs from there now and starts at login. You can delete the downloaded copy.")
	os.Exit(0)
	return nil
}

// promptWriteInPlace runs a plain first-run in the current folder (fallback / when
// already installed).
func promptWriteInPlace(cfgPath string) *Config {
	dir := filepath.Dir(cfgPath)
	cfg := defaultConfig()
	if s := strings.TrimSpace(psInputBox("VulcanTunes Print Agent — setup", "Back-office URL:", prodURL)); s != "" {
		cfg.BaseURL = strings.TrimRight(s, "/")
	}
	cfg.AgentKey = strings.TrimSpace(psInputBox("VulcanTunes Print Agent — setup",
		"Paste the agent key (Back office → System → Print Agents → New agent):", ""))
	cfg.SumatraPath = filepath.Join(dir, "SumatraPDF.exe")
	if err := saveConfig(cfgPath, cfg); err != nil {
		log.Printf("save config: %v", err)
	}
	if _, err := os.Stat(cfg.SumatraPath); err != nil {
		_ = downloadFile(strings.TrimRight(cfg.BaseURL, "/")+"/downloads/SumatraPDF.exe", cfg.SumatraPath)
	}
	_ = setAutostart(true)
	return cfg
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// psInputBox pops a Windows input dialog via PowerShell and returns the entry.
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

func psMsg(title, text string) {
	esc := func(s string) string { return strings.ReplaceAll(s, "'", "''") }
	script := fmt.Sprintf(
		"[void][Reflection.Assembly]::LoadWithPartialName('Microsoft.VisualBasic');"+
			"[Microsoft.VisualBasic.Interaction]::MsgBox('%s',0,'%s')",
		esc(text), esc(title),
	)
	_ = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Run()
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
