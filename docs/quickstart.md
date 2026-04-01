# Quickstart

## Run from prebuilt binaries

Download the latest release from:

- <https://github.com/dontpanicx-13/recon/releases>

Steps:

- Extract the archive
- Linux/macOS: `chmod +x recon`
- Run `./recon` (Windows: `recon.exe`)

No Go installation is required for prebuilt binaries.

## Prerequisites (source build)

- Go 1.24+
- Linux or macOS
- A terminal that supports ANSI escape sequences (most modern terminals)

## Build and Run (local)

```bash
go run -ldflags "-X main.name=recon -X main.version=v1.0.0" ./cmd/recon/main.go
```

## Build a binary

```bash
go build -ldflags "-X main.name=recon -X main.version=v1.0.0" -o recon ./cmd/recon
```

## Notes on version display

- If you build without `-ldflags`, the UI shows `version=dev`.
- Releases published from CI embed the correct version automatically.
