package httpapi

import (
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"testing"
)

func TestHandleV1MITMStatus_MethodNotAllowed(t *testing.T) {
	deps := &Deps{Cfg: cfgpkg.Config{}}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mitm/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleV1MITMStatus_MITMDisabled(t *testing.T) {
	deps := &Deps{
		Cfg: cfgpkg.Config{
			MITMEnabled: false,
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["enabled"] != false {
		t.Error("enabled should be false")
	}
	if resp["hasCA"] != false {
		t.Error("hasCA should be false")
	}
}

func TestHandleV1MITMStatus_MITMEnabledNoCA(t *testing.T) {
	deps := &Deps{
		Cfg: cfgpkg.Config{
			MITMEnabled: true,
		},
		MITM: nil,
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["enabled"] != true {
		t.Error("enabled should be true")
	}
	if resp["hasCA"] != false {
		t.Error("hasCA should be false when MITM is nil")
	}
}

func TestHandleV1MITMStatus_WithAllowDeny(t *testing.T) {
	deps := &Deps{
		Cfg: cfgpkg.Config{
			MITMEnabled:      true,
			MITMDomainsAllow: []string{"example.com", "test.org"},
			MITMDomainsDeny:  []string{"bad.com"},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	allow, ok := resp["allow"].([]interface{})
	if !ok || len(allow) != 2 {
		t.Error("allow domains not properly encoded")
	}

	deny, ok := resp["deny"].([]interface{})
	if !ok || len(deny) != 1 {
		t.Error("deny domains not properly encoded")
	}
}

func TestHandleV1MITMStatus_WithCA(t *testing.T) {
	// Generate a test CA
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	deps := &Deps{
		Cfg: cfgpkg.Config{
			MITMEnabled: true,
		},
		MITM: &MITM{CA: ca},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["hasCA"] != true {
		t.Error("hasCA should be true")
	}

	// Check that CA details are present
	if resp["caSubject"] == nil || resp["caSubject"] == "" {
		t.Error("caSubject should be present")
	}
	if resp["caSerial"] == nil || resp["caSerial"] == "" {
		t.Error("caSerial should be present")
	}
	if resp["caFingerprint"] == nil || resp["caFingerprint"] == "" {
		t.Error("caFingerprint should be present")
	}
}

func TestHandleV1MITMStatus_CAWithoutSerialNumber(t *testing.T) {
	// Generate a test CA
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	// Set serial number to nil to test that branch
	ca.RootCertificate().SerialNumber = nil

	deps := &Deps{
		Cfg: cfgpkg.Config{
			MITMEnabled: true,
		},
		MITM: &MITM{CA: ca},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["hasCA"] != true {
		t.Error("hasCA should be true")
	}
}

func TestHandleV1MITMGetCA_MethodNotAllowed(t *testing.T) {
	deps := &Deps{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mitm/ca", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMGetCA(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleV1MITMGetCA_NoCA(t *testing.T) {
	deps := &Deps{MITM: nil}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/ca", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMGetCA(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleV1MITMGetCA_Success(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	deps := &Deps{MITM: &MITM{CA: ca}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/ca", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMGetCA(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/x-pem-file" {
		t.Errorf("Content-Type = %s, want application/x-pem-file", ct)
	}

	if cd := w.Header().Get("Content-Disposition"); cd == "" {
		t.Error("Content-Disposition should be set")
	}

	// Check that PEM is present in body
	body := w.Body.String()
	if !contains(body, "-----BEGIN CERTIFICATE-----") {
		t.Error("response should contain PEM certificate")
	}
}

func TestHandleV1MITMGenerate_MethodNotAllowed(t *testing.T) {
	deps := &Deps{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/generate", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMGenerate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleV1MITMGenerate_DefaultCN(t *testing.T) {
	deps := &Deps{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mitm/generate", nil)
	w := httptest.NewRecorder()

	deps.handleV1MITMGenerate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["certPEM"] == "" {
		t.Error("certPEM should not be empty")
	}
	if resp["keyPEM"] == "" {
		t.Error("keyPEM should not be empty")
	}

	// Check that CA was loaded into deps
	if deps.MITM == nil || deps.MITM.CA == nil {
		t.Error("CA should be loaded into deps.MITM")
	}
}

func TestPemEncodeCert_ValidDER(t *testing.T) {
	// Generate a simple certificate
	certPEM, _, err := GenerateDevCA("Test", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Extract DER from PEM
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode generated certificate PEM")
	}

	w := httptest.NewRecorder()
	err = pemEncodeCert(w, block.Bytes)

	if err != nil {
		t.Errorf("pemEncodeCert failed: %v", err)
	}

	body := w.Body.String()
	if !contains(body, "-----BEGIN CERTIFICATE-----") {
		t.Error("output should contain BEGIN marker")
	}
	if !contains(body, "-----END CERTIFICATE-----") {
		t.Error("output should contain END marker")
	}
}

func TestPemEncodeCert_LineWrapping(t *testing.T) {
	// Create a larger DER to test line wrapping
	largeDER := make([]byte, 200)
	for i := range largeDER {
		largeDER[i] = byte(i % 256)
	}

	w := httptest.NewRecorder()
	err := pemEncodeCert(w, largeDER)

	if err != nil {
		t.Errorf("pemEncodeCert failed: %v", err)
	}

	body := w.Body.String()
	lines := splitLines(body)

	// Check that lines are wrapped at 64 characters (excluding newline)
	for i, line := range lines {
		if i == 0 || i == len(lines)-1 {
			continue // Skip BEGIN/END markers
		}
		if len(line) > 64 && !contains(line, "-----") {
			t.Errorf("line %d has length %d, should be <= 64", i, len(line))
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper to split string by newlines
func splitLines(s string) []string {
	var lines []string
	current := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
