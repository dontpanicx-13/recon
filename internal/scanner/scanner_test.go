package scanner

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

type eventRecorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *eventRecorder) onEvent(evt Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, evt)
}

func (r *eventRecorder) kinds() []EventKind {
	r.mu.Lock()
	defer r.mu.Unlock()
	kinds := make([]EventKind, 0, len(r.events))
	for _, evt := range r.events {
		kinds = append(kinds, evt.Kind)
	}
	return kinds
}

func (r *eventRecorder) eventsForPort(port int) []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Event
	for _, evt := range r.events {
		if evt.Port == port {
			out = append(out, evt)
		}
	}
	return out
}

func TestValidateConfig_NilArgs(t *testing.T) {
	if err := validateConfig(nil, nil); err == nil {
		t.Fatalf("expected error for nil args")
	}

	ctx := context.Background()
	if err := validateConfig(&ctx, nil); err == nil {
		t.Fatalf("expected error for nil config")
	}
}

func TestValidateConfig_DefaultsAndGuards(t *testing.T) {
	var ctx context.Context
	cfg := ScanConfig{Targets: []string{"127.0.0.1"}, Ports: []int{80}, Concurrency: 0}

	if err := validateConfig(&ctx, &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx == nil {
		t.Fatalf("expected ctx to be initialized")
	}
	if cfg.Concurrency != 1 {
		t.Fatalf("expected concurrency default 1, got %d", cfg.Concurrency)
	}
}

func TestValidateConfig_RejectsEmptyTargets(t *testing.T) {
	ctx := context.Background()
	cfg := ScanConfig{Ports: []int{80}}
	if err := validateConfig(&ctx, &cfg); err == nil {
		t.Fatalf("expected error for empty targets")
	}
}

func TestValidateConfig_RejectsEmptyPorts(t *testing.T) {
	ctx := context.Background()
	cfg := ScanConfig{Targets: []string{"127.0.0.1"}}
	if err := validateConfig(&ctx, &cfg); err == nil {
		t.Fatalf("expected error for empty ports")
	}
}

func TestResolveTimeout(t *testing.T) {
	if got := resolveTimeout(0); got != time.Second {
		t.Fatalf("expected 1s, got %v", got)
	}
	if got := resolveTimeout(-10); got != time.Second {
		t.Fatalf("expected 1s for negative, got %v", got)
	}
	if got := resolveTimeout(1500); got != 1500*time.Millisecond {
		t.Fatalf("expected 1500ms, got %v", got)
	}
}

func TestNewSummary(t *testing.T) {
	cfg := ScanConfig{Targets: []string{"a", "b"}, Ports: []int{1, 2, 3}}
	summary := newSummary(cfg)
	if summary.HostsTotal != 2 {
		t.Fatalf("expected HostsTotal 2, got %d", summary.HostsTotal)
	}
	if summary.PortsTotal != 6 {
		t.Fatalf("expected PortsTotal 6, got %d", summary.PortsTotal)
	}
}

func TestBuildResult_StatusDateOrderAndFilter(t *testing.T) {
	start := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	cfg := ScanConfig{Targets: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}, Ports: []int{80, 22}}
	hostStates := map[string]*hostState{
		"10.0.0.1": {
			open:  true,
			ports: []PortState{{Port: 80, State: PortOpen}, {Port: 22, State: PortOpen}},
		},
		"10.0.0.2": {open: false},
		"10.0.0.3": {
			open:  true,
			ports: []PortState{{Port: 443, State: PortOpen}, {Port: 21, State: PortOpen}},
		},
	}

	recorder := &eventRecorder{}
	result := (&Scanner{}).buildResult(context.Background(), cfg, "scan", start, hostStates, ScanSummary{}, recorder.onEvent)

	if result.Meta.Status != StatusCompleted {
		t.Fatalf("expected status completed, got %s", result.Meta.Status)
	}
	if result.Meta.Date != start.Format(time.RFC3339) {
		t.Fatalf("expected date %s, got %s", start.Format(time.RFC3339), result.Meta.Date)
	}
	if len(result.Hosts) != 2 {
		t.Fatalf("expected 2 hosts after filtering, got %d", len(result.Hosts))
	}
	if result.Hosts[0].IP != "10.0.0.1" || result.Hosts[1].IP != "10.0.0.3" {
		t.Fatalf("host order mismatch: %v", []string{result.Hosts[0].IP, result.Hosts[1].IP})
	}
	if len(result.Hosts[1].Ports) != 2 || result.Hosts[1].Ports[0].Port != 21 {
		t.Fatalf("ports not sorted or missing")
	}
	kinds := recorder.kinds()
	if len(kinds) == 0 || kinds[len(kinds)-1] != EventScanDone {
		t.Fatalf("expected scan_done event")
	}
}

func TestBuildResult_StatusAborted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := (&Scanner{}).buildResult(ctx, ScanConfig{}, "scan", time.Now(), map[string]*hostState{}, ScanSummary{}, nil)
	if result.Meta.Status != StatusAborted {
		t.Fatalf("expected status aborted, got %s", result.Meta.Status)
	}
}

