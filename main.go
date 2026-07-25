package main

import (
	"flag"
	"log"
)

func main() {
	noTray := flag.Bool("no-tray", false, "run headless, without a system-tray icon")
	cfgPath := flag.String("config", defaultConfigPath(), "path to config.json")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		// First run: on Windows this prompts for server + key, downloads SumatraPDF,
		// and registers autostart; elsewhere it writes a template.
		cfg = firstRun(*cfgPath)
	}
	logToFile(*cfgPath)
	log.Printf("vt-print-agent %s starting (base_url=%s)", version, cfg.BaseURL)

	agent := newAgent(cfg, *cfgPath)
	statusCh := make(chan string, 8)
	go runAgent(agent, statusCh)

	if *noTray {
		select {} // block forever; the agent loop runs in its goroutine
	}
	runTray(agent, statusCh)
}
