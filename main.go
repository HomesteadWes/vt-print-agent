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
		// First run (or unreadable): drop a template the operator can fill in.
		cfg = defaultConfig()
		if serr := saveConfig(*cfgPath, cfg); serr == nil {
			log.Printf("wrote template config to %s — set base_url + agent_key, then restart", *cfgPath)
		} else {
			log.Printf("could not read or create config at %s: %v", *cfgPath, err)
		}
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
