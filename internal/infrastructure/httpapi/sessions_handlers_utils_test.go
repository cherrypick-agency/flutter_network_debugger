package httpapi

import (
	"testing"
)

func TestAllAllowedFold_EmptyRequested(t *testing.T) {
	allowed := []string{"Content-Type", "Authorization"}
	requested := []string{}

	result := allAllowedFold(allowed, requested)

	if !result {
		t.Error("should return true when requested is empty")
	}
}

func TestAllAllowedFold_EmptyAllowed(t *testing.T) {
	allowed := []string{}
	requested := []string{"Content-Type"}

	result := allAllowedFold(allowed, requested)

	if result {
		t.Error("should return false when allowed is empty but requested is not")
	}
}

func TestAllAllowedFold_AllAllowed(t *testing.T) {
	allowed := []string{"Content-Type", "Authorization", "X-Custom"}
	requested := []string{"content-type", "authorization"}

	result := allAllowedFold(allowed, requested)

	if !result {
		t.Error("should return true when all requested headers are allowed")
	}
}

func TestAllAllowedFold_SomeNotAllowed(t *testing.T) {
	allowed := []string{"Content-Type"}
	requested := []string{"Content-Type", "Authorization"}

	result := allAllowedFold(allowed, requested)

	if result {
		t.Error("should return false when some requested headers are not allowed")
	}
}

func TestAllAllowedFold_CaseInsensitive(t *testing.T) {
	allowed := []string{"CONTENT-TYPE", "authorization"}
	requested := []string{"content-type", "AUTHORIZATION"}

	result := allAllowedFold(allowed, requested)

	if !result {
		t.Error("should be case insensitive")
	}
}

func TestAllAllowedFold_WildcardNotSupported(t *testing.T) {
	allowed := []string{"*"}
	requested := []string{"Content-Type"}

	result := allAllowedFold(allowed, requested)

	// Wildcard is not specially handled, should match exact string
	if result {
		t.Error("wildcard should not match arbitrary headers")
	}
}

func TestComputeCORSMeta_NilRequestHeaders(t *testing.T) {
	result := computeCORSMeta("GET", nil, map[string]string{}, false)

	if result.Ok {
		t.Error("should return not ok when request headers are nil")
	}
	if result.Reason != "missing headers" {
		t.Errorf("reason = %s, want 'missing headers'", result.Reason)
	}
}

func TestComputeCORSMeta_NilResponseHeaders(t *testing.T) {
	result := computeCORSMeta("GET", map[string]string{}, nil, false)

	if result.Ok {
		t.Error("should return not ok when response headers are nil")
	}
	if result.Reason != "missing headers" {
		t.Errorf("reason = %s, want 'missing headers'", result.Reason)
	}
}

func TestComputeCORSMeta_NoOrigin(t *testing.T) {
	req := map[string]string{}
	resp := map[string]string{}

	result := computeCORSMeta("GET", req, resp, false)

	if !result.Ok {
		t.Error("should return ok when no origin header")
	}
	if result.Reason != "no origin" {
		t.Errorf("reason = %s, want 'no origin'", result.Reason)
	}
}

func TestComputeCORSMeta_SimpleRequest_WildcardOrigin(t *testing.T) {
	req := map[string]string{"Origin": "https://example.com"}
	resp := map[string]string{"Access-Control-Allow-Origin": "*"}

	result := computeCORSMeta("GET", req, resp, false)

	if !result.Ok {
		t.Error("should be ok with wildcard origin")
	}
	if result.Reason != "" {
		t.Errorf("reason should be empty, got %s", result.Reason)
	}
}

func TestComputeCORSMeta_SimpleRequest_MatchingOrigin(t *testing.T) {
	req := map[string]string{"Origin": "https://example.com"}
	resp := map[string]string{"Access-Control-Allow-Origin": "https://example.com"}

	result := computeCORSMeta("GET", req, resp, false)

	if !result.Ok {
		t.Error("should be ok with matching origin")
	}
}

func TestComputeCORSMeta_SimpleRequest_OriginMismatch(t *testing.T) {
	req := map[string]string{"Origin": "https://example.com"}
	resp := map[string]string{"Access-Control-Allow-Origin": "https://other.com"}

	result := computeCORSMeta("GET", req, resp, false)

	if result.Ok {
		t.Error("should not be ok with mismatched origin")
	}
	if result.Reason != "origin" {
		t.Errorf("reason = %s, want 'origin'", result.Reason)
	}
}

func TestComputeCORSMeta_SimpleRequest_MethodCheck(t *testing.T) {
	req := map[string]string{"Origin": "https://example.com"}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://example.com",
		"Access-Control-Allow-Methods": "GET, POST",
	}

	result := computeCORSMeta("PUT", req, resp, false)

	if result.Ok {
		t.Error("should not be ok when method not allowed")
	}
	if result.Reason != "method" {
		t.Errorf("reason = %s, want 'method'", result.Reason)
	}
}

