package httpapi

import (
    "bytes"
    "compress/flate"
    "compress/gzip"
    "encoding/json"
    "io"
    "crypto/tls"
    "crypto/x509"
    "net/http"
    "net/http/httputil"
    "net/http/httptrace"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "strconv"
    "time"

    "go-proxy/internal/domain"
    "go-proxy/pkg/shared/id"
    "go-proxy/pkg/shared/redact"
    "mime/multipart"
    "bufio"
)

// handleHTTPProxy implements a simple reverse proxy that forwards requests to the `target` upstream.
// Path after /httpproxy is appended to target path. Query parameters (except `target`) are passed through.
func (d *Deps) handleHTTPProxy(w http.ResponseWriter, r *http.Request) {
    tgt := r.URL.Query().Get("target")
    if tgt == "" {
        // fallback to default target from config
        if d.Cfg.DefaultTarget != "" {
            tgt = d.Cfg.DefaultTarget
        } else {
            writeError(w, http.StatusBadRequest, "MISSING_TARGET", "missing target", nil)
            return
        }
    }
    u, err := url.Parse(tgt)
    if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
        writeError(w, http.StatusBadRequest, "INVALID_TARGET", "invalid target", map[string]any{"target": tgt})
        return
    }

    // Build upstream URL by joining path suffix after /httpproxy or /proxy
    prefix := "/httpproxy"
    if strings.HasPrefix(r.URL.Path, "/proxy") { prefix = "/proxy" }
    suffix := strings.TrimPrefix(r.URL.Path, prefix)
    if !strings.HasPrefix(suffix, "/") { suffix = "/" + suffix }
    upstream := *u
    // Join paths
    upstream.Path = strings.TrimRight(upstream.Path, "/") + suffix

    // Filter query params (drop `target`)
    qp := r.URL.Query()
    qp.Del("target")
    upstream.RawQuery = qp.Encode()
    if upstream.RawQuery == "" {
        upstream.ForceQuery = false
    }

    sessionID := id.New()
    sess := domain.Session{
        ID:        sessionID,
        Target:    upstream.String(),
        ClientAddr: clientHost(r.RemoteAddr),
        StartedAt: time.Now().UTC(),
        Kind:      "http",
    }
    if err := d.Svc.Create(r.Context(), sess); err != nil {
        writeError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", err.Error(), nil)
        return
    }
    d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
    d.Metrics.ActiveSessions.Inc()

    // Create reverse proxy
    director := func(req *http.Request) {
        req.URL = &upstream
        req.Host = upstream.Host
        // Clean hop-by-hop headers; httputil will remove most, but ensure here for clarity
        removeHopHeaders(req.Header)
    }

    transport := newTransport(d.Cfg)
    // timings via httptrace
    var tStart = time.Now()
    var tDNS, tConnStart, tTLSStart, tFirstByte time.Time
    hadError := false
    proxy := &httputil.ReverseProxy{
        Director:  director,
        Transport: transport,
        ModifyResponse: func(resp *http.Response) error {
            // Artificial response delay (to visualize timeline)
            sleepResponseDelay(d.Cfg)
            // Log response frame with timings embedded
            basePreview := buildHTTPResponsePreview(resp)
            ttfb := durationMs(tStart, tFirstByte)
            total := durationMs(tStart, time.Now())
            preview := augmentPreviewWithTimings(basePreview, ttfb, total)
            fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: int(resp.ContentLength), Preview: preview}
            _ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
            d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
            d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

            // Persist HTTP transaction summary
            tx := domain.HTTPTransaction{
                ID: id.New(), SessionID: sessionID, Method: r.Method, URL: strings.TrimSuffix(upstream.String(), "?"),
                Status: resp.StatusCode,
                ReqSize: int(r.ContentLength), RespSize: int(resp.ContentLength),
                StartedAt: tStart, EndedAt: time.Now().UTC(),
                Timings: domain.HTTPTimings{
                    DNS:     durationMs(tDNS, tConnStart),
                    Connect: durationMs(tConnStart, useOrFallback(tTLSStart, tFirstByte)),
                    TLS:     durationMs(useOrFallback(tTLSStart, tFirstByte), tFirstByte),
                    TTFB:    durationMs(tStart, tFirstByte),
                    Total:   durationMs(tStart, time.Now()),
                },
            }
            // Best-effort content-type
            if ct := resp.Header.Get("Content-Type"); ct != "" { tx.ContentType = ct }
            // Optional body spooling
            if d.Cfg.CaptureBodies {
                if f, err := d.spoolBody(resp.Body, int64(d.Cfg.BodyMaxBytes), "resp"); err == nil && f != "" {
                    tx.RespBodyFile = f
                }
            }
            _ = d.Svc.AddHTTPTransaction(contextWithNoCancel(), tx)
            d.Monitor.Broadcast(MonitorEvent{Type: "http_tx_added", ID: sessionID, Ref: tx.ID})
            return nil
        },
        ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
            hadError = true
            _ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
            d.Logger.Error().Err(err).Msg("reverse proxy error")
            writeError(rw, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error(), map[string]any{"target": upstream.String()})
        },
    }

    // Emit lightweight session-start heartbeat frame so UI can draw in-progress bar immediately.
    {
        hb := map[string]any{"type": "http_progress", "phase": "started"}
        b, _ := json.Marshal(hb)
        fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: len(b), Preview: string(b)}
        _ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
        d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
    }

    // Safely peek a small portion of request body and keep stream intact for upstream.
    var reqBodyBuf []byte
    if r.Body != nil {
        peekSize := previewMaxBytes
        if peekSize <= 0 { peekSize = 65536 }
        if peekSize > 65536 { peekSize = 65536 }
        peek := make([]byte, peekSize)
        n, _ := io.ReadFull(r.Body, peek)
        if n > 0 {
            reqBodyBuf = peek[:n]
            r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(reqBodyBuf), r.Body))
        }
    }
    // Optional request body spooling
    if d.Cfg.CaptureBodies && r.Body != nil {
        if f, err := d.spoolBody(r.Body, int64(d.Cfg.BodyMaxBytes), "req"); err == nil && f != "" {
            // rewind spooled for upstream
            if fd, err2 := os.Open(f); err2 == nil {
                r.Body = fd // upstream will read from file; fd will be closed by transport
            }
        }
    }
    // For preview, show the real upstream URL (not the /httpproxy path)
    rPrev := *r
    rPrev.URL = &upstream
    reqPreview := buildHTTPRequestPreview(&rPrev, reqBodyBuf)
    fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(r.ContentLength), Preview: reqPreview}
    _ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
    d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
    d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

    // Also broadcast a lightweight event for frontend session_started consistency in HTTP flows
    // (ws flow already broadcasts in wsproxy)
    // d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID}) // already sent above

    // Attach httptrace to catch milestones
    r = r.WithContext(httptrace.WithClientTrace(r.Context(), &httptrace.ClientTrace{
        DNSStart: func(info httptrace.DNSStartInfo) { tDNS = time.Now() },
        ConnectStart: func(network, addr string) { tConnStart = time.Now() },
        TLSHandshakeStart: func() { tTLSStart = time.Now() },
        GotFirstResponseByte: func() { tFirstByte = time.Now() },
    }))

    // Standard forwarding headers (useful for logs/upstream)
    if ip := clientHost(r.RemoteAddr); ip != "" { r.Header.Set("X-Forwarded-For", ip) }
    if r.TLS != nil { r.Header.Set("X-Forwarded-Proto", "https") } else { r.Header.Set("X-Forwarded-Proto", "http") }
    r.Header.Set("Via", "go-proxy")

    // Serve
    proxy.ServeHTTP(w, r)
    if !hadError {
        _ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
    }
    d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
    d.Metrics.ActiveSessions.Dec()
}

