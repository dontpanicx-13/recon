package scanner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	SchemaVersion = 1

	StatusCompleted = "completed"
	StatusAborted   = "aborted"
	StatusFailed    = "failed"
)

type EventKind string

const (
	EventScanStart EventKind = "scan_start"
	EventHostStart EventKind = "host_start"
	EventPort      EventKind = "port_result"
	EventHostDone  EventKind = "host_done"
	EventScanDone  EventKind = "scan_done"
)

type Event struct {
	Kind           EventKind
	Host           string
	Port           int
	State          string
	Service        string
	Banner         *string
	TLS            *TLSInfo
	Err            error
	PortsProbed    int
	PortsTotal     int
	HostsCompleted int
	HostsTotal     int
	Elapsed        time.Duration
}

type Scanner struct {
	BannerGrabber  BannerGrabber
	TLSInspector   TLSInspector
	ServiceGuesser ServiceGuesser
	ReverseLookup  func(ctx context.Context, host string) (string, error)
	Now            func() time.Time
}

type hostState struct {
	ports    []PortState
	probed   int
	open     bool
	hostname string
}

func NewScanner() *Scanner {
	return &Scanner{
		Now:           time.Now,
		ReverseLookup: defaultReverseLookup,
	}
}

func (s *Scanner) Scan(ctx context.Context, cfg ScanConfig, onEvent func(Event)) (ScanResult, error) {
	if err := validateConfig(&ctx, &cfg); err != nil {
		return ScanResult{}, err
	}

	timeout := resolveTimeout(cfg.TimeoutMS)
	now := s.now()
	scanID := newScanID(now)
	summary := newSummary(cfg)

	s.emit(onEvent, Event{
		Kind:       EventScanStart,
		PortsTotal: summary.PortsTotal,
		HostsTotal: summary.HostsTotal,
	})

	tasks := make(chan Task)
	results := make(chan Result)

	s.startWorkers(ctx, cfg, timeout, tasks, results)
	s.startProducer(ctx, cfg, tasks, onEvent)

	hostStates := s.collectResults(ctx, cfg, results, &summary, onEvent)

	return s.buildResult(ctx, cfg, scanID, now, hostStates, summary, onEvent), nil
}

func validateConfig(ctx *context.Context, cfg *ScanConfig) error {
	if ctx == nil || cfg == nil {
		return errors.New("invalid arguments")
	}
	if *ctx == nil {
		*ctx = context.Background()
	}
	if len(cfg.Targets) == 0 {
		return errors.New("no targets provided")
	}
	if len(cfg.Ports) == 0 {
		return errors.New("no ports provided")
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}

	return nil
}

func resolveTimeout(timeoutMS int) time.Duration {
	timeout := time.Duration(timeoutMS) * time.Millisecond
	if timeout <= 0 {
		return time.Second
	}
	return timeout
}

func newSummary(cfg ScanConfig) ScanSummary {
	return ScanSummary{
		HostsTotal: len(cfg.Targets),
		PortsTotal: len(cfg.Targets) * len(cfg.Ports),
	}
}

func (s *Scanner) emit(onEvent func(Event), evt Event) {
	if onEvent == nil {
		return
	}
	onEvent(evt)
}

func (s *Scanner) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Scanner) reverseLookup(ctx context.Context, host string) (string, error) {
	if s != nil && s.ReverseLookup != nil {
		return s.ReverseLookup(ctx, host)
	}
	return defaultReverseLookup(ctx, host)
}

func defaultReverseLookup(ctx context.Context, host string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	names, err := net.LookupAddr(host)
	if err != nil || len(names) == 0 {
		return "", err
	}
	return names[0], nil
}

func (s *Scanner) startWorkers(
	ctx context.Context,
	cfg ScanConfig,
	timeout time.Duration,
	tasks <-chan Task,
	results chan<- Result,
) {
	var wg sync.WaitGroup

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(tasks, results, WorkerConfig{
				Context:        ctx,
				Timeout:        timeout,
				BannerEnabled:  cfg.BannerGrabbing,
				TLSEnabled:     cfg.TLSAnalysis,
				BannerGrabber:  s.BannerGrabber,
				TLSInspector:   s.TLSInspector,
				ServiceGuesser: s.ServiceGuesser,
			})
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()
}

