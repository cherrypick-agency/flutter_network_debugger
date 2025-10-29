package httpapi

import (
	"net/http"
	"net/http/httptest"
	mem "network-debugger/internal/adapters/storage/memory"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
	"testing"
)

func TestWithForwardProxy_InterceptsCONNECT(t *testing.T) {
	d := &Deps{Monitor: NewMonitorHub(), Metrics: obs.NewMetrics()}
	store := mem.NewStore(8, 8, 0)
	d.Svc = usecase.NewSessionService(store, store, store)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Inner", "1")
		w.WriteHeader(http.StatusOK)
	})
	h := withForwardProxy(d, inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	h.ServeHTTP(rr, req)
	if rr.Header().Get("X-Inner") != "" {
		t.Fatalf("inner handler should not be called")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 from proxy path, got %d", rr.Code)
	}
}
