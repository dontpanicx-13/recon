package report

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"recon/internal/scanner"
)

func Generate(scan scanner.ScanResult) string {
	var b bytes.Buffer

	header := formatTargets(scan.Config)
	fmt.Fprintf(&b, "# Scan Detail\n\n")
	fmt.Fprintf(&b, "**Targets:** %s\n\n", header)
	fmt.Fprintf(&b, "**Date:** %s\n\n", scan.Meta.Date)
	fmt.Fprintf(&b, "**Status:** %s\n\n", scan.Meta.Status)
	fmt.Fprintf(&b, "**Duration:** %d ms\n\n", scan.Meta.DurationMS)

	fmt.Fprintf(&b, "## Config\n\n")
	if scan.Config.Profile != "" {
		fmt.Fprintf(&b, "- Profile: %s\n", scan.Config.Profile)
	}
	fmt.Fprintf(&b, "- Concurrency: %d\n", scan.Config.Concurrency)
	fmt.Fprintf(&b, "- Timeout: %d ms\n", scan.Config.TimeoutMS)
	fmt.Fprintf(&b, "- Ports: %s\n", formatPorts(scan.Config.Ports))
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "## Options\n\n")
	fmt.Fprintf(&b, "- Banner grabbing: %s\n", yesNo(scan.Config.BannerGrabbing))
	fmt.Fprintf(&b, "- TLS analysis: %s\n", yesNo(scan.Config.TLSAnalysis))
	fmt.Fprintf(&b, "- Reverse DNS: %s\n\n", yesNo(scan.Config.ReverseDNS))

	fmt.Fprintf(&b, "## Summary\n\n")
	fmt.Fprintf(&b, "- Hosts total: %d\n", scan.Summary.HostsTotal)
	fmt.Fprintf(&b, "- Hosts found: %d\n", scan.Summary.HostsFound)
	fmt.Fprintf(&b, "- Hosts completed: %d\n", scan.Summary.HostsCompleted)
	fmt.Fprintf(&b, "- Ports total: %d\n", scan.Summary.PortsTotal)
	fmt.Fprintf(&b, "- Ports probed: %d\n", scan.Summary.PortsProbed)
	fmt.Fprintf(&b, "- Open ports: %d\n\n", scan.Summary.OpenPorts)

	hosts := append([]scanner.Host(nil), scan.Hosts...)
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].IP < hosts[j].IP })

	for _, host := range hosts {
		fmt.Fprintf(&b, "## %s", host.IP)
		if host.Hostname != "" {
			fmt.Fprintf(&b, "  %s", host.Hostname)
		}
		fmt.Fprint(&b, "\n\n")

		fmt.Fprintf(&b, "| Port | State | Service | Banner |\n")
		fmt.Fprintf(&b, "| --- | --- | --- | --- |\n")

		ports := append([]scanner.PortState(nil), host.Ports...)
		sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
		for _, port := range ports {
			banner := "—"
			if port.Banner != nil && *port.Banner != "" {
				banner = sanitizeInline(*port.Banner)
			}
			service := "—"
			if port.ServiceGuess != "" {
				service = sanitizeInline(port.ServiceGuess)
			}
			fmt.Fprintf(&b, "| %d | %s | %s | %s |\n", port.Port, port.State, service, banner)
		}
		fmt.Fprint(&b, "\n")

		for _, port := range ports {
			if port.TLS == nil {
				continue
			}
			fmt.Fprintf(&b, "### TLS — %d\n\n", port.Port)
			fmt.Fprintf(&b, "- CN: %s\n", orDash(port.TLS.CommonName))
			fmt.Fprintf(&b, "- SAN: %s\n", orDash(strings.Join(port.TLS.SAN, ", ")))
			fmt.Fprintf(&b, "- Issuer: %s\n", orDash(port.TLS.Issuer))
			fmt.Fprintf(&b, "- Expires: %s\n", orDash(port.TLS.Expires))
			fmt.Fprintf(&b, "- Version: %s / %s\n", orDash(port.TLS.TLSVersion), orDash(port.TLS.Cipher))
			if port.TLS.Note != "" {
				fmt.Fprintf(&b, "- Note: %s\n", sanitizeInline(port.TLS.Note))
			}
			fmt.Fprint(&b, "\n")
		}
	}

	return b.String()
}

func formatTargets(cfg scanner.ScanConfig) string {
	if len(cfg.Targets) == 0 {
		return "—"
	}
	return strings.Join(cfg.Targets, ", ")
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "—"
	}
	if len(ports) <= 20 {
		var out []string
		for _, p := range ports {
			out = append(out, fmt.Sprintf("%d", p))
		}
		return strings.Join(out, ", ")
	}
	return fmt.Sprintf("%d ports", len(ports))
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func sanitizeInline(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.TrimSpace(value)
	if value == "" {
		return "—"
	}
	return value
}

func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
