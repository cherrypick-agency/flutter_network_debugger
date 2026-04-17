package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bufio"
	proxyhttp "github.com/777genius/proxykit/proxyhttp"
	reverseproxy "github.com/777genius/proxykit/reverse"
	"mime/multipart"
	"network-debugger/internal/domain"
	processdomain "network-debugger/internal/features/process/domain"
	"network-debugger/pkg/shared/id"
	"network-debugger/pkg/shared/redact"
	"sync/atomic"
)

// handleHTTPProxy implements a simple reverse proxy that forwards requests to the `_target` upstream.
// Path after /httpproxy is appended to target path. Query parameters (except `_target`) are passed through.
func (d *Deps) handleHTTPProxy(w http.ResponseWriter, r *http.Request) {
	// Log incoming request for diagnostics
	d.Logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Str("target_param", r.URL.Query().Get("_target")).
		Msg("[HTTPPROXY] Incoming request")

	// Check for _resetCapture parameter and reset if requested
	if r.URL.Query().Get("_resetCapture") == "true" {
		d.resetCaptureBeforeRequest()
	}

	// Offline mode simulation
	if d.Cfg.ThrottleOffline {
		writeError(w, http.StatusServiceUnavailable, "OFFLINE", "proxy offline (simulated)", nil)
		return
	}
	prefix := "/httpproxy"
	if strings.HasPrefix(r.URL.Path, "/proxy") {
		prefix = "/proxy"
	}
	resolver := reverseproxy.QueryTargetResolver{
		Param:         "_target",
		DefaultTarget: d.Cfg.DefaultTarget,
		DropParams:    []string{"_resetCapture"},
		MountPath:     prefix,
	}
	upstream, err := resolver.Resolve(r)
	if err != nil {
		switch {
		case errors.Is(err, reverseproxy.ErrMissingTarget):
			writeError(w, http.StatusBadRequest, "MISSING_TARGET", "missing target", nil)
		case errors.Is(err, reverseproxy.ErrInvalidTarget):
			writeError(w, http.StatusBadRequest, "INVALID_TARGET", "invalid target", nil)
		default:
			writeError(w, http.StatusInternalServerError, "TARGET_RESOLVE_FAILED", err.Error(), nil)
		}
		return
	}

	// Resolve cookie/stealth options (can be overridden per-request via query)
	cookieMode := d.Cfg.Cookies.Mode
	if v := strings.TrimSpace(r.URL.Query().Get("_cookie_mode")); v != "" {
		lv := strings.ToLower(v)
		if lv == CookieModeIsolate || lv == CookieModeAuto || lv == CookieModeOff {
			cookieMode = lv
		}
	}
	stealth := d.Cfg.StealthHeaders
	if v := strings.TrimSpace(r.URL.Query().Get("_stealth")); v != "" {
		lv := strings.ToLower(v)
		if lv == "0" || lv == "false" {
			stealth = false
		}
		if lv == "1" || lv == "true" {
			stealth = true
		}
	}

	ns := computeNamespaceFromURL(upstream)
	opts := CookieRewriteOptions{
		Mode:            cookieMode,
		DomainStrategy:  d.Cfg.Cookies.DomainStrategy,
		PathStrategy:    d.Cfg.Cookies.PathStrategy,
		ProxyHost:       sanitizeHost(r.Host),
		ProxyPathPrefix: prefix,
		HTTPS:           r.TLS != nil,
		Namespace:       ns,
	}

	// Check if target should be monitored (exclude monitoring endpoints)
	shouldMonitor := d.shouldMonitorTarget(upstream.Path)
	d.Logger.Info().
		Str("upstream_url", upstream.String()).
		Str("upstream_path", upstream.Path).
		Bool("should_monitor", shouldMonitor).
		Msg("[HTTPPROXY] shouldMonitorTarget check")

	sessionID := id.New()
	sess := domain.Session{
		ID:         sessionID,
		Target:     upstream.String(),
		ClientAddr: r.RemoteAddr,
		StartedAt:  time.Now().UTC(),
		Kind:       "http",
		// For reverse-proxy sessions, don't run local process detection.
		// Otherwise UI app (e.g., "Runner") will be displayed, which is confusing.
		ProcessInfo: &processdomain.ProcessInfo{Name: "Reverse Proxy"},
	}
	// Always create session regardless of shouldMonitor
	d.Logger.Info().Str("session_id", sessionID).Str("target", upstream.String()).Msg("[HTTPPROXY] Creating session")
	if err := d.Svc.Create(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", err.Error(), nil)
		return
	}
	d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID})
	d.Logger.Info().Str("session_id", sessionID).Msg("[HTTPPROXY] Broadcasted session_started event")
	d.Metrics.ActiveSessions.Inc()

	// Mapping (Map Remote/Local) — evaluate against final upstream URL
	mappedPreserveHost := false
	d.CfgMu.RLock()
	mappingEnabled := d.Cfg.MappingEnabled
	d.CfgMu.RUnlock()
	if mappingEnabled && d.MapRt != nil {
		prevReq := r.Clone(r.Context())
		u2 := *upstream
		prevReq.URL = &u2
		prevReq.Host = u2.Host
		if dec, ok := d.MapRt.EvalRequest(prevReq); ok {
			if dec.Kind == "local" {
				var bodyAll []byte
				var readErr error
				if dec.LocalFilePath != nil && *dec.LocalFilePath != "" {
					bodyAll, readErr = os.ReadFile(*dec.LocalFilePath)
				} else if dec.LocalBlobPath != nil && *dec.LocalBlobPath != "" {
					bodyAll, readErr = os.ReadFile(*dec.LocalBlobPath)
				}
				if readErr != nil {
					msg := []byte("mapping local source is not readable")
					h := http.Header{}
					h.Set("Content-Type", "text/plain; charset=utf-8")
					h.Set("X-ND-Mapped", "local")
					h.Set("X-ND-Rule", dec.RuleID)
					h.Set("X-ND-Mapping-Error", "read_failed")
					resp := &http.Response{StatusCode: http.StatusBadGateway, Status: strconv.Itoa(http.StatusBadGateway) + " " + http.StatusText(http.StatusBadGateway), Header: h, Body: io.NopCloser(bytes.NewReader(msg)), ContentLength: int64(len(msg))}
					basePreview := buildHTTPResponsePreview(resp)
					preview := augmentPreviewWithTimings(basePreview, 0, durationMs(time.Now(), time.Now()))
					fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: len(msg), Preview: preview}
					_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
					d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
					d.broadcastMonitorEvent(domain.MonitorEvent{Type: "mapping_applied", ID: sessionID, Ref: dec.RuleID})
					if d.Metrics != nil && d.Metrics.MappingAppliedTotal != nil {
						d.Metrics.MappingAppliedTotal.WithLabelValues("local").Inc()
					}
					d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

					copyHeader(w.Header(), h)
					w.Header().Set("Connection", "close")
					w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
					w.WriteHeader(http.StatusBadGateway)
					_, _ = w.Write(msg)
					_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
					d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
					d.Metrics.ActiveSessions.Dec()
					return
				}
				h := http.Header{}
				if dec.ContentTypeOverride != "" {
					h.Set("Content-Type", dec.ContentTypeOverride)
				} else if len(bodyAll) > 0 {
					h.Set("Content-Type", http.DetectContentType(bodyAll))
				} else {
					h.Set("Content-Type", "application/octet-stream")
				}
				h.Set("X-ND-Mapped", "local")
				h.Set("X-ND-Rule", dec.RuleID)
				resp := &http.Response{StatusCode: dec.StatusOverride, Status: strconv.Itoa(dec.StatusOverride) + " " + http.StatusText(dec.StatusOverride), Header: h, Body: io.NopCloser(bytes.NewReader(bodyAll)), ContentLength: int64(len(bodyAll))}
				basePreview := buildHTTPResponsePreview(resp)
				preview := augmentPreviewWithTimings(basePreview, 0, durationMs(time.Now(), time.Now()))
				// Always log mapping local response frame
				fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: len(bodyAll), Preview: preview}
				_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
				// mapping_applied (local)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "mapping_applied", ID: sessionID, Ref: dec.RuleID})
				if d.Metrics != nil && d.Metrics.MappingAppliedTotal != nil {
					d.Metrics.MappingAppliedTotal.WithLabelValues("local").Inc()
				}
				d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

				copyHeader(w.Header(), h)
				w.Header().Set("Connection", "close")
				w.Header().Set("Content-Length", strconv.Itoa(len(bodyAll)))
				w.WriteHeader(dec.StatusOverride)
				if len(bodyAll) > 0 {
					_, _ = w.Write(bodyAll)
				}
				// Always close session after local mapping response
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
				return
			}
			if u, err := url.Parse(dec.RemoteURL); err == nil {
				upstream = u
				mappedPreserveHost = dec.PreserveHost
				// mapping_applied (remote) - always broadcast
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "mapping_applied", ID: sessionID, Ref: dec.RuleID})
				if d.Metrics != nil && d.Metrics.MappingAppliedTotal != nil {
					d.Metrics.MappingAppliedTotal.WithLabelValues("remote").Inc()
				}
			}
		}
	}

	startedAt := time.Now()
	flow := &reverseProxyFlow{
		deps:         d,
		prefix:       prefix,
		upstream:     upstream,
		preserveHost: mappedPreserveHost,
		cookieOpts:   opts,
		stealth:      stealth,
		sessionID:    sessionID,
		startedAt:    startedAt,
	}

	// Safely peek a small portion of request body and keep stream intact for upstream.
	// This must be done before scripts/interception and before the public reverse handler clones the request.
	if r.Body != nil {
		peekSize := int(previewMaxBytes.Load())
		if peekSize <= 0 {
			peekSize = 65536
		}
		if peekSize > 65536 {
			peekSize = 65536
		}
		peek := make([]byte, peekSize)
		n, _ := io.ReadFull(r.Body, peek)
		if n > 0 {
			flow.reqBodyBuf = peek[:n]
			flow.scriptReqBodyBuf = append([]byte(nil), flow.reqBodyBuf...)
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(flow.reqBodyBuf), r.Body))
		}
	}

	// Emit lightweight session-start heartbeat frame so UI can draw in-progress bar immediately.
	hb := map[string]any{"type": "http_progress", "phase": "started"}
	b, _ := json.Marshal(hb)
	fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: len(b), Preview: string(b)}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
	d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})

	// Optional request body spooling
	if d.Cfg.CaptureBodies && r.Body != nil {
		if f, err := d.spoolBody(r.Body, int64(d.Cfg.BodyMaxBytes), "req"); err == nil && f != "" {
			// track for cleanup
			d.Svc.AddSpoolFile(contextWithNoCancel(), sessionID, f)
			// rewind spooled for upstream
			if fd, err2 := os.Open(f); err2 == nil {
				r.Body = fd // upstream will read from file; fd will be closed by transport
			}
		}
	}
	// For preview, show the real upstream URL (not the /httpproxy path)
	rPrev := *r
	rPrev.URL = cloneURL(upstream)
	reqPreview := buildHTTPRequestPreview(&rPrev, flow.reqBodyBuf)
	// Always log request frame
	flow.reqFrameID = id.New()
	fr = domain.Frame{ID: flow.reqFrameID, Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(r.ContentLength), Preview: reqPreview}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
	d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
	d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

	// Attach httptrace to catch milestones (write times atomically; parallel dial may trigger concurrently)
	r = r.WithContext(httptrace.WithClientTrace(r.Context(), &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			atomic.CompareAndSwapInt64(&flow.dnsStartNs, 0, time.Now().UnixNano())
		},
		ConnectStart: func(network, addr string) {
			atomic.CompareAndSwapInt64(&flow.connectStartNs, 0, time.Now().UnixNano())
		},
		TLSHandshakeStart: func() {
			atomic.CompareAndSwapInt64(&flow.tlsStartNs, 0, time.Now().UnixNano())
		},
		GotFirstResponseByte: func() {
			atomic.CompareAndSwapInt64(&flow.firstResponseNs, 0, time.Now().UnixNano())
		},
	}))
	flow.request = r
	flow.serve(w)
}

