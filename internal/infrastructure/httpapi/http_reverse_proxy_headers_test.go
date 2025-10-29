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

func TestHTTPProxy_SetsForwardHeaders_WhenStealthDisabled(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"xff": r.Header.Get("X-Forwarded-For"),
			"xfp": r.Header.Get("X-Forwarded-Proto"),
			"via": r.Header.Get("Via"),
		})
	}))
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 64, ExposeSensitiveHeaders: false, PreviewDecompress: true, StealthHeaders: false}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL+"/hdr", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status %d", rr.Code)
	}
}