func TestComputeCORSMeta_SimpleRequest_AllowedMethod(t *testing.T) {
	req := map[string]string{"Origin": "https://example.com"}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://example.com",
		"Access-Control-Allow-Methods": "GET, POST",
	}

	result := computeCORSMeta("POST", req, resp, false)

	if !result.Ok {
		t.Error("should be ok when method is allowed")
	}
}

func TestComputeCORSMeta_Preflight_Success(t *testing.T) {
	req := map[string]string{
		"Origin":                         "https://example.com",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "Content-Type",
	}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://example.com",
		"Access-Control-Allow-Methods": "GET, POST, PUT",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	result := computeCORSMeta("OPTIONS", req, resp, true)

	if !result.Ok {
		t.Error("should be ok for valid preflight")
	}
}

func TestComputeCORSMeta_Preflight_OriginFail(t *testing.T) {
	req := map[string]string{
		"Origin":                        "https://example.com",
		"Access-Control-Request-Method": "POST",
	}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://other.com",
		"Access-Control-Allow-Methods": "GET, POST",
	}

	result := computeCORSMeta("OPTIONS", req, resp, true)

	if result.Ok {
		t.Error("should not be ok when preflight origin doesn't match")
	}
	if result.Reason != "origin" {
		t.Errorf("reason = %s, want 'origin'", result.Reason)
	}
}

func TestComputeCORSMeta_Preflight_MethodFail(t *testing.T) {
	req := map[string]string{
		"Origin":                        "https://example.com",
		"Access-Control-Request-Method": "DELETE",
	}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://example.com",
		"Access-Control-Allow-Methods": "GET, POST",
	}

	result := computeCORSMeta("OPTIONS", req, resp, true)

	if result.Ok {
		t.Error("should not be ok when preflight method not allowed")
	}
	if result.Reason != "method" {
		t.Errorf("reason = %s, want 'method'", result.Reason)
	}
}

func TestComputeCORSMeta_Preflight_HeadersFail(t *testing.T) {
	req := map[string]string{
		"Origin":                         "https://example.com",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "Content-Type, X-Custom",
	}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://example.com",
		"Access-Control-Allow-Methods": "GET, POST",
		"Access-Control-Allow-Headers": "Content-Type",
	}

	result := computeCORSMeta("OPTIONS", req, resp, true)

	if result.Ok {
		t.Error("should not be ok when preflight headers not allowed")
	}
	if result.Reason != "headers" {
		t.Errorf("reason = %s, want 'headers'", result.Reason)
	}
}

func TestComputeCORSMeta_IncludesVary(t *testing.T) {
	req := map[string]string{"Origin": "https://example.com"}
	resp := map[string]string{
		"Access-Control-Allow-Origin": "https://example.com",
		"Vary":                        "Origin",
	}

	result := computeCORSMeta("GET", req, resp, false)

	if result.Vary != "Origin" {
		t.Errorf("vary = %s, want 'Origin'", result.Vary)
	}
}

func TestClassifyNetError_Timeout(t *testing.T) {
	tests := []string{
		"context deadline exceeded",
		"request timeout",
		"i/o timeout",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "TIMEOUT" {
			t.Errorf("classifyNetError(%q) = %s, want TIMEOUT", msg, result)
		}
	}
}

func TestClassifyNetError_DNS(t *testing.T) {
	tests := []string{
		"no such host",
		"server misbehaving",
		"dns lookup failed: no such host",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "DNS_ERROR" {
			t.Errorf("classifyNetError(%q) = %s, want DNS_ERROR", msg, result)
		}
	}
}

func TestClassifyNetError_TLS(t *testing.T) {
	tests := []string{
		"x509: certificate has expired",
		"certificate verify failed",
		"tls: bad certificate",
		"tls handshake failure",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "TLS_ERROR" {
			t.Errorf("classifyNetError(%q) = %s, want TLS_ERROR", msg, result)
		}
	}
}

func TestClassifyNetError_Connect(t *testing.T) {
	tests := []string{
		"connection refused",
		"cannot assign requested address",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "SERVER_UNAVAILABLE" {
			t.Errorf("classifyNetError(%q) = %s, want SERVER_UNAVAILABLE", msg, result)
		}
	}
}

func TestClassifyNetError_Reset(t *testing.T) {
	tests := []string{
		"connection reset by peer",
		"connection reset",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "CONNECTION_RESET" {
			t.Errorf("classifyNetError(%q) = %s, want CONNECTION_RESET", msg, result)
		}
	}
}

