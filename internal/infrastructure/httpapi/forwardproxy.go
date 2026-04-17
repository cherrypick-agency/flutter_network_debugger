package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	connectproxy "github.com/777genius/proxykit/connect"
	forwardproxy "github.com/777genius/proxykit/forward"
	"github.com/777genius/proxykit/observe"
	proxyhttp "github.com/777genius/proxykit/proxyhttp"
	"network-debugger/internal/domain"
	mappingrt "network-debugger/internal/features/mapping/runtime"
	"network-debugger/pkg/shared/id"
)

var errForwardInterceptDropped = errors.New("forward request dropped by interceptor")

// handleForwardOrNotFound routes absolute-URI and CONNECT requests as a standard forward proxy.
// Non-proxy requests fall back to 404 so that REST/WS routes are handled by other handlers.
func (d *Deps) handleForwardOrNotFound(w http.ResponseWriter, r *http.Request) {
	// For server requests in net/http, absolute URI may come in RequestURI,
	// while r.URL.Scheme/Host may be empty. Handle both cases.
	if r.Method == http.MethodConnect ||
		(r.URL != nil && r.URL.Scheme != "" && r.URL.Host != "") ||
		proxyhttp.IsAbsoluteURL(r.RequestURI) {
		// If URL is not parsed yet but RequestURI is absolute, restore r.URL
		if r.Method != http.MethodConnect && (r.URL == nil || r.URL.Scheme == "" || r.URL.Host == "") && proxyhttp.IsAbsoluteURL(r.RequestURI) {
			if u, err := url.Parse(r.RequestURI); err == nil {
				r.URL = u
			}
		}
		d.handleForwardProxy(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
}

func (d *Deps) handleForwardProxy(w http.ResponseWriter, r *http.Request) {
	// Offline simulation
	if d.Cfg.ThrottleOffline {
		writeError(w, http.StatusServiceUnavailable, "OFFLINE", "proxy offline (simulated)", nil)
		return
	}
	if r.Method == http.MethodConnect {
		// If MITM is enabled and domain matches — intercept TLS
		if d.MITM != nil && d.MITM.CA != nil && d.MITM.shouldIntercept(r.Host) {
			d.handleConnectMITM(w, r)
			return
		}
		d.handleConnectTunnel(w, r)
		return
	}
	// WebSocket upgrade over plain HTTP (ws://) via forward proxy
	// If scheme is ws — treat it as WS upgrade, even if Upgrade header is non-standard.
	if r.URL != nil && r.URL.Scheme == "ws" {
		d.handleHTTPForwardWebSocket(w, r)
		return
	}
	// Or explicit Upgrade to ws over plain http absolute URL
	if isWebSocketRequest(r) && (r.URL != nil && r.URL.Scheme == "http") {
		d.handleHTTPForwardWebSocket(w, r)
		return
	}
	// Forward regular HTTP request with absolute URI in r.URL
	d.handleHTTPForwardRequest(w, r)
}

// handleHTTPForwardWebSocket proxies WebSocket Upgrade (ws://) in forward-proxy mode.
// Hijacks the client connection, establishes connection to upstream,
// forwards the original Upgrade request (origin-form) and after 101 pipes bytes bidirectionally.
func (d *Deps) handleHTTPForwardWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check if target should be monitored (exclude monitoring endpoints)
	shouldMonitor := d.shouldMonitorTarget(r.URL.Path)

	// Create session (ws)
	sessionID := id.New()
	if shouldMonitor {
		_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, Target: r.URL.String(), ClientAddr: r.RemoteAddr, StartedAt: time.Now().UTC(), Kind: "ws"})
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID})
		d.Metrics.ActiveSessions.Inc()
	}

	// Hijack client
	hj, ok := w.(http.Hijacker)
	if !ok {
		writeError(w, http.StatusInternalServerError, "HIJACK_NOT_SUPPORTED", "proxy: hijacking not supported", nil)
		if shouldMonitor {
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr("hijack not supported"))
			d.Metrics.ActiveSessions.Dec()
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
		}
		return
	}
	clientConn, clientBuf, err := hj.Hijack()
	if err != nil {
		if shouldMonitor {
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
			d.Metrics.ActiveSessions.Dec()
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
		}
		return
	}

	// Dial to upstream (ws:// => tcp:80)
	upstreamAddr := r.URL.Host
	if !strings.Contains(upstreamAddr, ":") {
		upstreamAddr = net.JoinHostPort(upstreamAddr, "80")
	}
	upstreamConn, err := net.DialTimeout("tcp", upstreamAddr, 10*time.Second)
	if err != nil {
		// Notify client with 502 and close
		_, _ = clientBuf.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = clientBuf.Flush()
		_ = clientConn.Close()
		if shouldMonitor {
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
			d.Metrics.ActiveSessions.Dec()
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
		}
		return
	}

	// Prepare outgoing request to upstream (origin-form)
	outReq := r.Clone(contextWithNoCancel())
	// Origin-form: empty RequestURI, URL points to target (scheme/host for Host header)
	outURL := *r.URL
	outURL.Scheme = "http"
	outReq.URL = &outURL
	outReq.Host = outURL.Host
	outReq.RequestURI = "" // important for correct origin-form writing
	// Remove hop-by-hop headers but keep Upgrade/Connection
	sanitizeForUpgrade(outReq.Header)

	// Light request preview (downstream frame)
	if shouldMonitor {
		reqPreview := buildHTTPRequestPreview(outReq, nil)
		fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: 0, Preview: reqPreview}
		_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
		d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()
	}

	// Write Upgrade request to upstream
	if err := outReq.Write(upstreamConn); err != nil {
		_, _ = clientBuf.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = clientBuf.Flush()
		_ = clientConn.Close()
		_ = upstreamConn.Close()
		if shouldMonitor {
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
			d.Metrics.ActiveSessions.Dec()
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
		}
		return
	}

	// Read upstream response and send to client as-is
	upr := bufio.NewReader(upstreamConn)
	resp, err := http.ReadResponse(upr, outReq)
	if err != nil {
		_, _ = clientBuf.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = clientBuf.Flush()
		_ = clientConn.Close()
		_ = upstreamConn.Close()
		if shouldMonitor {
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
			d.Metrics.ActiveSessions.Dec()
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
		}
		return
	}
	// Send status/headers to client
	_ = resp.Write(clientBuf)
	_ = clientBuf.Flush()

	// 101 — switch to bidirectional WebSocket frame piping
	if resp.StatusCode == http.StatusSwitchingProtocols {
		// Account for bytes possibly already buffered in upr (first WS frames) and at client
		if n := upr.Buffered(); n > 0 {
			if b, _ := upr.Peek(n); len(b) > 0 {
				upstreamConn = &prependConn{Conn: upstreamConn, r: bytes.NewReader(append([]byte(nil), b...))}
			}
		}
		if n := clientBuf.Reader.Buffered(); n > 0 {
			if b, _ := clientBuf.Reader.Peek(n); len(b) > 0 {
				clientConn = &prependConn{Conn: clientConn, r: bytes.NewReader(append([]byte(nil), b...))}
			}
		}
		// Signal monitor about upgrade fact (minimal event),
		// full events will come from pipeWSMessages during SIO parsing
		if shouldMonitor {
			e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: "/_sys", Name: "upgraded", ArgsPreview: "[]"}
			_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
		}

		go d.pipeWSMessages(sessionID, clientConn, upstreamConn, domain.DirectionClientToUpstream, shouldMonitor)
		d.pipeWSMessages(sessionID, upstreamConn, clientConn, domain.DirectionUpstreamToClient, shouldMonitor)
		return
	}

	// Not 101 — consider the Upgrade attempt completed
	_ = clientConn.Close()
	_ = upstreamConn.Close()
	if shouldMonitor {
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
		d.Metrics.ActiveSessions.Dec()
	}
}

