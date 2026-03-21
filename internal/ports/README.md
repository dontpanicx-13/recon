# Port Lists

This package uses curated port lists derived from the official Nmap `nmap-services` data file.

Source datafile (official raw):
`https://raw.githubusercontent.com/nmap/nmap/master/nmap-services`

## Why this approach

The Nmap `nmap-services` file includes per-port frequency data used to rank “most common” ports.
Generating `top100.txt` and `top1000.txt` from that file keeps the lists reproducible and
traceable back to a known, authoritative source.

## Update workflow

1. Download the latest `nmap-services` from the official Nmap repository.
2. Generate the lists using `portsgen`.

```bash
curl -L -o /tmp/nmap-services https://raw.githubusercontent.com/nmap/nmap/master/nmap-services
go run ./cmd/portsgen -in /path/to/nmap-services
```

This will overwrite:
- `internal/ports/top100.txt`
- `internal/ports/top1000.txt`

## Notes

- The files in this directory are embedded into the binary at build time.
- If the lists are empty, the code falls back to simple ranges (1-100 / 1-1000).
