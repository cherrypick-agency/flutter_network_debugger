package httpapi

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
)

type proxyErrorInfo struct {
	Category    string
	Code        string
	UserMessage string
}

func classifyProxyError(err error) proxyErrorInfo {
	if err == nil {
		return proxyErrorInfo{
			Category:    "unknown",
			Code:        "UNKNOWN_ERROR",
			UserMessage: "Unknown error",
		}
	}

	// Type-based checks first (they are more reliable than string parsing).
	if errors.Is(err, context.Canceled) {
		return proxyErrorInfo{
			Category:    "cancel",
			Code:        "CANCELED",
			UserMessage: "Request cancelled",
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return proxyErrorInfo{
			Category:    "timeout",
			Code:        "TIMEOUT",
			UserMessage: "Request timeout",
		}
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && strings.Contains(strings.ToLower(dnsErr.Err), "no such host") {
		name := strings.TrimSpace(dnsErr.Name)
		if name != "" {
			return proxyErrorInfo{
				Category:    "dns",
				Code:        "DNS_ERROR",
				UserMessage: "Domain not found: " + name,
			}
		}
		return proxyErrorInfo{
			Category:    "dns",
			Code:        "DNS_ERROR",
			UserMessage: "Domain not found",
		}
	}
	if errors.Is(err, io.EOF) {
		return proxyErrorInfo{
			Category:    "eof",
			Code:        "CONNECTION_CLOSED",
			UserMessage: "Server closed connection unexpectedly",
		}
	}

	// Fallback: string-based classification.
	return classifyProxyErrorString(err.Error())
}

func classifyProxyErrorString(raw string) proxyErrorInfo {
	s := strings.TrimSpace(raw)
	if s == "" {
		return proxyErrorInfo{
			Category:    "unknown",
			Code:        "UNKNOWN_ERROR",
			UserMessage: "Unknown error",
		}
	}
	low := strings.ToLower(s)

	// Cancellation (most often: client aborted the request).
	if strings.Contains(low, "context canceled") ||
		strings.Contains(low, "context cancelled") ||
		strings.Contains(low, "operation was canceled") ||
		strings.Contains(low, "operation was cancelled") ||
		strings.Contains(low, "request canceled") ||
		strings.Contains(low, "request cancelled") ||
		strings.Contains(low, "client canceled") ||
		strings.Contains(low, "client cancelled") {
		return proxyErrorInfo{
			Category:    "cancel",
			Code:        "CANCELED",
			UserMessage: "Request cancelled",
		}
	}

	// Timeout.
	if strings.Contains(low, "context deadline exceeded") ||
		strings.Contains(low, "i/o timeout") ||
		(strings.Contains(low, "timeout") && !strings.Contains(low, "without timeout")) {
		return proxyErrorInfo{
			Category:    "timeout",
			Code:        "TIMEOUT",
			UserMessage: "Connection timeout",
		}
	}

	// DNS.
	if strings.Contains(low, "no such host") || strings.Contains(low, "server misbehaving") {
		return proxyErrorInfo{
			Category:    "dns",
			Code:        "DNS_ERROR",
			UserMessage: "Domain not found",
		}
	}

	// TLS.
	if strings.Contains(low, "first record does not look like a tls handshake") {
		return proxyErrorInfo{
			Category:    "tls",
			Code:        "TLS_HANDSHAKE_FAILED",
			UserMessage: "TLS handshake failed - target server may not support TLS",
		}
	}
	if strings.Contains(low, "x509") || strings.Contains(low, "certificate") || strings.Contains(low, "tls") {
		return proxyErrorInfo{
			Category:    "tls",
			Code:        "TLS_ERROR",
			UserMessage: "SSL/TLS certificate error",
		}
	}

	// Connectivity / routing.
	if strings.Contains(low, "network is unreachable") {
		return proxyErrorInfo{
			Category:    "network",
			Code:        "NETWORK_UNREACHABLE",
			UserMessage: "Network unreachable",
		}
	}

	// TCP connect.
	if strings.Contains(low, "connection refused") || strings.Contains(low, "cannot assign requested address") || strings.Contains(low, "cannot assign") {
		return proxyErrorInfo{
			Category:    "connect",
			Code:        "SERVER_UNAVAILABLE",
			UserMessage: "Server unavailable (connection refused)",
		}
	}

	// Reset.
	if strings.Contains(low, "connection reset") || strings.Contains(low, "reset by peer") {
		return proxyErrorInfo{
			Category:    "reset",
			Code:        "CONNECTION_RESET",
			UserMessage: "Connection reset by server",
		}
	}

	// EOF-ish.
	if low == "eof" ||
		strings.Contains(low, "unexpected eof") ||
		strings.Contains(low, "early eof") ||
		strings.Contains(low, "before full header") {
		return proxyErrorInfo{
			Category:    "eof",
			Code:        "CONNECTION_CLOSED",
			UserMessage: "Server closed connection unexpectedly",
		}
	}

	// Redirect loops.
	if strings.Contains(low, "stopped after") && strings.Contains(low, "redirect") {
		return proxyErrorInfo{
			Category:    "redirect",
			Code:        "TOO_MANY_REDIRECTS",
			UserMessage: "Too many redirects",
		}
	}

	// Default: keep original message, but provide stable code+category.
	return proxyErrorInfo{
		Category:    "upstream",
		Code:        "UPSTREAM_ERROR",
		UserMessage: s,
	}
}
