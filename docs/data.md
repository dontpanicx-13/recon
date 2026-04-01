# Data & Storage

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

- `manifest.json` stores a lightweight index for fast UI load.
- Individual scan results are stored per scan ID and loaded on demand.
