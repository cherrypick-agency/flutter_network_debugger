package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/777genius/proxykit/observe"
	proxyhttp "github.com/777genius/proxykit/proxyhttp"
	reverseproxy "github.com/777genius/proxykit/reverse"
	"network-debugger/internal/domain"
	scriptdomain "network-debugger/internal/features/scripting/domain"
	"network-debugger/pkg/shared/id"
)

var errReverseInterceptDropped = errors.New("reverse request dropped by interceptor")

type reverseProxyFlow struct {
	deps             *Deps
	request          *http.Request
	prefix           string
	upstream         *url.URL
	preserveHost     bool
	cookieOpts       CookieRewriteOptions
	stealth          bool
	sessionID        string
	reqBodyBuf       []byte
	scriptReqBodyBuf []byte
	reqFrameID       string
	startedAt        time.Time
	dnsStartNs       int64
	connectStartNs   int64
	tlsStartNs       int64
	firstResponseNs  int64
}

func (f *reverseProxyFlow) serve(w http.ResponseWriter) {
	handler, err := reverseproxy.New(reverseproxy.Options{
		Resolver: reverseproxy.ResolveTargetFunc(func(*http.Request) (*url.URL, error) {
			return cloneURL(f.upstream), nil
		}),
		RoundTripper:      newTransport(f.deps.Cfg),
		GenerateSessionID: func() string { return f.sessionID },
		WriteError:        f.writeError,
		MutateRequest:     f.mutateRequest,
		MutateResponse:    f.mutateResponse,
		PreserveHost:      f.preserveHost,
		Hooks: observe.Hooks{
			OnError:        f.onError,
			OnSessionClose: f.onSessionClose,
		},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "REVERSE_PROXY_INIT_FAILED", err.Error(), nil)
		return
	}
	handler.ServeHTTP(w, f.request)
}

func (f *reverseProxyFlow) writeError(w http.ResponseWriter, _ *http.Request, _ int, err error) {
	switch {
	case errors.Is(err, errReverseInterceptDropped):
		writeError(w, http.StatusForbidden, "INTERCEPT_DROPPED", "request dropped by interceptor", nil)
	default:
		info := classifyProxyError(err)
		status := http.StatusBadGateway
		if info.Code == "CANCELED" {
			status = 499
		}
		writeError(w, status, info.Code, info.UserMessage, map[string]any{
			"target": f.upstream.String(),
			"raw":    err.Error(),
		})
	}
}

