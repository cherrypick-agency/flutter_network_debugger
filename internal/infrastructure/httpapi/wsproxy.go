package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	sio "network-debugger/internal/adapters/decoders/socketio"
	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
	"network-debugger/pkg/shared/redact"
)

func (d *Deps) handleWSProxy(w http.ResponseWriter, r *http.Request) {
	tgt := r.URL.Query().Get("_target")
	if tgt == "" {
		if d.Cfg.DefaultTarget != "" {
			tgt = d.Cfg.DefaultTarget
		} else {
			writeError(w, http.StatusBadRequest, "MISSING_TARGET", "missing target", nil)
			return
		}
	}
	u, err := url.Parse(tgt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TARGET", "invalid target", map[string]any{"target": tgt})
		return
	}
	// Авто-нормализация схемы: если клиент прислал http(s) — переводим в ws(s).
	// Это удобно для Socket.IO ссылок вида https://host/socket.io?EIO=4&transport=websocket
	switch u.Scheme {
	case "ws", "wss":
		// уже корректная схема
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		writeError(w, http.StatusBadRequest, "INVALID_TARGET", "invalid target", map[string]any{"target": tgt})
		return
	}

	sessionID := id.New()
	sess := domain.Session{
		ID:         sessionID,
		Target:     u.String(),
		ClientAddr: clientHost(r.RemoteAddr),
		StartedAt:  time.Now().UTC(),
		Kind:       "ws",
	}
	if err := d.Svc.Create(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", err.Error(), nil)
		return
	}
	d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()
	d.Logger.Info().Str("session", sessionID).Str("target", u.String()).Str("client", sess.ClientAddr).Msg("network-debugger: incoming WS session")

	upgrader := websocket.Upgrader{
		CheckOrigin:  func(r *http.Request) bool { return true },
		Subprotocols: []string{r.Header.Get("Sec-WebSocket-Protocol")},
	}
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorCode, errorMessage := humanizeProxyError(err)
		d.Logger.Error().Err(err).Str("errorCode", errorCode).Msg(errorMessage)

		// Broadcast error to frontend
		d.Monitor.Broadcast(MonitorEvent{
			Type: "session_error",
			ID:   sessionID,
			Error: &ErrorDetails{
				Code:    errorCode,
				Message: "WebSocket upgrade failed: " + errorMessage,
				Raw:     err.Error(),
				Target:  tgt,
				Method:  "WS",
			},
		})

		_ = d.Svc.SetClosed(r.Context(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		return
	}
	d.Logger.Info().Str("session", sessionID).Msg("network-debugger: client upgraded to WebSocket")

	// Ограничиваем время рукопожатия/диала к апстриму, чтобы не вешать клиента при недоступном апстриме
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		NetDialContext:   (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}
	if u.Scheme == "wss" && d.Cfg.InsecureTLS {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	hdr := http.Header{}
	// whitelist selected headers
	copyHeaderIfPresent(&hdr, r.Header, "Authorization")
	copyHeaderIfPresent(&hdr, r.Header, "Cookie")
	copyHeaderIfPresent(&hdr, r.Header, "Origin")
	copyHeaderIfPresent(&hdr, r.Header, "User-Agent")
	copyHeaderIfPresent(&hdr, r.Header, "Referer")
	if sp := r.Header.Get("Sec-WebSocket-Protocol"); sp != "" {
		hdr.Set("Sec-WebSocket-Protocol", sp)
	}
	// Some upstreams require Origin; if none from client, synthesize from target host
	if hdr.Get("Origin") == "" {
		origin := "http://" + u.Host
		if u.Scheme == "wss" {
			origin = "https://" + u.Host
		}
		hdr.Set("Origin", origin)
	}

	upstreamConn, resp, err := dialer.Dial(u.String(), hdr)
	if err != nil {
		// Get human-readable error message
		errorCode, errorMessage := humanizeProxyError(err)

		// Log handshake debug details when available
		if resp != nil {
			status := resp.Status
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
			d.Logger.Error().Err(err).Str("status", status).Str("errorCode", errorCode).Msg(errorMessage)
		} else {
			d.Logger.Error().Err(err).Str("errorCode", errorCode).Msg(errorMessage)
		}

		// Broadcast error to frontend
		d.Monitor.Broadcast(MonitorEvent{
			Type: "session_error",
			ID:   sessionID,
			Error: &ErrorDetails{
				Code:    errorCode,
				Message: errorMessage,
				Raw:     err.Error(),
				Target:  u.String(),
				Method:  "WS",
			},
		})

		_ = clientConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, errorMessage), time.Now().Add(2*time.Second))
		_ = clientConn.Close()
		_ = d.Svc.SetClosed(r.Context(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		return
	}
	d.Logger.Info().Str("session", sessionID).Str("upstream", u.String()).Msg("network-debugger: connected to upstream")
	// Register live session for API injection
	if d.Live != nil {
		d.Live.Register(sessionID, clientConn, upstreamConn)
	}

	// Start pumps
	go d.pipe(sessionID, clientConn, upstreamConn, domain.DirectionClientToUpstream)
	go d.pipe(sessionID, upstreamConn, clientConn, domain.DirectionUpstreamToClient)
}

func (d *Deps) pipe(sessionID string, src, dst *websocket.Conn, direction domain.Direction) {
	loggedFirst := false
	loggedFirstUpstreamText := false
	var lastErr error
	defer func() {
		_ = src.Close()
		_ = dst.Close()
		// Закрываем сессию один раз, не затирая потенциальную ошибку
		ctx := contextWithNoCancel()
		sess, ok, _ := d.Svc.Get(ctx, sessionID)
		if !ok || sess.ClosedAt == nil {
			var errPtr *string
			if lastErr != nil {
				s := lastErr.Error()
				errPtr = &s
			}
			_ = d.Svc.SetClosed(ctx, sessionID, time.Now().UTC(), errPtr)
			d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
			d.Metrics.ActiveSessions.Dec()
		}
		d.Logger.Info().Str("session", sessionID).Str("direction", string(direction)).Msg("network-debugger: stream closed")
		if d.Live != nil {
			d.Live.Unregister(sessionID)
		}
	}()
	for {
		mt, data, err := src.ReadMessage()
		if err != nil {
			lastErr = err
			return
		}
		_ = dst.SetWriteDeadline(time.Now().Add(15 * time.Second))
		if err := dst.WriteMessage(mt, data); err != nil {
			lastErr = err
			return
		}

		// log frame
		opcode := opcodeFromType(mt)
		preview := buildPreview(opcode, data)
		fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: direction, Opcode: opcode, Size: len(data), Preview: preview}
		_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
		d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
		d.Metrics.FramesTotal.WithLabelValues(string(direction), string(opcode)).Inc()

		// best-effort Socket.IO event decoding for text frames
		if !loggedFirst {
			d.Logger.Info().Str("session", sessionID).Str("direction", string(direction)).Str("opcode", string(opcode)).Int("size", len(data)).Msg("network-debugger: first frame proxied")
			loggedFirst = true
		}
		if opcode == domain.OpcodeText {
			// Parse from raw text (not from preview), to preserve SIO prefixes 42/43
			raw := strings.TrimSpace(string(data))
			if direction == domain.DirectionUpstreamToClient && !loggedFirstUpstreamText {
				// emit lightweight probe (no payload exposure)
				// payload: {dir:"upstream", prefix:"<first 6>", len:n}
				pref := raw
				if len(pref) > 6 {
					pref = pref[:6]
				}
				_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, domain.Event{
					ID: id.New(), Ts: time.Now().UTC(), Namespace: "", Name: "sio_probe", AckID: nil,
					ArgsPreview: "{\"dir\":\"upstream\",\"prefix\":\"" + pref + "\",\"len\":" + strconv.Itoa(len(raw)) + "}",
					FrameIDs:    []string{fr.ID},
				})
				d.Monitor.Broadcast(MonitorEvent{Type: "sio_probe", ID: sessionID, Ref: pref})
				loggedFirstUpstreamText = true
			}
			if nsp, ev, argsJSON, ok := sio.ParseEvent(raw); ok {
				var ack *int64
				if a := tryExtractAckID(raw); a >= 0 {
					ack = &a
				}
				e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: nsp, Name: ev, AckID: ack, ArgsPreview: argsJSON, FrameIDs: []string{fr.ID}}
				_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
				d.Monitor.Broadcast(MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
			} else {
				// Fallbacks for common forms to avoid missing events in e2e
				if strings.HasPrefix(raw, "43") {
					if a := tryExtractAckID(raw); a >= 0 {
						aa := a
						e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: "", Name: "ack", AckID: &aa, ArgsPreview: "[]", FrameIDs: []string{fr.ID}}
						_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
						d.Monitor.Broadcast(MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
					}
				} else if strings.HasPrefix(raw, "42/") || strings.HasPrefix(raw, "42[") || strings.HasPrefix(raw, "42,") || strings.HasPrefix(raw, "42") {
					if nsp, ev, argsJSON, ok2 := sio.ParseEvent(raw); ok2 {
						var ack *int64
						if a := tryExtractAckID(raw); a >= 0 {
							ack = &a
						}
						e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: nsp, Name: ev, AckID: ack, ArgsPreview: argsJSON, FrameIDs: []string{fr.ID}}
						_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
						d.Monitor.Broadcast(MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
					}
				}
			}
		}
	}
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

func buildPreview(op domain.Opcode, data []byte) string {
	if op == domain.OpcodeText {
		max := previewMaxBytes
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
	max := previewMaxBytes
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
	// optional comma directly after type (e.g., 43,5[])
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
