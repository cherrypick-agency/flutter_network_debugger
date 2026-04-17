package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	pkgwsproxy "github.com/777genius/proxykit/wsproxy"
	"github.com/gorilla/websocket"

	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
	"network-debugger/pkg/shared/redact"
)

func (d *Deps) handleWSProxy(w http.ResponseWriter, r *http.Request) {
	if d.Cfg.ThrottleOffline {
		writeError(w, http.StatusServiceUnavailable, "OFFLINE", "proxy offline (simulated)", nil)
		return
	}
	loggedFirst := false
	loggedFirstUpstreamText := false
	handler, err := pkgwsproxy.New(pkgwsproxy.Options{
		Resolver: pkgwsproxy.QueryTargetResolver{
			Param:           "_target",
			DefaultTarget:   d.Cfg.DefaultTarget,
			DropParams:      []string{"_resetCapture"},
			NormalizeScheme: true,
		},
		ObserveTarget: func(target *url.URL) bool {
			return d.shouldMonitorTarget(target.Path)
		},
		GenerateSessionID:      id.New,
		HandshakeTimeout:       10 * time.Second,
		InsecureTLS:            d.Cfg.InsecureTLS,
		AllowPlaintextFallback: true,
		BeforeForward: func(direction pkgwsproxy.Direction, _ pkgwsproxy.MessageType, size int) pkgwsproxy.ForwardDecision {
			d.throttleSleepWS(domain.Direction(direction), size)
			if d.Cfg.ThrottleEnabled && d.Cfg.ThrottlePacketLoss > 0 {
				if randInt := time.Now().UnixNano() % 100; int(randInt) < d.Cfg.ThrottlePacketLoss {
					return pkgwsproxy.ForwardDecision{Drop: true}
				}
			}
			return pkgwsproxy.ForwardDecision{}
		},
		SynthesizeOrigin: true,
		WriteError: func(w http.ResponseWriter, _ *http.Request, status int, err error) {
			switch {
			case errors.Is(err, pkgwsproxy.ErrMissingTarget):
				writeError(w, status, "MISSING_TARGET", "missing target", nil)
			case errors.Is(err, pkgwsproxy.ErrInvalidTarget):
				writeError(w, status, "INVALID_TARGET", "invalid target", nil)
			default:
				writeError(w, status, "WS_PROXY_ERROR", err.Error(), nil)
			}
		},
		Hooks: pkgwsproxy.Hooks{
			OnSessionOpen: func(ctx context.Context, session pkgwsproxy.Session) error {
				sess := domain.Session{
					ID:         session.ID,
					Target:     session.Target,
					ClientAddr: session.ClientAddr,
					StartedAt:  session.StartedAt,
					Kind:       "ws",
				}
				if err := d.Svc.Create(ctx, sess); err != nil {
					return err
				}
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: session.ID})
				d.Metrics.ActiveSessions.Inc()
				d.Logger.Info().Str("session", session.ID).Str("target", session.Target).Str("client", session.ClientAddr).Msg("network-debugger: incoming WS session")
				return nil
			},
			OnSessionConnected: func(_ context.Context, session pkgwsproxy.Session, clientConn, upstreamConn *websocket.Conn) {
				d.Logger.Info().Str("session", session.ID).Str("upstream", session.Target).Msg("network-debugger: connected to upstream")
				if d.Live != nil {
					d.Live.Register(session.ID, clientConn, upstreamConn)
				}
			},
			OnFrame: func(ctx context.Context, frame pkgwsproxy.Frame) error {
				opcode := opcodeFromMessageType(frame.Type)
				payload := frame.Payload
				preview := buildPreview(opcode, payload)
				fr := domain.Frame{
					ID:        id.New(),
					Ts:        frame.At,
					Direction: domain.Direction(frame.Direction),
					Opcode:    opcode,
					Size:      len(payload),
					Preview:   preview,
				}
				_ = d.Svc.AddFrame(ctx, frame.SessionID, fr)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: frame.SessionID, Ref: fr.ID})
				d.Metrics.FramesTotal.WithLabelValues(string(frame.Direction), string(opcode)).Inc()

				if !loggedFirst {
					d.Logger.Info().Str("session", frame.SessionID).Str("direction", string(frame.Direction)).Str("opcode", string(opcode)).Int("size", len(payload)).Msg("network-debugger: first frame proxied")
					loggedFirst = true
				}
				if opcode == domain.OpcodeText {
					raw := strings.TrimSpace(string(payload))
					if frame.Direction == pkgwsproxy.DirectionUpstreamToClient && !loggedFirstUpstreamText {
						pref := raw
						if len(pref) > 6 {
							pref = pref[:6]
						}
						_ = d.Svc.AddEvent(ctx, frame.SessionID, domain.Event{
							ID:          id.New(),
							Ts:          time.Now().UTC(),
							Namespace:   "",
							Name:        "sio_probe",
							AckID:       nil,
							ArgsPreview: "{\"dir\":\"upstream\",\"prefix\":\"" + pref + "\",\"len\":" + strconv.Itoa(len(raw)) + "}",
							FrameIDs:    []string{fr.ID},
						})
						d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sio_probe", ID: frame.SessionID, Ref: pref})
						loggedFirstUpstreamText = true
					}
					_ = d.recordSIOIfAny(frame.SessionID, raw, fr.ID)
				}
				return nil
			},
			OnError: func(_ context.Context, info pkgwsproxy.ErrorInfo) {
				classified := classifyProxyError(info.Err)
				errorMessage := classified.UserMessage
				if info.Stage == "upgrade_client" {
					errorMessage = "WebSocket upgrade failed: " + errorMessage
				}
				d.Logger.Error().Err(info.Err).Str("errorCode", classified.Code).Str("stage", info.Stage).Msg(errorMessage)
				d.broadcastMonitorEvent(domain.MonitorEvent{
					Type: "session_error",
					ID:   info.SessionID,
					Error: &domain.ErrorDetails{
						Category:    classified.Category,
						Code:        classified.Code,
						UserMessage: classified.UserMessage,
						Message:     errorMessage,
						Raw:         info.Err.Error(),
						Target:      info.Target,
						Method:      "WS",
					},
				})
			},
			OnSessionClose: func(ctx context.Context, info pkgwsproxy.CloseInfo) {
				if d.Live != nil {
					d.Live.Unregister(info.SessionID)
				}
				var errPtr *string
				if info.Err != nil && !shouldSuppressWSCloseError(info.Err) {
					s := info.Err.Error()
					errPtr = &s
				}
				_ = d.Svc.SetClosed(ctx, info.SessionID, info.At, errPtr)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: info.SessionID})
				d.Metrics.ActiveSessions.Dec()
				d.Logger.Info().Str("session", info.SessionID).Msg("network-debugger: stream closed")
			},
		},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "WS_PROXY_INIT_FAILED", err.Error(), nil)
		return
	}
	handler.ServeHTTP(w, r)
}

