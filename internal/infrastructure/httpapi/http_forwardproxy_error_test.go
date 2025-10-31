package httpapi

import (
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
)

func TestWithForwardProxy_AbsoluteURI_Error502(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 64, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := withForwardProxy(d, NewRouterWithoutForwardProxy(d))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1/", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("forward proxy error -> 502 expected, got %d", rr.Code)
	}
}
