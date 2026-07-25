package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// version is stamped into heartbeats (override at build time with -ldflags).
var version = "0.1.0-dev"

// runAgent is the main loop: enumerate printers → heartbeat → poll → print. It
// pushes short status strings onto statusCh for the tray to display.
func runAgent(cfg *Config, statusCh chan<- string) {
	if cfg.AgentKey == "" {
		setStatus(statusCh, "Not configured — set agent_key")
		log.Printf("no agent_key configured; edit the config and restart")
		return
	}
	client := newClient(cfg.BaseURL, cfg.AgentKey)
	poll := time.Duration(cfg.PollSeconds) * time.Second

	for {
		printers, err := enumeratePrinters()
		if err != nil {
			log.Printf("enumerate printers: %v", err)
		}
		if err := client.Heartbeat(version, cfg.ActiveUser, cfg.Location, printers); err != nil {
			setStatus(statusCh, "Offline — "+err.Error())
			log.Printf("heartbeat: %v", err)
			time.Sleep(poll)
			continue
		}

		jobs, err := client.PollJobs()
		if err != nil {
			setStatus(statusCh, "Poll error — "+err.Error())
			log.Printf("poll: %v", err)
			time.Sleep(poll)
			continue
		}

		if len(jobs) == 0 {
			setStatus(statusCh, fmt.Sprintf("Online · %d printer(s)", len(printers)))
		}
		for _, j := range jobs {
			processJob(client, cfg, statusCh, j)
		}

		time.Sleep(poll)
	}
}

// processJob downloads the job's PDF, prints it, and reports the outcome. The
// server already moved the job to 'printing' when it handed it to us, so we only
// report the terminal result.
func processJob(client *Client, cfg *Config, statusCh chan<- string, j Job) {
	label := j.OrderNumber
	if label == "" {
		label = fmt.Sprintf("job %d", j.ID)
	}
	setStatus(statusCh, "Printing "+label+"…")
	log.Printf("job %d: %s → printer %q", j.ID, j.DocumentType, j.Printer)

	pdf, err := client.download(j.DocumentURL)
	if err != nil {
		reportFail(client, j.ID, "download: "+err.Error())
		return
	}
	tmp, err := writeTemp(pdf)
	if err != nil {
		reportFail(client, j.ID, "temp file: "+err.Error())
		return
	}
	defer os.Remove(tmp)

	if err := printPDF(cfg, j.Printer, tmp, j.Copies); err != nil {
		reportFail(client, j.ID, "print: "+err.Error())
		return
	}
	if err := client.ReportStatus(j.ID, "printed", ""); err != nil {
		log.Printf("job %d: report printed failed: %v", j.ID, err)
	}
	log.Printf("job %d: printed", j.ID)
}

func reportFail(client *Client, id int, msg string) {
	log.Printf("job %d failed: %s", id, msg)
	if err := client.ReportStatus(id, "failed", msg); err != nil {
		log.Printf("job %d: report failed status errored: %v", id, err)
	}
}

func writeTemp(pdf []byte) (string, error) {
	f, err := os.CreateTemp("", "vt-label-*.pdf")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(pdf); err != nil {
		f.Close()
		return "", err
	}
	return f.Name(), f.Close()
}

func setStatus(ch chan<- string, s string) {
	if ch == nil {
		return
	}
	select {
	case ch <- s:
	default: // drop if the tray isn't draining fast enough
	}
}

// logToFile mirrors log output to a file next to the config for field debugging.
func logToFile(configPath string) {
	f, err := os.OpenFile(filepath.Join(filepath.Dir(configPath), "vt-print-agent.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		log.SetOutput(f)
	}
}
