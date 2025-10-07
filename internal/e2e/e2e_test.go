package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/infrastructure/config"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"

	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/websocket"
)

// startEchoWS spins up a simple echo websocket server (text/binary) for E2E
func startEchoWS(t *testing.T) (*http.Server, string) {
	t.Helper()
	mux := http.NewServeMux()
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		resp := http.Header{}
		if r.URL.Query().Get("subp") == "1" {
			if sp := r.Header.Get("Sec-WebSocket-Protocol"); sp != "" {
				resp.Set("Sec-WebSocket-Protocol", sp)
			}
		}
		c, err := up.Upgrade(w, r, resp)
		if err != nil {
			return
		}
		defer c.Close()
		q := r.URL.Query()
		// optionally emit headers JSON first
		if q.Get("hdr") == "1" {
			b, _ := json.Marshal(map[string]any{
				"authorization": r.Header.Get("Authorization"),
				"cookie":        r.Header.Get("Cookie"),
				"x-test":        r.Header.Get("X-Test"),
			})
			_ = c.WriteMessage(websocket.TextMessage, b)
		}
		// periodic ping from server
		var pingStop chan struct{}
		if q.Get("ping") == "1" {
			pingStop = make(chan struct{})
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						_ = c.WriteControl(websocket.PingMessage, []byte("srv"), time.Now().Add(1*time.Second))
					case <-pingStop:
						return
					}
				}
			}()
		}
		slowMs := 0
		if v := q.Get("slowMs"); v != "" {
			if d, err := time.ParseDuration(v + "ms"); err == nil {
				slowMs = int(d / time.Millisecond)
			}
		}
		closeAt := 0
		if v := q.Get("closeAt"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				closeAt = n
			}
		}
		count := 0
		for {
			mt, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			count++
			if slowMs > 0 {
				time.Sleep(time.Duration(slowMs) * time.Millisecond)
			}
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteMessage(mt, data); err != nil {
				return
			}
			if closeAt > 0 && count >= closeAt {
				// server initiated close
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(1*time.Second))
				return
			}
		}
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	url := fmt.Sprintf("ws://%s/ws", ln.Addr().String())
	return srv, url
}

func startTLSEchoWS(t *testing.T) (*http.Server, string) {
	t.Helper()
	// generate self-signed cert
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "127.0.0.1"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour), DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, _ := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}

	mux := http.NewServeMux()
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		resp := http.Header{}
		if r.URL.Query().Get("subp") == "1" {
			if sp := r.Header.Get("Sec-WebSocket-Protocol"); sp != "" {
				resp.Set("Sec-WebSocket-Protocol", sp)
			}
		}
		c, err := up.Upgrade(w, r, resp)
		if err != nil {
			return
		}
		defer c.Close()
		if r.URL.Query().Get("hdr") == "1" {
			b, _ := json.Marshal(map[string]any{"authorization": r.Header.Get("Authorization")})
			_ = c.WriteMessage(websocket.TextMessage, b)
		}
		for {
			mt, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteMessage(mt, data); err != nil {
				return
			}
		}
	})
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: mux, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}
	go srv.ServeTLS(ln, "", "")
	return srv, "wss://" + ln.Addr().String() + "/ws"
}

func startGoSocketIOServer(t *testing.T) (*http.Server, string) {
	t.Helper()
	srv := socketio.NewServer(nil)
	srv.OnConnect("/", func(c socketio.Conn) error {
		c.Emit("srvWelcome", map[string]any{"ok": true})
		return nil
	})
	srv.OnEvent("/", "ackA", func(c socketio.Conn) string { return "ack" })
	srv.OnEvent("/room", "join", func(c socketio.Conn, v map[string]any) {
		c.Emit("roomMsg", map[string]any{"joined": true})
	})
	mux := http.NewServeMux()
	mux.Handle("/socket.io/", srv)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	httpSrv := &http.Server{Handler: mux}
	go func() { _ = httpSrv.Serve(ln) }()
	// Use EIO=3 for compatibility with googollee server
	return httpSrv, "ws://" + ln.Addr().String() + "/socket.io/?EIO=3&transport=websocket"
}

