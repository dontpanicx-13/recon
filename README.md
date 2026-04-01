# recon

Terminal-based network scanner in Go. Fast TCP Connect scans with optional banner grabbing, TLS inspection, and reverse DNS in a focused TUI.

Docs: 

Repository: https://github.com/dontpanicx-13/recon

Releases: https://github.com/dontpanicx-13/recon/releases

Issues: https://github.com/dontpanicx-13/recon/issues


## Highlights

- Native Go scanner with concurrent workers (no external scan tooling)
- Multiple target types: IP, domain, CIDR, lists, and files
- Port modes: presets, ranges, and explicit lists
- JSON persistence and Markdown report export
- Reproducible presets derived from Nmap `nmap-services`

![Main UI](docs/assets/main-ui.png)
![Scan Details](docs/assets/scan-details.png)

## Quickstart

### Prebuilt binaries (no Go required)

Download the latest release from:

- <https://github.com/dontpanicx-13/recon/releases>

Steps:

- Extract the archive
- Linux/macOS: `chmod +x recon`
- Run `./recon` (Windows: `recon.exe`)

### Prerequisites (source build)

- Go 1.24+
- Linux or macOS
- A terminal that supports ANSI escape sequences

### Build and run (local)

```bash
go run -ldflags "-X main.name=recon -X main.version=v1.0.0" ./cmd/recon/main.go
```

### Build a binary

```bash
go build -ldflags "-X main.name=recon -X main.version=v1.0.0" -o recon ./cmd/recon
```

```bash
./recon
```

Notes on version display:

- If you build without `-ldflags`, the UI shows `version=dev`.
- Releases published from CI embed the correct version automatically.

## Configuration

All scan configuration is done from the TUI. Values are validated before the scan starts.

### Targets

`recon` accepts multiple target formats in a single input, separated by commas:

- Single IPv4 address: `192.168.1.10`
- Domain name: `example.com`
- CIDR range: `10.0.0.0/24`
- File path: `/path/to/targets.txt`
- Mixed list: `10.0.0.10, example.com, 10.0.1.0/24, targets.txt`

Behavior:

- IPv6 is not supported.
- CIDR expansion excludes network and broadcast addresses when possible.
- File lines can include `#` comments and blank lines; invalid lines are skipped with warnings.
- Targets with inner whitespace are rejected.

### Ports

You can choose one of three port modes:

| Mode | Format | Example | Notes |
| --- | --- | --- | --- |
| Preset | Built-in list | `Top 100` | Derived from Nmap `nmap-services` |
| Range | `start-end` | `1-1024` | Valid range `1-65535` |
| List | Comma-separated | `22,80,443` | Duplicates removed, valid `1-65535` |

Presets are derived from the official Nmap `nmap-services` file:

- `Top 100`
- `Top 1000`
- `All` (1-65535)

If the embedded lists are empty at build time, the scanner falls back to ranges `1-100` and `1-1000`.

Official releases embed these lists at build time, so prebuilt binaries include Top 100 and Top 1000 without any extra setup.

Updating preset lists (maintainers):

This workflow is for maintainers only. It requires the repo and Go installed, then the lists are rebuilt and a new release is published.

```bash
curl -L -o /tmp/nmap-services https://raw.githubusercontent.com/nmap/nmap/master/nmap-services
go run ./cmd/portsgen -in /tmp/nmap-services
```

This overwrites:

- `internal/ports/top100.txt`
- `internal/ports/top1000.txt`

Range mode: use a numeric range like `1-1024`. Values must be within `1-65535`.

List mode: use a comma-separated list like `22,80,443`. Duplicates are removed; valid range `1-65535`.

### Profiles

Profiles set recommended defaults for concurrency and timeout. You can still edit them manually.

| Profile | Concurrency | Timeout |
| --- | --- | --- |
| `Quick` | `200` | `300ms` |
| `Default` | `100` | `1000ms` |
| `Full` | `50` | `3000ms` |
| `Custom` | (set manually) | (set manually) |

### Concurrency

- Positive integer
- Controls number of Go workers processing port checks in parallel
- Higher values increase speed but may generate more load on the target and your host

