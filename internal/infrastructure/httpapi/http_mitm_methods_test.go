package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMITM_MethodNotAllowed(t *testing.T) {
	d := &Deps{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mitm/status", nil)
	d.handleV1MITMStatus(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status 405")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/mitm/ca", nil)
	d.handleV1MITMGetCA(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("ca 405")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/mitm/ca/generate", nil)
	d.handleV1MITMGenerate(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("generate 405")
	}
}
