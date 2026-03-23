package banner

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

type Strategy int

const (
	StrategyPassive Strategy = iota
	StrategyActive
)

type Probe struct {
	Strategy Strategy
	Payload  []byte
}

var activeProbes = map[int]Probe{
	80:    {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	443:   {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	3000:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	5000:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	5001:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	5678:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	6006:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	7001:  {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	8080:  {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	8000:  {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	8001:  {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	8443:  {Strategy: StrategyActive, Payload: []byte("HEAD / HTTP/1.0\r\n\r\n")},
	8888:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	7860:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	11434: {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	6379:  {Strategy: StrategyActive, Payload: []byte("PING\r\n")},
	9200:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	9300:  {Strategy: StrategyActive, Payload: []byte("GET / HTTP/1.0\r\n\r\n")},
	5432:  {Strategy: StrategyActive, Payload: []byte{0x00, 0x00, 0x00, 0x08, 0x04, 0xD2, 0x16, 0x2F}},
	27017: {Strategy: StrategyActive, Payload: []byte{0x3a, 0x00, 0x00, 0x00, 0x3a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd4, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x24, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x69, 0x73, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x00, 0x00, 0x00, 0x00, 0x00}},
	389:   {Strategy: StrategyActive, Payload: []byte{0x30, 0x1c, 0x02, 0x01, 0x01, 0x60, 0x17, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	445:   {Strategy: StrategyActive, Payload: []byte{0x00, 0x00, 0x00, 0x85, 0xff, 0x53, 0x4d, 0x42, 0x72, 0x00, 0x00, 0x00, 0x00, 0x18, 0x53, 0xc8, 0x00, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
}

var passivePorts = map[int]struct{}{
	21:   {},
	22:   {},
	23:   {},
	25:   {},
	110:  {},
	143:  {},
	3306: {},
	5900: {},
}

func Grab(ctx context.Context, host string, port int, timeout time.Duration) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	strategy := StrategyPassive
	probe := Probe{Strategy: StrategyPassive}
	if p, ok := activeProbes[port]; ok {
		strategy = p.Strategy
		probe = p
	} else if _, ok := passivePorts[port]; ok {
		strategy = StrategyPassive
	} else {
		strategy = StrategyPassive
	}

	dialer := net.Dialer{}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	if strategy == StrategyActive && len(probe.Payload) > 0 {
		if _, err := conn.Write(probe.Payload); err != nil {
			return "", err
		}
	}

	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return "", nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", nil
		}
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return "", nil
		}
	}

	if n <= 0 {
		return "", nil
	}
	return sanitizeBanner(string(buf[:n])), nil
}

func sanitizeBanner(banner string) string {
	banner = strings.ReplaceAll(banner, "\r", "")
	banner = strings.ReplaceAll(banner, "\n", " ")
	banner = strings.TrimSpace(banner)
	if len(banner) > 1024 {
		banner = banner[:1024]
	}
	return banner
}