// sanitizeForUpgrade removes hop-by-hop headers, keeping Upgrade/Connection.
func sanitizeForUpgrade(h http.Header) {
	// Basic list of hop-by-hop headers without Upgrade/Connection
	hop := []string{"Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding"}
	for _, k := range hop {
		h.Del(k)
	}
	// Keep Upgrade and Connection as-is
}

func (d *Deps) handleConnectTunnel(w http.ResponseWriter, r *http.Request) {
	sessionID := id.New()
	handler := connectproxy.New(connectproxy.Options{
		GenerateSessionID: func() string { return sessionID },
		WriteError: func(rw http.ResponseWriter, _ *http.Request, status int, err error) {
			if errors.Is(err, connectproxy.ErrHijackNotSupported) {
				writeError(rw, http.StatusInternalServerError, "HIJACK_NOT_SUPPORTED", "proxy: hijacking not supported", nil)
				return
			}
			writeError(rw, status, "CONNECT_TUNNEL_FAILED", err.Error(), nil)
		},
		Hooks: observe.Hooks{
			OnSessionOpen: func(_ context.Context, session observe.Session) error {
				_ = d.Svc.Create(contextWithNoCancel(), domain.Session{
					ID:         sessionID,
					Target:     "connect://" + r.Host,
					ClientAddr: r.RemoteAddr,
					StartedAt:  session.StartedAt,
					Kind:       string(observe.SessionKindConnect),
				})
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID})
				d.Metrics.ActiveSessions.Inc()
				return nil
			},
			OnProtocolEvent: func(_ context.Context, event observe.ProtocolEvent) error {
				if event.Name != "tunnel_established" {
					return nil
				}
				e := domain.Event{
					ID:          id.New(),
					Ts:          time.Now().UTC(),
					Namespace:   "/_sys",
					Name:        event.Name,
					ArgsPreview: "{}",
				}
				_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
				return nil
			},
			OnSessionClose: func(_ context.Context, info observe.CloseInfo) {
				var errMsg *string
				if info.Err != nil {
					msg := info.Err.Error()
					errMsg = &msg
				}
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), errMsg)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
			},
		},
	})
	handler.ServeHTTP(w, r)
}