func (s *Scanner) startProducer(
	ctx context.Context,
	cfg ScanConfig,
	tasks chan<- Task,
	onEvent func(Event),
) {
	go func() {
		defer close(tasks)

		for _, host := range cfg.Targets {
			if ctx.Err() != nil {
				return
			}

			s.emit(onEvent, Event{
				Kind:       EventHostStart,
				Host:       host,
				PortsTotal: len(cfg.Ports),
			})

			for _, port := range cfg.Ports {
				select {
				case <-ctx.Done():
					return
				case tasks <- Task{Host: host, Port: port}:
				}
			}
		}
	}()
}

func (s *Scanner) collectResults(
	ctx context.Context,
	cfg ScanConfig,
	results <-chan Result,
	summary *ScanSummary,
	onEvent func(Event),
) map[string]*hostState {
	hostStates := make(map[string]*hostState, len(cfg.Targets))

	for res := range results {
		hs := hostStates[res.Host]
		if hs == nil {
			hs = &hostState{}
			hostStates[res.Host] = hs
		}

		hs.probed++
		summary.PortsProbed++

		if res.Port.State == PortOpen {
			summary.OpenPorts++
			hs.open = true
			hs.ports = append(hs.ports, res.Port)
		}

		s.emit(onEvent, Event{
			Kind:        EventPort,
			Host:        res.Host,
			Port:        res.Port.Port,
			State:       res.Port.State,
			Service:     res.Port.ServiceGuess,
			Banner:      res.Port.Banner,
			TLS:         res.Port.TLS,
			Err:         res.Err,
			PortsProbed: summary.PortsProbed,
			PortsTotal:  summary.PortsTotal,
		})

		if hs.probed == len(cfg.Ports) {
			s.handleHostComplete(ctx, cfg, res.Host, hs, summary, onEvent)
		}
	}

	return hostStates
}

func (s *Scanner) handleHostComplete(
	ctx context.Context,
	cfg ScanConfig,
	host string,
	hs *hostState,
	summary *ScanSummary,
	onEvent func(Event),
) {
	summary.HostsCompleted++

	if hs.open {
		summary.HostsFound++
		if cfg.ReverseDNS {
			if name, err := s.reverseLookup(ctx, host); err == nil {
				hs.hostname = name
			}
		}
	}

	s.emit(onEvent, Event{
		Kind:           EventHostDone,
		Host:           host,
		PortsProbed:    summary.PortsProbed,
		PortsTotal:     summary.PortsTotal,
		HostsCompleted: summary.HostsCompleted,
		HostsTotal:     summary.HostsTotal,
	})
}

func (s *Scanner) buildResult(
	ctx context.Context,
	cfg ScanConfig,
	scanID string,
	startTime time.Time,
	hostStates map[string]*hostState,
	summary ScanSummary,
	onEvent func(Event),
) ScanResult {
	status := StatusCompleted
	if ctx.Err() != nil {
		status = StatusAborted
	}
	duration := time.Since(startTime)

	result := ScanResult{
		SchemaVersion: SchemaVersion,
		ScanID:        scanID,
		Config:        cfg,
		Meta: ScanMeta{
			Date:       startTime.UTC().Format(time.RFC3339),
			Status:     status,
			DurationMS: duration.Milliseconds(),
		},
		Summary: summary,
	}

	for _, host := range cfg.Targets {
		hs := hostStates[host]
		if hs == nil || !hs.open {
			continue
		}
		sort.Slice(hs.ports, func(i, j int) bool {
			return hs.ports[i].Port < hs.ports[j].Port
		})
		result.Hosts = append(result.Hosts, Host{
			IP:       host,
			Hostname: hs.hostname,
			Ports:    hs.ports,
		})
	}

	s.emit(onEvent, Event{
		Kind:           EventScanDone,
		PortsProbed:    summary.PortsProbed,
		PortsTotal:     summary.PortsTotal,
		HostsCompleted: summary.HostsCompleted,
		HostsTotal:     summary.HostsTotal,
		Elapsed:        duration,
	})

	return result
}

func newScanID(now time.Time) string {
	suffix := "0000"
	if randSuffix, err := randomHex(2); err == nil {
		suffix = randSuffix
	}
	return now.UTC().Format("20060102-150405") + "-" + suffix
}

func randomHex(bytes int) (string, error) {
	if bytes <= 0 {
		return "", errors.New("bytes must be positive")
	}
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
