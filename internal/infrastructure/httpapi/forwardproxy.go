package httpapi

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
)

// handleForwardOrNotFound routes absolute-URI and CONNECT requests as a standard forward proxy.
// Non-proxy requests fall back to 404 so that REST/WS routes are handled by other handlers.
func (d *Deps) handleForwardOrNotFound(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect || (r.URL != nil && r.URL.Scheme != "" && r.URL.Host != "") {
		d.handleForwardProxy(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
}

func (d *Deps) handleForwardProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		d.handleConnectTunnel(w, r)
		return
	}
	// Forward regular HTTP request with absolute URI in r.URL
	d.handleHTTPForwardRequest(w, r)
}

func (d *Deps) handleConnectTunnel(w http.ResponseWriter, r *http.Request) {
	// r.Host contains host:port of upstream
	upstream := r.Host
	hj, ok := w.(http.Hijacker)
	if !ok {
		writeError(w, http.StatusInternalServerError, "HIJACK_NOT_SUPPORTED", "proxy: hijacking not supported", nil)
		return
	}
	clientConn, bufrw, err := hj.Hijack()
	if err != nil {
		return
	}
	// Dial upstream
	upstreamConn, err := net.DialTimeout("tcp", upstream, 10*time.Second)
	if err != nil {
		_, _ = bufrw.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = bufrw.Flush()
		_ = clientConn.Close()
		return
	}
	// Respond 200 and start tunneling
	_, _ = bufrw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	_ = bufrw.Flush()

	// minimal session for CONNECT (no payload introspection)
	sessionID := id.New()
	_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, Target: "connect://" + upstream, ClientAddr: clientHost(r.RemoteAddr), StartedAt: time.Now().UTC()})
	d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()

	// bidirectional copy
	go func() {
		_, _ = io.Copy(upstreamConn, clientConn)
		_ = upstreamConn.Close()
	}()
	_, _ = io.Copy(clientConn, upstreamConn)
	_ = clientConn.Close()

	_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
	d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
	d.Metrics.ActiveSessions.Dec()
}

func (d *Deps) handleHTTPForwardRequest(w http.ResponseWriter, r *http.Request) {
	// r.URL is absolute here (scheme+host+path)
	// Create session for logging
	sessionID := id.New()
	_ = d.Svc.Create(r.Context(), domain.Session{ID: sessionID, Target: r.URL.String(), ClientAddr: clientHost(r.RemoteAddr), StartedAt: time.Now().UTC()})
	d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()

	// Prepare outbound request: clone the original request but with absolute URL
	outURL := *r.URL
	outReq := r.Clone(r.Context())
	outReq.URL = &outURL
	outReq.Host = outURL.Host
	// Hop-by-hop headers must be removed
	outReq.Header = cloneHeader(outReq.Header)
	// Remove hop headers
	removeHopHeaders(outReq.Header)
	// Standard forwarding headers
	if ip := clientHost(r.RemoteAddr); ip != "" {
		outReq.Header.Set("X-Forwarded-For", ip)
	}
	if r.TLS != nil {
		outReq.Header.Set("X-Forwarded-Proto", "https")
	} else {
		outReq.Header.Set("X-Forwarded-Proto", "http")
	}
	outReq.Header.Set("Via", "network-debugger")

	// Safely peek a small portion of request body and keep stream intact
	var reqBodyBuf []byte
	if outReq.Body != nil {
		peekSize := previewMaxBytes
		if peekSize <= 0 {
			peekSize = 65536
		}
		if peekSize > 65536 {
			peekSize = 65536
		}
		peek := make([]byte, peekSize)
		n, _ := io.ReadFull(outReq.Body, peek)
		if n > 0 {
			reqBodyBuf = peek[:n]
			outReq.Body = io.NopCloser(io.MultiReader(bytes.NewReader(reqBodyBuf), outReq.Body))
		}
	}
	// For preview, use the real upstream absolute URL
	rPrev := *r
	rPrev.URL = &outURL
	reqPreview := buildHTTPRequestPreview(&rPrev, reqBodyBuf)
	fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(r.ContentLength), Preview: reqPreview}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
	d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
	d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

	// Send using unified transport
	tr := newTransport(d.Cfg)
	resp, err := tr.RoundTrip(outReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error(), map[string]any{"target": outURL.String()})
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		d.Metrics.ActiveSessions.Dec()
		d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
		return
	}
	defer resp.Body.Close()

	// Build response preview and keep body intact for client
	preview := buildHTTPResponsePreview(resp)
	fr2 := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: int(resp.ContentLength), Preview: preview}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr2)
	d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr2.ID})
	d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

	// Optional artificial response delay
	sleepResponseDelay(d.Cfg)
	// Write back to client
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)

	_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
	d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
	d.Metrics.ActiveSessions.Dec()
}

func cloneHeader(h http.Header) http.Header {
	dst := make(http.Header, len(h))
	for k, vv := range h {
		cp := make([]string, len(vv))
		copy(cp, vv)
		dst[k] = cp
	}
	return dst
}

func copyHeader(dst http.Header, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// isAbsoluteURL returns true if s looks like an absolute URI.
func isAbsoluteURL(s string) bool {
	if u, err := url.Parse(s); err == nil {
		return u.Scheme != "" && u.Host != ""
	}
	return false
}
