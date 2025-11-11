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

func TestHandleHTTPForwardWebSocket_UpgradeSuccess(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != "websocket" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Sec-WebSocket-Accept", "test")
		w.WriteHeader(http.StatusSwitchingProtocols)
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
	req := httptest.NewRequest(http.MethodGet, "ws://"+upstream.URL[7:]+"/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Version", "13")

	d.handleHTTPForwardWebSocket(rr, req)
}

func TestHandleHTTPForwardWebSocket_NoHijack(t *testing.T) {
	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "ws://example.com/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")

	d.handleHTTPForwardWebSocket(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardWebSocket_UpstreamError(t *testing.T) {
	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := &testHijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodGet, "ws://127.0.0.1:1/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")

	d.handleHTTPForwardWebSocket(rr, req)

	if !bytes.Contains(rr.out.Bytes(), []byte("502 Bad Gateway")) {
		t.Fatalf("expected 502 error")
	}
}

func TestHandleHTTPForwardWebSocket_NonUpgradeResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not upgraded"))
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
	req := httptest.NewRequest(http.MethodGet, "ws://"+upstream.URL[7:]+"/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	d.handleHTTPForwardWebSocket(rr, req)
}

func TestHandleHTTPForwardWebSocket_HTTPUpgrade(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.WriteHeader(http.StatusSwitchingProtocols)
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
	req := httptest.NewRequest(http.MethodGet, "http://"+upstream.URL[7:]+"/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	d.handleHTTPForwardWebSocket(rr, req)
}

func TestHandleHTTPForwardWebSocket_WriteError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusSwitchingProtocols)
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
	req := httptest.NewRequest(http.MethodGet, "ws://"+upstream.URL[7:]+"/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	d.handleHTTPForwardWebSocket(rr, req)
}
