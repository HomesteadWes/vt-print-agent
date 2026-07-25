# vt-print-agent

A tiny cross-platform system-tray print agent for the VulcanTunes back office. It
sits on a warehouse / DC PC and:

1. **Enumerates** the machine's printers and **registers** them with the back
   office (tagged to the agent's warehouse/location).
2. **Polls** the back office print queue for jobs targeting its printers.
3. **Downloads** each label/document PDF, **prints** it to the target printer, and
   **reports the result** (printed / failed) back up.

It only makes **outbound HTTPS** calls — no inbound ports, no localhost web server,
so there's nothing to open in a firewall and no browser mixed-content issues.

## How it talks to the back office

All requests carry the device key in an `X-Agent-Key` header.

| Call | Purpose |
|---|---|
| `POST /api/print/agent/heartbeat` | register/refresh printers + liveness |
| `GET  /api/print/agent/jobs` | claim queued jobs for this agent's printers |
| `POST /api/print/agent/jobs/{id}/status` | report `printed` / `failed` |

The job payload includes a `document_url` the agent fetches to get the PDF.

## Provisioning

1. In the back office: **System → Printers → Agents → New agent**, pick the
   warehouse/location. It shows a **device key once** — copy it.
2. Put the key + base URL in the agent config (see below), then start the agent.
   Its printers appear in the back office after the first heartbeat; set each
   printer's **role** (label / document) and the default there.

## Config

**Easiest:** run the agent, then in its **tray menu → “Paste agent key from
clipboard”** (after copying the device key from the back office). No file editing.
The tray can also set the server URL from the clipboard and open the config folder.

Otherwise, edit the JSON directly — at the OS config dir (`%AppData%\vt-print-agent\config.json` on Windows,
`~/.config/vt-print-agent/config.json` on Linux, `~/Library/Application
Support/vt-print-agent/config.json` on macOS), or pass `--config <path>`. A
template is written on first run:

```json
{
  "base_url": "https://dev.vulcantunes.com",
  "agent_key": "",
  "location": "",
  "active_user": "",
  "poll_seconds": 15,
  "sumatra_path": "SumatraPDF.exe"
}
```

## Printing

- **Windows:** shells out to **SumatraPDF** (portable) for silent PDF printing —
  `SumatraPDF.exe -print-to "<printer>" -silent -print-settings "<n>x,noscale"`.
  Bundle `SumatraPDF.exe` next to the agent (or set `sumatra_path`). Labels print
  at 100% (`noscale`) so 4×6 thermal labels aren't shrunk.
- **macOS / Linux:** shells out to CUPS — `lp -d <printer> -n <copies> -o
  fit-to-page=false`.

## Build

```bash
go mod tidy
go build ./...                       # host platform
GOOS=windows GOARCH=amd64 go build   # cross-compile for the warehouse PCs
```

Run headless (no tray) for testing: `./vt-print-agent --no-tray`.

## Status / roadmap

MVP scaffold. TODO before production: tray login → warehouse binding, autostart on
boot, code-signing + installer (MSI / pkg), auto-update, richer tray UI (recent
jobs, test print), stale-job reclaim.
