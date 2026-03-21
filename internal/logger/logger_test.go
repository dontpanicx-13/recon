package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoggerWritesJSONLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recon.log")

	log, err := NewAt(path)
	if err != nil {
		t.Fatalf("new logger failed: %v", err)
	}
	t.Cleanup(func() { _ = log.Close() })

	log.Info("scan_start", map[string]any{"scan_id": "abc"})

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open log failed: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatalf("expected one log line")
	}
	line := scanner.Text()
	if !strings.Contains(line, "\"level\":\"info\"") {
		t.Fatalf("expected level info, got %s", line)
	}
	if !strings.Contains(line, "\"event\":\"scan_start\"") {
		t.Fatalf("expected event scan_start, got %s", line)
	}
	if !strings.Contains(line, "\"scan_id\":\"abc\"") {
		t.Fatalf("expected scan_id in log, got %s", line)
	}
}
