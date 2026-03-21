package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type portRank struct {
	Port int
	Freq float64
}

func main() {
	inPath := flag.String("in", "", "Path to nmap-services file")
	flag.Parse()

	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "missing -in path to nmap-services")
		os.Exit(2)
	}

	ranked, err := loadRanks(*inPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse nmap-services:", err)
		os.Exit(1)
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Freq == ranked[j].Freq {
			return ranked[i].Port < ranked[j].Port
		}
		return ranked[i].Freq > ranked[j].Freq
	})

	if err := writeList("internal/ports/top100.txt", ranked, 100); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write top100:", err)
		os.Exit(1)
	}
	if err := writeList("internal/ports/top1000.txt", ranked, 1000); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write top1000:", err)
		os.Exit(1)
	}
}

func loadRanks(path string) ([]portRank, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	seen := make(map[int]struct{})
	var ranked []portRank

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		portProto := fields[1]
		parts := strings.Split(portProto, "/")
		if len(parts) != 2 || parts[1] != "tcp" {
			continue
		}

		port, err := strconv.Atoi(parts[0])
		if err != nil || port < 1 || port > 65535 {
			continue
		}

		freq, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			continue
		}

		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		ranked = append(ranked, portRank{Port: port, Freq: freq})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ranked, nil
}

func writeList(path string, ranked []portRank, n int) error {
	if n > len(ranked) {
		n = len(ranked)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < n; i++ {
		if _, err := fmt.Fprintf(file, "%d\n", ranked[i].Port); err != nil {
			return err
		}
	}
	return nil
}
