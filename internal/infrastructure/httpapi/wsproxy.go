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

	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
	"network-debugger/pkg/shared/redact"
)

func (d *Deps) handleWSProxy(w http.ResponseWriter, r *http.Request) {
	if d.Cfg.ThrottleOffline {
		writeError(w, http.StatusServiceUnavailable, "OFFLINE", "proxy offline (simulated)", nil)
		return
	}
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

	// Объединяем query из таргета и входящего запроса (кроме служебных параметров `_target`, `_resetCapture`)
	// Параметры из входящего запроса имеют приоритет поверх параметров таргета
	targetQ := u.Query()
	incomingQ := r.URL.Query()
	incomingQ.Del("_target")
	incomingQ.Del("_resetCapture")

	// Объединяем: входящие query перезаписывают query из таргета
	for k, vv := range incomingQ {
		targetQ.Del(k)
		for _, v := range vv {
			targetQ.Add(k, v)
		}
	}
	u.RawQuery = targetQ.Encode()

	// Check if this target should be monitored (exclude monitoring endpoints)
	shouldMonitor := d.shouldMonitorTarget(u.Path)

	sessionID := id.New()
	sess := domain.Session{
		ID:         sessionID,
		Target:     u.String(),
		ClientAddr: r.RemoteAddr,
		StartedAt:  time.Now().UTC(),
		Kind:       "ws",
	}

	// Only create session in storage and broadcast if this is not a monitoring endpoint
	if shouldMonitor {
		if err := d.Svc.Create(r.Context(), sess); err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", err.Error(), nil)
			return
		}
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sessionID})
		d.Metrics.ActiveSessions.Inc()
		d.Logger.Info().Str("session", sessionID).Str("target", u.String()).Str("client", sess.ClientAddr).Msg("network-debugger: incoming WS session")
	} else {
		d.Logger.Debug().Str("session", sessionID).Str("target", u.String()).Msg("network-debugger: skipping monitoring for internal endpoint")
	}

	upgrader := websocket.Upgrader{
		CheckOrigin:  func(r *http.Request) bool { return true },
		Subprotocols: []string{r.Header.Get("Sec-WebSocket-Protocol")},
	}
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorCode, errorMessage := humanizeProxyError(err)
		d.Logger.Error().Err(err).Str("errorCode", errorCode).Msg(errorMessage)

		// Broadcast error to frontend (only if monitoring)
		if shouldMonitor {
			d.broadcastMonitorEvent(domain.MonitorEvent{
				Type: "session_error",
				ID:   sessionID,
				Error: &domain.ErrorDetails{
					Code:    errorCode,
					Message: "WebSocket upgrade failed: " + errorMessage,
					Raw:     err.Error(),
					Target:  tgt,
					Method:  "WS",
				},
			})
			_ = d.Svc.SetClosed(r.Context(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		}
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

	upstreamConn, resp, err := d.dialWithFallback(&dialer, u, hdr)
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

		// Broadcast error to frontend (only if monitoring)
		if shouldMonitor {
			d.broadcastMonitorEvent(domain.MonitorEvent{
				Type: "session_error",
				ID:   sessionID,
				Error: &domain.ErrorDetails{
					Code:    errorCode,
					Message: errorMessage,
					Raw:     err.Error(),
					Target:  u.String(),
					Method:  "WS",
				},
			})
			_ = d.Svc.SetClosed(r.Context(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		}

		_ = clientConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, errorMessage), time.Now().Add(2*time.Second))
		_ = clientConn.Close()
		return
	}
	d.Logger.Info().Str("session", sessionID).Str("upstream", u.String()).Msg("network-debugger: connected to upstream")
	// Register live session for API injection
	if d.Live != nil {
		d.Live.Register(sessionID, clientConn, upstreamConn)
	}

	// Start pumps
	go d.pipe(sessionID, clientConn, upstreamConn, domain.DirectionClientToUpstream, shouldMonitor)
	go d.pipe(sessionID, upstreamConn, clientConn, domain.DirectionUpstreamToClient, shouldMonitor)
}

// dialWithFallback attempts to connect to the upstream WebSocket server.
// If the connection fails due to a TLS handshake error (e.g., "first record does not look like a TLS handshake"),
// it automatically retries with plain ws:// instead of wss://.
func (d *Deps) dialWithFallback(dialer *websocket.Dialer, u *url.URL, hdr http.Header) (*websocket.Conn, *http.Response, error) {
	conn, resp, err := dialer.Dial(u.String(), hdr)
	if err != nil && u.Scheme == "wss" && strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
		// Log warning about TLS fallback
		d.Logger.Warn().
			Str("original_url", u.String()).
			Msg("TLS handshake failed, attempting fallback to plain ws:// (target server may not support TLS)")

		// Fallback to plain ws://
		fallbackURL := *u
		fallbackURL.Scheme = "ws"

		// Remove TLS config for plain connection
		fallbackDialer := *dialer
		fallbackDialer.TLSClientConfig = nil

		conn, resp, err = fallbackDialer.Dial(fallbackURL.String(), hdr)
		if err == nil {
			d.Logger.Info().
				Str("fallback_url", fallbackURL.String()).
				Msg("Successfully connected via plain ws:// after TLS fallback")
		}
	}
	return conn, resp, err
}

func (d *Deps) pipe(sessionID string, src, dst *websocket.Conn, direction domain.Direction, shouldMonitor bool) {
	loggedFirst := false
	loggedFirstUpstreamText := false
	var lastErr error
	defer func() {
		_ = src.Close()
		_ = dst.Close()
		// Закрываем сессию один раз, не затирая потенциальную ошибку (only if monitoring)
		if shouldMonitor {
			ctx := contextWithNoCancel()
			sess, ok, _ := d.Svc.Get(ctx, sessionID)
			if !ok || sess.ClosedAt == nil {
				var errPtr *string
				if lastErr != nil {
					s := lastErr.Error()
					errPtr = &s
				}
				_ = d.Svc.SetClosed(ctx, sessionID, time.Now().UTC(), errPtr)
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
			}
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
		// Троттлинг WS на уровне сообщений
		d.throttleSleepWS(direction, len(data))
		// Симуляция потерь: по вероятности пропускаем запись
		if d.Cfg.ThrottleEnabled && d.Cfg.ThrottlePacketLoss > 0 {
			if randInt := time.Now().UnixNano() % 100; int(randInt) < d.Cfg.ThrottlePacketLoss { // лёгкий детерминизм без глобального rand
				// пропустим запись — как будто кадр потерян
				continue
			}
		}
		_ = dst.SetWriteDeadline(time.Now().Add(15 * time.Second))
		if err := dst.WriteMessage(mt, data); err != nil {
			lastErr = err
			return
		}

		// log frame (only if monitoring)
		if shouldMonitor {
			opcode := opcodeFromType(mt)
			preview := buildPreview(opcode, data)
			fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: direction, Opcode: opcode, Size: len(data), Preview: preview}
			_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
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
					d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sio_probe", ID: sessionID, Ref: pref})
					loggedFirstUpstreamText = true
				}
				_ = d.recordSIOIfAny(sessionID, raw, fr.ID)
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
