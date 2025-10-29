package httpapi

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    cfgpkg "network-debugger/internal/infrastructure/config"
    obs "network-debugger/internal/infrastructure/observability"
    uc "network-debugger/internal/usecase"
)

func TestWithForwardProxy_AbsoluteURI_GET(t *testing.T) {
    // upstream echo
    upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ _ = json.NewEncoder(w).Encode(map[string]any{"ok":true}) }))
    defer upstream.Close()
    s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
    d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 512, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
    h := NewRouterWithDeps(d)
    rr := httptest.NewRecorder()
    // absolute-URI request triggers forward proxy
    req := httptest.NewRequest(http.MethodGet, upstream.URL+"/echo", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("forward proxy GET failed: %d", rr.Code) }
}