func TestCollectResults_EventsAndSummary(t *testing.T) {
	cfg := ScanConfig{Targets: []string{"host"}, Ports: []int{80, 81}}
	summary := newSummary(cfg)
	recorder := &eventRecorder{}

	results := make(chan Result, 2)
	results <- Result{Host: "host", Port: PortState{Port: 80, State: PortOpen}}
	results <- Result{Host: "host", Port: PortState{Port: 81, State: PortClosed}}
	close(results)

	s := &Scanner{}
	hostStates := s.collectResults(context.Background(), cfg, results, &summary, recorder.onEvent)

	if summary.PortsProbed != 2 {
		t.Fatalf("expected PortsProbed 2, got %d", summary.PortsProbed)
	}
	if summary.OpenPorts != 1 {
		t.Fatalf("expected OpenPorts 1, got %d", summary.OpenPorts)
	}
	if summary.HostsCompleted != 1 {
		t.Fatalf("expected HostsCompleted 1, got %d", summary.HostsCompleted)
	}

	hs := hostStates["host"]
	if hs == nil || !hs.open || len(hs.ports) != 1 {
		t.Fatalf("expected open host with one port")
	}

	kinds := recorder.kinds()
	portEvents := 0
	hostDoneEvents := 0
	for _, kind := range kinds {
		if kind == EventPort {
			portEvents++
		}
		if kind == EventHostDone {
			hostDoneEvents++
		}
	}
	if portEvents != 2 || hostDoneEvents != 1 {
		t.Fatalf("expected 2 port events and 1 host_done, got %d and %d", portEvents, hostDoneEvents)
	}
}

func TestHandleHostComplete_ReverseLookup(t *testing.T) {
	s := &Scanner{ReverseLookup: func(ctx context.Context, host string) (string, error) {
		return "example.local", nil
	}}
	cfg := ScanConfig{ReverseDNS: true}
	summary := &ScanSummary{}
	hs := &hostState{open: true}
	recorder := &eventRecorder{}

	s.handleHostComplete(context.Background(), cfg, "host", hs, summary, recorder.onEvent)

	if summary.HostsCompleted != 1 {
		t.Fatalf("expected HostsCompleted 1, got %d", summary.HostsCompleted)
	}
	if summary.HostsFound != 1 {
		t.Fatalf("expected HostsFound 1, got %d", summary.HostsFound)
	}
	if hs.hostname != "example.local" {
		t.Fatalf("expected hostname to be set")
	}

	kinds := recorder.kinds()
	if len(kinds) != 1 || kinds[0] != EventHostDone {
		t.Fatalf("expected single host_done event")
	}
}

func TestEventFlowOrdering(t *testing.T) {
	cfg := ScanConfig{Targets: []string{"host"}, Ports: []int{80, 81}}
	summary := newSummary(cfg)
	recorder := &eventRecorder{}
	s := &Scanner{}

	results := make(chan Result, 2)
	s.emit(recorder.onEvent, Event{Kind: EventScanStart})
	s.emit(recorder.onEvent, Event{Kind: EventHostStart, Host: "host", PortsTotal: len(cfg.Ports)})
	results <- Result{Host: "host", Port: PortState{Port: 80, State: PortOpen}}
	results <- Result{Host: "host", Port: PortState{Port: 81, State: PortOpen}}
	close(results)

	hostStates := s.collectResults(context.Background(), cfg, results, &summary, recorder.onEvent)
	s.buildResult(context.Background(), cfg, "scan", time.Now(), hostStates, summary, recorder.onEvent)

	kinds := recorder.kinds()
	want := []EventKind{EventScanStart, EventHostStart, EventPort, EventPort, EventHostDone, EventScanDone}
	if len(kinds) != len(want) {
		t.Fatalf("expected %d events, got %d", len(want), len(kinds))
	}
	for i, kind := range want {
		if kinds[i] != kind {
			t.Fatalf("event order mismatch at %d: expected %s got %s", i, kind, kinds[i])
		}
	}
}

func TestContextCancellationAbortsScan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	recorder := &eventRecorder{}
	s := NewScanner()
	cfg := ScanConfig{Targets: []string{"127.0.0.1"}, Ports: []int{80}, Concurrency: 1}

	result, err := s.Scan(ctx, cfg, recorder.onEvent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Meta.Status != StatusAborted {
		t.Fatalf("expected status aborted, got %s", result.Meta.Status)
	}

	kinds := recorder.kinds()
	if len(kinds) == 0 || kinds[len(kinds)-1] != EventScanDone {
		t.Fatalf("expected scan_done event on cancellation")
	}
}

func TestIntegrationLite_LocalListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("listen not permitted in this environment: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	openPort := listener.Addr().(*net.TCPAddr).Port

	closedListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	closedPort := closedListener.Addr().(*net.TCPAddr).Port
	_ = closedListener.Close()

	recorder := &eventRecorder{}
	s := NewScanner()
	cfg := ScanConfig{Targets: []string{"127.0.0.1"}, Ports: []int{openPort, closedPort}, Concurrency: 1, TimeoutMS: 200}

	result, err := s.Scan(context.Background(), cfg, recorder.onEvent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hosts) != 1 {
		t.Fatalf("expected one host in results, got %d", len(result.Hosts))
	}
	if len(result.Hosts[0].Ports) != 1 || result.Hosts[0].Ports[0].Port != openPort {
		t.Fatalf("expected open port %d in results", openPort)
	}

	closedEvents := recorder.eventsForPort(closedPort)
	if len(closedEvents) == 0 {
		t.Fatalf("expected port_result for closed port")
	}
	state := closedEvents[0].State
	if state != PortClosed && state != PortFiltered {
		t.Fatalf("expected closed or filtered state, got %s", state)
	}
}