func removeHopHeaders(h http.Header) {
	proxyhttp.RemoveHopHeaders(h)
}

// (moved to preview.go) var previewMaxBytes = 1024

func buildHTTPRequestPreview(r *http.Request, body []byte) string {
	// redact sensitive headers
	hdr := map[string]string{}
	hdrRaw := map[string]string{}
	for k, v := range r.Header {
		if len(v) == 0 {
			continue
		}
		lk := strings.ToLower(k)
		val := v[0]
		// expose raw optionally (disabled by default via config)
		// NOTE: no direct config access here; use package-level default (off)
		if lk == "authorization" || lk == "cookie" || strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.Contains(lk, "apikey") || strings.Contains(lk, "api-key") {
			hdr[k] = "***"
		} else {
			hdr[k] = val
		}
		if exposeSensitiveHeaders.Load() {
			hdrRaw[k] = val
		}
	}
	// attempt to decode gzip if any; however requests usually are plain
	preview := map[string]any{
		"type":    "http_request",
		"method":  r.Method,
		"url":     r.URL.String(),
		"headers": hdr,
	}
	if exposeSensitiveHeaders.Load() {
		preview["headersRaw"] = hdrRaw
	}
	// headersRaw currently disabled in preview helpers to avoid config deps
	max := int(previewMaxBytes.Load())
	if len(body) > 0 {
		// Store raw size BEFORE any transformations (JSON compaction, redaction)
		// This allows frontend to know if body was fully read vs truncated
		preview["bodyRawSize"] = len(body)

		// Best-effort: decompress request preview if Content-Encoding set
		b := body
		enc := strings.ToLower(r.Header.Get("Content-Encoding"))
		if previewDecompress.Load() && (enc == "gzip" || enc == "deflate") {
			if dec, ok := tryDecompress(b, enc); ok {
				b = dec
			}
		}

		// Try to parse form-data/urlencoded by Content-Type
		ct := strings.ToLower(r.Header.Get("Content-Type"))
		// urlencoded: parse key-value pairs
		if strings.Contains(ct, "application/x-www-form-urlencoded") {
			// For preview, parsing first bytes is sufficient
			// Note: body may be truncated, this is normal for preview
			vals, err := url.ParseQuery(string(b))
			if err == nil && len(vals) > 0 {
				fields := make([]map[string]any, 0, len(vals))
				for k, vv := range vals {
					for _, v := range vv {
						fields = append(fields, map[string]any{
							"name":  k,
							"value": v,
						})
					}
				}
				preview["form"] = map[string]any{
					"type":   "urlencoded",
					"fields": fields,
				}
			}
		} else if strings.Contains(ct, "multipart/form-data") {
			// multipart: iterate over parts, read a little from each part
			if mr := multipartReaderFrom(ct, b); mr != nil {
				const perPartLimit = 2048 // ~2KB per part for preview
				fields := make([]map[string]any, 0, 8)
				files := make([]map[string]any, 0, 4)
				for {
					part, err := mr.NextPart()
					if err != nil {
						break
					}
					name := part.FormName()
					filename := part.FileName()
					pct := part.Header.Get("Content-Type")

					// Read limited number of bytes to avoid bloating the preview
					buf := make([]byte, perPartLimit+1)
					n, _ := io.ReadFull(part, buf)
					truncated := n > perPartLimit
					previewSize := n
					if previewSize > perPartLimit {
						previewSize = perPartLimit
					}
					if n > perPartLimit {
						n = perPartLimit
					}
					data := buf[:n]

					if filename == "" {
						// Regular text form field
						fields = append(fields, map[string]any{
							"name":         name,
							"valuePreview": string(data),
							"truncated":    truncated,
						})
					} else {
						// File: show metadata and small preview for text types
						file := map[string]any{
							"name":        name,
							"filename":    filename,
							"contentType": pct,
							"truncated":   truncated,
							"previewSize": previewSize,
						}
						lct := strings.ToLower(pct)
						if lct == "" || strings.Contains(lct, "text") || strings.Contains(lct, "+json") || strings.Contains(lct, "json") || strings.Contains(lct, "xml") || strings.Contains(lct, "csv") {
							file["valuePreview"] = string(data)
						}
						files = append(files, file)
					}
				}
				if len(fields) > 0 || len(files) > 0 {
					preview["form"] = map[string]any{
						"type":   "multipart",
						"fields": fields,
						"files":  files,
					}
				}
			}
		}

		// Also save raw body in preview (truncated) to preserve original
		if tryJSON := tryCompactJSON(b); tryJSON != "" {
			if len(tryJSON) > max {
				tryJSON = tryJSON[:max]
			}
			preview["body"] = tryJSON
		} else {
			if len(b) > max {
				b = b[:max]
			}
			preview["body"] = string(b)
		}
	}
	b, _ := json.Marshal(preview)
	return string(b)
}