// handleConnectMITM: establishes TLS with client using a leaf certificate from local CA,
// and simultaneously initiates an outgoing connection to upstream (TLS). All HTTP/1.1 requests/responses
// inside TLS are decrypted and can be instrumented similarly to reverse proxy.
func (d *Deps) handleConnectMITM(w http.ResponseWriter, r *http.Request) {
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
	// Respond to client that tunnel is established
	_, _ = bufrw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	_ = bufrw.Flush()

	// Get certificate for this host
	leaf, err := d.MITM.CA.IssueFor(upstream)
	if err != nil {
		_ = clientConn.Close()
		return
	}
	// Configure TLS server for client
	tlsSrv := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{leaf},
		NextProtos:   []string{"http/1.1"}, // simplified: only H1 inside
	})
	if err := tlsSrv.Handshake(); err != nil {
		_ = tlsSrv.Close()
		return
	}

	// Dial to upstream TCP, then TLS client
	upstreamTCP, err := net.DialTimeout("tcp", upstream, 10*time.Second)
	if err != nil {
		_ = tlsSrv.Close()
		return
	}
	// TLS client side to real server
	serverName := upstream
	if h, _, err := net.SplitHostPort(upstream); err == nil {
		serverName = h
	}
	tlsCli := tls.Client(upstreamTCP, &tls.Config{ServerName: serverName, InsecureSkipVerify: d.Cfg.InsecureTLS})
	if err := tlsCli.Handshake(); err != nil {
		_ = tlsCli.Close()
		_ = tlsSrv.Close()
		return
	}

	// Create session (type http), will log requests/responses
	sessionID := id.New()
	_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, Target: "mitm://" + upstream, ClientAddr: r.RemoteAddr, StartedAt: time.Now().UTC(), Kind: "http"})
	d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()

	// Simple loop: read HTTP requests from client, send to upstream, read response, send back.
	// Works for keep-alive request sequences.
	go func() {
		defer func() {
			_ = tlsCli.Close()
			_ = tlsSrv.Close()
			ctx := contextWithNoCancel()
			if sess, ok, _ := d.Svc.Get(ctx, sessionID); !ok || sess.ClosedAt == nil {
				_ = d.Svc.SetClosed(ctx, sessionID, time.Now().UTC(), nil)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
			}
		}()

		// Upload throttling in MITM tunnel (client -> upstream)
		var srcConn net.Conn = tlsSrv
		if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleUpKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) {
			srcConn = wrapConnForThrottle(&d.Cfg, srcConn, "up")
		}
		clientBR := bufio.NewReader(srcConn)
		serverBR := bufio.NewReader(tlsCli)
		for {
			// Read request from client
			req, err := http.ReadRequest(clientBR)
			if err != nil {
				return
			}
			// Rewrite scheme/host for upstream
			req.URL.Scheme = "https"
			req.URL.Host = upstream
			req.RequestURI = ""
			removeHopHeaders(req.Header)
			// Standard forwarding headers (can be disabled in stealth mode)
			if ip := clientHost(r.RemoteAddr); ip != "" {
				req.Header.Set("X-Forwarded-For", ip)
			}
			req.Header.Set("Via", "network-debugger")

			// For preview: carefully peek body
			var reqBodyBuf []byte
			if req.Body != nil {
				peekSize := int(previewMaxBytes.Load())
				if peekSize <= 0 {
					peekSize = 65536
				}
				if peekSize > 65536 {
					peekSize = 65536
				}
				peek := make([]byte, peekSize)
				n, _ := io.ReadFull(req.Body, peek)
				if n > 0 {
					reqBodyBuf = peek[:n]
					req.Body = io.NopCloser(io.MultiReader(bytes.NewReader(reqBodyBuf), req.Body))
				}
			}
			// Log request like in reverse proxy
			rPrev := &http.Request{Method: req.Method, URL: req.URL, Header: req.Header}
			reqPreview := buildHTTPRequestPreview(rPrev, reqBodyBuf)
			fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(req.ContentLength), Preview: reqPreview}
			_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
			d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

			// Mapping (MITM): evaluate and apply before sending to upstream
			d.CfgMu.RLock()
			mappingEnabled := d.Cfg.MappingEnabled
			d.CfgMu.RUnlock()
			if mappingEnabled && d.MapRt != nil {
				if dec, ok := d.MapRt.EvalRequest(req); ok {
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
							resp := &http.Response{
								StatusCode:    http.StatusBadGateway,
								Status:        strconv.Itoa(http.StatusBadGateway) + " " + http.StatusText(http.StatusBadGateway),
								Header:        h,
								Body:          io.NopCloser(bytes.NewReader(msg)),
								ContentLength: int64(len(msg)),
							}
							preview := buildHTTPResponsePreview(resp)
							fr2 := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: len(msg), Preview: preview}
							_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr2)
							d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr2.ID})
							d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()
							_ = resp.Write(tlsSrv)
							continue
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
						preview := buildHTTPResponsePreview(resp)
						fr2 := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: len(bodyAll), Preview: preview}
						_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr2)
						d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr2.ID})
						d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()
						_ = resp.Write(tlsSrv)
						continue
					}
					if u, err := url.Parse(dec.RemoteURL); err == nil {
						req.URL = u
						if !dec.PreserveHost {
							req.Host = u.Host
						}
					}
				}
			}

			// Interception: request inside MITM tunnel
			if d.InterceptSvc != nil {
				mgr := d.InterceptSvc.Manager()
				mgrCfg := mgr.Config()
				if mgrCfg.Enabled && mgrCfg.Requests {
					capBody := reqBodyBuf
					if max := mgrCfg.BodyMaxBytes; max > 0 && len(capBody) > max {
						capBody = capBody[:max]
					}
					origEnc := strings.ToLower(req.Header.Get("Content-Encoding"))
					decCap, _ := decodeForIntercept(capBody, origEnc, mgrCfg.BodyMaxBytes)
					ct := strings.ToLower(req.Header.Get("Content-Type"))
					input := toRequestMatchInput(req)
					input.BodyPreview = string(decCap)
					if dec, _ := mgr.InterceptRequest(contextWithNoCancel(), sessionID, input, decCap, ct); dec != nil {
						if strings.ToLower(dec.Action) == "drop" {
							return
						}
						if dec.Method != "" {
							req.Method = dec.Method
						}
						if dec.Headers != nil {
							req.Header = http.Header(dec.Headers)
						}
						if dec.Body != nil {
							bodyToWrite := dec.Body
							if mgrCfg.Reencode && (origEnc == "gzip" || origEnc == "deflate") {
								if encBody, ok := encodeForIntercept(dec.Body, origEnc); ok {
									bodyToWrite = encBody
									req.Header.Set("Content-Encoding", origEnc)
								} else {
									req.Header.Del("Content-Encoding")
								}
							} else {
								req.Header.Del("Content-Encoding")
							}
							req.Body = io.NopCloser(bytes.NewReader(bodyToWrite))
							req.ContentLength = int64(len(bodyToWrite))
							req.Header.Del("Transfer-Encoding")
							req.Header.Set("Content-Length", strconv.Itoa(len(bodyToWrite)))
						}
					}
				}
			}

			// Send request to upstream
			if err := req.Write(tlsCli); err != nil {
				return
			}
			// Read response
			resp, err := http.ReadResponse(serverBR, req)
			if err != nil {
				return
			}
			// If upgrade (e.g., WebSocket) — after writing 101 switch to raw byte piping
			preview := buildHTTPResponsePreview(resp)
			fr2 := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: int(resp.ContentLength), Preview: preview}
			_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr2)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr2.ID})
			d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

			// Limit upstream to client download speed
			if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleDownKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && resp.Body != nil {
				bps := kbpsToBytesPerSec(d.Cfg.ThrottleDownKbps)
				resp.Body = io.NopCloser(wrapReaderThrottleLoss(resp.Body, bps, d.Cfg.ThrottlePacketLoss))
			}
			// Send response to client
			if err := resp.Write(tlsSrv); err != nil {
				return
			}

			// Interception: response inside MITM tunnel (before possible WS upgrade handling below)
			if d.InterceptSvc != nil {
				mgr := d.InterceptSvc.Manager()
				mgrCfg := mgr.Config()
				if mgrCfg.Enabled && mgrCfg.Responses {
					var capBuf []byte
					if resp.Body != nil {
						lim := mgrCfg.BodyMaxBytes
						if lim <= 0 {
							lim = 1 << 20
						}
						buf := make([]byte, lim)
						if n, _ := io.ReadFull(resp.Body, buf); n > 0 {
							capBuf = append(capBuf[:0], buf[:n]...)
							resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(capBuf), resp.Body))
						}
					}
					ct := strings.ToLower(resp.Header.Get("Content-Type"))
					respInput := toResponseMatchInput(resp)
					respInput.BodyPreview = string(capBuf)
					if dec, _ := mgr.InterceptResponse(contextWithNoCancel(), sessionID, respInput, capBuf, ct); dec != nil {
						if dec.Status > 0 {
							resp.StatusCode = dec.Status
							if txt := http.StatusText(dec.Status); txt != "" {
								resp.Status = strconv.Itoa(dec.Status) + " " + txt
							} else {
								resp.Status = strconv.Itoa(dec.Status)
							}
						}
						if dec.Headers != nil {
							resp.Header = http.Header(dec.Headers)
						}
						if dec.Body != nil {
							resp.Header.Del("Content-Encoding")
							resp.Body = io.NopCloser(bytes.NewReader(dec.Body))
							resp.ContentLength = int64(len(dec.Body))
							resp.Header.Set("Content-Length", strconv.Itoa(len(dec.Body)))
						}
					}
				}
			}

			if resp.StatusCode == http.StatusSwitchingProtocols {
				// After 101 HTTP is no longer present — switch to WS piping with frame logging.
				// Check if this specific request path should be monitored
				shouldMonitor := d.shouldMonitorTarget(req.URL.Path)
				go d.pipeWSMessages(sessionID, tlsSrv, tlsCli, domain.DirectionClientToUpstream, shouldMonitor)
				d.pipeWSMessages(sessionID, tlsCli, tlsSrv, domain.DirectionUpstreamToClient, shouldMonitor)
				return
			}
		}
	}()
}

