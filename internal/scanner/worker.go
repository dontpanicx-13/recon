package scanner

import (
	"context"
	"errors"
	"net"
	"strconv"
	"syscall"
	"time"
)

type BannerGrabber func(ctx context.Context, host string, port int, timeout time.Duration) (string, error)
type TLSInspector func(ctx context.Context, host string, port int, timeout time.Duration) (*TLSInfo, error)
type ServiceGuesser func(port int) string

type WorkerConfig struct {
	Context       context.Context
	Timeout       time.Duration
	BannerEnabled bool
	TLSEnabled    bool

	BannerGrabber  BannerGrabber
	TLSInspector   TLSInspector
	ServiceGuesser ServiceGuesser
}

func worker(tasks <-chan Task, results chan<- Result, cfg WorkerConfig) {
	ctx := cfg.Context
	if ctx == nil {
		ctx = context.Background()
	}

	for task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result := Result{Host: task.Host}
		state, err := probePort(cfg, task.Host, task.Port, &result.Port)
		if err != nil {
			result.Err = err
		}
		result.Port.Port = task.Port
		result.Port.State = state

		if cfg.ServiceGuesser != nil {
			result.Port.ServiceGuess = cfg.ServiceGuesser(task.Port)
		}

		results <- result
	}
}

func probePort(cfg WorkerConfig, host string, port int, portState *PortState) (string, error) {
	ctx := cfg.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	address := net.JoinHostPort(host, intToString(port))
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return PortFiltered, err
		}
		return classifyDialError(err), nil
	}
	conn.Close()

	if cfg.BannerEnabled && cfg.BannerGrabber != nil {
		if banner, bannerErr := cfg.BannerGrabber(ctx, host, port, cfg.Timeout); bannerErr == nil {
			if banner != "" {
				portState.Banner = &banner
			}
		}
	}

	if cfg.TLSEnabled && cfg.TLSInspector != nil {
		tlsInfo, tlsErr := cfg.TLSInspector(ctx, host, port, cfg.Timeout)
		if tlsErr != nil {
			portState.TLS = &TLSInfo{Note: tlsErr.Error()}
		} else {
			portState.TLS = tlsInfo
		}
	}

	return PortOpen, nil
}

func classifyDialError(err error) string {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return PortFiltered
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return PortClosed
	}
	return PortFiltered
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