func buildHTTPResponsePreview(resp *http.Response) string {
	hdr := map[string]string{}
	hdrRaw := map[string]string{}
	for k, v := range resp.Header {
		if len(v) == 0 {
			continue
		}
		lk := strings.ToLower(k)
		val := v[0]
		// see note above
		if lk == "set-cookie" || strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.Contains(lk, "authorization") || strings.Contains(lk, "api-key") {
			hdr[k] = "***"
		} else {
			hdr[k] = val
		}
		if exposeSensitiveHeaders.Load() {
			hdrRaw[k] = val
		}
	}
	preview := map[string]any{
		"type":    "http_response",
		"status":  resp.StatusCode,
		"headers": hdr,
	}
	if exposeSensitiveHeaders.Load() {
		preview["headersRaw"] = hdrRaw
	}
	// TLS/security summary
	if resp.TLS != nil {
		preview["tls"] = map[string]any{
			"version":          tlsVersionString(resp.TLS.Version),
			"cipherSuite":      cipherSuiteString(resp.TLS.CipherSuite),
			"alpn":             resp.TLS.NegotiatedProtocol,
			"serverName":       resp.TLS.ServerName,
			"peerCertificates": certsSummary(resp.TLS.PeerCertificates),
		}
	}
	// Cookie flags summary (do not expose values)
	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) > 0 {
		var nSecure, nHttpOnly, nLax, nStrict, nNone int
		for _, c := range cookies {
			lc := strings.ToLower(c)
			if strings.Contains(lc, "secure") {
				nSecure++
			}
			if strings.Contains(lc, "httponly") {
				nHttpOnly++
			}
			if strings.Contains(lc, "samesite=lax") {
				nLax++
			}
			if strings.Contains(lc, "samesite=strict") {
				nStrict++
			}
			if strings.Contains(lc, "samesite=none") {
				nNone++
			}
		}
		preview["cookieSummary"] = map[string]any{
			"setCookieCount": len(cookies),
			"secure":         nSecure,
			"httpOnly":       nHttpOnly,
			"sameSiteLax":    nLax,
			"sameSiteStrict": nStrict,
			"sameSiteNone":   nNone,
		}
	}
	// see note above
	// best-effort: peek limited bytes and reattach back to resp.Body
	var bodyBuf []byte
	if resp.Body != nil {
		// If gzip encoded, we do not decompress to avoid corrupting stream. We just sample raw bytes.
		// Read a small chunk and then reattach it in front so client receives original body in full.
		peekSize := int(previewMaxBytes.Load())
		if peekSize <= 0 {
			peekSize = 65536
		}
		if peekSize > 65536 {
			peekSize = 65536
		}
		peek := make([]byte, peekSize)
		n, _ := io.ReadFull(resp.Body, peek)
		if n > 0 {
			bodyBuf = peek[:n]
			resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyBuf), resp.Body))
		}
	}
	max := int(previewMaxBytes.Load())
	if len(bodyBuf) > 0 {
		// Store raw size BEFORE any transformations (JSON compaction, redaction)
		// This allows frontend to know if body was fully read vs truncated
		preview["bodyRawSize"] = len(bodyBuf)

		// Best-effort decompress for gzip/deflate for preview only
		b := bodyBuf
		enc := strings.ToLower(resp.Header.Get("Content-Encoding"))
		if previewDecompress.Load() && (enc == "gzip" || enc == "deflate") {
			if dec, ok := tryDecompress(b, enc); ok {
				b = dec
			}
		}
		if tryJSON := tryCompactJSON(b); tryJSON != "" {
			sanitized := redact.RedactJSON(tryJSON)
			if max > 0 && len(sanitized) > max {
				sanitized = sanitized[:max]
			}
			preview["body"] = sanitized
		} else {
			if max > 0 && len(b) > max {
				b = b[:max]
			}
			preview["body"] = string(b)
		}
	}
	b, _ := json.Marshal(preview)
	return string(b)
}

