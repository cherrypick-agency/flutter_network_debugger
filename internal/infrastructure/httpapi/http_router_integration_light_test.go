package httpapi

import (
    "net/http/httptest"
    "testing"
    cfgpkg "network-debugger/internal/infrastructure/config"
    obs "network-debugger/internal/infrastructure/observability"
    uc "network-debugger/internal/usecase"
)

func TestRouter_BasicEndpoints(t *testing.T) {
    // deps with stubs
    s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
    d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*", PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
    h := NewRouterWithoutForwardProxy(d)

    // healthz
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/healthz", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("healthz") }

    // readyz
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/readyz", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("readyz") }

    // api version
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/api/version", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("version") }

    // settings get
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/_api/v1/settings", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("settings get") }

    // mitm status
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/_api/v1/mitm/status", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("mitm status") }
}

func TestRouter_SessionsEndpointsAndMetrics(t *testing.T) {
    s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
    d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
    h := NewRouterWithoutForwardProxy(d)

    // list sessions
    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/_api/v1/sessions", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("list sessions") }

    // get session by id (not found)
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/_api/v1/sessions/unknown", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 404 { t.Fatalf("session by id 404 expected") }

    // metrics
    rr = httptest.NewRecorder()
    req = httptest.NewRequest("GET", "/metrics", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("metrics") }
}


