package tlsinfo

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"recon/internal/scanner"
)

func Inspect(ctx context.Context, host string, port int, timeout time.Duration) (*scanner.TLSInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	tlsConn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	})
	defer tlsConn.Close()

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, err
	}

	state := tlsConn.ConnectionState()
	info := &scanner.TLSInfo{
		TLSVersion: tlsVersionString(state.Version),
		Cipher:     tls.CipherSuiteName(state.CipherSuite),
	}

	if len(state.PeerCertificates) == 0 {
		info.Note = "no peer certificate"
		return info, nil
	}

	cert := state.PeerCertificates[0]
	info.CommonName = cert.Subject.CommonName
	info.SAN = cert.DNSNames
	if len(cert.Issuer.Organization) > 0 {
		info.Issuer = cert.Issuer.Organization[0]
	} else if cert.Issuer.CommonName != "" {
		info.Issuer = cert.Issuer.CommonName
	}
	info.Expires = cert.NotAfter.Format("2006-01-02")

	return info, nil
}

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}
