package ports

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed top100.txt
var top100Raw string

//go:embed top1000.txt
var top1000Raw string

// Top100 returns the curated top 100 TCP ports.
// If the embedded list is empty, it falls back to ports 1-100.
func Top100() []int {
	ports := parseLines(top100Raw)
	if len(ports) == 0 {
		return rangePorts(1, 100)
	}
	return ports
}

// Top1000 returns the curated top 1000 TCP ports.
// If the embedded list is empty, it falls back to ports 1-1000.
func Top1000() []int {
	ports := parseLines(top1000Raw)
	if len(ports) == 0 {
		return rangePorts(1, 1000)
	}
	return ports
}

// All returns all TCP ports 1-65535.
func All() []int {
	return rangePorts(1, 65535)
}

func parseLines(raw string) []int {
	lines := strings.Split(raw, "\n")
	ports := make([]int, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		port, err := strconv.Atoi(line)
		if err != nil || port < 1 || port > 65535 {
			continue
		}
		ports = append(ports, port)
	}
	return ports
}

func rangePorts(start, end int) []int {
	if start < 1 {
		start = 1
	}
	if end > 65535 {
		end = 65535
	}
	if start > end {
		return nil
	}
	ports := make([]int, 0, end-start+1)
	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}
	return ports
}