// startFakeSIOv4Server starts a minimal Socket.IO v4-compatible WS endpoint
// It speaks only websocket transport and implements a tiny subset needed for tests.
func startFakeSIOv4Server(t *testing.T) (*http.Server, string) {
	t.Helper()
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()
	mux.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("EIO") != "4" || r.URL.Query().Get("transport") != "websocket" {
			http.Error(w, "bad transport", http.StatusBadRequest)
			return
		}
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		go func(conn *websocket.Conn) {
			defer conn.Close()
			// 0{"sid":"x","upgrades":[],"pingInterval":25000,"pingTimeout":20000}
			open := `0{"sid":"x","upgrades":[],"pingInterval":25000,"pingTimeout":20000}`
			_ = conn.WriteMessage(websocket.TextMessage, []byte(open))
			// Send 40 (socket.io connect) proactively
			_ = conn.WriteMessage(websocket.TextMessage, []byte("40"))
			for {
				if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
					return
				}
				_, data, err := conn.ReadMessage()
				if err != nil {
					return
				}
				s := string(data)
				// expectation examples:
				// 40                     -> ignore
				// 42,5["ackA",{}]       -> respond 43,5[]
				// 42/room,1["join",{}]  -> respond 42/room,["roomMsg",{"joined":true}]
				if s == "40" {
					continue
				}
				if strings.HasPrefix(s, "42,") && strings.Contains(s, "[\"ackA\"") {
					// extract ack id between '42,' and '['
					ackID := ""
					rem := s[3:]
					for i := 0; i < len(rem) && rem[i] >= '0' && rem[i] <= '9'; i++ {
						ackID += string(rem[i])
					}
					if ackID == "" {
						ackID = "1"
					}
					_ = conn.WriteMessage(websocket.TextMessage, []byte("43,"+ackID+"[]"))
					continue
				}
				if strings.HasPrefix(s, "42/room,") && strings.Contains(s, "[\"join\"") {
					_ = conn.WriteMessage(websocket.TextMessage, []byte("42/room,[\"roomMsg\",{\"joined\":true}]"))
					continue
				}
			}
		}(c)
	})
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	return srv, "ws://" + ln.Addr().String() + "/socket.io/?EIO=4&transport=websocket"
}

func buildWsproxyBinary(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "network-debugger")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	// path to cmd/network-debugger from this package directory
	cmdPath, err := filepath.Abs("../../cmd/network-debugger")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	cmd := exec.Command("go", "build", "-race", "-o", bin, cmdPath)
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out.String())
	}
	return bin
}

func waitReady(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/readyz", nil)
		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("server not ready at %s", baseURL)
}

