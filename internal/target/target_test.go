package target

import (
	"net"
	"os"
	"testing"
)

func TestParseSingleIPv4Valid(t *testing.T) {
	result, errs := Parse("192.168.1.1", Options{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
	if len(result.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "192.168.1.1" {
		t.Fatalf("expected target 192.168.1.1, got %q", result.Targets[0])
	}
}

func TestParseSingleIPv4Invalid(t *testing.T) {
	result, errs := Parse("999.1.1.1", Options{})
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestParseDomainResolvableIPv4(t *testing.T) {
	lookupCalls := 0
	opts := Options{
		LookupIP: func(host string) ([]net.IP, error) {
			lookupCalls++
			if host != "example.com" {
				t.Fatalf("unexpected lookup host %q", host)
			}
			return []net.IP{
				net.ParseIP("93.184.216.34"),
				net.ParseIP("2606:2800:220:1:248:1893:25c8:1946"),
			}, nil
		},
	}

	result, errs := Parse("example.com", opts)
	if lookupCalls != 1 {
		t.Fatalf("expected 1 lookup call, got %d", lookupCalls)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 1 {
		t.Fatalf("expected 1 IPv4 target, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "93.184.216.34" {
		t.Fatalf("expected target 93.184.216.34, got %q", result.Targets[0])
	}
}

func TestParseDomainNotResolvable(t *testing.T) {
	lookupCalls := 0
	opts := Options{
		LookupIP: func(host string) ([]net.IP, error) {
			lookupCalls++
			if host != "noexiste.invalid" {
				t.Fatalf("unexpected lookup host %q", host)
			}
			return nil, &net.DNSError{Err: "no such host", Name: host}
		},
	}

	result, errs := Parse("noexiste.invalid", opts)
	if lookupCalls != 1 {
		t.Fatalf("expected 1 lookup call, got %d", lookupCalls)
	}
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestParseCIDR24StandardLAN(t *testing.T) {
	result, errs := Parse("192.168.1.0/24", Options{ExcludeNetworkBroadcast: true})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 254 {
		t.Fatalf("expected 254 targets, got %d", len(result.Targets))
	}
	if result.Targets[0] != "192.168.1.1" {
		t.Fatalf("expected first target 192.168.1.1, got %q", result.Targets[0])
	}
	if result.Targets[len(result.Targets)-1] != "192.168.1.254" {
		t.Fatalf("expected last target 192.168.1.254, got %q", result.Targets[len(result.Targets)-1])
	}
}

func TestParseCIDR31PointToPoint(t *testing.T) {
	result, errs := Parse("10.0.0.0/31", Options{ExcludeNetworkBroadcast: true})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "10.0.0.0" {
		t.Fatalf("expected first target 10.0.0.0, got %q", result.Targets[0])
	}
	if result.Targets[1] != "10.0.0.1" {
		t.Fatalf("expected last target 10.0.0.1, got %q", result.Targets[1])
	}
}

func TestParseCIDR32SingleHost(t *testing.T) {
	result, errs := Parse("10.0.0.5/32", Options{ExcludeNetworkBroadcast: true})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "10.0.0.5" {
		t.Fatalf("expected target 10.0.0.5, got %q", result.Targets[0])
	}
}

func TestParseIPv6SingleUnsupported(t *testing.T) {
	result, errs := Parse("2001:db8::1", Options{})
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestParseIPv6CIDRUnsupported(t *testing.T) {
	result, errs := Parse("2001:db8::/64", Options{})
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestParseCommaSeparatedWithSpaces(t *testing.T) {
	result, errs := Parse("10.0.0.1, 10.0.0.2", Options{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "10.0.0.1" {
		t.Fatalf("expected first target 10.0.0.1, got %q", result.Targets[0])
	}
	if result.Targets[1] != "10.0.0.2" {
		t.Fatalf("expected second target 10.0.0.2, got %q", result.Targets[1])
	}
}

func TestParseDeduplication(t *testing.T) {
	result, errs := Parse("10.0.0.1,10.0.0.1", Options{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "10.0.0.1" {
		t.Fatalf("expected target 10.0.0.1, got %q", result.Targets[0])
	}
}

func TestParseFileInputValid(t *testing.T) {
	dir := t.TempDir()
	path := dir + string(os.PathSeparator) + "targets.txt"
	contents := "10.0.0.1\n192.168.1.0/30\nexample.com\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	lookupCalls := 0
	opts := Options{
		ExcludeNetworkBroadcast: true,
		LookupIP: func(host string) ([]net.IP, error) {
			lookupCalls++
			if host != "example.com" {
				t.Fatalf("unexpected lookup host %q", host)
			}
			return []net.IP{net.ParseIP("93.184.216.34")}, nil
		},
	}

	result, errs := Parse(path, opts)
	if lookupCalls != 1 {
		t.Fatalf("expected 1 lookup call, got %d", lookupCalls)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
	if len(result.Targets) != 4 {
		t.Fatalf("expected 4 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "10.0.0.1" {
		t.Fatalf("expected first target 10.0.0.1, got %q", result.Targets[0])
	}
	if result.Targets[1] != "192.168.1.1" {
		t.Fatalf("expected second target 192.168.1.1, got %q", result.Targets[1])
	}
	if result.Targets[2] != "192.168.1.2" {
		t.Fatalf("expected third target 192.168.1.2, got %q", result.Targets[2])
	}
	if result.Targets[3] != "93.184.216.34" {
		t.Fatalf("expected fourth target 93.184.216.34, got %q", result.Targets[3])
	}
}

func TestParseFileInputInvalidLines(t *testing.T) {
	dir := t.TempDir()
	path := dir + string(os.PathSeparator) + "targets.txt"
	contents := "# comment\n10.0.0.1\nbad domain\n10.0.0.1 10.0.0.2\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	opts := Options{
		LookupIP: func(host string) ([]net.IP, error) {
			return nil, &net.DNSError{Err: "no such host", Name: host}
		},
	}

	result, errs := Parse(path, opts)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
	if len(result.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d: %v", len(result.Targets), result.Targets)
	}
	if result.Targets[0] != "10.0.0.1" {
		t.Fatalf("expected target 10.0.0.1, got %q", result.Targets[0])
	}
	if len(result.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestParseFileInputUnreadable(t *testing.T) {
	result, errs := Parse("/path/does/not/exist.txt", Options{})
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestParseTokenLooksLikePathStatFails(t *testing.T) {
	result, errs := Parse("./missing.txt", Options{})
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestParseTokenWithWhitespace(t *testing.T) {
	result, errs := Parse("10.0.0.1 10.0.0.2", Options{})
	if len(result.Targets) != 0 {
		t.Fatalf("expected 0 targets, got %d: %v", len(result.Targets), result.Targets)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}
