package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mem "network-debugger/internal/adapters/storage/memory"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

func TestHandleHTTPForwardRequest_POST(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"received": string(body)})
	}))
	defer upstream.Close()

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	body := bytes.NewReader([]byte(`{"test": "data"}`))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, upstream.URL+"/api/test", body)
	req.Header.Set("Content-Type", "application/json")
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_PUT(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("updated"))
	}))
	defer upstream.Close()

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	body := bytes.NewReader([]byte("update data"))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, upstream.URL+"/resource", body)
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_DELETE(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, upstream.URL+"/resource/123", nil)
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_Headers(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Response-Header", "response-value")
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

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	req.Header.Set("X-Custom-Header", "test-value")
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Response-Header") != "response-value" {
		t.Errorf("expected X-Response-Header header")
	}
}

func TestHandleHTTPForwardRequest_UpstreamError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer upstream.Close()

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_Redirect(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "/target")
			w.WriteHeader(http.StatusFound)
			return
		}
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

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, upstream.URL+"/redirect", nil)
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_LargeBody(t *testing.T) {
	largeBody := make([]byte, 100000)
	for i := range largeBody {
		largeBody[i] = byte(i % 256)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != len(largeBody) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, upstream.URL, bytes.NewReader(largeBody))
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_QueryParameters(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("param1") != "value1" || query.Get("param2") != "value2" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, upstream.URL+"/path?param1=value1&param2=value2", nil)
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_HTTPS(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 512, InsecureTLS: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, upstream.URL, nil)
	d.handleHTTPForwardRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPForwardRequest_OfflineMode(t *testing.T) {
	store := mem.NewStore(16, 16, 0)
	d := &Deps{
		Cfg:     cfgpkg.Config{ThrottleOffline: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Svc:     usecase.NewSessionService(store, store, store),
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	d.handleForwardProxy(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}