### Timeout (ms)

- Positive integer, in milliseconds
- Per-port timeout. If not set, the scanner falls back to `1000ms`

### Options

- Banner grabbing: reads a short application banner from open TCP ports (capped to 1024 bytes and sanitized)
- TLS analysis: attempts a TLS handshake on open ports and records version, cipher, issuer, CN/SAN, and expiry
- Reverse DNS: resolves hostnames for hosts with at least one open port

### Banner Grabbing Scope

When enabled, banner grabbing runs on every open TCP port discovered during the scan. It uses two strategies:

- Active probes: send a small protocol-appropriate payload first, then read a response
- Passive read: connect and read without sending a payload

Why this design:

- Safety: small payloads reduce the chance of unintended side effects
- Speed: short reads keep scans fast and predictable
- Consistency: banners are capped to 1024 bytes and sanitized to avoid noisy output

Active probe ports:

| Ports | Payload |
| --- | --- |
| 80, 443, 7001, 8000, 8001, 8080, 8443 | HTTP `HEAD` |
| 3000, 5000, 5001, 5678, 6006, 7860, 8888, 11434, 9200, 9300 | HTTP `GET` |
| 6379 | Redis `PING` |
| 5432 | Postgres startup packet |
| 27017 | MongoDB isMaster |
| 389 | LDAP bind request |
| 445 | SMB negotiation |

Passive-only ports:

| Ports | Service |
| --- | --- |
| 21, 22, 23, 25 | ftp, ssh, telnet, smtp |
| 110, 143, 3306, 5900 | pop3, imap, mysql, vnc |

### Service Guessing

`recon` includes a small, explicit port-to-service map used to label common services. This is a best-effort hint, not full protocol detection.

Ports covered include:

| Ports | Labels |
| --- | --- |
| 21, 22, 23 | ftp, ssh, telnet |
| 25, 465, 587 | smtp, smtps |
| 53, 69 | dns, tftp |
| 80, 3000, 5000, 8000, 8080, 9000 | http-alt |
| 110, 995 | pop3, pop3s |
| 143, 993 | imap, imaps |
| 161 | snmp |
| 389, 636 | ldap, ldaps |
| 443, 8443 | https, https-alt |
| 445 | smb |
| 3389 | rdp |
| 5900 | vnc |
| 3306 | mysql |
| 5432 | postgres |
| 1433 | mssql |
| 1521 | oracle |
| 2049 | nfs |
| 2375 | docker |
| 2379, 2380 | etcd |
| 6379 | redis |
| 9092 | kafka |
| 9200, 9300 | elasticsearch |
| 10250 | kubelet |
| 11211 | memcached |
| 15672 | rabbitmq-mgmt |
| 5672 | amqp |
| 5678 | n8n |
| 6443 | k8s-api |
| 7001 | weblogic |
| 8888 | jupyter |
| 11434 | ollama |
| 6006 | tensorboard |

If a port is not listed, the service label is left empty.

## Controls

| Context | Key | Action |
| --- | --- | --- |
| Global | `Q` / `Ctrl+C` | Quit the app |
| Global | `Alt+Up` | Focus New Scan |
| Global | `Alt+Right` | Focus Running / Logs |
| Global | `Alt+Down` | Focus Scan History |
| New Scan | `Up` / `Down` | Move between fields |
| New Scan | `Left` / `Right` | Change select fields (mode/preset/profile) |
| New Scan | `Enter` | Toggle options, open file picker, or start scan |
| File picker | `Up` / `Down` | Move |
| File picker | `Left` | Back |
| File picker | `Right` | Open |
| File picker | `Enter` | Select |
| File picker | `Esc` | Close |
| Running / Logs | `C` / `Enter` | Cancel active scan |
| Running / Logs | `Up` / `Down` | Scroll logs by 1 |
| Running / Logs | `PgUp` / `PgDown` | Scroll by 5 |
| Running / Logs | `Home` | Jump to top |
| Running / Logs | `End` | Follow latest logs |
| Scan History | `Up` / `Down` | Move selection |
| Scan History | `PgUp` / `PgDown` | Jump by 5 |
| Scan History | `Home` / `End` | Jump to top/bottom |
| Scan History | `Enter` | Open scan details |
| Scan History | `D` | Delete selected scan (confirmation required) |
| Scan Details | `Esc` | Close details |
| Scan Details | `J` | Export JSON (writes `<scan_id>.json` in current directory) |
| Scan Details | `M` | Export Markdown (writes `<scan_id>.md` in current directory) |
| Scan Details | `W` / `S` | Scroll details by 1 |
| Scan Details | `Shift+W` / `Shift+S` | Scroll by 5 |
| Scan Details | `Up` / `Down` | Move host selection |
| Scan Details | `Home` / `End` | Jump to top/bottom |

