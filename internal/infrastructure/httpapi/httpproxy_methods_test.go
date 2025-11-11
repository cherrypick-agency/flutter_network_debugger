package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"

	"github.com/rs/zerolog"
)

func TestHandleHTTPProxy_POST(t *testing.T) {
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

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	body := bytes.NewReader([]byte(`{"test": "data"}`))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/httpproxy/api/test?_target="+upstream.URL, body)
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_PUT(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("updated"))
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	body := bytes.NewReader([]byte("update data"))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/httpproxy/resource?_target="+upstream.URL, body)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_DELETE(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/httpproxy/resource/123?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_PATCH(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("patched"))
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	body := bytes.NewReader([]byte(`{"field": "value"}`))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/httpproxy/resource?_target="+upstream.URL, body)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_HeadersForwarding(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Response-Header", "response-value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(customHeader))
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, StealthHeaders: false},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	req.Header.Set("X-Custom-Header", "test-value")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Response-Header") != "response-value" {
		t.Errorf("expected X-Response-Header header")
	}
}

func TestHandleHTTPProxy_QueryParameters(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("param1") != "value1" || query.Get("param2") != "value2" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy/path?param1=value1&param2=value2&_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_UpstreamError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_UpstreamTimeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestTimeout)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestTimeout {
		t.Fatalf("expected 408, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_Redirect(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "/target")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy/redirect?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_DefaultTarget(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, DefaultTarget: upstream.URL},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy/path", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_InvalidTargetScheme(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target=ftp://example.com", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_OfflineMode(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{ThrottleOffline: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target=http://example.com", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_LargeBody(t *testing.T) {
	largeBody := strings.Repeat("a", 100000)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != len(largeBody) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/httpproxy?_target="+upstream.URL, strings.NewReader(largeBody))
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_StealthMode(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-For") != "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, StealthHeaders: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHTTPProxy_PathPrefix(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     s,
	}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy/api/v1/users?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
