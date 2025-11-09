package httpapi

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	scriptdomain "network-debugger/internal/features/scripting/domain"
)

// toScriptHTTPRequest converts http.Request to domain.HTTPRequest for scripts
func toScriptHTTPRequest(r *http.Request, body []byte) *scriptdomain.HTTPRequest {
	headers := make(map[string][]string)
	for k, v := range r.Header {
		headers[k] = v
	}

	// Extract the actual path that will be sent to upstream
	// Remove /httpproxy or /proxy prefix to get the real path for pattern matching
	urlPath := r.URL.Path
	if strings.HasPrefix(urlPath, "/httpproxy") {
		urlPath = strings.TrimPrefix(urlPath, "/httpproxy")
	} else if strings.HasPrefix(urlPath, "/proxy") {
		urlPath = strings.TrimPrefix(urlPath, "/proxy")
	}
	// Ensure path starts with /
	if urlPath == "" || !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}

	return &scriptdomain.HTTPRequest{
		Method:  r.Method,
		URL:     urlPath,
		Headers: headers,
		Body:    body,
	}
}

// toScriptHTTPResponse converts http.Response to domain.HTTPResponse for scripts
func toScriptHTTPResponse(resp *http.Response, body []byte) *scriptdomain.HTTPResponse {
	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}

	return &scriptdomain.HTTPResponse{
		Status:  resp.StatusCode,
		Headers: headers,
		Body:    body,
	}
}

// applyScriptRequestModifications applies modifications from script execution to http.Request
func applyScriptRequestModifications(r *http.Request, modified *scriptdomain.HTTPRequest) (*http.Request, []byte) {
	// Clone the request
	newReq := r.Clone(r.Context())

	// Apply method changes
	if modified.Method != "" && modified.Method != r.Method {
		newReq.Method = modified.Method
	}

	// Apply header changes
	if modified.Headers != nil {
		newReq.Header = make(http.Header)
		for k, v := range modified.Headers {
			newReq.Header[k] = v
		}
	}

	// Return modified body
	return newReq, modified.Body
}

// applyScriptResponseModifications applies modifications from script execution to http.Response
func applyScriptResponseModifications(resp *http.Response, modified *scriptdomain.HTTPResponse) *http.Response {
	// Apply status code changes
	if modified.Status > 0 {
		resp.StatusCode = modified.Status
		resp.Status = http.StatusText(modified.Status)
	}

	// Apply header changes
	if modified.Headers != nil {
		resp.Header = make(http.Header)
		for k, v := range modified.Headers {
			resp.Header[k] = v
		}
	}

	// Apply body changes
	if modified.Body != nil {
		resp.Body = io.NopCloser(bytes.NewReader(modified.Body))
		resp.ContentLength = int64(len(modified.Body))
	}

	return resp
}
