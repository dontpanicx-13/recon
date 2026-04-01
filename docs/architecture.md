# Architecture

## Components

- `cmd/recon`: program entry point and UI bootstrap.
- `internal/ui`: TUI layout, controls, and rendering.
- `internal/scanner`: worker pool, task scheduling, result aggregation.
- `internal/target`: parsing of IPs, domains, CIDR, and files.
- `internal/ports`: curated port lists and port utilities.
- `internal/banner`: TCP banner grabbing.
- `internal/tlsinfo`: TLS handshake inspection.
- `internal/dns`: reverse DNS lookups.
- `internal/store`: persistence of scan manifests and results.
- `internal/report`: Markdown report generation.

## Scan Pipeline (high level)

1. Parse targets and ports.
2. Spawn worker pool (`concurrency`).
3. Enqueue all `host:port` tasks.
4. Workers probe ports, optionally grab banners and TLS metadata.
5. Results aggregate into per-host summaries and final report.

## Port Preset Generation

Port presets are generated from the official Nmap `nmap-services` file using `cmd/portsgen`. See `Configuration` for the exact update workflow.
