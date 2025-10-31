package httpapi

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
)

// handleForwardOrNotFound routes absolute-URI and CONNECT requests as a standard forward proxy.
// Non-proxy requests fall back to 404 so that REST/WS routes are handled by other handlers.
func (d *Deps) handleForwardOrNotFound(w http.ResponseWriter, r *http.Request) {
	// Для серверных запросов в net/http абсолютный URI может прийти в RequestURI,
	// при этом r.URL.Scheme/Host могут быть пустыми. Учтём оба случая.
	if r.Method == http.MethodConnect ||
		(r.URL != nil && r.URL.Scheme != "" && r.URL.Host != "") ||
		isAbsoluteURL(r.RequestURI) {
		// Если URL ещё не разобран, но RequestURI — абсолютный, восстановим r.URL
		if r.Method != http.MethodConnect && (r.URL == nil || r.URL.Scheme == "" || r.URL.Host == "") && isAbsoluteURL(r.RequestURI) {
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
		// Если MITM включен и домен подходит — перехватываем TLS
		if d.MITM != nil && d.MITM.CA != nil && d.MITM.shouldIntercept(r.Host) {
			d.handleConnectMITM(w, r)
			return
		}
		d.handleConnectTunnel(w, r)
		return
	}
	// WebSocket upgrade over plain HTTP (ws://) via forward proxy
	// Если схема ws — считаем это WS‑апгрейдом, даже если Upgrade заголовок нестандартный.
	if r.URL != nil && r.URL.Scheme == "ws" {
		d.handleHTTPForwardWebSocket(w, r)
		return
	}
	// Либо явный Upgrade на ws поверх обычного http абсолютного URL
	if isWebSocketRequest(r) && (r.URL != nil && r.URL.Scheme == "http") {
		d.handleHTTPForwardWebSocket(w, r)
		return
	}
	// Forward regular HTTP request with absolute URI in r.URL
	d.handleHTTPForwardRequest(w, r)
}

// handleHTTPForwardWebSocket проксирует WebSocket Upgrade (ws://) в режиме forward‑proxy.
// Делает hijack клиентского соединения, устанавливает соединение к апстриму,
// передаёт исходный Upgrade‑запрос (origin‑form) и после 101 прокачивает байты в обе стороны.
func (d *Deps) handleHTTPForwardWebSocket(w http.ResponseWriter, r *http.Request) {
	// Создаём сессию (ws)
	sessionID := id.New()
	_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, Target: r.URL.String(), ClientAddr: clientHost(r.RemoteAddr), StartedAt: time.Now().UTC(), Kind: "ws"})
	d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()

	// Hijack клиента
	hj, ok := w.(http.Hijacker)
	if !ok {
		writeError(w, http.StatusInternalServerError, "HIJACK_NOT_SUPPORTED", "proxy: hijacking not supported", nil)
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr("hijack not supported"))
		d.Metrics.ActiveSessions.Dec()
		d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
		return
	}
	clientConn, clientBuf, err := hj.Hijack()
	if err != nil {
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		d.Metrics.ActiveSessions.Dec()
		d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
		return
	}

	// Dial к апстриму (ws:// => tcp:80)
	upstreamAddr := r.URL.Host
	if !strings.Contains(upstreamAddr, ":") {
		upstreamAddr = net.JoinHostPort(upstreamAddr, "80")
	}
	upstreamConn, err := net.DialTimeout("tcp", upstreamAddr, 10*time.Second)
	if err != nil {
		// Сообщаем клиенту о 502 и закрываем
		_, _ = clientBuf.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = clientBuf.Flush()
		_ = clientConn.Close()
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		d.Metrics.ActiveSessions.Dec()
		d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
		return
	}

	// Подготовим исходящий запрос к апстриму (origin-form)
	outReq := r.Clone(contextWithNoCancel())
	// Origin-form: пустой RequestURI, URL указывает на цель (scheme/host для Host заголовка)
	outURL := *r.URL
	outURL.Scheme = "http"
	outReq.URL = &outURL
	outReq.Host = outURL.Host
	outReq.RequestURI = "" // важно для корректной записи origin-form
	// Удаляем hop-by-hop, но сохраняем Upgrade/Connection
	sanitizeForUpgrade(outReq.Header)

	// Лёгкое превью запроса (кадр вниз по потоку)
	reqPreview := buildHTTPRequestPreview(outReq, nil)
	fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: 0, Preview: reqPreview}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
	d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
	d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

	// Пишем апстриму Upgrade‑запрос
	if err := outReq.Write(upstreamConn); err != nil {
		_, _ = clientBuf.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = clientBuf.Flush()
		_ = clientConn.Close()
		_ = upstreamConn.Close()
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		d.Metrics.ActiveSessions.Dec()
		d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
		return
	}

	// Читаем ответ апстрима и отдаем клиенту «как есть»
	upr := bufio.NewReader(upstreamConn)
	resp, err := http.ReadResponse(upr, outReq)
	if err != nil {
		_, _ = clientBuf.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = clientBuf.Flush()
		_ = clientConn.Close()
		_ = upstreamConn.Close()
		_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr(err.Error()))
		d.Metrics.ActiveSessions.Dec()
		d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
		return
	}
	// Отдадим клиенту статус/заголовки
	_ = resp.Write(clientBuf)
	_ = clientBuf.Flush()

	// 101 — переключаемся на двунаправленную прокачку WebSocket кадров
	if resp.StatusCode == http.StatusSwitchingProtocols {
		// Учтём возможные байты, уже буферизированные в upr (первые WS‑кадры) и у клиента
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
		// Сигнализируем в монитор о факте апгрейда (минимальное событие),
		// далее полноценные события будут приходить из pipeWSMessages при разборе SIO
		e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: "/_sys", Name: "upgraded", ArgsPreview: "[]"}
		_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
		d.Monitor.Broadcast(MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})

		go d.pipeWSMessages(sessionID, clientConn, upstreamConn, domain.DirectionClientToUpstream)
		d.pipeWSMessages(sessionID, upstreamConn, clientConn, domain.DirectionUpstreamToClient)
		return
	}

	// Не 101 — считаем завершённой попытку Upgrade
	_ = clientConn.Close()
	_ = upstreamConn.Close()
	_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
	d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
	d.Metrics.ActiveSessions.Dec()
}

