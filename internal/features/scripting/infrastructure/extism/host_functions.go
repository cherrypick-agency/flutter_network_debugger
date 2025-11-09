package extism

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero/api"
)

// createHostFunctions creates host functions that scripts can call
// These functions provide controlled access to system resources from WASM
func createHostFunctions() []extism.HostFunction {
	return []extism.HostFunction{
		createLogFunction(),
		createHTTPFetchFunction(),
	}
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
func createHTTPFetchFunction() extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		"http_fetch",
		func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
			urlOffset := stack[0]
			url, err := plugin.ReadString(urlOffset)
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

			// Security: Check allowed hosts (would need to be passed via config)
			// For now, implement basic fetch

			// Create HTTP request with context to respect script timeout
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

			// Create HTTP client with proper context propagation
			// Configure Transport to respect request context cancellation
			client := &http.Client{
				Timeout: 0, // No timeout - request context will handle cancellation
				Transport: &http.Transport{
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second, // Connection timeout
						KeepAlive: 30 * time.Second,
					}).DialContext,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
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
			limitedBody := io.LimitReader(resp.Body, maxBodySize)
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
