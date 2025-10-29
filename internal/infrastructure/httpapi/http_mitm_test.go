package httpapi

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleV1MITMStatus_AndGetCA_NoCA(t *testing.T) {
	d := &Deps{}
	// GET status
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/status", nil)
	d.handleV1MITMStatus(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status code")
	}

	// GET CA when absent -> 404
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/mitm/ca", nil)
	d.handleV1MITMGetCA(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when no CA")
	}
}

func TestHandleV1MITMGetCA_WithCA(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Skip("skip if crypto unavailable: ", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}
	d := &Deps{MITM: &MITM{CA: ca}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mitm/ca", nil)
	d.handleV1MITMGetCA(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !bytes.HasPrefix(rr.Body.Bytes(), []byte("-----BEGIN CERTIFICATE-----")) {
		t.Fatalf("pem header missing")
	}
}

func TestMITM_ShouldIntercept(t *testing.T) {
	m := &MITM{CA: &CertAuthority{}, AllowSuffix: []string{".ok"}}
	if !m.shouldIntercept("api.ok") {
		t.Fatalf("allow should pass")
	}
	if m.shouldIntercept("api.bad") {
		t.Fatalf("deny by default when allow present")
	}
	m = &MITM{CA: &CertAuthority{}, DenySuffix: []string{".bad"}}
	if m.shouldIntercept("x.bad") {
		t.Fatalf("deny suffix should block")
	}
	if !m.shouldIntercept("x.good") {
		t.Fatalf("no allow -> allow all")
	}
}

func TestPEMEncodeCert_WritesPEM(t *testing.T) {
	// возьмём валидный DER из сгенерированного dev CA
	certPEM, _, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Skip("skip crypto: ", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatalf("no pem block")
	}
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	rr := httptest.NewRecorder()
	if err := pemEncodeCert(rr, block.Bytes); err != nil {
		t.Fatalf("pemEncodeCert: %v", err)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("END CERTIFICATE")) {
		t.Fatalf("pem tail missing")
	}
}

func TestWriteRaw_WritesHTTPResponse(t *testing.T) {
	resp := &http.Response{StatusCode: 200, Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1, Header: http.Header{"X": []string{"1"}}, Body: io.NopCloser(bytes.NewReader([]byte("body")))}
	var buf bytes.Buffer
	if err := writeRaw(&buf, resp); err != nil {
		t.Fatalf("writeRaw: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("HTTP/1.1 200 OK")) {
		t.Fatalf("status line missing: %q", buf.String())
	}
}
