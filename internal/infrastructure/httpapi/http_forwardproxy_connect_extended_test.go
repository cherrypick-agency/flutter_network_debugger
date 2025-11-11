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

// func TestHandleConnectTunnel_Success(t *testing.T) {
// 	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		_, _ = w.Write([]byte("ok"))
// 	}))
// 	defer upstream.Close()

// 	store := mem.NewStore(16, 16, 0)
// 	d := &Deps{
// 		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
// 		Metrics: obs.NewMetrics(),
// 		Monitor: NewMonitorHub(),
// 		Svc:     usecase.NewSessionService(store, store, store),
// 	}

// 	rr := &testHijackableWriter{out: &bytes.Buffer{}}
// 	req := httptest.NewRequest(http.MethodConnect, "http://"+upstream.URL[7:], nil)
// 	req.Host = upstream.URL[7:]

// 	d.handleConnectTunnel(rr, req)

// 	if !bytes.Contains(rr.out.Bytes(), []byte("200 Connection Established")) {
// 		t.Fatalf("expected 200 Connection Established")
// 	}
// }

func TestHandleConnectTunnel_NoHijack(t *testing.T) {
	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)

	d.handleConnectTunnel(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestHandleConnectTunnel_InvalidHost(t *testing.T) {
	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "http://invalid-host:99999", nil)
	req.Host = "invalid-host:99999"

	d.handleConnectTunnel(rr, req)

	if !bytes.Contains(rr.out.Bytes(), []byte("502 Bad Gateway")) {
		t.Fatalf("expected 502 error")
	}
}

func TestHandleConnectTunnel_UnreachableHost(t *testing.T) {
	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "http://127.0.0.1:1", nil)
	req.Host = "127.0.0.1:1"

	d.handleConnectTunnel(rr, req)

	if !bytes.Contains(rr.out.Bytes(), []byte("502 Bad Gateway")) {
		t.Fatalf("expected 502 error")
	}
}

func TestHandleConnectTunnel_HTTPSHost(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "https://example.com:443", nil)
	req.Host = "example.com:443"

	d.handleConnectTunnel(rr, req)
}

// func TestHandleConnectTunnel_DataTransfer(t *testing.T) {
// 	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		_, _ = w.Write([]byte("response"))
// 	}))
// 	defer upstream.Close()

// 	store := mem.NewStore(16, 16, 0)
// 	d := &Deps{
// 		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
// 		Metrics: obs.NewMetrics(),
// 		Monitor: NewMonitorHub(),
// 		Svc:     usecase.NewSessionService(store, store, store),
// 	}

// 	rr := &testHijackableWriter{out: &bytes.Buffer{}}
// 	req := httptest.NewRequest(http.MethodConnect, "http://"+upstream.URL[7:], nil)
// 	req.Host = upstream.URL[7:]

// 	go func() {
// 		time.Sleep(100 * time.Millisecond)
// 	}()

// 	d.handleConnectTunnel(rr, req)

// 	if !bytes.Contains(rr.out.Bytes(), []byte("200 Connection Established")) {
// 		t.Fatalf("expected 200 Connection Established")
// 	}
// }
