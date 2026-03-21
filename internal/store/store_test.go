package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"recon/internal/scanner"
)

func TestSaveAndLoadManifest(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	result := scanner.ScanResult{
		SchemaVersion: scanner.SchemaVersion,
		ScanID:        "scan-1",
		Config: scanner.ScanConfig{
			Targets: []string{"127.0.0.1"},
		},
		Meta: scanner.ScanMeta{
			Date:       "2026-03-05T14:00:00Z",
			Status:     scanner.StatusCompleted,
			DurationMS: 1234,
		},
		Summary: scanner.ScanSummary{
			HostsTotal:  1,
			HostsFound:  1,
			PortsTotal:  1,
			PortsProbed: 1,
			OpenPorts:   1,
		},
	}

	fileAbs, err := s.SaveScan(result, "label")
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, err := os.Stat(fileAbs); err != nil {
		t.Fatalf("expected scan file to exist: %v", err)
	}

	manifest, err := s.LoadManifest()
	if err != nil {
		t.Fatalf("load manifest failed: %v", err)
	}
	if len(manifest.Scans) != 1 {
		t.Fatalf("expected 1 manifest entry, got %d", len(manifest.Scans))
	}
	item := manifest.Scans[0]
	if item.ScanID != "scan-1" {
		t.Fatalf("unexpected scan id: %s", item.ScanID)
	}
	if item.TargetsText != "127.0.0.1" {
		t.Fatalf("unexpected targets text: %s", item.TargetsText)
	}
}

func TestLoadScan(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	result := scanner.ScanResult{
		SchemaVersion: scanner.SchemaVersion,
		ScanID:        "scan-2",
		Config: scanner.ScanConfig{
			Targets: []string{"127.0.0.1"},
		},
		Meta: scanner.ScanMeta{
			Date:       "2026-03-05T14:00:00Z",
			Status:     scanner.StatusCompleted,
			DurationMS: 100,
		},
	}

	if _, err := s.SaveScan(result, ""); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := s.LoadScan("scan-2")
	if err != nil {
		t.Fatalf("load scan failed: %v", err)
	}
	if loaded.ScanID != "scan-2" {
		t.Fatalf("unexpected scan id: %s", loaded.ScanID)
	}
}

func TestLoadManifestMissingReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	manifest, err := s.LoadManifest()
	if err != nil {
		t.Fatalf("load manifest failed: %v", err)
	}
	if manifest.SchemaVersion != SchemaVersion {
		t.Fatalf("expected schema version %d, got %d", SchemaVersion, manifest.SchemaVersion)
	}
	if len(manifest.Scans) != 0 {
		t.Fatalf("expected empty scans")
	}
}

func TestSaveScanCreatesDirectories(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	s := New(dir)

	result := scanner.ScanResult{
		SchemaVersion: scanner.SchemaVersion,
		ScanID:        "scan-3",
		Meta: scanner.ScanMeta{
			Date:       "2026-03-05T14:00:00Z",
			Status:     scanner.StatusCompleted,
			DurationMS: 100,
		},
	}

	if _, err := s.SaveScan(result, ""); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "scans", "scan-3.json")); err != nil {
		t.Fatalf("expected scan file to exist: %v", err)
	}
}

func TestSaveScan_WritesScanJSON(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	result := scanner.ScanResult{
		SchemaVersion: scanner.SchemaVersion,
		ScanID:        "scan-json",
		Config: scanner.ScanConfig{
			Targets: []string{"127.0.0.1"},
			Ports:   []int{22, 80},
		},
		Meta: scanner.ScanMeta{
			Date:       "2026-03-05T14:00:00Z",
			Status:     scanner.StatusCompleted,
			DurationMS: 250,
		},
		Summary: scanner.ScanSummary{
			HostsTotal:  1,
			HostsFound:  1,
			PortsTotal:  2,
			PortsProbed: 2,
			OpenPorts:   1,
		},
	}

	if _, err := s.SaveScan(result, ""); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "scans", "scan-json.json"))
	if err != nil {
		t.Fatalf("read scan file failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload["scan_id"] != "scan-json" {
		t.Fatalf("unexpected scan_id: %v", payload["scan_id"])
	}
	if payload["schema_version"] == nil {
		t.Fatalf("missing schema_version")
	}
}

func TestSaveScan_WritesManifestJSON(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	result := scanner.ScanResult{
		SchemaVersion: scanner.SchemaVersion,
		ScanID:        "scan-manifest",
		Config: scanner.ScanConfig{
			Targets: []string{"127.0.0.1"},
		},
		Meta: scanner.ScanMeta{
			Date:       "2026-03-05T14:00:00Z",
			Status:     scanner.StatusCompleted,
			DurationMS: 100,
		},
		Summary: scanner.ScanSummary{
			HostsTotal:  1,
			HostsFound:  1,
			PortsTotal:  1,
			PortsProbed: 1,
			OpenPorts:   1,
		},
	}

	if _, err := s.SaveScan(result, "label"); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload["schema_version"] == nil {
		t.Fatalf("missing schema_version")
	}
	scans, ok := payload["scans"].([]any)
	if !ok || len(scans) != 1 {
		t.Fatalf("expected one scan entry")
	}
	first, ok := scans[0].(map[string]any)
	if !ok {
		t.Fatalf("invalid scan entry")
	}
	if first["scan_id"] != "scan-manifest" {
		t.Fatalf("unexpected scan_id: %v", first["scan_id"])
	}
}