func (d *Deps) handleHTTPForwardRequest(w http.ResponseWriter, r *http.Request) {
	// r.URL is absolute here (scheme+host+path)
	// Check if target should be monitored (exclude monitoring endpoints)
	shouldMonitor := d.shouldMonitorTarget(r.URL.Path)

	// Create session for logging
	sessionID := id.New()
	if shouldMonitor {
		_ = d.Svc.Create(r.Context(), domain.Session{ID: sessionID, Target: r.URL.String(), ClientAddr: r.RemoteAddr, StartedAt: time.Now().UTC(), Kind: "http"})
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID})
		d.Metrics.ActiveSessions.Inc()
	}

	outURL := *r.URL
	// Safely peek a small portion of request body and keep stream intact
	var reqBodyBuf []byte
	var reqFrameID string
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
			reqBodyBuf = peek[:n]
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(reqBodyBuf), r.Body))
		}
	}
	// For preview, use the real upstream absolute URL
	if shouldMonitor {
		rPrev := *r
		rPrev.URL = &outURL
		reqPreview := buildHTTPRequestPreview(&rPrev, reqBodyBuf)
		reqFrameID = id.New()
		fr := domain.Frame{ID: reqFrameID, Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(r.ContentLength), Preview: reqPreview}
		_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
		d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()
	}

	sampleLimit := previewBodySampleLimit()
	transport := newTransport(d.Cfg)
	var localMappedResp *http.Response
	var respProto string
	var respContentType string

	handler := forwardproxy.New(forwardproxy.Options{
		RoundTripper: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if localMappedResp != nil {
				resp := localMappedResp
				localMappedResp = nil
				return resp, nil
			}
			return transport.RoundTrip(req)
		}),
		GenerateSessionID: func() string { return sessionID },
		ObserveRequest:    func(*http.Request) bool { return shouldMonitor },
		ForwardedHeaders: proxyhttp.ForwardedHeaderConfig{
			ClientIP: clientHost(r.RemoteAddr),
			Proto:    forwardedProto(r),
			Via:      "network-debugger",
		},
		WriteError: func(rw http.ResponseWriter, _ *http.Request, _ int, err error) {
			if errors.Is(err, errForwardInterceptDropped) {
				writeError(rw, http.StatusForbidden, "INTERCEPT_DROPPED", "request dropped by interceptor", nil)
				return
			}
			writeError(rw, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error(), map[string]any{"target": outURL.String()})
		},
		MutateRequest: func(_ context.Context, req *http.Request) error {
			d.CfgMu.RLock()
			mappingEnabledFwd := d.Cfg.MappingEnabled
			d.CfgMu.RUnlock()
			if mappingEnabledFwd && d.MapRt != nil {
				if dec, ok := d.MapRt.EvalRequest(req); ok {
					if shouldMonitor {
						d.broadcastMonitorEvent(domain.MonitorEvent{Type: "mapping_applied", ID: sessionID, Ref: dec.RuleID})
						if d.Metrics != nil && d.Metrics.MappingAppliedTotal != nil {
							d.Metrics.MappingAppliedTotal.WithLabelValues(string(dec.Kind)).Inc()
						}
					}
					if dec.Kind == "local" {
						localMappedResp = buildForwardLocalMappingResponse(dec)
						return nil
					}
					if u, err := url.Parse(dec.RemoteURL); err == nil {
						req.URL = u
						if !dec.PreserveHost {
							req.Host = u.Host
						}
					}
				}
			}

			if d.InterceptSvc != nil {
				mgr := d.InterceptSvc.Manager()
				mgrCfg := mgr.Config()
				if mgrCfg.Enabled && mgrCfg.Requests {
					capBody := reqBodyBuf
					if max := mgrCfg.BodyMaxBytes; max > 0 && len(capBody) > max {
						capBody = capBody[:max]
					}
					origEnc := strings.ToLower(req.Header.Get("Content-Encoding"))
					decCap, _ := decodeForIntercept(capBody, origEnc, mgrCfg.BodyMaxBytes)
					ct := strings.ToLower(req.Header.Get("Content-Type"))
					input := toRequestMatchInput(req)
					input.BodyPreview = string(decCap)
					if dec, _ := mgr.InterceptRequest(r.Context(), sessionID, input, decCap, ct); dec != nil {
						if strings.ToLower(dec.Action) == "drop" {
							return errForwardInterceptDropped
						}
						if dec.Method != "" {
							req.Method = dec.Method
						}
						if dec.URL != "" {
							if u, err := url.Parse(dec.URL); err == nil {
								if u.Scheme != "" && u.Host != "" {
									req.URL = u
									req.Host = u.Host
								} else {
									newURL := *req.URL
									newURL.Path = u.Path
									newURL.RawQuery = u.RawQuery
									req.URL = &newURL
								}
							}
						}
						if dec.Headers != nil {
							req.Header = http.Header(dec.Headers)
						}
						if dec.Body != nil {
							bodyToWrite := dec.Body
							if mgrCfg.Reencode && (origEnc == "gzip" || origEnc == "deflate") {
								if encBody, ok := encodeForIntercept(dec.Body, origEnc); ok {
									bodyToWrite = encBody
									req.Header.Set("Content-Encoding", origEnc)
								} else {
									req.Header.Del("Content-Encoding")
								}
							} else {
								req.Header.Del("Content-Encoding")
							}
							req.Body = io.NopCloser(bytes.NewReader(bodyToWrite))
							req.ContentLength = int64(len(bodyToWrite))
							req.Header.Del("Transfer-Encoding")
							req.Header.Set("Content-Length", strconv.Itoa(len(bodyToWrite)))
						}
					}
				}
			}

			if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleUpKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && req.Body != nil {
				bps := kbpsToBytesPerSec(d.Cfg.ThrottleUpKbps)
				req.Body = io.NopCloser(wrapReaderThrottleLoss(req.Body, bps, d.Cfg.ThrottlePacketLoss))
			}
			return nil
		},
		MutateResponse: func(_ context.Context, _ *http.Request, resp *http.Response) error {
			if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleDownKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && resp.Body != nil {
				bps := kbpsToBytesPerSec(d.Cfg.ThrottleDownKbps)
				resp.Body = io.NopCloser(wrapReaderThrottleLoss(resp.Body, bps, d.Cfg.ThrottlePacketLoss))
			}

			if d.InterceptSvc != nil {
				mgr := d.InterceptSvc.Manager()
				mgrCfg := mgr.Config()
				if mgrCfg.Enabled && mgrCfg.Responses {
					var capBuf []byte
					if resp.Body != nil {
						lim := mgrCfg.BodyMaxBytes
						if lim <= 0 {
							lim = 1 << 20
						}
						buf := make([]byte, lim)
						if n, _ := io.ReadFull(resp.Body, buf); n > 0 {
							capBuf = append(capBuf[:0], buf[:n]...)
							resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(capBuf), resp.Body))
						}
					}
					origEnc := strings.ToLower(resp.Header.Get("Content-Encoding"))
					decCap, _ := decodeForIntercept(capBuf, origEnc, mgrCfg.BodyMaxBytes)
					ct := strings.ToLower(resp.Header.Get("Content-Type"))
					respInput := toResponseMatchInput(resp)
					respInput.BodyPreview = string(decCap)
					if dec, _ := mgr.InterceptResponse(r.Context(), sessionID, respInput, capBuf, ct); dec != nil {
						if dec.Status > 0 {
							resp.StatusCode = dec.Status
							if txt := http.StatusText(dec.Status); txt != "" {
								resp.Status = strconv.Itoa(dec.Status) + " " + txt
							} else {
								resp.Status = strconv.Itoa(dec.Status)
							}
						}
						if dec.Headers != nil {
							resp.Header = http.Header(dec.Headers)
						}
						if dec.Body != nil {
							bodyToWrite := dec.Body
							if mgrCfg.Reencode && (origEnc == "gzip" || origEnc == "deflate") {
								if encBody, ok := encodeForIntercept(dec.Body, origEnc); ok {
									bodyToWrite = encBody
									resp.Header.Set("Content-Encoding", origEnc)
								} else {
									resp.Header.Del("Content-Encoding")
								}
							} else {
								resp.Header.Del("Content-Encoding")
							}
							resp.Body = io.NopCloser(bytes.NewReader(bodyToWrite))
							resp.ContentLength = int64(len(bodyToWrite))
							resp.Header.Set("Content-Length", strconv.Itoa(len(bodyToWrite)))
						}
					}
				}
			}

			respProto = resp.Proto
			respContentType = resp.Header.Get("Content-Type")
			sleepResponseDelay(d.Cfg)
			return nil
		},
		SampleRequestBodyBytes:  sampleLimit,
		SampleResponseBodyBytes: sampleLimit,
		Hooks: observe.Hooks{
			OnHTTPResponse: func(_ context.Context, resp observe.HTTPResponse) error {
				if !shouldMonitor {
					return nil
				}
				preview := buildObservedHTTPResponsePreview(resp)
				fr := domain.Frame{
					ID:        id.New(),
					Ts:        time.Now().UTC(),
					Direction: domain.DirectionUpstreamToClient,
					Opcode:    domain.OpcodeText,
					Size:      observedBodySize(resp.Body),
					Preview:   preview,
				}
				_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
				d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()
				return nil
			},
			OnHTTPRoundTrip: func(_ context.Context, rt observe.HTTPRoundTrip) error {
				if !shouldMonitor {
					return nil
				}
				tx := domain.HTTPTransaction{
					ID:        id.New(),
					SessionID: sessionID,
					Method:    r.Method,
					URL:       outURL.String(),
					Status:    rt.Response.StatusCode,
					ReqSize:   observedBodySize(rt.Request.Body),
					RespSize:  observedBodySize(rt.Response.Body),
					StartedAt: rt.Session.StartedAt,
					EndedAt:   rt.Session.StartedAt.Add(rt.Timings.Total),
					Timings: domain.HTTPTimings{
						TTFB:  int64(rt.Timings.TTFB / time.Millisecond),
						Total: int64(rt.Timings.Total / time.Millisecond),
					},
					ReqHeaders:      cloneHeader(rt.Request.Header),
					RespHeaders:     cloneHeader(rt.Response.Header),
					Cookies:         r.Cookies(),
					QueryParams:     outURL.Query(),
					ReqHTTPVersion:  r.Proto,
					RespHTTPVersion: respProto,
					ContentType:     respContentType,
				}
				if d.Cfg.CaptureBodies && len(rt.Request.Body.Bytes) > 0 {
					if f, err := d.spoolBodyBytes(rt.Request.Body.Bytes, "req"); err == nil && f != "" {
						tx.ReqBodyFile = f
						d.Svc.AddSpoolFile(contextWithNoCancel(), sessionID, f)
						if reqFrameID != "" {
							_ = d.Svc.UpdateFrameBodyFile(contextWithNoCancel(), sessionID, reqFrameID, f)
						}
					}
				}
				_ = d.Svc.AddHTTPTransaction(contextWithNoCancel(), tx)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "http_tx_added", ID: sessionID, Ref: tx.ID})
				return nil
			},
			OnSessionClose: func(_ context.Context, info observe.CloseInfo) {
				if !shouldMonitor {
					return
				}
				var errMsg *string
				if info.Err != nil {
					msg := info.Err.Error()
					errMsg = &msg
				}
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), errMsg)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
			},
		},
	})
	handler.ServeHTTP(w, r)
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

