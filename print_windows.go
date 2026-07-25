//go:build windows

package main

import (
	"fmt"
	"os/exec"
)

// printPDF prints a PDF silently on Windows via SumatraPDF. "noscale" keeps 4×6
// thermal labels at 100% so they aren't shrunk to a letter page.
func printPDF(cfg *Config, printerName, pdfPath string, copies int) error {
	if copies < 1 {
		copies = 1
	}
	sumatra := cfg.SumatraPath
	if sumatra == "" {
		sumatra = "SumatraPDF.exe"
	}
	settings := fmt.Sprintf("%dx,noscale", copies)
	cmd := exec.Command(sumatra,
		"-print-to", printerName,
		"-silent",
		"-print-settings", settings,
		"-exit-when-done",
		pdfPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("SumatraPDF: %v: %s", err, string(out))
	}
	return nil
}
