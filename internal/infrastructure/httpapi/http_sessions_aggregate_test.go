package httpapi

import (
    "net/http"
    "net/http/httptest"
    "testing"
    mem "network-debugger/internal/adapters/storage/memory"
    uc "network-debugger/internal/usecase"
    cfgpkg "network-debugger/internal/infrastructure/config"
    obs "network-debugger/internal/infrastructure/observability"
    "network-debugger/internal/domain"
    "time"
)

func TestV1SessionsAggregate(t *testing.T) {
    store := mem.NewStore(100, 100, 0)
    s := uc.NewSessionService(store, store, store)
    d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
    _ = s.Create(contextWithNoCancel(), domain.Session{ID: "a", Target: "http://a.example/x", StartedAt: time.Unix(0,0).UTC()})
    _ = s.Create(contextWithNoCancel(), domain.Session{ID: "b", Target: "http://b.example/x", StartedAt: time.Unix(0,0).UTC()})
    _ = s.Create(contextWithNoCancel(), domain.Session{ID: "a2", Target: "http://a.example/y", StartedAt: time.Unix(0,0).UTC()})
    h := NewRouterWithoutForwardProxy(d)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/aggregate", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("aggregate") }
}


