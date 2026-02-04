package extism

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

var httpFetchTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, // Connection timeout
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   10,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func withMaxTimeout(ctx context.Context, max time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), max)
	}
	if deadline, ok := ctx.Deadline(); ok {
		// If it's already shorter — leave it as is so we don't "extend" the timeout.
		if time.Until(deadline) <= max {
			return context.WithCancel(ctx)
		}
	}
	return context.WithTimeout(ctx, max)
}

func newHTTPFetchClient(allowedHosts []string) *http.Client {
	return &http.Client{
		Timeout:   0, // manage timeout via request context
		Transport: httpFetchTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// By default net/http limits redirects, but here we add our own checks.
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			if req.URL == nil {
				return errors.New("redirect URL is nil")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return fmt.Errorf("redirect to unsupported scheme: %q", req.URL.Scheme)
			}
			if strings.TrimSpace(req.URL.Host) == "" {
				return errors.New("redirect URL host is empty")
			}
			if !isHostAllowed(req.URL.Host, allowedHosts) {
				return fmt.Errorf("redirect to disallowed host: %s", req.URL.Host)
			}
			return nil
		},
	}
}

// createHostFunctions creates host functions that scripts can call
// These functions provide controlled access to system resources from WASM
// allowedHosts: list of hosts that scripts are allowed to fetch (nil means all hosts allowed)
func createHostFunctions(allowedHosts []string) []extism.HostFunction {
	return []extism.HostFunction{
		createLogFunction(),
		createHTTPFetchFunction(allowedHosts),
	}
}

// isHostAllowed checks if the target host is in the allowed hosts list
// Returns true if allowedHosts is nil/empty (all hosts allowed) or host matches
func isHostAllowed(targetHost string, allowedHosts []string) bool {
	// If no restrictions, allow all hosts
	if len(allowedHosts) == 0 {
		return true
	}

	// Normalize target host (remove port if present)
	host := targetHost
	if h, _, err := net.SplitHostPort(targetHost); err == nil {
		host = h
	}

	// Check if host matches any allowed host pattern
	for _, allowed := range allowedHosts {
		// Exact match
		if host == allowed {
			return true
		}
		// Wildcard subdomain match (e.g., "*.example.com" matches "api.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:] // remove "*."
			if host == domain || strings.HasSuffix(host, "."+domain) {
				return true
			}
		}
	}

	return false
}

// createLogFunction allows scripts to log messages
func createLogFunction() extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		"log",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			offset := stack[0]
			msg, err := plugin.ReadString(offset)
			if err != nil {
				log.Printf("[Script Log Error] Failed to read message: %v", err)
				return
			}

			// Log message from script
			log.Printf("[Script] %s", msg)
		},
		[]api.ValueType{api.ValueTypeI64}, // Input: message offset
		[]api.ValueType{},                 // Output: none
	)
}

// createHTTPFetchFunction allows scripts to make HTTP requests
// Security: Respects AllowedHosts from script config
// allowedHosts: list of allowed hosts (nil/empty means all hosts allowed)
func createHTTPFetchFunction(allowedHosts []string) extism.HostFunction {
	client := newHTTPFetchClient(allowedHosts)

	return extism.NewHostFunctionWithStack(
		"http_fetch",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			urlOffset := stack[0]
			urlStr, err := plugin.ReadString(urlOffset)
			if err != nil {
				errJSON := fmt.Sprintf(`{"error": "Failed to read URL: %v"}`, err)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}

			// Parse URL to extract host
			parsedURL, err := url.Parse(urlStr)
			if err != nil {
				errJSON := fmt.Sprintf(`{"error": "Invalid URL: %v"}`, err)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}

			// Limit protocols: only HTTP(S)
			if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
				errJSON := fmt.Sprintf(`{"error": "Unsupported URL scheme: %q"}`, parsedURL.Scheme)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}
			if strings.TrimSpace(parsedURL.Host) == "" {
				errJSON := `{"error": "URL host is empty"}`
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}

			// Security: Check if host is allowed
			if !isHostAllowed(parsedURL.Host, allowedHosts) {
				errJSON := fmt.Sprintf(`{"error": "Host '%s' is not in allowed hosts list"}`, parsedURL.Host)
				log.Printf("[Script Security] Blocked HTTP fetch to disallowed host: %s", parsedURL.Host)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}

			// Create HTTP request with context to respect script timeout
			// Even if ctx has no deadline (depends on Extism), set an upper limit.
			reqCtx, cancel := withMaxTimeout(ctx, 30*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, urlStr, nil)
			if err != nil {
				errJSON := fmt.Sprintf(`{"error": "Failed to create request: %v"}`, err)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				errJSON := fmt.Sprintf(`{"error": "HTTP request failed: %v"}`, err)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}
			defer resp.Body.Close()

			// Read response body with 10MB limit to prevent DoS
			const maxBodySize = 10 * 1024 * 1024 // 10 MB
			limitedBody := io.LimitReader(resp.Body, maxBodySize+1)
			body, err := io.ReadAll(limitedBody)
			if err != nil {
				errJSON := fmt.Sprintf(`{"error": "Failed to read response: %v"}`, err)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}
			if len(body) > maxBodySize {
				errJSON := fmt.Sprintf(`{"error": "Response body exceeds %d bytes"}`, maxBodySize)
				offset, writeErr := plugin.WriteString(errJSON)
				if writeErr != nil {
					log.Printf("[Script Error] Failed to write error response: %v", writeErr)
					stack[0] = 0
					return
				}
				stack[0] = offset
				return
			}

			// Return response as JSON
			responseJSON := fmt.Sprintf(
				`{"status": %d, "body": %q}`,
				resp.StatusCode,
				string(body),
			)
			offset, writeErr := plugin.WriteString(responseJSON)
			if writeErr != nil {
				log.Printf("[Script Error] Failed to write response: %v", writeErr)
				stack[0] = 0
				return
			}
			stack[0] = offset
		},
		[]api.ValueType{api.ValueTypeI64}, // Input: URL offset
		[]api.ValueType{api.ValueTypeI64}, // Output: response offset
	)
}
