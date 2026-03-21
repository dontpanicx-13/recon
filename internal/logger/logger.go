package logger

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	mu   sync.Mutex
	file *os.File
}

func NewDefault() (*Logger, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(base, "recon", "recon.log")
	return NewAt(path)
}

func NewAt(path string) (*Logger, error) {
	if path == "" {
		return nil, errors.New("log path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

func (l *Logger) Info(event string, fields map[string]any) {
	l.write("info", event, fields)
}

func (l *Logger) Warn(event string, fields map[string]any) {
	l.write("warn", event, fields)
}

func (l *Logger) Error(event string, fields map[string]any) {
	l.write("error", event, fields)
}

func (l *Logger) write(level, event string, fields map[string]any) {
	if l == nil || l.file == nil {
		return
	}
	payload := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": level,
		"event": event,
	}
	for k, v := range fields {
		payload[k] = v
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.file.Write(append(data, '\n'))
}
