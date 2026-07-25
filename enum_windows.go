//go:build windows

package main

import "github.com/alexbrainman/printer"

// enumeratePrinters lists installed Windows printers via the spooler.
func enumeratePrinters() ([]Printer, error) {
	names, err := printer.ReadNames()
	if err != nil {
		return nil, err
	}
	def, _ := printer.Default()
	out := make([]Printer, 0, len(names))
	for _, n := range names {
		out = append(out, Printer{SystemName: n, Name: n, IsDefault: n == def})
	}
	return out, nil
}
