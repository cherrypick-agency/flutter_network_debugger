package httpapi

import (
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
)

func TestV1Sessions_Endpoints(t *testing.T) {
	s := uc.NewSessionService(apiStubRepo{}, apiStubRepo{}, apiStubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 list")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/s1", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 get")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/s1/frames", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 frames")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/s1/events", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 events")
	}

	// http list may be absent in v1 mapping; skip
}