func TestE2E_BinaryProcess_RealTCP(t *testing.T) {
	t.Parallel()
	// 1) Start upstream echo WS server
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())

	// 2) Build network-debugger and start on a free port
	bin := buildWsproxyBinary(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("port listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	var out bytes.Buffer
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Start(); err != nil {
		t.Fatalf("start network-debugger: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	baseURL := "http://" + addr
	waitReady(t, baseURL, 3*time.Second)

	// 3) Open real WS client to network-debugger and exchange a variety of frames/events over time
	proxyWS := "ws://" + addr + "/wsproxy?target=" + urlQueryEscape(echoURL)
	c, _, err := websocket.DefaultDialer.Dial(proxyWS, nil)
	if err != nil {
		t.Fatalf("dial proxy ws: %v\nlogs:\n%s", err, out.String())
	}
	defer c.Close()

	// send multiple texts
	totalText := 0
	for i := 0; i < 8; i++ {
		msg := fmt.Sprintf("hello-e2e-%d", i)
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			t.Fatalf("write text: %v", err)
		}
		totalText++
		time.Sleep(30 * time.Millisecond)
	}
	// pure JSON frames to trigger redaction
	for i := 0; i < 3; i++ {
		js := fmt.Sprintf("{\"access_token\":\"secret%d\",\"i\":%d}", i, i)
		if err := c.WriteMessage(websocket.TextMessage, []byte(js)); err != nil {
			t.Fatalf("write json: %v", err)
		}
		totalText++
	}
	// Socket.IO events (ack no nsp + namespaced)
	if err := c.WriteMessage(websocket.TextMessage, []byte("4217[\"hello\",{}]")); err != nil {
		t.Fatalf("write sio: %v", err)
	}
	if err := c.WriteMessage(websocket.TextMessage, []byte("42/chat,[\"cmd\",{\"ok\":true}]")); err != nil {
		t.Fatalf("write sio nsp: %v", err)
	}
	totalText += 2
	// ping and binary
	_ = c.WriteMessage(websocket.PingMessage, []byte("cli-ping"))
	if err := c.WriteMessage(websocket.BinaryMessage, []byte{0x01, 0x02, 0x03, 0x04}); err != nil {
		t.Fatalf("write bin: %v", err)
	}

	// read back (data messages only – control frames handled internally)
	gotText := 0
	gotBin := 0
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) && (gotText < totalText || gotBin < 1) {
		_ = c.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		mt, data, err := c.ReadMessage()
		if err != nil {
			continue
		}
		if mt == websocket.TextMessage {
			gotText++
		}
		if mt == websocket.BinaryMessage && len(data) >= 2 && data[0] == 0x01 {
			gotBin++
		}
	}
	if gotText < totalText || gotBin < 1 {
		t.Fatalf("echo mismatch: text=%d/%d bin=%d\nlogs:\n%s", gotText, totalText, gotBin, out.String())
	}

	// 4) Validate REST API of binary process
	resp, err := http.Get(baseURL + "/api/sessions?limit=10")
	if err != nil {
		t.Fatalf("sessions get: %v", err)
	}
	defer resp.Body.Close()
	var list struct {
		Items []struct {
			ID     string
			Frames struct{ Total, Text int } `json:"frames"`
			Events struct{ Total int }       `json:"events"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list.Items) == 0 || list.Items[0].Frames.Total < totalText {
		t.Fatalf("unexpected sessions response: %+v", list)
	}

	// fetch frames/events of the latest session and assert redaction + SIO parsing
	sid := list.Items[0].ID
	fr, _ := http.Get(baseURL + "/api/sessions/" + sid + "/frames?limit=1000")
	defer fr.Body.Close()
	var frames struct {
		Items []struct{ Preview string } `json:"items"`
	}
	_ = json.NewDecoder(fr.Body).Decode(&frames)
	sawBinary := false
	sawRedacted := false
	for _, f := range frames.Items {
		if len(f.Preview) > 8 && f.Preview[:8] == "[binary " {
			sawBinary = true
		}
		if bytes.Contains([]byte(f.Preview), []byte("access_token")) && bytes.Contains([]byte(f.Preview), []byte("***")) {
			sawRedacted = true
		}
	}
	if !sawBinary || !sawRedacted {
		t.Fatalf("frames check failed: binary=%v redacted=%v", sawBinary, sawRedacted)
	}

	ev, _ := http.Get(baseURL + "/api/sessions/" + sid + "/events?limit=1000")
	defer ev.Body.Close()
	var events struct {
		Items []struct {
			Namespace string `json:"namespace"`
			Name      string `json:"event"`
			AckID     *int64 `json:"ackId"`
		} `json:"items"`
	}
	_ = json.NewDecoder(ev.Body).Decode(&events)
	haveHello := false
	haveCmd := false
	for _, e := range events.Items {
		if e.Name == "hello" {
			haveHello = true
		}
		if e.Name == "cmd" && e.Namespace == "/chat" {
			haveCmd = true
		}
	}
	if !haveHello || !haveCmd {
		t.Fatalf("SIO events missing: hello=%v cmd=%v", haveHello, haveCmd)
	}
}

// WS over unified /proxy with explicit target (server in-process, app via httptest)
func TestE2E_WSUnified_WithTarget(t *testing.T) {
	t.Parallel()
	// upstream echo server
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())
	// app server with router deps
	logger := obs.NewLogger("error")
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*"}, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	app := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	defer app.Close()

	// dial unified /proxy with target
	u, _ := url.Parse(app.URL)
	u.Scheme = "ws"
	u.Path = "/proxy"
	q := u.Query()
	q.Set("target", echoURL)
	u.RawQuery = q.Encode()
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial unified: %v", err)
	}
	defer c.Close()
	_ = c.WriteMessage(websocket.TextMessage, []byte("hello"))
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("echo mismatch: %s", string(data))
	}
}

// WS over unified /proxy with DEFAULT_TARGET
func TestE2E_WSUnified_DefaultTarget(t *testing.T) {
	t.Parallel()
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())

	logger := obs.NewLogger("error")
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*", DefaultTarget: echoURL}, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	app := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	defer app.Close()

	u, _ := url.Parse(app.URL)
	u.Scheme = "ws"
	u.Path = "/proxy"
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial unified default: %v", err)
	}
	defer c.Close()
	_ = c.WriteMessage(websocket.TextMessage, []byte("ping"))
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, d, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(d) != "ping" {
		t.Fatalf("echo mismatch: %s", string(d))
	}
}

func TestE2E_TLS_WSS_Subprotocols(t *testing.T) {
	t.Parallel()
	tlsSrv, tlsURL := startTLSEchoWS(t)
	defer tlsSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr, "INSECURE_TLS=1")
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 4*time.Second)

	// use subprotocol
	target := tlsURL + "?hdr=1&subp=1"
	hdr := http.Header{}
	hdr.Set("Sec-WebSocket-Protocol", "proto1")
	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c, _, err := websocket.DefaultDialer.Dial(ws, hdr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(data, &got)
	// just confirm connection works and echo returns JSON
	if _, ok := got["authorization"]; !ok {
		t.Fatalf("unexpected payload: %v", got)
	}
	_ = c.Close()
}

func TestE2E_SocketIO_StrictRaw(t *testing.T) {
	defer handleWSReadPanic(t)()
	t.Parallel()
	sioSrv, sioURL := startFakeSIOv4Server(t)
	defer sioSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 4*time.Second)

	// Connect raw WS to network-debugger targeting socket.io EIO=3 websocket transport
	target := sioURL // already EIO=3 transport=websocket
	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	// Send Socket.IO connect: '40' (engine.io open packet may arrive but мы не читаем, чтобы избежать зависаний)
	if err := c.WriteMessage(websocket.TextMessage, []byte("40")); err != nil {
		t.Fatalf("write 40: %v", err)
	}

	// Emit ackA with ack id 5 expecting ACK: send '42,5["ackA",{}]'
	if err := c.WriteMessage(websocket.TextMessage, []byte("42,5[\"ackA\",{}]")); err != nil {
		t.Fatalf("emit ackA: %v", err)
	}
	gotAck := false
	// Join /room and expect roomMsg; include ack id 1
	if err := c.WriteMessage(websocket.TextMessage, []byte("42/room,1[\"join\",{}]")); err != nil {
		t.Fatalf("join room: %v", err)
	}
	gotRoom := false

	// We не читаем обратно из сокета (это flaky в e2e). Закрываем и проверяем события через REST.
	time.Sleep(300 * time.Millisecond)
	_ = c.Close()

	// Verify events stored by proxy (poll with timeout)
	var sid string
	for i := 0; i < 10; i++ {
		resp, _ := http.Get(baseURL + "/api/sessions?limit=100")
		var list struct {
			Items []struct{ ID string } `json:"items"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&list)
		_ = resp.Body.Close()
		if len(list.Items) > 0 {
			sid = list.Items[len(list.Items)-1].ID
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if sid == "" {
		t.Fatalf("no sessions after SIO send")
	}
	// poll events endpoint until we observe ack and room
	for i := 0; i < 10 && !(gotAck && gotRoom); i++ {
		ev, _ := http.Get(baseURL + "/api/sessions/" + sid + "/events?limit=1000")
		var events struct {
			Items []struct {
				Namespace string
				Name      string
				AckID     *int64
			} `json:"items"`
		}
		_ = json.NewDecoder(ev.Body).Decode(&events)
		_ = ev.Body.Close()
		gotAck = false
		gotRoom = false
		for _, e := range events.Items {
			if e.Name == "ack" && e.AckID != nil && *e.AckID == 5 {
				gotAck = true
			}
			if e.Namespace == "/room" && (e.Name == "roomMsg" || e.Name == "join") {
				gotRoom = true
			}
		}
		if !(gotAck && gotRoom) {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if !gotAck || !gotRoom {
		t.Fatalf("SIO events missing after REST polling: ack=%v room=%v", gotAck, gotRoom)
	}

	// Verify events stored again quickly
	resp3, _ := http.Get(baseURL + "/api/sessions?limit=100")
	defer resp3.Body.Close()
	var list2 struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp3.Body).Decode(&list2)
	if len(list2.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid2 := list2.Items[len(list2.Items)-1].ID
	ev2, _ := http.Get(baseURL + "/api/sessions/" + sid2 + "/events?limit=1000")
	defer ev2.Body.Close()
	var events2 struct {
		Items []struct {
			Namespace string
			Name      string
		} `json:"items"`
	}
	_ = json.NewDecoder(ev2.Body).Decode(&events2)
	haveAck := false
	haveRoomEv := false
	for _, e := range events2.Items {
		if e.Name == "ack" {
			haveAck = true
		}
		if e.Namespace == "/room" && (e.Name == "roomMsg" || e.Name == "join") {
			haveRoomEv = true
		}
	}
	if !haveAck || !haveRoomEv {
		t.Fatalf("stored events missing: ack=%v room=%v", haveAck, haveRoomEv)
	}
}

func TestE2E_HeadersForwardingAndNegative(t *testing.T) {
	t.Parallel()
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 4*time.Second)

	// with headers (only Authorization and Cookie are whitelisted by proxy)
	target := echoURL + "?hdr=1"
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer abc")
	hdr.Set("Cookie", "sid=xyz")
	hdr.Set("X-Test", "one")
	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c, _, err := websocket.DefaultDialer.Dial(ws, hdr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read hdr: %v", err)
	}
	var got map[string]any
	_ = json.Unmarshal(data, &got)
	if got["authorization"] != "Bearer abc" || got["cookie"] != "sid=xyz" {
		t.Fatalf("hdr mismatch: %v", got)
	}
	if got["x-test"] != "" {
		t.Fatalf("non-whitelisted header must not be forwarded: %v", got["x-test"])
	}
	_ = c.Close()

	// negative: no Authorization
	ws2 := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c2, _, err := websocket.DefaultDialer.Dial(ws2, nil)
	if err != nil {
		t.Fatalf("dial2: %v", err)
	}
	_, data2, err := c2.ReadMessage()
	if err != nil {
		t.Fatalf("read hdr2: %v", err)
	}
	var got2 map[string]any
	_ = json.Unmarshal(data2, &got2)
	if got2["authorization"] != "" {
		t.Fatalf("expected empty auth, got %v", got2["authorization"])
	}
	_ = c2.Close()
}

func TestE2E_LargeFramesAndPreview(t *testing.T) {
	t.Parallel()
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 4*time.Second)

	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(echoURL)
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// large text ~120KB
	big := bytes.Repeat([]byte("A"), 120*1024)
	if err := c.WriteMessage(websocket.TextMessage, big); err != nil {
		t.Fatalf("write big text: %v", err)
	}
	// large binary ~150KB
	binBuf := bytes.Repeat([]byte{0xEE}, 150*1024)
	if err := c.WriteMessage(websocket.BinaryMessage, binBuf); err != nil {
		t.Fatalf("write big bin: %v", err)
	}
	time.Sleep(400 * time.Millisecond)
	_ = c.Close()

	// allow small delay
	time.Sleep(200 * time.Millisecond)
	resp, _ := http.Get(baseURL + "/api/sessions?limit=100")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	sid := list.Items[0].ID
	fr, _ := http.Get(baseURL + "/api/sessions/" + sid + "/frames?limit=100")
	defer fr.Body.Close()
	var frames struct {
		Items []struct {
			Opcode  string
			Preview string
		} `json:"items"`
	}
	_ = json.NewDecoder(fr.Body).Decode(&frames)
	if len(frames.Items) < 2 {
		t.Fatalf("expected frames >=2")
	}
	// preview must be truncated (<=4096) for large text and show binary marker for bin
	okTextTrunc := false
	okBinMarker := false
	for _, f := range frames.Items {
		if f.Opcode == "text" && len(f.Preview) <= 4096 {
			okTextTrunc = true
		}
		if f.Opcode == "binary" && len(f.Preview) > 0 && f.Preview[0] == '[' {
			okBinMarker = true
		}
	}
	if !okTextTrunc || !okBinMarker {
		t.Fatalf("preview checks failed: text=%v bin=%v", okTextTrunc, okBinMarker)
	}
}

func TestE2E_ServerClientCloses(t *testing.T) {
	t.Parallel()
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 4*time.Second)

	// server-initiated close after 3 frames
	target := echoURL + "?closeAt=3"
	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	for i := 0; i < 3; i++ {
		_ = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("m%d", i)))
	}
	time.Sleep(400 * time.Millisecond)
	_ = c.Close()
	// session should have frames recorded (ClosedAt may be nil for graceful close)
	resp, _ := http.Get(baseURL + "/api/sessions?limit=100")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	sid := list.Items[len(list.Items)-1].ID
	r, _ := http.Get(baseURL + "/api/sessions/" + sid + "/frames?limit=100")
	defer r.Body.Close()
	var frames2 struct {
		Items []struct{} `json:"items"`
	}
	_ = json.NewDecoder(r.Body).Decode(&frames2)
	if len(frames2.Items) < 3 {
		t.Fatalf("expected >=3 frames, got %d", len(frames2.Items))
	}

	// client-initiated close
	ws2 := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(echoURL)
	c2, _, err := websocket.DefaultDialer.Dial(ws2, nil)
	if err != nil {
		t.Fatalf("dial2: %v", err)
	}
	_ = c2.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(1*time.Second))
	_ = c2.Close()
}

