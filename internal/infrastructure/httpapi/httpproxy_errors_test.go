package httpapi

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestHumanizeProxyError_Mappings(t *testing.T) {
	cases := []struct {
		in   error
		code string
	}{
		{ioEOF(), "CONNECTION_CLOSED"},
		{errors.New("connection refused"), "SERVER_UNAVAILABLE"},
		{dnsNoSuchHost(), "DNS_ERROR"},
		{errors.New("tls: handshake failure"), "TLS_ERROR"},
		{errors.New("network is unreachable"), "NETWORK_UNREACHABLE"},
		{errors.New("connection reset by peer"), "CONNECTION_RESET"},
		{errors.New("stopped after 10 redirects"), "TOO_MANY_REDIRECTS"},
		{context.DeadlineExceeded, "TIMEOUT"},
		{context.Canceled, "CANCELED"},
		{errors.New("i/o timeout"), "TIMEOUT"},
	}
	for _, tc := range cases {
		code, _ := humanizeProxyError(tc.in)
		if code != tc.code {
			t.Fatalf("%v => %s (got %s)", tc.in, tc.code, code)
		}
	}
}

// helpers to craft specific error types without importing io directly in table
func ioEOF() error { return errors.New("EOF") }

func dnsNoSuchHost() error {
	return &net.DNSError{Err: "no such host", Name: "bad.host"}
}