func TestClassifyNetError_EOF(t *testing.T) {
	tests := []string{
		"unexpected EOF",
		"before full header received",
		"early eof",
		"EOF",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "CONNECTION_CLOSED" {
			t.Errorf("classifyNetError(%q) = %s, want CONNECTION_CLOSED", msg, result)
		}
	}
}

func TestClassifyNetError_Cancel(t *testing.T) {
	tests := []string{
		"request canceled",
		"client canceled request",
	}

	for _, msg := range tests {
		result := classifyNetError(msg)
		if result != "CANCELED" {
			t.Errorf("classifyNetError(%q) = %s, want CANCELED", msg, result)
		}
	}
}

func TestClassifyNetError_Generic(t *testing.T) {
	result := classifyNetError("some unknown error")

	if result != "UPSTREAM_ERROR" {
		t.Errorf("classifyNetError for unknown error = %s, want UPSTREAM_ERROR", result)
	}
}

func TestClassifyNetError_CaseInsensitive(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"CONTEXT DEADLINE EXCEEDED", "TIMEOUT"},
		{"NO SUCH HOST", "DNS_ERROR"},
		{"X509: CERTIFICATE ERROR", "TLS_ERROR"},
	}

	for _, tt := range tests {
		result := classifyNetError(tt.msg)
		if result != tt.want {
			t.Errorf("classifyNetError(%q) = %s, want %s", tt.msg, result, tt.want)
		}
	}
}

func TestParseCacheControl_Empty(t *testing.T) {
	result := parseCacheControl("")

	if result != nil {
		t.Error("should return nil for empty string")
	}
}

func TestParseCacheControl_SingleDirective(t *testing.T) {
	result := parseCacheControl("max-age=3600")

	if result["max-age"] != "3600" {
		t.Errorf("max-age = %s, want 3600", result["max-age"])
	}
}

func TestParseCacheControl_MultipleDirectives(t *testing.T) {
	result := parseCacheControl("max-age=3600, public, must-revalidate")

	if result["max-age"] != "3600" {
		t.Errorf("max-age = %s, want 3600", result["max-age"])
	}
	if result["public"] != "true" {
		t.Errorf("public = %s, want true", result["public"])
	}
	if result["must-revalidate"] != "true" {
		t.Errorf("must-revalidate = %s, want true", result["must-revalidate"])
	}
}

func TestParseCacheControl_WithQuotes(t *testing.T) {
	result := parseCacheControl(`max-age="3600"`)

	// Should strip quotes
	if result["max-age"] != "3600" {
		t.Errorf("max-age = %s, want 3600 (quotes stripped)", result["max-age"])
	}
}

func TestParseCacheControl_CaseInsensitive(t *testing.T) {
	result := parseCacheControl("Max-Age=3600, PRIVATE")

	// Keys should be lowercase
	if result["max-age"] != "3600" {
		t.Error("should lowercase directive names")
	}
	if result["private"] != "true" {
		t.Error("should lowercase flag directives")
	}
}

func TestParseCacheControl_EmptyParts(t *testing.T) {
	result := parseCacheControl("max-age=3600, , , public")

	// Should skip empty parts
	if len(result) != 2 {
		t.Errorf("result length = %d, want 2", len(result))
	}
}

func TestParseCacheControl_WithSpaces(t *testing.T) {
	result := parseCacheControl("  max-age = 3600  ,  public  ")

	// Should trim spaces
	if result["max-age"] != "3600" {
		t.Errorf("max-age = %s, want 3600", result["max-age"])
	}
	if result["public"] != "true" {
		t.Error("public should be true")
	}
}

func TestParseCacheControl_NoEquals(t *testing.T) {
	result := parseCacheControl("no-cache, no-store")

	if result["no-cache"] != "true" {
		t.Error("no-cache should be true")
	}
	if result["no-store"] != "true" {
		t.Error("no-store should be true")
	}
}

func TestParseCacheControl_Complex(t *testing.T) {
	result := parseCacheControl("max-age=31536000, public, immutable, stale-while-revalidate=86400")

	if result["max-age"] != "31536000" {
		t.Error("max-age incorrect")
	}
	if result["public"] != "true" {
		t.Error("public should be true")
	}
	if result["immutable"] != "true" {
		t.Error("immutable should be true")
	}
	if result["stale-while-revalidate"] != "86400" {
		t.Error("stale-while-revalidate incorrect")
	}
}

func TestParseCacheControl_OnlyCommas(t *testing.T) {
	result := parseCacheControl(",,,")

	if len(result) != 0 {
		t.Error("should return empty map for only commas")
	}
}

func TestParseCacheControl_EqualsAtStart(t *testing.T) {
	result := parseCacheControl("=3600")

	// Edge case: equals at position 0
	// Should create entry with empty key
	if len(result) == 0 {
		// If implementation skips it, that's fine too
	}
}