// spoolBody writes up to max bytes from r into a temp file and returns the file path.
// If BodySpoolDir is empty, uses os.CreateTemp default.
func (d *Deps) spoolBody(r io.Reader, max int64, kind string) (string, error) {
	dir := d.Cfg.BodySpoolDir
	if dir == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "gpx-"+kind+"-*.bin")
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Sync() }()
	// limit copy
	if _, err := io.CopyN(f, r, max); err != nil && err != io.EOF {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()
	// return path
	abs, _ := filepath.Abs(f.Name())
	return abs, nil
}

func (d *Deps) spoolBodyBytes(data []byte, kind string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	dir := d.Cfg.BodySpoolDir
	if dir == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "gpx-"+kind+"-*.bin")
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Sync() }()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return "", err
	}
	_ = f.Close()
	abs, _ := filepath.Abs(f.Name())
	return abs, nil
}

func tryCompactJSON(b []byte) string {
	var js any
	if json.Unmarshal(b, &js) == nil {
		out, _ := json.Marshal(js)
		return string(out)
	}
	return ""
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	out := *u
	return &out
}

// augmentPreviewWithTimings injects {timings:{ttfbMs,totalMs}} into JSON preview.
func augmentPreviewWithTimings(preview string, ttfb, total int64) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(preview), &m); err != nil {
		return preview
	}
	m["timings"] = map[string]any{"ttfbMs": ttfb, "totalMs": total}
	b, err := json.Marshal(m)
	if err != nil {
		return preview
	}
	return string(b)
}

