package ports

import "testing"

func TestTop100HasExpectedLength(t *testing.T) {
	ports := Top100()
	if len(ports) != 100 {
		t.Fatalf("expected 100 ports, got %d", len(ports))
	}
}

func TestTop1000HasExpectedLength(t *testing.T) {
	ports := Top1000()
	if len(ports) != 1000 {
		t.Fatalf("expected 1000 ports, got %d", len(ports))
	}
}

func TestPortsUniqueAndOrdered(t *testing.T) {
	ports := Top1000()
	seen := make(map[int]struct{}, len(ports))
	for _, port := range ports {
		if port < 1 || port > 65535 {
			t.Fatalf("invalid port %d", port)
		}
		if _, ok := seen[port]; ok {
			t.Fatalf("duplicate port %d", port)
		}
		seen[port] = struct{}{}
	}
}
