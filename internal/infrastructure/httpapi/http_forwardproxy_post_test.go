package httpapi

import (
	"bytes"
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

func TestWithForwardProxy_AbsoluteURI_POST_WithBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 128, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Logger: &logger, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := withForwardProxy(d, NewRouterWithoutForwardProxy(d))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, upstream.URL+"/echo", bytes.NewReader([]byte(`{"x":1}`)))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("forward proxy POST failed: %d", rr.Code)
	}
}
