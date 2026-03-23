package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"recon/internal/scanner"
)

const (
	SchemaVersion = 1
)

type Manifest struct {
	SchemaVersion int            `json:"schema_version"`
	Scans         []ManifestItem `json:"scans"`
}

type ManifestItem struct {
	ScanID      string   `json:"scan_id"`
	File        string   `json:"file"`
	Targets     []string `json:"targets"`
	Date        string   `json:"date"`
	Status      string   `json:"status"`
	HostsFound  int      `json:"hosts_found"`
	OpenPorts   int      `json:"open_ports"`
	DurationMS  int64    `json:"duration_ms"`
	Label       string   `json:"label,omitempty"`
	HostsTotal  int      `json:"hosts_total,omitempty"`
	PortsTotal  int      `json:"ports_total,omitempty"`
	PortsProbed int      `json:"ports_probed,omitempty"`
	TargetsText string   `json:"targets_text,omitempty"`
}

type Store struct {
	BaseDir string
}

func New(baseDir string) *Store {
	return &Store{BaseDir: baseDir}
}

func Default() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Store{BaseDir: filepath.Join(base, "recon")}, nil
}

func (s *Store) Ensure() error {
	if s == nil {
		return errors.New("store is nil")
	}
	if s.BaseDir == "" {
		return errors.New("store base dir is empty")
	}
	if err := os.MkdirAll(s.scansDir(), 0o755); err != nil {
		return err
	}
	return nil
}

func (s *Store) SaveScan(result scanner.ScanResult, label string) (string, error) {
	if s == nil {
		return "", errors.New("store is nil")
	}
	if err := s.Ensure(); err != nil {
		return "", err
	}
	if result.ScanID == "" {
		return "", errors.New("scan id is empty")
	}

	fileRel := filepath.Join("scans", result.ScanID+".json")
	fileAbs := filepath.Join(s.BaseDir, fileRel)

	if err := writeJSON(fileAbs, result); err != nil {
		return "", err
	}

	item := ManifestItem{
		ScanID:      result.ScanID,
		File:        fileRel,
		Targets:     result.Config.Targets,
		Date:        result.Meta.Date,
		Status:      result.Meta.Status,
		HostsFound:  result.Summary.HostsFound,
		OpenPorts:   result.Summary.OpenPorts,
		DurationMS:  result.Meta.DurationMS,
		Label:       label,
		HostsTotal:  result.Summary.HostsTotal,
		PortsTotal:  result.Summary.PortsTotal,
		PortsProbed: result.Summary.PortsProbed,
		TargetsText: strings.Join(result.Config.Targets, ", "),
	}

	if err := s.appendManifest(item); err != nil {
		return "", err
	}

	return fileAbs, nil
}

func (s *Store) LoadManifest() (Manifest, error) {
	if s == nil {
		return Manifest{}, errors.New("store is nil")
	}
	path := s.manifestPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Manifest{SchemaVersion: SchemaVersion}, nil
		}
		return Manifest{}, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	if manifest.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("unsupported manifest schema: %d", manifest.SchemaVersion)
	}
	return manifest, nil
}

func (s *Store) LoadScan(scanID string) (scanner.ScanResult, error) {
	if s == nil {
		return scanner.ScanResult{}, errors.New("store is nil")
	}
	if scanID == "" {
		return scanner.ScanResult{}, errors.New("scan id is empty")
	}
	path := filepath.Join(s.BaseDir, "scans", scanID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return scanner.ScanResult{}, err
	}
	var result scanner.ScanResult
	if err := json.Unmarshal(data, &result); err != nil {
		return scanner.ScanResult{}, err
	}
	if result.SchemaVersion != scanner.SchemaVersion {
		return scanner.ScanResult{}, fmt.Errorf("unsupported scan schema: %d", result.SchemaVersion)
	}
	return result, nil
}

func (s *Store) DeleteScan(scanID string) error {
	if s == nil {
		return errors.New("store is nil")
	}
	if scanID == "" {
		return errors.New("scan id is empty")
	}
	if err := s.Ensure(); err != nil {
		return err
	}

	manifest, err := s.LoadManifest()
	if err != nil {
		return err
	}

	var (
		found   bool
		fileRel string
	)
	next := make([]ManifestItem, 0, len(manifest.Scans))
	for _, item := range manifest.Scans {
		if item.ScanID == scanID {
			found = true
			fileRel = item.File
			continue
		}
		next = append(next, item)
	}
	if !found {
		return fmt.Errorf("scan not found: %s", scanID)
	}
	manifest.Scans = next
	if err := writeJSON(s.manifestPath(), manifest); err != nil {
		return err
	}

	if fileRel == "" {
		fileRel = filepath.Join("scans", scanID+".json")
	}
	fileAbs := filepath.Join(s.BaseDir, fileRel)
	if err := os.Remove(fileAbs); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *Store) appendManifest(item ManifestItem) error {
	manifest, err := s.LoadManifest()
	if err != nil {
		return err
	}
	if manifest.SchemaVersion == 0 {
		manifest.SchemaVersion = SchemaVersion
	}
	manifest.Scans = append([]ManifestItem{item}, manifest.Scans...)
	return writeJSON(s.manifestPath(), manifest)
}

func (s *Store) manifestPath() string {
	return filepath.Join(s.BaseDir, "manifest.json")
}

func (s *Store) scansDir() string {
	return filepath.Join(s.BaseDir, "scans")
}

func writeJSON(path string, v any) error {
	tmp := path + ".tmp"
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
