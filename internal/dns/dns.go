package dns

import (
	"context"
	"net"
	"strings"
)

func ReverseLookup(ctx context.Context, host string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	names, err := net.DefaultResolver.LookupAddr(ctx, host)
	if err != nil || len(names) == 0 {
		return "", err
	}
	return normalizeHostname(names[0]), nil
}

func normalizeHostname(name string) string {
	return strings.TrimSuffix(name, ".")
}