// sanitizeForUpgrade удаляет hop-by-hop заголовки, оставляя Upgrade/Connection.
func sanitizeForUpgrade(h http.Header) {
	// Базовый список hop‑by‑hop без Upgrade/Connection
	hop := []string{"Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding"}
	for _, k := range hop {
		h.Del(k)
	}
	// Оставляем Upgrade и Connection как есть
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

	// minimal session for CONNECT (add a synthetic event for observability)
	sessionID := id.New()
	_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, Target: "connect://" + upstream, ClientAddr: clientHost(r.RemoteAddr), StartedAt: time.Now().UTC()})
	d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()
	// Synthetic event to signal established tunnel (useful for tests and UI)
	_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: "/_sys", Name: "tunnel_established", ArgsPreview: "{}"})
	d.Monitor.Broadcast(MonitorEvent{Type: "event_added", ID: sessionID})

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

// handleConnectMITM: устанавливает TLS с клиентом, используя leaf-сертификат от локального CA,
// и параллельно инициирует исходящее соединение к upstream (TLS). Все HTTP/1.1 запросы/ответы
// внутри TLS расшифрованы и могут быть проинструментированы аналогично reverse proxy.
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
	// Отвечаем клиенту, что туннель установлен
	_, _ = bufrw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	_ = bufrw.Flush()

	// Получаем сертификат под этот host
	leaf, err := d.MITM.CA.IssueFor(upstream)
	if err != nil {
		_ = clientConn.Close()
		return
	}
	// Настраиваем TLS сервер для клиента
	tlsSrv := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{leaf},
		NextProtos:   []string{"http/1.1"}, // упрощаем: только H1 внутри
	})
	if err := tlsSrv.Handshake(); err != nil {
		_ = tlsSrv.Close()
		return
	}

	// Dial к upstream TCP, затем TLS клиент
	upstreamTCP, err := net.DialTimeout("tcp", upstream, 10*time.Second)
	if err != nil {
		_ = tlsSrv.Close()
		return
	}
	// Клиентская сторона TLS к реальному серверу
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

	// Создаем сессию (тип http), будем логировать запросы/ответы
	sessionID := id.New()
	_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, Target: "mitm://" + upstream, ClientAddr: clientHost(r.RemoteAddr), StartedAt: time.Now().UTC(), Kind: "http"})
	d.Monitor.Broadcast(MonitorEvent{Type: "session_started", ID: sessionID})
	d.Metrics.ActiveSessions.Inc()

	// Простой цикл: читаем HTTP запросы от клиента, отправляем к апстриму, читаем ответ, отдаем назад.
	// Работает для keep-alive последовательности запросов.
	go func() {
		defer func() {
			_ = tlsCli.Close()
			_ = tlsSrv.Close()
			ctx := contextWithNoCancel()
			if sess, ok, _ := d.Svc.Get(ctx, sessionID); !ok || sess.ClosedAt == nil {
				_ = d.Svc.SetClosed(ctx, sessionID, time.Now().UTC(), nil)
				d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
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
			// Читаем запрос от клиента
			req, err := http.ReadRequest(clientBR)
			if err != nil {
				return
			}
			// Переписываем схему/хост для апстрима
			req.URL.Scheme = "https"
			req.URL.Host = upstream
			req.RequestURI = ""
			removeHopHeaders(req.Header)
			if ip := clientHost(r.RemoteAddr); ip != "" {
				req.Header.Set("X-Forwarded-For", ip)
			}
			req.Header.Set("Via", "network-debugger")

			// Для превью: аккуратно пикнем тело
			var reqBodyBuf []byte
			if req.Body != nil {
				peekSize := previewMaxBytes
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
			// Логируем запрос как в reverse proxy
			rPrev := &http.Request{Method: req.Method, URL: req.URL, Header: req.Header}
			reqPreview := buildHTTPRequestPreview(rPrev, reqBodyBuf)
			fr := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionClientToUpstream, Opcode: domain.OpcodeText, Size: int64ToInt(req.ContentLength), Preview: reqPreview}
			_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr)
			d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr.ID})
			d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionClientToUpstream), string(domain.OpcodeText)).Inc()

			// Interception: request inside MITM tunnel
			if d.Interceptor != nil && d.Cfg.InterceptEnabled && d.Cfg.InterceptRequests {
				capBody := reqBodyBuf
				if max := d.Cfg.InterceptBodyMaxBytes; max > 0 && len(capBody) > max {
					capBody = capBody[:max]
				}
				origEnc := strings.ToLower(req.Header.Get("Content-Encoding"))
				decCap, _ := decodeForIntercept(capBody, origEnc, d.Cfg.InterceptBodyMaxBytes)
				ct := strings.ToLower(req.Header.Get("Content-Type"))
				if dec, _ := d.Interceptor.InterceptRequest(contextWithNoCancel(), sessionID, req, string(decCap), decCap, ct); dec != nil {
					if strings.ToLower(dec.Action) == "drop" {
						return
					}
					if dec.Method != "" {
						req.Method = dec.Method
					}
					if dec.Headers != nil {
						req.Header = cloneHeader(dec.Headers)
					}
					if dec.Body != nil {
						bodyToWrite := dec.Body
						if d.Cfg.InterceptReencode && (origEnc == "gzip" || origEnc == "deflate") {
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

			// Отправляем запрос к апстриму
			if err := req.Write(tlsCli); err != nil {
				return
			}
			// Читаем ответ
			resp, err := http.ReadResponse(serverBR, req)
			if err != nil {
				return
			}
			// Если апгрейд (например, WebSocket) — после записи 101 переключаемся на тупой прокач байтов
			preview := buildHTTPResponsePreview(resp)
			fr2 := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: int(resp.ContentLength), Preview: preview}
			_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr2)
			d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr2.ID})
			d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

			// Ограничим скорость выгрузки апстрима клиенту
			if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleDownKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && resp.Body != nil {
				bps := kbpsToBytesPerSec(d.Cfg.ThrottleDownKbps)
				resp.Body = io.NopCloser(wrapReaderThrottleLoss(resp.Body, bps, d.Cfg.ThrottlePacketLoss))
			}
			// Отдаём ответ клиенту
			if err := resp.Write(tlsSrv); err != nil {
				return
			}

			// Interception: response inside MITM tunnel (before possible WS upgrade handling below)
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
				ct := strings.ToLower(resp.Header.Get("Content-Type"))
				if dec, _ := d.Interceptor.InterceptResponse(contextWithNoCancel(), sessionID, resp, string(capBuf), capBuf, ct); dec != nil {
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
						resp.Header.Del("Content-Encoding")
						resp.Body = io.NopCloser(bytes.NewReader(dec.Body))
						resp.ContentLength = int64(len(dec.Body))
						resp.Header.Set("Content-Length", strconv.Itoa(len(dec.Body)))
					}
				}
			}

			if resp.StatusCode == http.StatusSwitchingProtocols {
				// После 101 HTTP больше нет — переключаемся на прокачку WS с логированием кадров.
				go d.pipeWSMessages(sessionID, tlsSrv, tlsCli, domain.DirectionClientToUpstream)
				d.pipeWSMessages(sessionID, tlsCli, tlsSrv, domain.DirectionUpstreamToClient)
				return
			}
		}
	}()
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

	// Interception: request (forward)
	if d.Interceptor != nil && d.Cfg.InterceptEnabled && d.Cfg.InterceptRequests {
		capBody := reqBodyBuf
		if max := d.Cfg.InterceptBodyMaxBytes; max > 0 && len(capBody) > max {
			capBody = capBody[:max]
		}
		origEnc := strings.ToLower(outReq.Header.Get("Content-Encoding"))
		decCap, _ := decodeForIntercept(capBody, origEnc, d.Cfg.InterceptBodyMaxBytes)
		ct := strings.ToLower(outReq.Header.Get("Content-Type"))
		if dec, _ := d.Interceptor.InterceptRequest(r.Context(), sessionID, outReq, string(decCap), decCap, ct); dec != nil {
			if strings.ToLower(dec.Action) == "drop" {
				writeError(w, http.StatusForbidden, "INTERCEPT_DROPPED", "request dropped by interceptor", nil)
				_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), strPtr("dropped by interceptor"))
				d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
				d.Metrics.ActiveSessions.Dec()
				return
			}
			if dec.Method != "" {
				outReq.Method = dec.Method
			}
			if dec.URL != "" {
				if u, err := url.Parse(dec.URL); err == nil {
					if u.Scheme != "" && u.Host != "" {
						outReq.URL = u
						outReq.Host = u.Host
					} else {
						// относительный путь
						newURL := *outReq.URL
						newURL.Path = u.Path
						newURL.RawQuery = u.RawQuery
						outReq.URL = &newURL
					}
				}
			}
			if dec.Headers != nil {
				outReq.Header = cloneHeader(dec.Headers)
			}
			if dec.Body != nil {
				bodyToWrite := dec.Body
				if d.Cfg.InterceptReencode && (origEnc == "gzip" || origEnc == "deflate") {
					if encBody, ok := encodeForIntercept(dec.Body, origEnc); ok {
						bodyToWrite = encBody
						outReq.Header.Set("Content-Encoding", origEnc)
					} else {
						outReq.Header.Del("Content-Encoding")
					}
				} else {
					outReq.Header.Del("Content-Encoding")
				}
				outReq.Body = io.NopCloser(bytes.NewReader(bodyToWrite))
				outReq.ContentLength = int64(len(bodyToWrite))
				outReq.Header.Del("Transfer-Encoding")
				outReq.Header.Set("Content-Length", strconv.Itoa(len(bodyToWrite)))
			}
		}
	}

	// Apply upload throttling for forward proxy
	if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleUpKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && outReq.Body != nil {
		bps := kbpsToBytesPerSec(d.Cfg.ThrottleUpKbps)
		outReq.Body = io.NopCloser(wrapReaderThrottleLoss(outReq.Body, bps, d.Cfg.ThrottlePacketLoss))
	}

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

	// Download throttling for forward proxy
	if d.Cfg.ThrottleEnabled && (d.Cfg.ThrottleDownKbps > 0 || d.Cfg.ThrottlePacketLoss > 0) && resp.Body != nil {
		bps := kbpsToBytesPerSec(d.Cfg.ThrottleDownKbps)
		resp.Body = io.NopCloser(wrapReaderThrottleLoss(resp.Body, bps, d.Cfg.ThrottlePacketLoss))
	}

	// Build response preview and keep body intact for client
	preview := buildHTTPResponsePreview(resp)
	fr2 := domain.Frame{ID: id.New(), Ts: time.Now().UTC(), Direction: domain.DirectionUpstreamToClient, Opcode: domain.OpcodeText, Size: int(resp.ContentLength), Preview: preview}
	_ = d.Svc.AddFrame(contextWithNoCancel(), sessionID, fr2)
	d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: sessionID, Ref: fr2.ID})
	d.Metrics.FramesTotal.WithLabelValues(string(domain.DirectionUpstreamToClient), string(domain.OpcodeText)).Inc()

	// Interception: response (forward)
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

	// Optional artificial response delay
	sleepResponseDelay(d.Cfg)
	// Fallback: standard ResponseWriter path (buffer response to set Content-Length)
	bodyAll, _ := io.ReadAll(resp.Body)
	copyHeader(w.Header(), resp.Header)
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Length", strconv.Itoa(len(bodyAll)))
	w.WriteHeader(resp.StatusCode)
	if len(bodyAll) > 0 {
		_, _ = w.Write(bodyAll)
	}

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

// prependConn сначала читает подготовленные байты из r, затем читает из базового соединения.
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
			// перейти к чтению из Conn
		} else if n > 0 || err != nil {
			return n, err
		}
	}
	return p.Conn.Read(b)
}

// isAbsoluteURL returns true if s looks like an absolute URI.
func isAbsoluteURL(s string) bool {
	if u, err := url.Parse(s); err == nil {
		return u.Scheme != "" && u.Host != ""
	}
	return false
}
