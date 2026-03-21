package target

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	ExcludeNetworkBroadcast bool
	LookupIP                func(host string) ([]net.IP, error)
	OnResolve               func(host string, ips []net.IP, err error)
}

type ParseResult struct {
	Targets  []string
	Warnings []string
}

func Parse(input string, opts Options) (ParseResult, []string) {
	var result ParseResult
	var errs []string
	seen := make(map[string]struct{})

	for _, raw := range splitTokens(input) {
		token := strings.TrimSpace(raw)
		if token == "" {
			errs = append(errs, "Empty target entry.")
			continue
		}
		if hasInnerWhitespace(token) {
			errs = append(errs, fmt.Sprintf("Invalid target %q: contains whitespace.", token))
			continue
		}

		if strings.Contains(token, "/") {
			if cidrTargets, ok, err := parseCIDR(token, opts); ok {
				if err != nil {
					errs = append(errs, err.Error())
					continue
				}
				appendTargets(&result, seen, cidrTargets)
				continue
			}
		}

		if fileInfo, statErr := os.Stat(token); statErr == nil {
			if fileInfo.IsDir() {
				errs = append(errs, fmt.Sprintf("%q is a directory, expected file.", token))
				continue
			}
			fileTargets, fileWarns, fileErr := parseFile(token, opts)
			if fileErr != nil {
				errs = append(errs, fileErr.Error())
				continue
			}
			for _, warn := range fileWarns {
				result.Warnings = append(result.Warnings, warn)
			}
			appendTargets(&result, seen, fileTargets)
			continue
		} else if statErr != nil && looksLikePath(token) {
			errs = append(errs, fmt.Sprintf("Could not access file %q: %v", token, statErr))
			continue
		}

		tokenTargets, tokenErrs := parseToken(token, opts)
		if len(tokenErrs) > 0 {
			errs = append(errs, tokenErrs...)
			continue
		}
		appendTargets(&result, seen, tokenTargets)
	}

	return result, errs
}

func splitTokens(input string) []string {
	parts := strings.Split(input, ",")
	if len(parts) == 0 {
		return nil
	}
	return parts
}

func parseToken(token string, opts Options) ([]string, []string) {
	if cidrTargets, ok, err := parseCIDR(token, opts); ok {
		if err != nil {
			return nil, []string{err.Error()}
		}
		return cidrTargets, nil
	}

	if ip := net.ParseIP(token); ip != nil {
		if ip.To4() == nil {
			return nil, []string{fmt.Sprintf("Invalid target %q: IPv6 is not supported.", token)}
		}
		return []string{ip.String()}, nil
	}

	lookup := opts.LookupIP
	if lookup == nil {
		lookup = net.LookupIP
	}
	ips, err := lookup(token)
	if opts.OnResolve != nil {
		opts.OnResolve(token, ips, err)
	}
	if err != nil {
		return nil, []string{fmt.Sprintf("Domain %q did not resolve.", token)}
	}
	var targets []string
	for _, ip := range ips {
		if ip.To4() == nil {
			continue
		}
		targets = append(targets, ip.String())
	}
	if len(targets) == 0 {
		return nil, []string{fmt.Sprintf("Domain %q has no IPv4 records.", token)}
	}
	return targets, nil
}

func parseFile(path string, opts Options) ([]string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not read file %q: %w", path, err)
	}
	defer file.Close()

	var targets []string
	var warns []string
	seen := make(map[string]struct{})

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if hasInnerWhitespace(line) {
			warns = append(warns, fmt.Sprintf("Skipped %q line %d: contains whitespace.", filepath.Base(path), lineNo))
			continue
		}
		lineTargets, lineErrs := parseToken(line, opts)
		if len(lineErrs) > 0 {
			warns = append(warns, fmt.Sprintf("Skipped %q line %d: %s", filepath.Base(path), lineNo, strings.Join(lineErrs, "; ")))
			continue
		}
		appendTargetsToSlice(&targets, seen, lineTargets)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("Could not read file %q: %w", path, err)
	}
	return targets, warns, nil
}

func parseCIDR(token string, opts Options) ([]string, bool, error) {
	ip, ipnet, err := net.ParseCIDR(token)
	if err != nil {
		return nil, false, nil
	}
	if ip.To4() == nil {
		return nil, true, fmt.Errorf("Invalid target %q: IPv6 CIDR is not supported.", token)
	}

	_, bits := ipnet.Mask.Size()
	if bits != 32 {
		return nil, true, fmt.Errorf("Invalid target %q: IPv6 CIDR is not supported.", token)
	}

	start := ipv4ToUint32(ipnet.IP)
	mask := maskToUint32(ipnet.Mask)
	network := start & mask
	broadcast := network | ^mask

	if opts.ExcludeNetworkBroadcast && broadcast-network > 1 {
		network++
		broadcast--
	}

	var targets []string
	for addr := network; addr <= broadcast; addr++ {
		targets = append(targets, uint32ToIPv4(addr))
	}
	return targets, true, nil
}

func appendTargets(result *ParseResult, seen map[string]struct{}, targets []string) {
	appendTargetsToSlice(&result.Targets, seen, targets)
}

func appendTargetsToSlice(dest *[]string, seen map[string]struct{}, targets []string) {
	for _, target := range targets {
		if _, exists := seen[target]; exists {
			continue
		}
		seen[target] = struct{}{}
		*dest = append(*dest, target)
	}
}

func hasInnerWhitespace(value string) bool {
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return true
		}
	}
	return false
}

func looksLikePath(value string) bool {
	if filepath.IsAbs(value) {
		return true
	}
	return strings.Contains(value, string(os.PathSeparator))
}

func ipv4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func maskToUint32(mask net.IPMask) uint32 {
	if len(mask) != 4 {
		return 0
	}
	return uint32(mask[0])<<24 | uint32(mask[1])<<16 | uint32(mask[2])<<8 | uint32(mask[3])
}

func uint32ToIPv4(value uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(value>>24),
		byte(value>>16),
		byte(value>>8),
		byte(value),
	)
}
