package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/extism/go-pdk"
)

// ScriptContext is passed from the debugger
type ScriptContext struct {
	Request *HTTPRequest `json:"request,omitempty"`
	Session *SessionInfo `json:"session,omitempty"`
}

// HTTPRequest represents an HTTP request
type HTTPRequest struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

// HTTPResponse for rate limit error
type HTTPResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

// SessionInfo contains client metadata
type SessionInfo struct {
	ID         string `json:"id"`
	ClientAddr string `json:"client_addr"`
}

// ScriptResult returned to debugger
type ScriptResult struct {
	Modified         bool          `json:"modified"`
	ModifiedResponse *HTTPResponse `json:"modifiedResponse,omitempty"`
	Logs             []string      `json:"logs"`
}

const (
	maxRequestsPerMinute = 10
	rateLimitWindow      = time.Minute
)

//export transform
func transform() int32 {
	logs := []string{"Script started: Rate Limiter"}

	// Read input from Extism host
	input := pdk.Input()
	var ctx ScriptContext
	if err := json.Unmarshal(input, &ctx); err != nil {
		logs = append(logs, fmt.Sprintf("Failed to parse input: %v", err))
		outputError(logs)
		return 1
	}

	// Get client IP from session
	clientIP := "unknown"
	if ctx.Session != nil {
		clientIP = ctx.Session.ClientAddr
	}
	logs = append(logs, fmt.Sprintf("Client IP: %s", clientIP))

	// Check rate limit
	count := incrementCounter(clientIP)
	logs = append(logs, fmt.Sprintf("Request count: %d/%d", count, maxRequestsPerMinute))

	if count > maxRequestsPerMinute {
		logs = append(logs, "RATE LIMIT EXCEEDED - Blocking request")

		// Return 429 Too Many Requests
		errorBody := fmt.Sprintf(`{"error": "Rate limit exceeded", "limit": %d, "window": "%s"}`,
			maxRequestsPerMinute, rateLimitWindow)

		response := &HTTPResponse{
			Status: 429,
			Headers: map[string][]string{
				"Content-Type":      {"application/json"},
				"Retry-After":       {"60"},
				"X-RateLimit-Limit": {fmt.Sprintf("%d", maxRequestsPerMinute)},
			},
			Body: []byte(errorBody),
		}

		result := ScriptResult{
			Modified:         true,
			ModifiedResponse: response,
			Logs:             logs,
		}

		output, _ := json.Marshal(result)
		pdk.OutputMemory(output)
		return 0
	}

	logs = append(logs, "Rate limit OK - Allowing request")

	// Allow request (no modification)
	result := ScriptResult{
		Modified: false,
		Logs:     logs,
	}

	output, _ := json.Marshal(result)
	pdk.OutputMemory(output)
	return 0
}

func outputError(logs []string) {
	result := ScriptResult{
		Modified: false,
		Logs:     logs,
	}
	output, _ := json.Marshal(result)
	pdk.OutputMemory(output)
}

func main() {}