## Data & Storage

Scan data is stored under the user configuration directory.

- Linux: `~/.config/recon/`
- macOS: `~/Library/Application Support/recon/`

Layout:

```
recon/
├── manifest.json
└── scans/
    ├── <scan_id>.json
    └── ...
```

- `manifest.json` stores a lightweight index for fast UI load
- Individual scan results are stored per scan ID and loaded on demand

## Theming

`recon` loads colors from `ui_colors.json` in the working directory. If the file is missing or invalid, defaults are used.

Create or edit `ui_colors.json` in the project root.

### Theme: Steel (cool blue)

```json
{
  "app_bg": "#0f172a",
  "accent_bg": "#38bdf8",
  "accent_fg": "#0f172a",
  "status_bg": "#1e293b",
  "status_fg": "#e2e8f0",
  "spinner_fg": "#22c55e",
  "controls_fg": "#60a5fa"
}
```

![Steel Theme](docs/assets/theme-steel.png)

### Theme: Sunset (warm)

```json
{
  "app_bg": "#1f1a24",
  "accent_bg": "#f97316",
  "accent_fg": "#1f1a24",
  "status_bg": "#2a2231",
  "status_fg": "#f5f3ff",
  "spinner_fg": "#facc15",
  "controls_fg": "#fb7185"
}
```

![Sunset Theme](docs/assets/theme-sunset.png)

### Theme: Signal (green)

```json
{
  "app_bg": "#1e2030",
  "accent_bg": "#7dffb5",
  "accent_fg": "#1e2030",
  "status_bg": "#2b2f45",
  "status_fg": "#e6e6e6",
  "spinner_fg": "#7dffb5",
  "controls_fg": "#5ee38f"
}
```

![Signal Theme](docs/assets/theme-signal.png)

Keys:

| Key | Description |
| --- | --- |
| `app_bg` | application background |
| `accent_bg` / `accent_fg` | focused controls and highlights |
| `status_bg` / `status_fg` | status bar and panel titles |
| `spinner_fg` | spinner color |
| `controls_fg` | control hints color |

## Architecture

Components:

- `cmd/recon`: program entry point and UI bootstrap
- `internal/ui`: TUI layout, controls, and rendering
- `internal/scanner`: worker pool, task scheduling, result aggregation
- `internal/target`: parsing of IPs, domains, CIDR, and files
- `internal/ports`: curated port lists and port utilities
- `internal/banner`: TCP banner grabbing
- `internal/tlsinfo`: TLS handshake inspection
- `internal/dns`: reverse DNS lookups
- `internal/store`: persistence of scan manifests and results
- `internal/report`: Markdown report generation

Scan pipeline (high level):

1. Parse targets and ports
2. Spawn worker pool (`concurrency`)
3. Enqueue all `host:port` tasks
4. Workers probe ports, optionally grab banners and TLS metadata
5. Results aggregate into per-host summaries and final report

Port presets are generated from the official Nmap `nmap-services` file using `cmd/portsgen`.

## Security & Ethics

`recon` is intended for authorized security testing and internal network assessment.

Guidelines:

- Only scan systems you own or have explicit permission to test
- Respect local laws and organizational policies
- Use conservative concurrency and timeouts on shared or sensitive networks

If you discover vulnerabilities or misconfigurations, report them responsibly.

## Releases

Versioning follows SemVer: `vMAJOR.MINOR.PATCH`.

Downloads:

- <https://github.com/dontpanicx-13/recon/releases>

## License

MIT
