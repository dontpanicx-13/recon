# Configuration

All scan configuration is done from the TUI. Values are validated before the scan starts.

## Targets

`recon` accepts multiple target formats in a single input, separated by commas:

- Single IPv4 address: `192.168.1.10`
- Domain name: `example.com`
- CIDR range: `10.0.0.0/24`
- File path: `/path/to/targets.txt`
- Mixed list: `10.0.0.10, example.com, 10.0.1.0/24, targets.txt`

Behavior:

- IPv6 is not supported (targets with IPv6 are rejected).
- CIDR expansion excludes network and broadcast addresses when possible.
- File lines can include `#` comments and blank lines; invalid lines are skipped with warnings.
- Targets with inner whitespace are rejected.

## Ports

You can choose one of three port modes:

| Mode | Format | Example | Notes |
| --- | --- | --- | --- |
| Preset | Built-in list | `Top 100` | Derived from Nmap `nmap-services` |
| Range | `start-end` | `1-1024` | Valid range `1-65535` |
| List | Comma-separated | `22,80,443` | Duplicates removed, valid `1-65535` |

### Presets

`recon` ships with curated Top 100 and Top 1000 lists derived from the official Nmap `nmap-services` data file. This keeps the preset lists traceable, reproducible, and aligned with real-world service frequency.

- `Top 100`: curated list derived from Nmap `nmap-services`.
- `Top 1000`: curated list derived from Nmap `nmap-services`.
- `All`: full range `1-65535`.

If the embedded lists are empty at build time, the scanner falls back to ranges `1-100` and `1-1000`.

Official releases embed these lists at build time, so prebuilt binaries include Top 100 and Top 1000 without any extra setup.

### Updating preset lists (maintainers)

This workflow is for maintainers only. It requires the repo and Go installed, then the lists are rebuilt and a new release is published.

The preset lists are generated from the official Nmap `nmap-services` file.

```bash
curl -L -o /tmp/nmap-services https://raw.githubusercontent.com/nmap/nmap/master/nmap-services
go run ./cmd/portsgen -in /tmp/nmap-services
```

This overwrites:

- `internal/ports/top100.txt`
- `internal/ports/top1000.txt`

### Range

Use a numeric range like `1-1024`. Values must be within `1-65535`.

### List

Use a comma-separated list like `22,80,443`. Duplicates are removed; valid range `1-65535`.

## Profiles

Profiles set recommended defaults for concurrency and timeout. You can still edit them manually.

| Profile | Concurrency | Timeout |
| --- | --- | --- |
| `Quick` | `200` | `300ms` |
| `Default` | `100` | `1000ms` |
| `Full` | `50` | `3000ms` |
| `Custom` | (set manually) | (set manually) |

## Concurrency

- Positive integer.
- Controls number of Go workers processing port checks in parallel.
- Higher values increase speed but may generate more load on the target and your host.

## Timeout (ms)

- Positive integer, in milliseconds.
- Per-port timeout. If not set, the scanner falls back to `1000ms`.

## Options

- **Banner grabbing**: reads a short application banner from open TCP ports (capped to 1024 bytes and sanitized).
- **TLS analysis**: attempts a TLS handshake on open ports and records version, cipher, issuer, CN/SAN, and expiry.
- **Reverse DNS**: resolves hostnames for hosts with at least one open port.

## Banner Grabbing Scope

When enabled, banner grabbing runs on every open TCP port discovered during the scan. It uses two strategies:

- **Active probes**: send a small protocol-appropriate payload first, then read a response.
- **Passive read**: connect and read without sending a payload.

Why this design:

- **Safety**: small payloads reduce the chance of unintended side effects.
- **Speed**: short reads keep scans fast and predictable.
- **Consistency**: banners are capped to 1024 bytes and sanitized to avoid noisy output.

### Active probe ports

Active payloads are sent only for specific ports:

| Ports | Payload |
| --- | --- |
| 80, 443, 7001, 8000, 8001, 8080, 8443 | HTTP `HEAD` |
| 3000, 5000, 5001, 5678, 6006, 7860, 8888, 11434, 9200, 9300 | HTTP `GET` |
| 6379 | Redis `PING` |
| 5432 | Postgres startup packet |
| 27017 | MongoDB isMaster |
| 389 | LDAP bind request |
| 445 | SMB negotiation |

### Passive-only ports

These ports are explicitly passive (no payload sent):

| Ports | Service |
| --- | --- |
| 21, 22, 23, 25 | ftp, ssh, telnet, smtp |
| 110, 143, 3306, 5900 | pop3, imap, mysql, vnc |

## Service Guessing

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
