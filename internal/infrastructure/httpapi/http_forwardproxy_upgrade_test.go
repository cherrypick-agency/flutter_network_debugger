package httpapi

import (
    "net/http"
    "net/http/httptest"
    cfgpkg "network-debugger/internal/infrastructure/config"
    mem "network-debugger/internal/adapters/storage/memory"
    obs "network-debugger/internal/infrastructure/observability"
    "network-debugger/internal/usecase"
    "testing"
)

func TestForwardProxy_Upgrade_NoHijack(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{}, Monitor: NewMonitorHub(), Metrics: obs.NewMetrics()}
    store := mem.NewStore(16, 16, 0)
    d.Svc = usecase.NewSessionService(store, store, store)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "http://example.com/ws", nil)
    req.Header.Set("Connection", "Upgrade")
    req.Header.Set("Upgrade", "websocket")
    d.handleForwardProxy(rr, req)
    if rr.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", rr.Code)
    }
    if !strContains(rr.Body.String(), "HIJACK_NOT_SUPPORTED") {
        t.Fatalf("expected HIJACK_NOT_SUPPORTED: %s", rr.Body.String())
    }
}

func TestForwardProxy_CONNECT_NoHijack(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{}, Monitor: NewMonitorHub(), Metrics: obs.NewMetrics()}
    store := mem.NewStore(16, 16, 0)
    d.Svc = usecase.NewSessionService(store, store, store)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
    d.handleForwardProxy(rr, req)
    if rr.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", rr.Code)
    }
    if !strContains(rr.Body.String(), "HIJACK_NOT_SUPPORTED") {
        t.Fatalf("expected HIJACK_NOT_SUPPORTED: %s", rr.Body.String())
    }
}


