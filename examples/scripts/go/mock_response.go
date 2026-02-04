// Example Go script for creating mock response
// Compilation: tinygo build -o mock_response.wasm -target wasi mock_response.go

//go:build tinygo.wasm || wasm
// +build tinygo.wasm wasm

package main

import (
	"encoding/json"
	"strings"

	"github.com/extism/go-pdk"
)

type HTTPRequest struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

type HTTPResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

//export process
func process() int32 {
	// Read input (request)
	input := pdk.Input()

	var req HTTPRequest
	if err := json.Unmarshal(input, &req); err != nil {
		pdk.Log(pdk.LogError, "Failed to parse request: "+err.Error())
		return 1
	}

	pdk.Log(pdk.LogInfo, "Processing: "+req.Method+" "+req.URL)

	// If URL contains /api/mock - return mock response
	if strings.Contains(req.URL, "/api/mock") {
		mockResp := HTTPResponse{
			Status: 200,
			Headers: map[string][]string{
				"Content-Type":  {"application/json"},
				"X-Mocked-By":   {"Go Script"},
				"Cache-Control": {"no-cache"},
			},
			Body: []byte(`{
				"success": true,
				"message": "This is a mocked response from Go script",
				"timestamp": "` + getCurrentTime() + `",
				"data": {
					"id": 12345,
					"name": "Mock User",
					"email": "mock@example.com"
				}
			}`),
		}

		output, _ := json.Marshal(mockResp)
		pdk.OutputMemory(output)
		return 0
	}

	// Otherwise pass request without changes
	output, _ := json.Marshal(req)
	pdk.OutputMemory(output)
	return 0
}

func getCurrentTime() string {
	// Simple implementation without time package (not supported in TinyGo WASI)
	return "2025-11-03T12:00:00Z"
}

func main() {}
