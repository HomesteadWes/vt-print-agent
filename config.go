package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config is the agent's on-disk settings. base_url + agent_key are required; the
// key is minted in the back office (System → Printers → Agents).
type Config struct {
	BaseURL     string `json:"base_url"`
	AgentKey    string `json:"agent_key"`
	Location    string `json:"location"`
	ActiveUser  string `json:"active_user"`
	PollSeconds int    `json:"poll_seconds"`
	SumatraPath string `json:"sumatra_path,omitempty"`
}

func defaultConfig() *Config {
	return &Config{
		BaseURL:     "https://www.vulcantunes.com",
		PollSeconds: 15,
		SumatraPath: "SumatraPDF.exe",
	}
}

// Agent holds the live config behind a lock so the tray can update it (e.g. paste
// a key) while the run loop reads it. Changes are saved to disk immediately.
type Agent struct {
	mu   sync.RWMutex
	cfg  *Config
	path string
}

func newAgent(cfg *Config, path string) *Agent { return &Agent{cfg: cfg, path: path} }

// snapshot returns a copy of the current config for the run loop.
func (a *Agent) snapshot() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return *a.cfg
}

func (a *Agent) configDir() string { return filepath.Dir(a.path) }

func (a *Agent) setKey(key string) error {
	a.mu.Lock()
	a.cfg.AgentKey = strings.TrimSpace(key)
	c := *a.cfg
	a.mu.Unlock()
	return saveConfig(a.path, &c)
}

func (a *Agent) setServer(url string) error {
	a.mu.Lock()
	a.cfg.BaseURL = strings.TrimRight(strings.TrimSpace(url), "/")
	c := *a.cfg
	a.mu.Unlock()
	return saveConfig(a.path, &c)
}

// defaultConfigPath returns the OS-appropriate config file location.
func defaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "vt-print-agent", "config.json")
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig()
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	if cfg.PollSeconds < 3 {
		cfg.PollSeconds = 15
	}
	return cfg, nil
}

func saveConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
