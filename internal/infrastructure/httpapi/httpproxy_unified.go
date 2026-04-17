package httpapi

import (
	"net/http"

	proxyhttp "github.com/777genius/proxykit/proxyhttp"
	"network-debugger/internal/infrastructure/config"
)

// handleUnifiedProxy dispatches to WS or HTTP reverse proxy based on Upgrade header.
// Single URL: /proxy. If headers indicate WebSocket Upgrade — use WS proxy.
// Otherwise — HTTP reverse. For simplicity, target can be omitted in URL if server
// is configured with DEFAULT_TARGET: then /proxy/.. will proxy to that target.
func (d *Deps) handleUnifiedProxy(w http.ResponseWriter, r *http.Request) {
	if isWebSocketRequest(r) {
		d.handleWSProxy(w, r)
		return
	}
	d.handleHTTPProxy(w, r)
}

func isWebSocketRequest(r *http.Request) bool {
	return proxyhttp.IsWebSocketRequest(r)
}

// newTransport centralizes http.Transport creation with TLS options/timeouts.
func newTransport(cfg config.Config) *http.Transport {
	return proxyhttp.NewTransport(proxyhttp.TransportConfig{InsecureTLS: cfg.InsecureTLS})
}