func previewBodySampleLimit() int {
	peekSize := int(previewMaxBytes.Load())
	if peekSize <= 0 {
		return 65536
	}
	if peekSize > 65536 {
		return 65536
	}
	return peekSize
}

func forwardedProto(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func buildForwardLocalMappingResponse(dec mappingrt.Decision) *http.Response {
	var bodyAll []byte
	var readErr error
	if dec.LocalFilePath != nil && *dec.LocalFilePath != "" {
		bodyAll, readErr = os.ReadFile(*dec.LocalFilePath)
	} else if dec.LocalBlobPath != nil && *dec.LocalBlobPath != "" {
		bodyAll, readErr = os.ReadFile(*dec.LocalBlobPath)
	}

	headers := http.Header{}
	if readErr != nil {
		bodyAll = []byte("mapping local source is not readable")
		headers.Set("Content-Type", "text/plain; charset=utf-8")
		headers.Set("X-ND-Mapped", "local")
		headers.Set("X-ND-Rule", dec.RuleID)
		headers.Set("X-ND-Mapping-Error", "read_failed")
		return &http.Response{
			StatusCode:    http.StatusBadGateway,
			Status:        strconv.Itoa(http.StatusBadGateway) + " " + http.StatusText(http.StatusBadGateway),
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        headers,
			Body:          io.NopCloser(bytes.NewReader(bodyAll)),
			ContentLength: int64(len(bodyAll)),
		}
	}

	if dec.ContentTypeOverride != "" {
		headers.Set("Content-Type", dec.ContentTypeOverride)
	} else if len(bodyAll) > 0 {
		headers.Set("Content-Type", http.DetectContentType(bodyAll))
	} else {
		headers.Set("Content-Type", "application/octet-stream")
	}
	headers.Set("X-ND-Mapped", "local")
	headers.Set("X-ND-Rule", dec.RuleID)
	status := dec.StatusOverride
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode:    status,
		Status:        strconv.Itoa(status) + " " + http.StatusText(status),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        headers,
		Body:          io.NopCloser(bytes.NewReader(bodyAll)),
		ContentLength: int64(len(bodyAll)),
	}
}

