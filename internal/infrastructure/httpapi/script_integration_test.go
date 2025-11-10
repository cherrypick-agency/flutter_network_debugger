package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	scriptdomain "network-debugger/internal/features/scripting/domain"
)

func TestToScriptHTTPRequest_Basic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "value")
	body := []byte("test body")

	result := toScriptHTTPRequest(req, body)

	if result.Method != http.MethodGet {
		t.Errorf("Expected method %s, got %s", http.MethodGet, result.Method)
	}
	if result.URL != "/api/test" {
		t.Errorf("Expected URL /api/test, got %s", result.URL)
	}
	if result.Headers["Content-Type"][0] != "application/json" {
		t.Errorf("Expected Content-Type header, got %v", result.Headers["Content-Type"])
	}
	if result.Headers["X-Custom"][0] != "value" {
		t.Errorf("Expected X-Custom header, got %v", result.Headers["X-Custom"])
	}
	if !bytes.Equal(result.Body, body) {
		t.Errorf("Expected body %v, got %v", body, result.Body)
	}
}

func TestToScriptHTTPRequest_WithHttpproxyPrefix(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/httpproxy/api/users", nil)
	body := []byte("{}")

	result := toScriptHTTPRequest(req, body)

	if result.URL != "/api/users" {
		t.Errorf("Expected URL /api/users, got %s", result.URL)
	}
}

func TestToScriptHTTPRequest_WithProxyPrefix(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/proxy/api/users", nil)
	body := []byte("{}")

	result := toScriptHTTPRequest(req, body)

	if result.URL != "/api/users" {
		t.Errorf("Expected URL /api/users, got %s", result.URL)
	}
}

func TestToScriptHTTPRequest_EmptyPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/httpproxy", nil)
	body := []byte("")

	result := toScriptHTTPRequest(req, body)

	if result.URL != "/" {
		t.Errorf("Expected URL /, got %s", result.URL)
	}
}

func TestToScriptHTTPResponse_Basic(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Custom":     []string{"value"},
		},
	}
	body := []byte(`{"key": "value"}`)

	result := toScriptHTTPResponse(resp, body)

	if result.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, result.Status)
	}
	if result.Headers["Content-Type"][0] != "application/json" {
		t.Errorf("Expected Content-Type header, got %v", result.Headers["Content-Type"])
	}
	if !bytes.Equal(result.Body, body) {
		t.Errorf("Expected body %v, got %v", body, result.Body)
	}
}

func TestApplyScriptRequestModifications_MethodChange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	modified := &scriptdomain.HTTPRequest{
		Method: http.MethodPost,
		URL:    "/test",
		Body:   []byte("modified body"),
	}

	newReq, newBody := applyScriptRequestModifications(req, modified)

	if newReq.Method != http.MethodPost {
		t.Errorf("Expected method %s, got %s", http.MethodPost, newReq.Method)
	}
	if !bytes.Equal(newBody, modified.Body) {
		t.Errorf("Expected body %v, got %v", modified.Body, newBody)
	}
}

func TestApplyScriptRequestModifications_HeaderChange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Original", "value")
	modified := &scriptdomain.HTTPRequest{
		Method: http.MethodGet,
		URL:    "/test",
		Headers: map[string][]string{
			"New-Header": {"new-value"},
		},
		Body: []byte(""),
	}

	newReq, _ := applyScriptRequestModifications(req, modified)

	if newReq.Header.Get("Original") != "" {
		t.Errorf("Expected original header to be removed, got %s", newReq.Header.Get("Original"))
	}
	if newReq.Header.Get("New-Header") != "new-value" {
		t.Errorf("Expected New-Header to be 'new-value', got %s", newReq.Header.Get("New-Header"))
	}
}

func TestApplyScriptResponseModifications_StatusChange(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{},
		Body:       http.NoBody,
	}
	modified := &scriptdomain.HTTPResponse{
		Status: http.StatusNotFound,
		Body:   []byte("Not found"),
	}

	result := applyScriptResponseModifications(resp, modified)

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, result.StatusCode)
	}
}

func TestApplyScriptResponseModifications_BodyChange(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       http.NoBody,
	}
	modified := &scriptdomain.HTTPResponse{
		Status: http.StatusOK,
		Body:   []byte("modified body"),
	}

	result := applyScriptResponseModifications(resp, modified)

	if result.ContentLength != int64(len(modified.Body)) {
		t.Errorf("Expected ContentLength %d, got %d", len(modified.Body), result.ContentLength)
	}
}