func removeHopHeaders(h http.Header) {
    hop := []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"}
    for _, k := range hop { h.Del(k) }
}

// (moved to preview.go) var previewMaxBytes = 1024

func buildHTTPRequestPreview(r *http.Request, body []byte) string {
    // redact sensitive headers
    hdr := map[string]string{}
    hdrRaw := map[string]string{}
    for k, v := range r.Header {
        if len(v) == 0 { continue }
        lk := strings.ToLower(k)
        val := v[0]
        // expose raw optionally (disabled by default via config)
        // NOTE: no direct config access here; use package-level default (off)
        if lk == "authorization" || lk == "cookie" || strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.Contains(lk, "apikey") || strings.Contains(lk, "api-key") {
            hdr[k] = "***"
        } else { hdr[k] = val }
        if exposeSensitiveHeaders { hdrRaw[k] = val }
    }
    // attempt to decode gzip if any; however requests usually are plain
    preview := map[string]any{
        "type": "http_request",
        "method": r.Method,
        "url": r.URL.String(),
        "headers": hdr,
    }
    if exposeSensitiveHeaders {
        preview["headersRaw"] = hdrRaw
    }
    // headersRaw currently disabled in preview helpers to avoid config deps
    max := previewMaxBytes
    if len(body) > 0 {
        // Best-effort: decompress request preview if Content-Encoding set
        b := body
        enc := strings.ToLower(r.Header.Get("Content-Encoding"))
        if previewDecompress && (enc == "gzip" || enc == "deflate") {
            if dec, ok := tryDecompress(b, enc); ok { b = dec }
        }
        if tryJSON := tryCompactJSON(b); tryJSON != "" {
            if len(tryJSON) > max { tryJSON = tryJSON[:max] }
            preview["body"] = tryJSON
        } else {
            if len(b) > max { b = b[:max] }
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
        if len(v) == 0 { continue }
        lk := strings.ToLower(k)
        val := v[0]
        // see note above
        if lk == "set-cookie" || strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.Contains(lk, "authorization") || strings.Contains(lk, "api-key") {
            hdr[k] = "***"
        } else { hdr[k] = val }
        if exposeSensitiveHeaders { hdrRaw[k] = val }
    }
    preview := map[string]any{
        "type": "http_response",
        "status": resp.StatusCode,
        "headers": hdr,
    }
    if exposeSensitiveHeaders { preview["headersRaw"] = hdrRaw }
    // TLS/security summary
    if resp.TLS != nil {
        preview["tls"] = map[string]any{
            "version": tlsVersionString(resp.TLS.Version),
            "cipherSuite": cipherSuiteString(resp.TLS.CipherSuite),
            "alpn": resp.TLS.NegotiatedProtocol,
            "serverName": resp.TLS.ServerName,
            "peerCertificates": certsSummary(resp.TLS.PeerCertificates),
        }
    }
    // Cookie flags summary (do not expose values)
    if cookies := resp.Header.Values("Set-Cookie"); len(cookies) > 0 {
        var nSecure, nHttpOnly, nLax, nStrict, nNone int
        for _, c := range cookies {
            lc := strings.ToLower(c)
            if strings.Contains(lc, "secure") { nSecure++ }
            if strings.Contains(lc, "httponly") { nHttpOnly++ }
            if strings.Contains(lc, "samesite=lax") { nLax++ }
            if strings.Contains(lc, "samesite=strict") { nStrict++ }
            if strings.Contains(lc, "samesite=none") { nNone++ }
        }
        preview["cookieSummary"] = map[string]any{
            "setCookieCount": len(cookies),
            "secure": nSecure,
            "httpOnly": nHttpOnly,
            "sameSiteLax": nLax,
            "sameSiteStrict": nStrict,
            "sameSiteNone": nNone,
        }
    }
    // see note above
    // best-effort: peek limited bytes and reattach back to resp.Body
    var bodyBuf []byte
    if resp.Body != nil {
        // If gzip encoded, we do not decompress to avoid corrupting stream. We just sample raw bytes.
        // Read a small chunk and then reattach it in front so client receives original body in full.
        peekSize := previewMaxBytes
        if peekSize <= 0 { peekSize = 65536 }
        if peekSize > 65536 { peekSize = 65536 }
        peek := make([]byte, peekSize)
        n, _ := io.ReadFull(resp.Body, peek)
        if n > 0 {
            bodyBuf = peek[:n]
            resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(bodyBuf), resp.Body))
        }
    }
    max := previewMaxBytes
    if len(bodyBuf) > 0 {
        // Best-effort decompress for gzip/deflate for preview only
        b := bodyBuf
        enc := strings.ToLower(resp.Header.Get("Content-Encoding"))
        if previewDecompress && (enc == "gzip" || enc == "deflate") {
            if dec, ok := tryDecompress(b, enc); ok { b = dec }
        }
        if tryJSON := tryCompactJSON(b); tryJSON != "" {
            sanitized := redact.RedactJSON(tryJSON)
            if max > 0 && len(sanitized) > max { sanitized = sanitized[:max] }
            preview["body"] = sanitized
        } else {
            if max > 0 && len(b) > max { b = b[:max] }
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
    if dir == "" { dir = os.TempDir() }
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    f, err := os.CreateTemp(dir, "gpx-"+kind+"-*.bin")
    if err != nil { return "", err }
    defer func() { _ = f.Sync() }()
    // limit copy
    if _, err := io.CopyN(f, r, max); err != nil && err != io.EOF { _ = f.Close(); return "", err }
    _ = f.Close()
    // return path
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

// augmentPreviewWithTimings injects {timings:{ttfbMs,totalMs}} into JSON preview.
func augmentPreviewWithTimings(preview string, ttfb, total int64) string {
    var m map[string]any
    if err := json.Unmarshal([]byte(preview), &m); err != nil { return preview }
    m["timings"] = map[string]any{"ttfbMs": ttfb, "totalMs": total}
    b, err := json.Marshal(m)
    if err != nil { return preview }
    return string(b)
}

// tryDecompress performs safe small-buffer decompression for preview
func tryDecompress(b []byte, enc string) ([]byte, bool) {
    // limit reader to avoid zip bombs
    const maxPreview = 1 << 20 // 1MB upper bound for preview
    switch enc {
    case "gzip":
        zr, err := gzip.NewReader(bytes.NewReader(b))
        if err != nil { return nil, false }
        defer zr.Close()
        r := io.LimitReader(zr, maxPreview)
        out, err := io.ReadAll(r)
        if err != nil { return nil, false }
        return out, true
    case "deflate":
        // support raw zlib/deflate
        fr := flate.NewReader(bytes.NewReader(b))
        if fr == nil { return nil, false }
        defer fr.Close()
        r := io.LimitReader(fr, maxPreview)
        out, err := io.ReadAll(r)
        if err != nil { return nil, false }
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
    if boundary == "" { return nil }
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
    if len(certs) == 0 { return nil }
    out := make([]map[string]any, 0, len(certs))
    for _, c := range certs {
        out = append(out, map[string]any{
            "subject": c.Subject.String(),
            "issuer": c.Issuer.String(),
            "notBefore": c.NotBefore.UTC(),
            "notAfter": c.NotAfter.UTC(),
            "dnsNames": c.DNSNames,
            "isCA": c.IsCA,
        })
    }
    return out
}

func int64ToInt(v int64) int {
    if v > int64(^uint(0)>>1) { return int(^uint(0)>>1) }
    if v < 0 { return 0 }
    return int(v)
}

func durationMs(from time.Time, to time.Time) int64 {
    if from.IsZero() || to.IsZero() { return 0 }
    return int64(to.Sub(from) / time.Millisecond)
}

// useOrFallback returns a if set, otherwise b. Helps when TLS phase отсутствует (plain HTTP).
func useOrFallback(a, b time.Time) time.Time { if a.IsZero() { return b }; return a }