// tryDecompress performs safe small-buffer decompression for preview
func tryDecompress(b []byte, enc string) ([]byte, bool) {
	// limit reader to avoid zip bombs
	const maxPreview = 1 << 20 // 1MB upper bound for preview
	switch enc {
	case "gzip":
		zr, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, false
		}
		defer zr.Close()
		r := io.LimitReader(zr, maxPreview)
		out, err := io.ReadAll(r)
		if err != nil {
			return nil, false
		}
		return out, true
	case "deflate":
		// support raw zlib/deflate
		fr := flate.NewReader(bytes.NewReader(b))
		if fr == nil {
			return nil, false
		}
		defer fr.Close()
		r := io.LimitReader(fr, maxPreview)
		out, err := io.ReadAll(r)
		if err != nil {
			return nil, false
		}
		return out, true
	default:
		_ = bufio.ErrAdvanceTooFar // keep import used
		return nil, false
	}
}

// multipartReaderFrom builds a multipart reader from content-type and raw body (preview-sized).
func multipartReaderFrom(ct string, body []byte) *multipart.Reader {
	boundary := ""
	parts := strings.Split(ct, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		lp := strings.ToLower(p)
		if strings.HasPrefix(lp, "boundary=") {
			boundary = strings.TrimPrefix(p, "boundary=")
			boundary = strings.Trim(boundary, "\"")
			break
		}
	}
	if boundary == "" {
		return nil
	}
	return multipart.NewReader(bytes.NewReader(body), boundary)
}

