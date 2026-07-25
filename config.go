package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		BaseURL:     "https://dev.vulcantunes.com",
		PollSeconds: 15,
		SumatraPath: "SumatraPDF.exe",
	}
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
