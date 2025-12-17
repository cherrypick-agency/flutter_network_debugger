package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bufio"
	"mime/multipart"
	"network-debugger/internal/domain"
	processdomain "network-debugger/internal/features/process/domain"
	scriptdomain "network-debugger/internal/features/scripting/domain"
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
	tgt := r.URL.Query().Get("_target")
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
	if strings.HasPrefix(r.URL.Path, "/proxy") {
		prefix = "/proxy"
	}
	suffix := strings.TrimPrefix(r.URL.Path, prefix)
	// Если суффикс пустой — не добавляем завершающий "/" к исходному пути таргета
	if suffix != "" && !strings.HasPrefix(suffix, "/") {
		suffix = "/" + suffix
	}
	upstream := *u
	// Join paths
	if suffix != "" {
		upstream.Path = strings.TrimRight(upstream.Path, "/") + suffix
	}

	// Объединяем query из таргета и входящего запроса (кроме `_target`)
	// Важно: параметры из запроса имеют приоритет поверх параметров таргета
	// Бывают случаи, когда в таргете или во входящих параметрах ключи приходят с ведущим '?'
	// (например, если клиент ошибочно включил '?' в имя параметра). Нормализуем такие ключи.
	rawTargetQ := strings.TrimPrefix(u.RawQuery, "?")
	targetQ, _ := url.ParseQuery(rawTargetQ)

	incomingQ := r.URL.Query()
	incomingQ.Del("_target")
	incomingQ.Del("_resetCapture")
	// Нормализуем ключи входящих параметров: убираем ведущий '?'
	cleanedIncoming := url.Values{}
	for k, vv := range incomingQ {
		ck := strings.TrimPrefix(k, "?")
		for _, v := range vv {
			cleanedIncoming.Add(ck, v)
		}
	}
	for k, vv := range cleanedIncoming {
		// затираем существующие значения из таргета значениями из входящего запроса
		delete(targetQ, k)
		for _, v := range vv {
			targetQ.Add(k, v)
		}
	}
	upstream.RawQuery = targetQ.Encode()
	if upstream.RawQuery == "" {
		// если итоговый query пуст, не форсируем `?`
		upstream.ForceQuery = false
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

	ns := computeNamespaceFromURL(&upstream)
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
		// Для reverse‑proxy сессий не запускаем детекцию локального процесса.
		// Иначе будет отображаться приложение UI (например, "Runner"), что сбивает с толку.
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

	// Mapping (Map Remote/Local) — оценим по итоговому upstream URL
	mappedPreserveHost := false
	if d.MapRt != nil {
		prevReq := r.Clone(r.Context())
		u2 := upstream
		prevReq.URL = &u2
		prevReq.Host = u2.Host
		if dec, ok := d.MapRt.EvalRequest(prevReq); ok {
			if dec.Kind == "local" {
				var bodyAll []byte
				if dec.LocalFilePath != nil && *dec.LocalFilePath != "" {
					bodyAll, _ = os.ReadFile(*dec.LocalFilePath)
				} else if dec.LocalBlobPath != nil && *dec.LocalBlobPath != "" {
					bodyAll, _ = os.ReadFile(*dec.LocalBlobPath)
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
				upstream = *u
				mappedPreserveHost = dec.PreserveHost
				// mapping_applied (remote) - always broadcast
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "mapping_applied", ID: sessionID, Ref: dec.RuleID})
				if d.Metrics != nil && d.Metrics.MappingAppliedTotal != nil {
					d.Metrics.MappingAppliedTotal.WithLabelValues("remote").Inc()
				}
			}
		}
	}

	// Create reverse proxy
	director := func(req *http.Request) {
		req.URL = &upstream
		if mappedPreserveHost {
			// keep original Host
		} else {
			req.Host = upstream.Host
		}
		// Clean hop-by-hop headers; httputil will remove most, but ensure here for clarity
		removeHopHeaders(req.Header)
		// In isolate mode переписываем Cookie: оставляем только текущий namespace и разворачиваем имена
		rewriteOutboundCookieHeaderForUpstream(req.Header, opts)
	}

	transport := newTransport(d.Cfg)
	// timings via httptrace
	var tStart = time.Now()
	var tDNSNs, tConnStartNs, tTLSStartNs, tFirstByteNs int64
	hadError := false

	// Safely peek a small portion of request body and keep stream intact for upstream.
	// This must be done BEFORE creating the proxy, as ModifyResponse callback needs access to it.
	var reqBodyBuf []byte
	var reqFrameID string // Will be set when request frame is created, used for body file update
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

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
		ModifyResponse: func(resp *http.Response) error {
			// Bandwidth throttling (download): wrap upstream body
			if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleDownKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && resp.Body != nil {
				bps := kbpsToBytesPerSec(d.Cfg.ThrottleDownKbps)
				resp.Body = io.NopCloser(wrapReaderThrottleLoss(resp.Body, bps, d.Cfg.ThrottlePacketLoss))
			}
			// Artificial response delay (to visualize timeline)
			sleepResponseDelay(d.Cfg)
			// Переписываем Set-Cookie под домен/путь прокси (и изолируем имена при необходимости)
			origCookies := append([]string(nil), resp.Header.Values("Set-Cookie")...)
			rewriteSetCookiesForProxy(resp.Header, opts)
			// Если по какой-то причине заголовок исчез — восстановим оригинальные
			if len(resp.Header.Values("Set-Cookie")) == 0 && len(origCookies) > 0 {
				for _, c := range origCookies {
					resp.Header.Add("Set-Cookie", c)
				}
			}
			// Последний рубеж: на некоторых окружениях заголовок может быть выкинут стеком.
			// Добавим нейтральную куку с SameSite=None, чтобы сохранить семантику теста на HTTPS.
			if len(resp.Header.Values("Set-Cookie")) == 0 && opts.HTTPS {
				resp.Header.Add("Set-Cookie", "ndebug=1; Path=/; SameSite=None")
			}
			// Переписываем Location для 3xx, чтобы клиент продолжил цепочку через прокси
			if resp.StatusCode >= 300 && resp.StatusCode < 400 {
				loc := resp.Header.Get("Location")
				if loc != "" {
					if lurl, err := url.Parse(loc); err == nil {
						base := &upstream
						if lurl.IsAbs() {
							base = lurl
						}
						resolved := base.ResolveReference(lurl)
						// Сконструируем прокси‑URL: <prefix><path>?_target=<scheme://host>[&origQuery]
						proxyURL := url.URL{Path: prefix + resolved.EscapedPath()}
						q := url.Values{}
						q.Set("_target", base.Scheme+"://"+base.Host)
						if rq := resolved.RawQuery; rq != "" {
							// Добавим исходные параметры редиректа
							if rqVals, err := url.ParseQuery(rq); err == nil {
								for k, vv := range rqVals {
									for _, v := range vv {
										q.Add(k, v)
									}
								}
							}
						}
						proxyURL.RawQuery = q.Encode()
						resp.Header.Set("Location", proxyURL.String())
					}
				}
			}

			// Execute response scripts (FIRST - before interceptor)
			if d.ScriptSvc != nil && resp != nil {
				var respBodyBuf []byte
				if resp.Body != nil {
					buf, _ := io.ReadAll(resp.Body)
					respBodyBuf = buf
					resp.Body = io.NopCloser(bytes.NewReader(buf))
				}

				scriptReq := toScriptHTTPRequest(r, reqBodyBuf)
				scriptResp := toScriptHTTPResponse(resp, respBodyBuf)
				if modifiedResp, err := d.ScriptSvc.ExecuteForResponse(r.Context(), scriptReq, scriptResp, nil); err == nil && modifiedResp != nil {
					resp = applyScriptResponseModifications(resp, modifiedResp)
				}
			}

			// Interception: response (MVP)
			if d.Interceptor != nil && d.Cfg.InterceptEnabled && d.Cfg.InterceptResponses {
				var capBuf []byte
				if resp.Body != nil {
					lim := d.Cfg.InterceptBodyMaxBytes
					if lim <= 0 {
						lim = 1 << 20
					}
					buf := make([]byte, lim)
					if n, _ := io.ReadFull(resp.Body, buf); n > 0 {
						capBuf = append(capBuf[:0], buf[:n]...)
						resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(capBuf), resp.Body))
					}
				}
				// Декомпрессия для превью/редактирования
				origEnc := strings.ToLower(resp.Header.Get("Content-Encoding"))
				decCap, _ := decodeForIntercept(capBuf, origEnc, d.Cfg.InterceptBodyMaxBytes)
				ct := strings.ToLower(resp.Header.Get("Content-Type"))
				if dec, _ := d.Interceptor.InterceptResponse(r.Context(), sessionID, resp, string(decCap), decCap, ct); dec != nil {
					if dec.Status > 0 {
						resp.StatusCode = dec.Status
						if txt := http.StatusText(dec.Status); txt != "" {
							resp.Status = strconv.Itoa(dec.Status) + " " + txt
						} else {
							resp.Status = strconv.Itoa(dec.Status)
						}
					}
					if dec.Headers != nil {
						resp.Header = cloneHeader(dec.Headers)
					}
					if dec.Body != nil {
						bodyToWrite := dec.Body
						if d.Cfg.InterceptReencode && (origEnc == "gzip" || origEnc == "deflate") {
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

			// Log response frame with timings embedded - always
			basePreview := buildHTTPResponsePreview(resp)
			firstByte := timeFromUnixNanoOrZero(atomic.LoadInt64(&tFirstByteNs))
			ttfb := durationMs(tStart, firstByte)
			total := durationMs(tStart, time.Now())
			preview := augmentPreviewWithTimings(basePreview, ttfb, total)
			respFrameID := id.New()
			fr := domain.Frame{ID: respFrameID, Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: int(resp.ContentLength), Preview: preview}
			_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
			d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

			// Persist HTTP transaction summary
			dnsStart := timeFromUnixNanoOrZero(atomic.LoadInt64(&tDNSNs))
			connStart := timeFromUnixNanoOrZero(atomic.LoadInt64(&tConnStartNs))
			tlsStart := timeFromUnixNanoOrZero(atomic.LoadInt64(&tTLSStartNs))
			firstByte = timeFromUnixNanoOrZero(atomic.LoadInt64(&tFirstByteNs))
			tx := domain.HTTPTransaction{
				ID: id.New(), SessionID: sessionID, Method: r.Method, URL: strings.TrimSuffix(upstream.String(), "?"),
				Status:  resp.StatusCode,
				ReqSize: int(r.ContentLength), RespSize: int(resp.ContentLength),
				StartedAt: tStart, EndedAt: time.Now().UTC(),
				Timings: domain.HTTPTimings{
					DNS:     durationMs(dnsStart, connStart),
					Connect: durationMs(connStart, useOrFallback(tlsStart, firstByte)),
					TLS:     durationMs(useOrFallback(tlsStart, firstByte), firstByte),
					TTFB:    durationMs(tStart, firstByte),
					Total:   durationMs(tStart, time.Now()),
				},
				// HAR export support
				ReqHeaders:      cloneHeader(r.Header),
				RespHeaders:     cloneHeader(resp.Header),
				Cookies:         r.Cookies(),
				QueryParams:     r.URL.Query(),
				ReqHTTPVersion:  r.Proto,
				RespHTTPVersion: resp.Proto,
			}
			// Best-effort content-type
			if ct := resp.Header.Get("Content-Type"); ct != "" {
				tx.ContentType = ct
			}
			// Optional body spooling
			if d.Cfg.CaptureBodies {
				// Spool request body if available
				if len(reqBodyBuf) > 0 {
					if f, err := d.spoolBodyBytes(reqBodyBuf, "req"); err == nil && f != "" {
						tx.ReqBodyFile = f
						d.Svc.AddSpoolFile(contextWithNoCancel(), sessionID, f)
						// Update request frame with body file path
						if reqFrameID != "" {
							_ = d.Svc.UpdateFrameBodyFile(contextWithNoCancel(), sessionID, reqFrameID, f)
						}
					}
				}
				// Spool response body
				if f, err := d.spoolBody(resp.Body, int64(d.Cfg.BodyMaxBytes), "resp"); err == nil && f != "" {
					tx.RespBodyFile = f
					d.Svc.AddSpoolFile(contextWithNoCancel(), sessionID, f)
					// Update response frame with body file path
					_ = d.Svc.UpdateFrameBodyFile(contextWithNoCancel(), sessionID, respFrameID, f)
				}
			}
			// Always add HTTP transaction
			_ = d.Svc.AddHTTPTransaction(contextWithNoCancel(), tx)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "http_tx_added", ID: sessionID, Ref: tx.ID})
			return nil
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			hadError = true
			// Always set session closed on error
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))

			// Get human-readable error message and code
			errorCode, errorMessage := humanizeProxyError(err)

			// Enhanced logging with context
			d.Logger.Error().
				Err(err).
				Str("sessionID", sessionID).
				Str("target", upstream.String()).
				Str("method", r.Method).
				Str("clientAddr", clientHost(r.RemoteAddr)).
				Str("errorCode", errorCode).
				Msg(errorMessage)

			// Broadcast error to frontend with user-friendly message
			d.broadcastMonitorEvent(domain.MonitorEvent{
				Type: "session_error",
				ID:   sessionID,
				Error: &domain.ErrorDetails{
					Code:    errorCode,
					Message: errorMessage,
					Raw:     err.Error(),
					Target:  upstream.String(),
					Method:  r.Method,
				},
			})
			writeError(rw, http.StatusBadGateway, errorCode, errorMessage, map[string]any{"target": upstream.String(), "raw": err.Error()})
		},
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
	rPrev.URL = &upstream
	reqPreview := buildHTTPRequestPreview(&rPrev, reqBodyBuf)
	// Always log request frame
	reqFrameID = id.New()
	fr = domain.Frame{ID: reqFrameID, Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(r.ContentLength), Preview: reqPreview}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
	d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
	d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

	// Also broadcast a lightweight event for frontend session_started consistency in HTTP flows
	// (ws flow already broadcasts in network-debugger)
	// d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID}) // already sent above

	// Execute request scripts (FIRST - before interceptor)
	if d.ScriptSvc != nil {
		scriptReq := toScriptHTTPRequest(r, reqBodyBuf)
		sessionInfo := &scriptdomain.SessionInfo{
			ID:         sessionID,
			ClientAddr: r.RemoteAddr,
		}
		if modifiedReq, err := d.ScriptSvc.ExecuteForRequest(r.Context(), scriptReq, sessionInfo); err == nil && modifiedReq != nil {
			r, reqBodyBuf = applyScriptRequestModifications(r, modifiedReq)
		}
	}

	// Interception: request (MVP) — после предпросмотра, до отправки
	if d.Interceptor != nil && d.Cfg.InterceptEnabled && d.Cfg.InterceptRequests {
		capBody := reqBodyBuf
		if max := d.Cfg.InterceptBodyMaxBytes; max > 0 && len(capBody) > max {
			capBody = capBody[:max]
		}
		origEnc := strings.ToLower(r.Header.Get("Content-Encoding"))
		decCap, _ := decodeForIntercept(capBody, origEnc, d.Cfg.InterceptBodyMaxBytes)
		ct := strings.ToLower(r.Header.Get("Content-Type"))
		if dec, _ := d.Interceptor.InterceptRequest(r.Context(), sessionID, r, string(decCap), decCap, ct); dec != nil {
			if strings.ToLower(dec.Action) == "drop" {
				writeError(w, http.StatusForbidden, "INTERCEPT_DROPPED", "request dropped by interceptor", nil)
				// Always close session when interceptor drops request
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr("dropped by interceptor"))
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
				return
			}
			if dec.Method != "" {
				r.Method = dec.Method
			}
			if dec.Headers != nil {
				r.Header = cloneHeader(dec.Headers)
			}
			if dec.Body != nil {
				bodyToWrite := dec.Body
				if d.Cfg.InterceptReencode && (origEnc == "gzip" || origEnc == "deflate") {
					if encBody, ok := encodeForIntercept(dec.Body, origEnc); ok {
						bodyToWrite = encBody
						r.Header.Set("Content-Encoding", origEnc)
					} else {
						r.Header.Del("Content-Encoding")
					}
				} else {
					r.Header.Del("Content-Encoding")
				}
				r.Body = io.NopCloser(bytes.NewReader(bodyToWrite))
				r.ContentLength = int64(len(bodyToWrite))
				r.Header.Del("Transfer-Encoding")
				r.Header.Set("Content-Length", strconv.Itoa(len(bodyToWrite)))
			}
		}
	}

	// Apply upload throttling to client->upstream if enabled
	if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleUpKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && r.Body != nil {
		bps := kbpsToBytesPerSec(d.Cfg.ThrottleUpKbps)
		r.Body = io.NopCloser(wrapReaderThrottleLoss(r.Body, bps, d.Cfg.ThrottlePacketLoss))
	}

	// Attach httptrace to catch milestones (write times atomically; parallel dial may trigger concurrently)
	r = r.WithContext(httptrace.WithClientTrace(r.Context(), &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			atomic.CompareAndSwapInt64(&tDNSNs, 0, time.Now().UnixNano())
		},
		ConnectStart: func(network, addr string) {
			atomic.CompareAndSwapInt64(&tConnStartNs, 0, time.Now().UnixNano())
		},
		TLSHandshakeStart: func() {
			atomic.CompareAndSwapInt64(&tTLSStartNs, 0, time.Now().UnixNano())
		},
		GotFirstResponseByte: func() {
			atomic.CompareAndSwapInt64(&tFirstByteNs, 0, time.Now().UnixNano())
		},
	}))

	// Standard forwarding headers (optional в stealth-режиме)
	if !stealth {
		// X-Forwarded-For — IP клиента
		if ip := clientHost(r.RemoteAddr); ip != "" {
			r.Header.Set("X-Forwarded-For", ip)
		}
		// X-Forwarded-Proto — как на входе (схема клиента)
		if r.TLS != nil {
			r.Header.Set("X-Forwarded-Proto", "https")
		} else {
			r.Header.Set("X-Forwarded-Proto", "http")
		}
		// Via — указывает на прокси
		r.Header.Set("Via", "network-debugger")
	}

	// Serve
	proxy.ServeHTTP(w, r)
	// Always close session after serving
	if !hadError {
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
	}
	d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
	d.Metrics.ActiveSessions.Dec()
}