// Helpers: TLS/cert pretty names
func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return ""
	}
}

func cipherSuiteString(id uint16) string {
	// Minimal mapping for common suites; fallback to code
	switch id {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	default:
		return "0x" + strings.ToUpper(strconv.FormatInt(int64(id), 16))
	}
}

func certsSummary(certs []*x509.Certificate) []map[string]any {
	if len(certs) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(certs))
	for _, c := range certs {
		out = append(out, map[string]any{
			"subject":   c.Subject.String(),
			"issuer":    c.Issuer.String(),
			"notBefore": c.NotBefore.UTC(),
			"notAfter":  c.NotAfter.UTC(),
			"dnsNames":  c.DNSNames,
			"isCA":      c.IsCA,
		})
	}
	return out
}

func int64ToInt(v int64) int {
	if v > int64(^uint(0)>>1) {
		return int(^uint(0) >> 1)
	}
	if v < 0 {
		return 0
	}
	return int(v)
}

func durationMs(from time.Time, to time.Time) int64 {
	if from.IsZero() || to.IsZero() {
		return 0
	}
	dur := to.Sub(from)
	ms := dur.Milliseconds()
	// If duration > 0 but less than 1ms, round up to 1ms to avoid showing "0 ms"
	if ms == 0 && dur > 0 {
		return 1
	}
	return ms
}

// useOrFallback returns a if set, otherwise b. Helps when TLS phase is absent (plain HTTP).
func useOrFallback(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	return a
}

// timeFromUnixNanoOrZero converts monotonic-independent nanoseconds to time.Time or returns zero time.
func timeFromUnixNanoOrZero(ns int64) time.Time {
	if ns <= 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// humanizeProxyError converts technical proxy errors into user-friendly messages with error codes
func humanizeProxyError(err error) (code string, message string) {
	info := classifyProxyError(err)
	return info.Code, info.UserMessage
}