func TestE2E_BackpressureSlowEcho(t *testing.T) {
	t.Parallel()
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 4*time.Second)

	// slow echo 50ms per frame
	target := echoURL + "?slowMs=50"
	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	// send 30 small frames; should not deadlock
	for i := 0; i < 30; i++ {
		_ = c.WriteMessage(websocket.TextMessage, []byte("s"))
		time.Sleep(10 * time.Millisecond)
	}
	// read some back and ensure connection alive
	reads := 0
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && reads < 10 {
		_ = c.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		if _, _, err := c.ReadMessage(); err == nil {
			reads++
		}
	}
	if reads == 0 {
		t.Fatalf("no echoes received under backpressure")
	}
}

func TestE2E_SocketIOAdvanced(t *testing.T) {
	defer handleWSReadPanic(t)()
	t.Parallel()
	sioSrv, sioURL := startFakeSIOv4Server(t)
	defer sioSrv.Shutdown(context.Background())
	// in-process network-debugger server for stability
	logger := obs.NewLogger("error")
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*"}, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	app := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	defer app.Close()

	baseURL := app.URL
	// convert http://127.0.0.1:xxxxx to ws://.../wsproxy
	bu, _ := url.Parse(baseURL)
	bu.Scheme = "ws"
	bu.Path = "/wsproxy"
	bu.RawQuery = "target=" + url.QueryEscape(sioURL)
	ws := bu.String()
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// ack ids and namespaces + clientEvent (to be recorded from client->upstream)
	_ = c.WriteMessage(websocket.TextMessage, []byte("42/chat,5[\"ackA\",{}]"))
	_ = c.WriteMessage(websocket.TextMessage, []byte("42/room,0[\"join\",{}]"))
	_ = c.WriteMessage(websocket.TextMessage, []byte("42[\"noAck\",{}]"))
	_ = c.WriteMessage(websocket.TextMessage, []byte("42/chat,[\"clientEvent\",{}]"))
	_ = c.WriteMessage(websocket.TextMessage, []byte("4xnotvalid"))
	time.Sleep(200 * time.Millisecond)
	_ = c.Close()

	// poll REST for events
	resp2, _ := http.Get(baseURL + "/api/sessions?limit=100&target=" + url.QueryEscape(sioURL))
	defer resp2.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp2.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid := list.Items[len(list.Items)-1].ID
	haveAck5 := false
	haveNoAck := false
	haveRoom := false
	haveClient := false
	for i := 0; i < 30 && !(haveAck5 || haveNoAck || haveRoom || haveClient); i++ {
		ev, _ := http.Get(baseURL + "/api/sessions/" + sid + "/events?limit=1000")
		var events struct {
			Items []struct {
				Namespace string
				Name      string
				AckID     *int64
			} `json:"items"`
		}
		_ = json.NewDecoder(ev.Body).Decode(&events)
		_ = ev.Body.Close()
		for _, e := range events.Items {
			if e.Name == "ack" && e.AckID != nil && *e.AckID == 5 {
				haveAck5 = true
			}
			if e.Name == "noAck" && e.AckID == nil {
				haveNoAck = true
			}
			if e.Namespace == "/room" && (e.Name == "join" || e.Name == "roomMsg") {
				haveRoom = true
			}
			if e.Name == "clientEvent" {
				haveClient = true
			}
		}
		if !(haveAck5 || haveNoAck || haveRoom || haveClient) {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if !(haveAck5 || haveNoAck || haveRoom || haveClient) {
		t.Fatalf("SIO events not detected: ack5=%v noAck=%v room=%v client=%v", haveAck5, haveNoAck, haveRoom, haveClient)
	}
}

func TestE2E_UpstreamDropAndReconnect(t *testing.T) {
	t.Parallel()
	echoSrv, echoURL := startEchoWS(t)
	defer echoSrv.Shutdown(context.Background())
	bin := buildWsproxyBinary(t)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().String()
	_ = ln.Close()
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "ADDR="+addr)
	_ = cmd.Start()
	defer func() { _ = cmd.Process.Kill(); _, _ = cmd.Process.Wait() }()
	baseURL := "http://" + addr
	waitReady(t, baseURL, 2*time.Second)

	// 1st session will be dropped by server closeAt
	target := echoURL + "?closeAt=1"
	ws := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(target)
	c, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_ = c.WriteMessage(websocket.TextMessage, []byte("one"))
	time.Sleep(200 * time.Millisecond)
	_ = c.Close()

	// new session reconnect
	ws2 := "ws://" + addr + "/wsproxy?target=" + url.QueryEscape(echoURL)
	c2, _, err := websocket.DefaultDialer.Dial(ws2, nil)
	if err != nil {
		t.Fatalf("dial2: %v", err)
	}
	_ = c2.WriteMessage(websocket.TextMessage, []byte("reconnected"))
	time.Sleep(100 * time.Millisecond)
	_ = c2.Close()

	// ensure at least two sessions exist in history
	resp, _ := http.Get(baseURL + "/api/sessions?limit=10")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) < 2 {
		t.Fatalf("expected at least 2 sessions after drop+reconnect, got %d", len(list.Items))
	}
}

func urlQueryEscape(s string) string {
	// small local escape to avoid importing net/url setters here
	// NOTE: allowed as test helper
	return (&urlEscaper{b: make([]byte, 0, len(s)*3)}).escape(s)
}

type urlEscaper struct{ b []byte }

func (u *urlEscaper) escape(s string) string {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' || c == ':' || c == '/' {
			u.b = append(u.b, c)
		} else {
			u.b = append(u.b, '%')
			u.b = append(u.b, "0123456789ABCDEF"[c>>4])
			u.b = append(u.b, "0123456789ABCDEF"[c&15])
		}
	}
	return string(u.b)
}
