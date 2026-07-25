//go:build !windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
)

// printPDF prints a PDF via CUPS (macOS / Linux). fit-to-page=false keeps label
// stock at its native size.
func printPDF(cfg *Config, printerName, pdfPath string, copies int) error {
	if copies < 1 {
		copies = 1
	}
	cmd := exec.Command("lp",
		"-d", printerName,
		"-n", strconv.Itoa(copies),
		"-o", "fit-to-page=false",
		pdfPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lp: %v: %s", err, string(out))
	}
	return nil
}
