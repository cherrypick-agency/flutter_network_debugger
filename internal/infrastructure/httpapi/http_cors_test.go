package httpapi

import (
    "net/http"
    "net/http/httptest"
    cfgpkg "network-debugger/internal/infrastructure/config"
    "testing"
)

func TestWithCORS_OptionsShortCircuit(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "http://x"}}
    called := false
    inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
    h := withCORS(d.Cfg, inner)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodOptions, "/any", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusNoContent { t.Fatalf("expected 204, got %d", rr.Code) }
    if called { t.Fatalf("inner handler should not be called on OPTIONS") }
    if rr.Header().Get("Access-Control-Allow-Origin") != "http://x" {
        t.Fatalf("CORS header not set")
    }
}


