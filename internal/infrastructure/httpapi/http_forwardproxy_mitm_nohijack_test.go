package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleConnectMITM_NoHijack(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Fatalf("ca gen: %v", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}
	d := &Deps{MITM: &MITM{CA: ca}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	d.handleConnectMITM(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if !strContains(rr.Body.String(), "HIJACK_NOT_SUPPORTED") {
		t.Fatalf("expected HIJACK_NOT_SUPPORTED: %s", rr.Body.String())
	}
}