func opcodeFromType(mt int) domain.Opcode {
	switch mt {
	case websocket.TextMessage:
		return domain.OpcodeText
	case websocket.BinaryMessage:
		return domain.OpcodeBinary
	case websocket.PingMessage:
		return domain.OpcodePing
	case websocket.PongMessage:
		return domain.OpcodePong
	case websocket.CloseMessage:
		return domain.OpcodeClose
	default:
		return domain.OpcodeBinary
	}
}

func opcodeFromMessageType(mt pkgwsproxy.MessageType) domain.Opcode {
	switch mt {
	case pkgwsproxy.MessageText:
		return domain.OpcodeText
	case pkgwsproxy.MessageBinary:
		return domain.OpcodeBinary
	case pkgwsproxy.MessagePing:
		return domain.OpcodePing
	case pkgwsproxy.MessagePong:
		return domain.OpcodePong
	case pkgwsproxy.MessageClose:
		return domain.OpcodeClose
	default:
		return domain.OpcodeBinary
	}
}

func buildPreview(op domain.Opcode, data []byte) string {
	if op == domain.OpcodeText {
		max := int(previewMaxBytes.Load())
		if max <= 0 {
			max = len(data)
		}
		if len(data) < max {
			max = len(data)
		}
		// try compact JSON
		var js any
		if json.Unmarshal(data[:max], &js) == nil {
			b, _ := json.Marshal(js)
			if max > 0 && len(b) > max {
				b = b[:max]
			}
			// redact known sensitive fields
			redacted := redact.RedactJSON(string(b))
			if max > 0 && len(redacted) > max {
				redacted = redacted[:max]
			}
			return redacted
		}
		return string(data[:max])
	}
	// Hex preview for binary
	max := int(previewMaxBytes.Load())
	if max <= 0 || max > len(data) {
		max = len(data)
	}
	if max > 256 {
		max = 256
	}
	return formatBinaryPreview(data[:max], max)
}

func copyHeaderIfPresent(dst *http.Header, src http.Header, key string) {
	if v := src.Get(key); v != "" {
		dst.Set(key, v)
	}
}

func strPtr(s string) *string { return &s }

func clientHost(remote string) string {
	if i := strings.LastIndexByte(remote, ':'); i > 0 {
		return remote[:i]
	}
	return remote
}

func shouldSuppressWSCloseError(err error) bool {
	return websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseNoStatusReceived,
		websocket.CloseGoingAway,
	)
}

// avoid context cancellation from HTTP request lifecycle for async logging
func contextWithNoCancel() context.Context { return context.Background() }

// tryExtractAckID extracts the numeric ack id between namespace and payload per SIO grammar.
// Example: 42/chat,17["message",{}] => 17
func tryExtractAckID(s string) int64 {
	// strip leading type code
	if len(s) < 2 {
		return -1
	}
	i := 2
	// Handle binary packets 45/46: skip attachments section like "<n>-"
	if s[0] == '4' && (s[1] == '5' || s[1] == '6') {
		k := i
		// read digits
		for k < len(s) && s[k] >= '0' && s[k] <= '9' {
			k++
		}
		if k < len(s) && s[k] == '-' {
			i = k + 1
		}
	}
	// optional comma directly after type or namespace (e.g., 43,5[])
	if i < len(s) && s[i] == ',' {
		i++
	}
	if i < len(s) && s[i] == '/' {
		// skip namespace until comma
		for i < len(s) && s[i] != ',' {
			i++
		}
		if i < len(s) && s[i] == ',' {
			i++
		}
	}
	// read digits until '['
	j := i
	for j < len(s) && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	if j == i {
		return -1
	}
	if j < len(s) && s[j] != '[' {
		return -1
	}
	val, err := strconv.ParseInt(s[i:j], 10, 64)
	if err != nil {
		return -1
	}
	return val
}

// handleWSSendText handles POST /api/sessions/{id}/ws/send with JSON body {direction, payload}
// Minimal MVP for injecting text frames from UI.
func (d *Deps) handleWSSendText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use POST", nil)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "ws" || parts[2] != "send" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
		return
	}
	id := parts[0]
	var req struct {
		Direction string `json:"direction"`
		Payload   string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
		return
	}
	if d.Live == nil {
		writeError(w, http.StatusServiceUnavailable, "LIVE_UNAVAILABLE", "live not ready", nil)
		return
	}
	if err := d.Live.SendText(id, req.Direction, req.Payload); err != nil {
		writeError(w, http.StatusBadRequest, "SEND_FAILED", err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