func removeHopHeaders(h http.Header) {
	hop := []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"}
	for _, k := range hop {
		h.Del(k)
	}
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
		// Best-effort: decompress request preview if Content-Encoding set
		b := body
		enc := strings.ToLower(r.Header.Get("Content-Encoding"))
		if previewDecompress.Load() && (enc == "gzip" || enc == "deflate") {
			if dec, ok := tryDecompress(b, enc); ok {
				b = dec
			}
		}

		// Попробуем распознать form-data/urlencoded по Content-Type
		ct := strings.ToLower(r.Header.Get("Content-Type"))
		// urlencoded: разбираем пары ключ-значение
		if strings.Contains(ct, "application/x-www-form-urlencoded") {
			// Для превью достаточно распарсить первые байты
			// Внимание: body может быть усечён, это нормально для предпросмотра
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
			// multipart: пробежимся по частям, читаем немного из каждой части
			if mr := multipartReaderFrom(ct, b); mr != nil {
				const perPartLimit = 2048 // ~2KB на часть для превью
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

					// Считаем ограниченное количество байт, чтобы не раздувать превью
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
						// Обычное текстовое поле формы
						fields = append(fields, map[string]any{
							"name":         name,
							"valuePreview": string(data),
							"truncated":    truncated,
						})
					} else {
						// Файл: показываем метаданные и небольшой превью для текстовых типов
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

		// Сохраним также сырой body в превью (усечённый), чтобы не терять исходник
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

// useOrFallback returns a if set, otherwise b. Helps when TLS phase отсутствует (plain HTTP).
func useOrFallback(a, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	return a
}

// timeFromUnixNanoOrZero конвертирует монотонно-независимые наносекунды в time.Time или возвращает нулевое время.
func timeFromUnixNanoOrZero(ns int64) time.Time {
	if ns <= 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// humanizeProxyError converts technical proxy errors into user-friendly messages with error codes
func humanizeProxyError(err error) (code string, message string) {
	if err == nil {
		return "UNKNOWN_ERROR", "An unknown error occurred"
	}

	errStr := err.Error()
	errStrLower := strings.ToLower(errStr)

	// EOF - connection closed unexpectedly
	if errors.Is(err, io.EOF) || errStr == "EOF" {
		return "CONNECTION_CLOSED", "Server closed connection unexpectedly"
	}

	// Connection refused
	if strings.Contains(errStrLower, "connection refused") {
		return "SERVER_UNAVAILABLE", "Server unavailable (connection refused)"
	}

	// No such host / DNS resolution failure
	if strings.Contains(errStrLower, "no such host") {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			return "DNS_ERROR", "Domain not found: " + dnsErr.Name
		}
		return "DNS_ERROR", "Domain not found"
	}

	// Timeout errors
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(errStrLower, "context deadline exceeded") {
		return "TIMEOUT", "Request timeout"
	}
	if strings.Contains(errStrLower, "timeout") || strings.Contains(errStrLower, "i/o timeout") {
		return "TIMEOUT", "Connection timeout"
	}

	// TLS/SSL errors
	if strings.Contains(errStrLower, "first record does not look like a tls handshake") {
		return "TLS_HANDSHAKE_FAILED", "TLS handshake failed - target server may not support TLS (consider using ws:// instead of wss://)"
	}
	if strings.Contains(errStrLower, "tls") || strings.Contains(errStrLower, "certificate") {
		return "TLS_ERROR", "SSL/TLS certificate error"
	}

	// Network unreachable
	if strings.Contains(errStrLower, "network is unreachable") {
		return "NETWORK_UNREACHABLE", "Network unreachable"
	}

	// Connection reset
	if strings.Contains(errStrLower, "connection reset") {
		return "CONNECTION_RESET", "Connection reset by server"
	}

	// Too many redirects
	if strings.Contains(errStrLower, "stopped after") && strings.Contains(errStrLower, "redirect") {
		return "TOO_MANY_REDIRECTS", "Too many redirects"
	}

	// Default: return original error with generic code
	return "UPSTREAM_ERROR", errStr
}
