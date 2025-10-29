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

func TestV1_DeleteByID_AndStream404(t *testing.T) {
    store := mem.NewStore(100, 100, 0)
    s := uc.NewSessionService(store, store, store)
    d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
    _ = s.Create(contextWithNoCancel(), domain.Session{ID: "s1", StartedAt: time.Unix(0,0).UTC()})
    h := NewRouterWithoutForwardProxy(d)
    // delete by id (v1)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions/s1", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusNoContent { t.Fatalf("v1 delete by id") }

    // sessions_stream missing id -> 404
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/api/sessions_stream/", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusNotFound { t.Fatalf("stream 404 expected") }
}