func buildObservedHTTPResponsePreview(resp observe.HTTPResponse) string {
	body := append([]byte(nil), resp.Body.Bytes...)
	contentLength := int64(len(body))
	if resp.Body.TotalSize >= 0 {
		contentLength = resp.Body.TotalSize
	}
	previewResp := &http.Response{
		StatusCode:    resp.StatusCode,
		Header:        cloneHeader(resp.Header),
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: contentLength,
	}
	return buildHTTPResponsePreview(previewResp)
}

func observedBodySize(body observe.InlineBody) int {
	if body.TotalSize >= 0 {
		return int(body.TotalSize)
	}
	return len(body.Bytes)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// prependConn first reads prepared bytes from r, then reads from the underlying connection.
type prependConn struct {
	net.Conn
	r io.Reader
}

func (p *prependConn) Read(b []byte) (int, error) {
	if p.r != nil {
		n, err := p.r.Read(b)
		if err == io.EOF {
			p.r = nil
			if n > 0 {
				return n, nil
			}
			// proceed to reading from Conn
		} else if n > 0 || err != nil {
			return n, err
		}
	}
	return p.Conn.Read(b)
}

// isAbsoluteURL returns true if s looks like an absolute URI.
func isAbsoluteURL(s string) bool {
	return proxyhttp.IsAbsoluteURL(s)
}
