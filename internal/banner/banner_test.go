package banner

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestGrab_PassiveBanner(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("listen not permitted in this environment: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("HELLO-BANNER\r\n"))
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	banner, err := Grab(context.Background(), "127.0.0.1", port, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if banner != "HELLO-BANNER" {
		t.Fatalf("expected banner, got %q", banner)
	}
}

func TestGrab_ActiveProbe(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("listen not permitted in this environment: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 128)
		_, _ = conn.Read(buf)
		_, _ = conn.Write([]byte("ACTIVE-BANNER\r\n"))
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	activeProbes[port] = Probe{Strategy: StrategyActive, Payload: []byte("PING\r\n")}
	t.Cleanup(func() { delete(activeProbes, port) })

	banner, err := Grab(context.Background(), "127.0.0.1", port, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if banner != "ACTIVE-BANNER" {
		t.Fatalf("expected banner, got %q", banner)
	}
}

func TestGrab_NoBannerDoesNotBlock(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("listen not permitted in this environment: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(200 * time.Millisecond)
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	start := time.Now()
	banner, err := Grab(context.Background(), "127.0.0.1", port, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if banner != "" {
		t.Fatalf("expected empty banner, got %q", banner)
	}
	if time.Since(start) > 300*time.Millisecond {
		t.Fatalf("grab took too long")
	}
}

func TestGrab_SanitizesBanner(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("listen not permitted in this environment: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("HELLO\r\nWORLD\r\n"))
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	banner, err := Grab(context.Background(), "127.0.0.1", port, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if banner != "HELLO WORLD" {
		t.Fatalf("expected sanitized banner, got %q", banner)
	}
}