func (f *reverseProxyFlow) mutateRequest(_ context.Context, req *http.Request) error {
	if !f.stealth {
		proxyhttp.ApplyForwardedHeaders(req.Header, proxyhttp.ForwardedHeaderConfig{
			ClientIP: clientHost(f.request.RemoteAddr),
			Proto:    forwardedProto(f.request),
			Via:      "network-debugger",
		})
	}
	rewriteOutboundCookieHeaderForUpstream(req.Header, f.cookieOpts)

	if f.deps.ScriptSvc != nil {
		scriptReq := toScriptHTTPRequest(req, f.scriptReqBodyBuf)
		sessionInfo := &scriptdomain.SessionInfo{
			ID:         f.sessionID,
			ClientAddr: f.request.RemoteAddr,
		}
		if modifiedReq, err := f.deps.ScriptSvc.ExecuteForRequest(req.Context(), scriptReq, sessionInfo); err == nil && modifiedReq != nil {
			updatedReq, updatedBody := applyScriptRequestModifications(req, modifiedReq)
			*req = *updatedReq
			f.scriptReqBodyBuf = updatedBody
		}
	}

	if f.deps.InterceptSvc != nil {
		mgr := f.deps.InterceptSvc.Manager()
		mgrCfg := mgr.Config()
		if mgrCfg.Enabled && mgrCfg.Requests {
			capBody := f.reqBodyBuf
			if max := mgrCfg.BodyMaxBytes; max > 0 && len(capBody) > max {
				capBody = capBody[:max]
			}
			origEnc := strings.ToLower(req.Header.Get("Content-Encoding"))
			decCap, _ := decodeForIntercept(capBody, origEnc, mgrCfg.BodyMaxBytes)
			ct := strings.ToLower(req.Header.Get("Content-Type"))
			input := toRequestMatchInput(req)
			input.BodyPreview = string(decCap)
			if dec, _ := mgr.InterceptRequest(req.Context(), f.sessionID, input, decCap, ct); dec != nil {
				if strings.ToLower(dec.Action) == "drop" {
					return errReverseInterceptDropped
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

	if f.deps.Cfg.ThrottleEnabled && (f.deps.Cfg.ThrottleUpKbps > 0 || f.deps.Cfg.ThrottlePacketLoss > 0) && req.Body != nil {
		bps := kbpsToBytesPerSec(f.deps.Cfg.ThrottleUpKbps)
		req.Body = io.NopCloser(wrapReaderThrottleLoss(req.Body, bps, f.deps.Cfg.ThrottlePacketLoss))
	}

	return nil
}

func (f *reverseProxyFlow) mutateResponse(_ context.Context, req *http.Request, resp *http.Response) error {
	if f.deps.Cfg.ThrottleEnabled && (f.deps.Cfg.ThrottleDownKbps > 0 || f.deps.Cfg.ThrottlePacketLoss > 0) && resp.Body != nil {
		bps := kbpsToBytesPerSec(f.deps.Cfg.ThrottleDownKbps)
		resp.Body = io.NopCloser(wrapReaderThrottleLoss(resp.Body, bps, f.deps.Cfg.ThrottlePacketLoss))
	}

	sleepResponseDelay(f.deps.Cfg)

	origCookies := append([]string(nil), resp.Header.Values("Set-Cookie")...)
	rewriteSetCookiesForProxy(resp.Header, f.cookieOpts)
	if len(resp.Header.Values("Set-Cookie")) == 0 && len(origCookies) > 0 {
		for _, c := range origCookies {
			resp.Header.Add("Set-Cookie", c)
		}
	}
	if len(resp.Header.Values("Set-Cookie")) == 0 && f.cookieOpts.HTTPS {
		resp.Header.Add("Set-Cookie", "ndebug=1; Path=/; SameSite=None")
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		if loc != "" {
			if rewritten, err := reverseproxy.RewriteRedirectLocation(loc, f.upstream, reverseproxy.RedirectRewriteOptions{
				MountPath: f.prefix,
			}); err == nil && rewritten != "" {
				resp.Header.Set("Location", rewritten)
			}
		}
	}

	if f.deps.ScriptSvc != nil && resp != nil {
		var respBodyBuf []byte
		if resp.Body != nil {
			buf, _ := io.ReadAll(resp.Body)
			respBodyBuf = buf
			resp.Body = io.NopCloser(bytes.NewReader(buf))
		}

		scriptReq := toScriptHTTPRequest(req, f.scriptReqBodyBuf)
		scriptResp := toScriptHTTPResponse(resp, respBodyBuf)
		if modifiedResp, err := f.deps.ScriptSvc.ExecuteForResponse(req.Context(), scriptReq, scriptResp, nil); err == nil && modifiedResp != nil {
			resp = applyScriptResponseModifications(resp, modifiedResp)
		}
	}

	if f.deps.InterceptSvc != nil {
		mgr := f.deps.InterceptSvc.Manager()
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
			if dec, _ := mgr.InterceptResponse(req.Context(), f.sessionID, respInput, capBuf, ct); dec != nil {
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

	basePreview := buildHTTPResponsePreview(resp)
	firstByte := timeFromUnixNanoOrZero(atomic.LoadInt64(&f.firstResponseNs))
	ttfb := durationMs(f.startedAt, firstByte)
	total := durationMs(f.startedAt, time.Now())
	preview := augmentPreviewWithTimings(basePreview, ttfb, total)
	respFrameID := id.New()
	frame := domain.Frame{
		ID:        respFrameID,
		Ts:        time.Now().UTC(),
		Direction: domain.DirectionUpstreamToClient,
		Opcode:    domain.OpcodeText,
		Size:      int(resp.ContentLength),
		Preview:   preview,
	}
	_ = f.deps.Svc.AddFrame(contextWithNoCancel(), f.sessionID, frame)
	f.deps.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: f.sessionID, Ref: frame.ID})
	f.deps.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

	dnsStart := timeFromUnixNanoOrZero(atomic.LoadInt64(&f.dnsStartNs))
	connStart := timeFromUnixNanoOrZero(atomic.LoadInt64(&f.connectStartNs))
	tlsStart := timeFromUnixNanoOrZero(atomic.LoadInt64(&f.tlsStartNs))
	firstByte = timeFromUnixNanoOrZero(atomic.LoadInt64(&f.firstResponseNs))
	tx := domain.HTTPTransaction{
		ID:        id.New(),
		SessionID: f.sessionID,
		Method:    f.request.Method,
		URL:       strings.TrimSuffix(f.upstream.String(), "?"),
		Status:    resp.StatusCode,
		ReqSize:   int(f.request.ContentLength),
		RespSize:  int(resp.ContentLength),
		StartedAt: f.startedAt,
		EndedAt:   time.Now().UTC(),
		Timings: domain.HTTPTimings{
			DNS:     durationMs(dnsStart, connStart),
			Connect: durationMs(connStart, useOrFallback(tlsStart, firstByte)),
			TLS:     durationMs(useOrFallback(tlsStart, firstByte), firstByte),
			TTFB:    durationMs(f.startedAt, firstByte),
			Total:   durationMs(f.startedAt, time.Now()),
		},
		ReqHeaders:      cloneHeader(req.Header),
		RespHeaders:     cloneHeader(resp.Header),
		Cookies:         f.request.Cookies(),
		QueryParams:     f.request.URL.Query(),
		ReqHTTPVersion:  f.request.Proto,
		RespHTTPVersion: resp.Proto,
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		tx.ContentType = ct
	}
	if f.deps.Cfg.CaptureBodies {
		if len(f.reqBodyBuf) > 0 {
			if path, err := f.deps.spoolBodyBytes(f.reqBodyBuf, "req"); err == nil && path != "" {
				tx.ReqBodyFile = path
				f.deps.Svc.AddSpoolFile(contextWithNoCancel(), f.sessionID, path)
				if f.reqFrameID != "" {
					_ = f.deps.Svc.UpdateFrameBodyFile(contextWithNoCancel(), f.sessionID, f.reqFrameID, path)
				}
			}
		}
		if resp.Body != nil {
			bodyData, readErr := io.ReadAll(io.LimitReader(resp.Body, int64(f.deps.Cfg.BodyMaxBytes)))
			if readErr == nil && len(bodyData) > 0 {
				resp.Body = io.NopCloser(bytes.NewReader(bodyData))
				if path, err := f.deps.spoolBodyBytes(bodyData, "resp"); err == nil && path != "" {
					tx.RespBodyFile = path
					f.deps.Svc.AddSpoolFile(contextWithNoCancel(), f.sessionID, path)
					_ = f.deps.Svc.UpdateFrameBodyFile(contextWithNoCancel(), f.sessionID, respFrameID, path)
				}
			}
		}
	}
	_ = f.deps.Svc.AddHTTPTransaction(contextWithNoCancel(), tx)
	f.deps.broadcastMonitorEvent(domain.MonitorEvent{Type: "http_tx_added", ID: f.sessionID, Ref: tx.ID})
	return nil
}

func (f *reverseProxyFlow) onError(_ context.Context, info observe.ErrorInfo) {
	if info.Err == nil || errors.Is(info.Err, errReverseInterceptDropped) {
		return
	}

	classified := classifyProxyError(info.Err)
	if classified.Code == "CANCELED" {
		f.deps.Logger.Info().
			Str("sessionID", f.sessionID).
			Str("target", f.upstream.String()).
			Str("method", f.request.Method).
			Str("clientAddr", clientHost(f.request.RemoteAddr)).
			Msg(classified.UserMessage)
		return
	}

	f.deps.Logger.Error().
		Err(info.Err).
		Str("sessionID", f.sessionID).
		Str("target", f.upstream.String()).
		Str("method", f.request.Method).
		Str("clientAddr", clientHost(f.request.RemoteAddr)).
		Str("errorCode", classified.Code).
		Msg(classified.UserMessage)

	f.deps.broadcastMonitorEvent(domain.MonitorEvent{
		Type: "session_error",
		ID:   f.sessionID,
		Error: &domain.ErrorDetails{
			Category:    classified.Category,
			Code:        classified.Code,
			UserMessage: classified.UserMessage,
			Message:     classified.UserMessage,
			Raw:         info.Err.Error(),
			Target:      f.upstream.String(),
			Method:      f.request.Method,
		},
	})
}

func (f *reverseProxyFlow) onSessionClose(_ context.Context, info observe.CloseInfo) {
	var errMsg *string
	switch {
	case info.Err == nil:
	case errors.Is(info.Err, errReverseInterceptDropped):
		errMsg = strPtr("dropped by interceptor")
	default:
		msg := info.Err.Error()
		errMsg = &msg
	}
	_ = f.deps.Svc.SetClosed(contextWithNoCancel(), f.sessionID, time.Now().UTC(), errMsg)
	f.deps.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: f.sessionID})
	f.deps.Metrics.ActiveSessions.Dec()
}
