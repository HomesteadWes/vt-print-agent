package main

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
)

// runTray shows the tray icon and lets the operator paste their agent key, flip
// between Production/Dev, toggle start-at-login, and open the config folder — no
// hand-editing config.json. It blocks (systray owns the main thread).
func runTray(a *Agent, statusCh <-chan string) {
	systray.Run(func() { onReady(a, statusCh) }, func() {})
}

func onReady(a *Agent, statusCh <-chan string) {
	systray.SetIcon(trayIcon)
	systray.SetTitle("") // icon-only
	systray.SetTooltip("VulcanTunes print agent")

	snap := a.snapshot()
	mStatus := systray.AddMenuItem("Starting…", "Agent status")
	mStatus.Disable()
	mServer := systray.AddMenuItem("Server: "+snap.BaseURL, "Back-office URL")
	mServer.Disable()
	systray.AddSeparator()

	mPasteKey := systray.AddMenuItem("Paste agent key from clipboard", "Copy the device key from the back office, then click this")
	mUseProd := systray.AddMenuItemCheckbox("Use production", "Point at "+prodURL, snap.BaseURL == prodURL)
	mUseDev := systray.AddMenuItemCheckbox("Use dev", "Point at "+devURL, snap.BaseURL == devURL)
	mPasteURL := systray.AddMenuItem("Paste server URL from clipboard", "Set a custom back-office URL")

	// Start-at-login is Windows-only; leave the channel nil elsewhere so its select
	// case never fires.
	var mStartup *systray.MenuItem
	var startupCh <-chan struct{}
	if runtime.GOOS == "windows" {
		mStartup = systray.AddMenuItemCheckbox("Start at login", "Launch automatically when you sign in", autostartEnabled())
		startupCh = mStartup.ClickedCh
	}

	mOpenCfg := systray.AddMenuItem("Open config folder", "Where config.json + the log live")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop the print agent")

	setServer := func(url string) {
		if err := a.setServer(url); err != nil {
			return
		}
		url = strings.TrimRight(url, "/")
		mServer.SetTitle("Server: " + url)
		if url == prodURL {
			mUseProd.Check()
		} else {
			mUseProd.Uncheck()
		}
		if url == devURL {
			mUseDev.Check()
		} else {
			mUseDev.Uncheck()
		}
		mStatus.SetTitle("Server saved — connecting…")
	}

	go func() {
		for {
			select {
			case s := <-statusCh:
				mStatus.SetTitle(s)
				systray.SetTooltip("VT Print — " + s)

			case <-mPasteKey.ClickedCh:
				key, err := clipboard.ReadAll()
				if err != nil {
					mStatus.SetTitle("Couldn't read the clipboard")
					break
				}
				key = strings.TrimSpace(key)
				if !isLikelyKey(key) {
					mStatus.SetTitle("Clipboard isn't a valid agent key")
					break
				}
				if err := a.setKey(key); err != nil {
					mStatus.SetTitle("Couldn't save the key")
					break
				}
				mStatus.SetTitle("Key saved — connecting…")

			case <-mUseProd.ClickedCh:
				setServer(prodURL)
			case <-mUseDev.ClickedCh:
				setServer(devURL)

			case <-mPasteURL.ClickedCh:
				u, err := clipboard.ReadAll()
				if err != nil {
					break
				}
				u = strings.TrimSpace(u)
				if !strings.HasPrefix(u, "http") {
					mStatus.SetTitle("Clipboard isn't a URL")
					break
				}
				setServer(u)

			case <-startupCh:
				if mStartup == nil {
					break
				}
				enable := !autostartEnabled()
				if err := setAutostart(enable); err == nil {
					if enable {
						mStartup.Check()
					} else {
						mStartup.Uncheck()
					}
				}

			case <-mOpenCfg.ClickedCh:
				openFolder(a.configDir())

			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

// isLikelyKey checks the pasted text looks like a 64-char hex device key.
func isLikelyKey(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func openFolder(dir string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	_ = cmd.Start()
}
