//go:build windows

package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// printPDF prints a PDF silently on Windows via SumatraPDF.
//
// Scaling matters for thermal labels: if the Zebra driver's configured stock
// isn't exactly the label's page size, "noscale" makes the driver reject the
// oversized page and StartDoc fails (SumatraPDF exits 1). So labels default to
// "fit" (scale to the driver's printable area) unless overridden in config.
//
// If the first attempt fails we retry once with no scaling token at all, letting
// the driver's own defaults handle the page — this rescues misconfigured drivers.
func printPDF(cfg *Config, printerName, pdfPath string, copies int, documentType string) error {
	if copies < 1 {
		copies = 1
	}
	sumatra := cfg.SumatraPath
	if sumatra == "" {
		sumatra = "SumatraPDF.exe"
	}

	scale := cfg.PrintSettings
	if scale == "" {
		if documentType == "label" {
			scale = "fit"
		} else {
			scale = "noscale"
		}
	}

	// Attempt 1: copies + scaling. Attempt 2 (fallback): copies only.
	attempts := []string{
		fmt.Sprintf("%dx,%s", copies, scale),
		fmt.Sprintf("%dx", copies),
	}
	var lastErr error
	for i, settings := range attempts {
		cmd := exec.Command(sumatra,
			"-print-to", printerName,
			"-silent",
			"-print-settings", settings,
			"-exit-when-done",
			pdfPath,
		)
		log.Printf("print attempt %d: %s -print-to %q -print-settings %q", i+1, sumatra, printerName, settings)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		lastErr = fmt.Errorf("SumatraPDF (settings %q): %v: %s", settings, err, strings.TrimSpace(string(out)))
		log.Printf("print attempt %d failed: %v", i+1, lastErr)
	}
	return lastErr
}
