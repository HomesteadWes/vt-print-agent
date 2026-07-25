//go:build !windows

package main

import (
	"os/exec"
	"strings"
)

// enumeratePrinters lists CUPS printers via lpstat (macOS / Linux).
func enumeratePrinters() ([]Printer, error) {
	out, err := exec.Command("lpstat", "-a").Output()
	if err != nil {
		return nil, err
	}
	var printers []Printer
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		printers = append(printers, Printer{SystemName: name, Name: name})
	}
	// Flag the system default, if any ("system default destination: NAME").
	if d, err := exec.Command("lpstat", "-d").Output(); err == nil {
		s := strings.TrimSpace(string(d))
		if i := strings.LastIndex(s, ": "); i >= 0 {
			def := strings.TrimSpace(s[i+2:])
			for k := range printers {
				if printers[k].SystemName == def {
					printers[k].IsDefault = true
				}
			}
		}
	}
	return printers, nil
}
