package httpapi

import (
    "net/http"
    "net/http/httptest"
    mem "network-debugger/internal/adapters/storage/memory"
    obs "network-debugger/internal/infrastructure/observability"
    "network-debugger/internal/usecase"
    "testing"
)

func TestHandleForwardOrNotFound_ProxyPath(t *testing.T) {
    d := &Deps{Monitor: NewMonitorHub(), Metrics: obs.NewMetrics()}
    store := mem.NewStore(8, 8, 0)
    d.Svc = usecase.NewSessionService(store, store, store)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
    d.handleForwardOrNotFound(rr, req)
    if rr.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", rr.Code)
    }
}


