package main

import (
	"encoding/base64"

	"github.com/getlantern/systray"
)

// trayIcon is a 1×1 placeholder PNG. Replace with a real brand icon (and note
// Windows tray icons want .ico) before shipping.
var trayIcon, _ = base64.StdEncoding.DecodeString(
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+ip1sAAAAASUVORK5CYII=")

// runTray shows the system-tray icon and reflects agent status. It blocks (systray
// owns the main thread), so main() calls it last.
func runTray(cfg *Config, statusCh <-chan string) {
	systray.Run(func() { onReady(cfg, statusCh) }, func() {})
}

func onReady(cfg *Config, statusCh <-chan string) {
	systray.SetIcon(trayIcon)
	systray.SetTitle("VT Print")
	systray.SetTooltip("VulcanTunes print agent")

	mStatus := systray.AddMenuItem("Starting…", "Agent status")
	mStatus.Disable()
	mUser := systray.AddMenuItem(userLine(cfg), "Signed-in user · location")
	mUser.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop the print agent")

	go func() {
		for {
			select {
			case s := <-statusCh:
				mStatus.SetTitle(s)
				systray.SetTooltip("VT Print — " + s)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func userLine(cfg *Config) string {
	u := cfg.ActiveUser
	if u == "" {
		u = "no user"
	}
	l := cfg.Location
	if l == "" {
		l = "no location"
	}
	return u + " · " + l
}
