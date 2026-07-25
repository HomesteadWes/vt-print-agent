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

1. In the back office: **System → Print Agents → New agent**, pick the
   warehouse/location. It shows a **device key once** — copy it.
2. Run the agent (below). Its printers appear in the back office after the first
   heartbeat; set each printer's **role** (label / document) and the default there.

## Windows — first-run installer

Just run `vt-print-agent-windows-amd64.exe` from wherever you downloaded it. On
first launch it **installs itself to `%AppData%\VulcanTunes\PrintAgent`**:

1. copies the exe into that folder,
2. prompts for the **back-office URL** (defaults to production) and the **agent key**,
3. writes `config.json` and **downloads SumatraPDF** into that same folder,
4. **registers autostart** pointing at the installed copy,
5. relaunches from there and exits — you can delete the downloaded copy.

After that it lives in the tray and starts at login. No files are left in your
Downloads folder, and nothing needs editing by hand.

## Config (other platforms / manual)

- **Tray:** run the agent, then **“Paste agent key from clipboard”** (after copying
  the key). The tray also has **Use production / Use dev**, **Paste server URL**,
  **Start at login** (Windows), and **Open config folder**.
- **JSON:** edit `config.json` directly. On Windows it sits next to the exe; on
  macOS `~/Library/Application Support/vt-print-agent/config.json`; on Linux
  `~/.config/vt-print-agent/config.json`. Or pass `--config <path>`.

```json
{
  "base_url": "https://www.vulcantunes.com",
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
