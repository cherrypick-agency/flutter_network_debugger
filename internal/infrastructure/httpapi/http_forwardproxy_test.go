package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"

	"github.com/rs/zerolog"
)

func TestWithForwardProxy_AbsoluteURI_GET(t *testing.T) {
	// upstream echo
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 512, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Logger: &logger, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := withForwardProxy(d, NewRouterWithoutForwardProxy(d))
	rr := httptest.NewRecorder()
	// absolute-URI request triggers forward proxy
	req := httptest.NewRequest(http.MethodGet, upstream.URL+"/echo", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("forward proxy GET failed: %d", rr.Code)
	}
}
