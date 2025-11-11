package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	mem "network-debugger/internal/adapters/storage/memory"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

func TestHandleConnectMITM_CertificateError(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Fatalf("ca gen: %v", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		MITM:    &MITM{CA: ca},
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "http://invalid-host:443", nil)
	req.Host = "invalid-host:443"

	d.handleConnectMITM(rr, req)
}

func TestHandleConnectMITM_UpstreamDialError(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Fatalf("ca gen: %v", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		MITM:    &MITM{CA: ca},
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "http://127.0.0.1:1", nil)
	req.Host = "127.0.0.1:1"

	d.handleConnectMITM(rr, req)
}

func TestHandleConnectMITM_ShouldIntercept(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Fatalf("ca gen: %v", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}

	mitm := &MITM{CA: ca, AllowSuffix: []string{"example.com"}}

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		MITM:    mitm,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	req.Host = "example.com:443"

	d.handleConnectMITM(rr, req)
}

func TestHandleConnectMITM_ShouldNotIntercept(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Fatalf("ca gen: %v", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}

	mitm := &MITM{CA: ca, AllowSuffix: []string{"example.com"}}

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		MITM:    mitm,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "http://other.com:443", nil)
	req.Host = "other.com:443"

	d.handleConnectMITM(rr, req)
}
