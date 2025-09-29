package httpapi

import (
    "crypto/tls"
    "net/http"
    "time"

    "go-proxy/internal/infrastructure/config"
    http2 "golang.org/x/net/http2"
)

// handleUnifiedProxy dispatches to WS or HTTP reverse proxy based on Upgrade header.
// Один URL: /proxy. Если заголовки указывают на WebSocket Upgrade — используем WS‑прокси.
// Иначе — HTTP reverse. Для простоты можно не указывать target в URL, если сервер
// сконфигурирован с DEFAULT_TARGET: тогда /proxy/.. будет проксировать на этот target.
func (d *Deps) handleUnifiedProxy(w http.ResponseWriter, r *http.Request) {
    if isWebSocketRequest(r) {
        d.handleWSProxy(w, r)
        return
    }
    d.handleHTTPProxy(w, r)
}

func isWebSocketRequest(r *http.Request) bool {
    if r.Header.Get("Upgrade") == "websocket" { return true }
    // Some clients use Sec-WebSocket-Key/Version headers as signal before upgrade
    if r.Header.Get("Sec-WebSocket-Key") != "" || r.Header.Get("Sec-WebSocket-Version") != "" {
        return true
    }
    return false
}

// newTransport centralizes http.Transport creation with TLS options/timeouts.
func newTransport(cfg config.Config) *http.Transport {
    tr := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        MaxIdleConns: 100,
        IdleConnTimeout: 90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    }
    if cfg.InsecureTLS {
        tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
    }
    // Enable HTTP/2 for outbound HTTPS where possible. Safe to ignore error and fall back to HTTP/1.1
    _ = http2.ConfigureTransport(tr)
    return tr
}


