# Releases

## Versioning

This project follows SemVer: `vMAJOR.MINOR.PATCH`.

## Downloads

- :fontawesome-brands-github: GitHub Releases: [github.com/dontpanicx-13/recon/releases](https://github.com/dontpanicx-13/recon/releases)

## Building from source

If you build locally, embed name and version using `-ldflags`:

```bash
go build -ldflags "-X main.name=recon -X main.version=v1.0.0" -o recon ./cmd/recon
```

If you omit `-ldflags`, the UI will display `version=dev`.
