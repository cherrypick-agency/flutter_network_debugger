package httpapi

import (
    "net/http"
    "time"
    "encoding/json"
    "strings"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/rs/zerolog"

    "go-proxy/internal/infrastructure/config"
    obs "go-proxy/internal/infrastructure/observability"
    "go-proxy/internal/usecase"
)

type Deps struct {
    Cfg     config.Config
    Logger  *zerolog.Logger
    Metrics *obs.Metrics
    Svc     *usecase.SessionService
    Monitor *MonitorHub
    Live    *LiveSessions
}

func NewRouter(cfg config.Config, logger *zerolog.Logger, metrics *obs.Metrics) http.Handler {
    // backward compatibility shim for early main.go; will be removed when deps used everywhere
    d := &Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Monitor: NewMonitorHub(), Live: NewLiveSessions()}
    return NewRouterWithDeps(d)
}

func NewRouterWithDeps(d *Deps) http.Handler {
    mux := buildBaseMux(d)
    // Wrap with forward-proxy OUTERMOST; then apply CORS to all non-CONNECT flows.
    return withForwardProxy(d, withCORS(d.Cfg, mux))
}

// NewRouterWithoutForwardProxy returns the same routes but without the forward-proxy wrapper.
// Useful for TLS server where we want HTTP/2 for REST/reverse, while keeping CONNECT on the plain server.
func NewRouterWithoutForwardProxy(d *Deps) http.Handler {
    // Build mux and apply CORS, but skip withForwardProxy
    return withCORS(d.Cfg, buildBaseMux(d))
}

// buildBaseMux constructs the mux with all routes, without wrappers.
func buildBaseMux(d *Deps) *http.ServeMux {
    mux := http.NewServeMux()

    // Apply preview limit from config (<=0 disables truncation)
    previewMaxBytes = d.Cfg.PreviewMaxBytes
    // Apply sensitive headers exposure flag
    exposeSensitiveHeaders = d.Cfg.ExposeSensitiveHeaders
    previewDecompress = d.Cfg.PreviewDecompress

    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })
    mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ready"))
    })

    mux.Handle("/metrics", promhttp.HandlerFor(d.Metrics.Registry(), promhttp.HandlerOpts{}))

    mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]any{
            "name": "go-proxy",
            "version": "0.1.0-mvp",
            "time": time.Now().UTC(),
        })
    })

    // REST sessions (legacy base)
    mux.HandleFunc("/api/sessions", d.handleListSessions)
    // Single handler for /api/sessions/* to avoid duplicate registrations
    mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, "/ws/send") {
            d.handleWSSendText(w, r); return
        }
        d.handleSessionByID(w, r)
    })

    // SSE stream for real-time updates (frames/events/httpTxs)
    mux.HandleFunc("/api/sessions_stream/", d.handleSessionStream)

    // Monitor WS (legacy)
    mux.HandleFunc("/api/monitor/ws", d.Monitor.HandleWS)

    // WS Proxy
    mux.HandleFunc("/wsproxy", d.handleWSProxy)
    mux.HandleFunc("/wsproxy/", d.handleWSProxy)

    // HTTP Reverse Proxy (prefix-based)
    // Usage examples:
    //  - GET /httpproxy?target=https://api.example.com               -> proxies to https://api.example.com/
    //  - GET /httpproxy/v1/users?target=https://api.example.com      -> proxies to https://api.example.com/v1/users
    //  Query params except `target` are forwarded to upstream.
    mux.HandleFunc("/httpproxy", d.handleHTTPProxy)
    mux.HandleFunc("/httpproxy/", d.handleHTTPProxy)
    // Unified endpoint for both HTTP reverse and WebSocket proxy
    // If the request is an upgrade to websocket, it will be handled by handleWSProxy-like flow.
    mux.HandleFunc("/proxy", d.handleUnifiedProxy)
    mux.HandleFunc("/proxy/", d.handleUnifiedProxy)

    // === V1 API ===
    mux.HandleFunc("/_api/v1/version", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]any{
            "name": "go-proxy",
            "version": "v1",
            "time": time.Now().UTC(),
        })
    })
    mux.HandleFunc("/_api/v1/sessions", d.handleV1ListSessions)
    mux.HandleFunc("/_api/v1/sessions/", d.handleV1SessionByID)
    mux.HandleFunc("/_api/v1/sessions/aggregate", d.handleV1SessionsAggregate)
    mux.HandleFunc("/_api/v1/monitor/ws", d.Monitor.HandleWS)
    mux.HandleFunc("/_api/v1/httpproxy", d.handleHTTPProxy)
    mux.HandleFunc("/_api/v1/httpproxy/", d.handleHTTPProxy)

    return mux
}

func withForwardProxy(d *Deps, h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Intercept standard proxy patterns: CONNECT and absolute-URI
        if r.Method == http.MethodConnect || (r.URL != nil && r.URL.Scheme != "" && r.URL.Host != "") {
            d.handleForwardProxy(w, r)
            return
        }
        h.ServeHTTP(w, r)
    })
}

func withCORS(cfg config.Config, h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", cfg.CORSAllowOrigin)
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie, Sec-WebSocket-Protocol")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        h.ServeHTTP(w, r)
    })
}


