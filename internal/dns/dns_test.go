package dns

import (
	"context"
	"testing"
)

func TestReverseLookup_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ReverseLookup(ctx, "127.0.0.1")
	if err == nil {
		t.Fatalf("expected error for canceled context")
	}
}

func TestNormalizeHostname(t *testing.T) {
	got := normalizeHostname("example.com.")
	if got != "example.com" {
		t.Fatalf("expected trimmed hostname, got %q", got)
	}
}
