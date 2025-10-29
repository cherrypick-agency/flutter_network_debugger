package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
)

func TestHTTPProxy_SimpleGET(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer upstream.Close()

	// deps
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL+"/echo", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("proxy status: %d", rr.Code)
	}
}

func TestHTTPProxy_InvalidTarget(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("missing target -> 400 expected")
	}
}
