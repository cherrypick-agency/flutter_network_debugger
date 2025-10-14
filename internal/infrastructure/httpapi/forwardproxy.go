package httpapi

import (
	"bufio"
	"bytes"
	"crypto/tls"
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
		// Если MITM включен и домен подходит — перехватываем TLS
		if d.MITM != nil && d.MITM.CA != nil && d.MITM.shouldIntercept(r.Host) {
			d.handleConnectMITM(w, r)
			return
		}
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
			_ = d.Svc.SetClosed(contextWithNoCancel(), sessionID, time.Now().UTC(), nil)
			d.Monitor.Broadcast(MonitorEvent{Type: "session_ended", ID: sessionID})
			d.Metrics.ActiveSessions.Dec()
		}()

		clientBR := bufio.NewReader(tlsSrv)
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

			// Отдаём ответ клиенту
			if err := resp.Write(tlsSrv); err != nil {
				return
			}

			if resp.StatusCode == http.StatusSwitchingProtocols {
				// После 101 HTTP больше нет — просто копируем байты в обе стороны до закрытия.
				go func() { _, _ = io.Copy(tlsCli, tlsSrv); _ = tlsCli.Close() }()
				_, _ = io.Copy(tlsSrv, tlsCli)
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
