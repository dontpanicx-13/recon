package report

import (
	"strings"
	"testing"

	"recon/internal/scanner"
)

func TestGenerate_Basic(t *testing.T) {
	banner := "SSH-2.0-OpenSSH_8.2\r\n"
	scan := scanner.ScanResult{
		ScanID: "scan-1",
		Config: scanner.ScanConfig{
			Targets: []string{"192.168.1.1"},
		},
		Meta: scanner.ScanMeta{
			Date:   "2026-03-05T14:00:00Z",
			Status: scanner.StatusCompleted,
		},
		Summary: scanner.ScanSummary{
			HostsTotal:     1,
			HostsFound:     1,
			HostsCompleted: 1,
			PortsTotal:     1,
			PortsProbed:    1,
			OpenPorts:      1,
		},
		Hosts: []scanner.Host{
			{
				IP:       "192.168.1.1",
				Hostname: "router.local",
				Ports: []scanner.PortState{
					{
						Port:         22,
						State:        scanner.PortOpen,
						ServiceGuess: "ssh",
						Banner:       &banner,
					},
				},
			},
		},
	}

	md := Generate(scan)
	if !strings.Contains(md, "# Scan Detail") {
		t.Fatalf("missing title")
	}
	if !strings.Contains(md, "router.local") {
		t.Fatalf("missing hostname")
	}
	if !strings.Contains(md, "| 22 | open | ssh | SSH-2.0-OpenSSH_8.2 |") {
		t.Fatalf("missing port row")
	}
}

func TestGenerate_TLSSection(t *testing.T) {
	scan := scanner.ScanResult{
		ScanID: "scan-2",
		Config: scanner.ScanConfig{
			Targets: []string{"10.0.0.1"},
		},
		Meta: scanner.ScanMeta{
			Date:   "2026-03-05T14:00:00Z",
			Status: scanner.StatusCompleted,
		},
		Hosts: []scanner.Host{
			{
				IP: "10.0.0.1",
				Ports: []scanner.PortState{
					{
						Port:  443,
						State: scanner.PortOpen,
						TLS: &scanner.TLSInfo{
							CommonName: "example.local",
							SAN:        []string{"example.local"},
							Issuer:     "Acme",
							Expires:    "2026-09-01",
							TLSVersion: "TLS 1.3",
							Cipher:     "AES_256_GCM_SHA384",
						},
					},
				},
			},
		},
	}

	md := Generate(scan)
	if !strings.Contains(md, "### TLS — 443") {
		t.Fatalf("missing tls section")
	}
	if !strings.Contains(md, "CN: example.local") {
		t.Fatalf("missing tls cn")
	}
	if !strings.Contains(md, "Version: TLS 1.3 / AES_256_GCM_SHA384") {
		t.Fatalf("missing tls version/cipher")
	}
}
