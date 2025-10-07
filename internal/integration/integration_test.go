package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/infrastructure/config"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

type monitorEvent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Ref  string `json:"ref,omitempty"`
}

func startEchoWSServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// echo back headers into first text message upon request
		hdrDump := r.Header.Get("X-Echo-Headers") == "1" || r.URL.Query().Get("dump") == "1"
		rich := r.URL.Query().Get("server") == "rich"
		sp := r.Header.Get("Sec-WebSocket-Protocol")
		respHdr := http.Header{}
		if sp != "" {
			respHdr.Set("Sec-WebSocket-Protocol", sp)
		}
		c, err := up.Upgrade(w, r, respHdr)
		if err != nil {
			return
		}
		defer c.Close()
		var wmu sync.Mutex
		write := func(mt int, data []byte) error {
			wmu.Lock()
			defer wmu.Unlock()
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			return c.WriteMessage(mt, data)
		}
		if hdrDump {
			b, _ := json.Marshal(map[string]any{
				"authorization": r.Header.Get("Authorization"),
				"cookie":        r.Header.Get("Cookie"),
				"subprotocol":   sp,
			})
			_ = write(websocket.TextMessage, b)
		}
		if rich {
			// initial welcome JSON with sensitive key to validate redaction
			_ = write(websocket.TextMessage, []byte(`{"access_token":"srv-secret","welcome":true}`))
			// server-initiated binary
			_ = write(websocket.BinaryMessage, []byte{0x10, 0x20, 0x30})
			// periodic server-initiated SIO events + ping frames
			go func() {
				ticker := time.NewTicker(80 * time.Millisecond)
				defer ticker.Stop()
				ticks := 0
				for range ticker.C {
					ticks++
					_ = write(websocket.TextMessage, []byte("42/chat,23[\"srv_event\",{\"tick\":"+strconv.Itoa(ticks)+"}]"))
					_ = write(websocket.PingMessage, []byte("srv-ping"))
					if ticks >= 5 {
						break
					}
				}
			}()
		}
		for {
			mt, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			if err := write(mt, data); err != nil {
				return
			}
		}
	})
	srv := httptest.NewServer(mux)
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"
	u.Path = "/ws"
	return srv, u.String()
}

func startAppServer(t *testing.T) (*httptest.Server, *httpapi.Deps) {
	t.Helper()
	logger := zerolog.New(io.Discard)
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*"}, Logger: &logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	srv := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	return srv, deps
}

func wsURLFromHTTP(base string, path string) string {
	b := base
	if strings.HasPrefix(b, "http://") {
		b = "ws://" + strings.TrimPrefix(b, "http://")
	} else if strings.HasPrefix(b, "https://") {
		b = "wss://" + strings.TrimPrefix(b, "https://")
	}
	p := path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if strings.HasSuffix(b, "/") {
		b = strings.TrimRight(b, "/")
	}
	return b + p
}

func readMonitorEvents(t *testing.T, c *websocket.Conn, ctx context.Context, out chan<- monitorEvent) {
	t.Helper()
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			var ev monitorEvent
			if err := json.Unmarshal(data, &ev); err == nil {
				out <- ev
			}
		}
	}()
}

func TestWSProxy_EndToEnd_SessionFramesEventsAndAPI(t *testing.T) {
	t.Parallel()

	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// Monitor WS
	monURL := wsURLFromHTTP(appSrv.URL, "/api/monitor/ws")
	monConn, _, err := websocket.DefaultDialer.Dial(monURL, nil)
	if err != nil {
		t.Fatalf("monitor dial failed: %v", err)
	}
	defer monConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	monCh := make(chan monitorEvent, 64)
	readMonitorEvents(t, monConn, ctx, monCh)
	var framesCount, eventsCount int

	// Proxy WS
	proxyURL := wsURLFromHTTP(appSrv.URL, "/wsproxy") + "?target=" + url.QueryEscape(echoWS)
	clientConn, _, err := websocket.DefaultDialer.Dial(proxyURL, nil)
	if err != nil {
		t.Fatalf("proxy dial failed: %v", err)
	}

	// ждём session_started
	var sessionID string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case ev := <-monCh:
			if ev.Type == "frame_added" {
				framesCount++
			}
			if ev.Type == "event_added" {
				eventsCount++
			}
			if ev.Type == "session_started" {
				sessionID = ev.ID
				goto started
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
started:
	if sessionID == "" {
		t.Fatal("no session_started received")
	}

	// Отправляем простой текст
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write hello failed: %v", err)
	}

	// Отправляем два Socket.IO события: с ack id и без
	msgAck := "42/chat,17[\"ack\",{\"foo\":\"bar\"}]"
	msgNoAck := "42/chat,[\"message\",{\"token\":\"secret\"}]"
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(msgAck)); err != nil {
		t.Fatalf("write sio ack failed: %v", err)
	}
	if err := clientConn.WriteMessage(websocket.TextMessage, []byte(msgNoAck)); err != nil {
		t.Fatalf("write sio no-ack failed: %v", err)
	}

	// Небольшая пауза, чтобы upstream эхо вернул ответы и всё записалось
	time.Sleep(300 * time.Millisecond)

	_ = clientConn.Close()

	// Ждём хотя бы один session_ended
	ended := false
	endWait := time.Now().Add(3 * time.Second)
	for time.Now().Before(endWait) && !ended {
		select {
		case ev := <-monCh:
			if ev.Type == "frame_added" {
				framesCount++
			}
			if ev.Type == "event_added" {
				eventsCount++
			}
			if ev.Type == "session_ended" && ev.ID == sessionID {
				ended = true
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
	if !ended {
		t.Fatalf("no session_ended for %s", sessionID)
	}
	if framesCount == 0 {
		t.Fatalf("monitor did not receive frame_added events")
	}
	if eventsCount == 0 {
		t.Fatalf("monitor did not receive event_added events")
	}

	// REST API проверки
	httpClient := appSrv.Client()

	// /api/sessions
	resp, err := httpClient.Get(appSrv.URL + "/api/sessions?limit=10&offset=0")
	if err != nil {
		t.Fatalf("sessions list failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("sessions list status: %d", resp.StatusCode)
	}
	var list struct {
		Items []struct {
			ID     string `json:"id"`
			Target string `json:"target"`
			Frames struct {
				Total int `json:"total"`
				Text  int `json:"text"`
			} `json:"frames"`
			Events struct {
				Total int `json:"total"`
			} `json:"events"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode sessions list failed: %v", err)
	}
	if list.Total == 0 {
		t.Fatalf("sessions total == 0")
	}
	found := false
	for _, it := range list.Items {
		if it.ID == sessionID {
			found = true
			if it.Frames.Total < 3 || it.Frames.Text == 0 {
				t.Fatalf("unexpected frame counters: %+v", it.Frames)
			}
			if it.Events.Total < 2 {
				t.Fatalf("expected >=2 events, got %d", it.Events.Total)
			}
		}
	}
	if !found {
		t.Fatalf("session %s not in list", sessionID)
	}

	// /api/sessions/{id}
	resp2, err := httpClient.Get(appSrv.URL + "/api/sessions/" + sessionID)
	if err != nil {
		t.Fatalf("session get failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("session get status: %d", resp2.StatusCode)
	}

	// frames listing
	resp3, err := httpClient.Get(appSrv.URL + "/api/sessions/" + sessionID + "/frames?limit=100")
	if err != nil {
		t.Fatalf("frames list failed: %v", err)
	}
	defer resp3.Body.Close()
	var framesOut struct {
		Items []struct {
			Opcode  string `json:"opcode"`
			Preview string `json:"preview"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp3.Body).Decode(&framesOut); err != nil {
		t.Fatalf("decode frames failed: %v", err)
	}
	if len(framesOut.Items) == 0 {
		t.Fatalf("no frames returned")
	}
	sawHello := false
	sawSIO := false
	for _, f := range framesOut.Items {
		if f.Preview == "hello" {
			sawHello = true
		}
		if strings.HasPrefix(f.Preview, "42/") && strings.Contains(f.Preview, "[\"ack\"") {
			sawSIO = true
		}
	}
	if !sawHello {
		t.Fatalf("did not see 'hello' frame preview")
	}
	if !sawSIO {
		t.Fatalf("did not see Socket.IO frame preview")
	}

	// events listing
	resp4, err := httpClient.Get(appSrv.URL + "/api/sessions/" + sessionID + "/events?limit=100")
	if err != nil {
		t.Fatalf("events list failed: %v", err)
	}
	defer resp4.Body.Close()
	var eventsOut struct {
		Items []struct {
			Namespace   string `json:"namespace"`
			Name        string `json:"event"`
			AckID       *int64 `json:"ackId"`
			ArgsPreview string `json:"argsPreview"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp4.Body).Decode(&eventsOut); err != nil {
		t.Fatalf("decode events failed: %v", err)
	}
	if len(eventsOut.Items) < 2 {
		t.Fatalf("expected >=2 events, got %d", len(eventsOut.Items))
	}
	haveAck := false
	haveMessage := false
	for _, e := range eventsOut.Items {
		if e.Name == "ack" && e.Namespace == "/chat" && e.AckID != nil && *e.AckID == 17 {
			haveAck = true
		}
		if e.Name == "message" && e.Namespace == "/chat" {
			haveMessage = true
		}
	}
	if !haveAck || !haveMessage {
		t.Fatalf("missing expected events: ack=%v message=%v", haveAck, haveMessage)
	}
}

func TestHeadersAndSubprotocolForwarding(t *testing.T) {
	t.Parallel()

	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	u := wsURLFromHTTP(appSrv.URL, "/wsproxy") + "?target=" + url.QueryEscape(echoWS+"?dump=1")
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer abc")
	hdr.Set("Cookie", "sid=xyz")
	hdr.Set("Sec-WebSocket-Protocol", "chat.v1")
	// X-Echo-Headers is not forwarded by proxy; we rely on target query dump=1
	c, _, err := websocket.DefaultDialer.Dial(u, hdr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer c.Close()
	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if got["authorization"] != "Bearer abc" || got["cookie"] != "sid=xyz" || got["subprotocol"] != "chat.v1" {
		t.Fatalf("headers not forwarded: %v", got)
	}
}

func TestBinaryCounters(t *testing.T) {
	t.Parallel()

	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	u := wsURLFromHTTP(appSrv.URL, "/wsproxy") + "?target=" + url.QueryEscape(echoWS)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// send binary
	if err := c.WriteMessage(websocket.BinaryMessage, []byte{0x01, 0x02}); err != nil {
		t.Fatalf("bin: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	_ = c.Close()

	// list sessions and assert counters
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct {
			ID     string
			Frames struct{ Total, Binary, Control int }
		} `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	ok := false
	for _, it := range list.Items {
		if it.Frames.Total >= 1 && it.Frames.Binary >= 1 {
			ok = true
			break
		}
	}
	if !ok {
		t.Fatalf("counters not updated: %+v", list.Items)
	}
}

func TestPaginationFramesAndEvents(t *testing.T) {
	t.Parallel()

	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// create session with many frames/events
	u := wsURLFromHTTP(appSrv.URL, "/wsproxy") + "?target=" + url.QueryEscape(echoWS)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	for i := 0; i < 5; i++ {
		payload := "42/chat,[\"event\",{\"i\":" + jsonInt(i) + "}]"
		_ = c.WriteMessage(websocket.TextMessage, []byte(payload))
	}
	time.Sleep(200 * time.Millisecond)
	_ = c.Close()

	// get last session id
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid := list.Items[len(list.Items)-1].ID

	// frames pagination
	var next string
	total := 0
	for {
		url := appSrv.URL + "/api/sessions/" + sid + "/frames?limit=2"
		if next != "" {
			url += "&from=" + next
		}
		r, _ := appSrv.Client().Get(url)
		var page struct {
			Items []struct{ ID string } `json:"items"`
			Next  string                `json:"next"`
		}
		_ = json.NewDecoder(r.Body).Decode(&page)
		r.Body.Close()
		total += len(page.Items)
		if page.Next == "" {
			break
		}
		next = page.Next
	}
	if total < 5 {
		t.Fatalf("pagination frames total=%d", total)
	}

	// events pagination
	next = ""
	total = 0
	for {
		url := appSrv.URL + "/api/sessions/" + sid + "/events?limit=2"
		if next != "" {
			url += "&from=" + next
		}
		r, _ := appSrv.Client().Get(url)
		var page struct {
			Items []struct{ ID string } `json:"items"`
			Next  string                `json:"next"`
		}
		_ = json.NewDecoder(r.Body).Decode(&page)
		r.Body.Close()
		total += len(page.Items)
		if page.Next == "" {
			break
		}
		next = page.Next
	}
	if total < 5 {
		t.Fatalf("pagination events total=%d", total)
	}
}

func jsonInt(i int) string { return strconv.Itoa(i) }

func TestAPI_VersionAndCORS(t *testing.T) {
	t.Parallel()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// /api/version
	resp, err := appSrv.Client().Get(appSrv.URL + "/api/version")
	if err != nil {
		t.Fatalf("version get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	var ver map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&ver); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if ver["name"] != "network-debugger" {
		t.Fatalf("unexpected name: %v", ver["name"])
	}

	// CORS preflight
	req, _ := http.NewRequest(http.MethodOptions, appSrv.URL+"/api/sessions", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp2, err := appSrv.Client().Do(req)
	if err != nil {
		t.Fatalf("preflight failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNoContent {
		t.Fatalf("preflight status: %d", resp2.StatusCode)
	}
	if ao := resp2.Header.Get("Access-Control-Allow-Origin"); ao == "" {
		t.Fatalf("no Access-Control-Allow-Origin header")
	}
	if ah := resp2.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(ah, "Sec-WebSocket-Protocol") {
		t.Fatalf("allow headers missing Sec-WebSocket-Protocol: %s", ah)
	}
}

func TestMonitor_MultipleClients(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// two monitor clients
	mon1, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/api/monitor/ws"), nil)
	if err != nil {
		t.Fatalf("mon1 dial: %v", err)
	}
	defer mon1.Close()
	mon2, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/api/monitor/ws"), nil)
	if err != nil {
		t.Fatalf("mon2 dial: %v", err)
	}
	defer mon2.Close()

	// create session
	c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	_ = c.WriteMessage(websocket.TextMessage, []byte("hi"))
	_ = c.Close()

	// both should receive at least one message
	got1 := false
	got2 := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !(got1 && got2) {
		mon1.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		if _, _, err := mon1.ReadMessage(); err == nil {
			got1 = true
		}
		mon2.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		if _, _, err := mon2.ReadMessage(); err == nil {
			got2 = true
		}
	}
	if !got1 || !got2 {
		t.Fatalf("both monitors must receive events: m1=%v m2=%v", got1, got2)
	}
}

func TestSocketIO_AckWithoutNamespace(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// ack id 21, no namespace
	_ = c.WriteMessage(websocket.TextMessage, []byte("421[\"hello\",{}]"))
	time.Sleep(200 * time.Millisecond)
	_ = c.Close()

	// last session
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid := list.Items[len(list.Items)-1].ID
	r, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/events?limit=100")
	defer r.Body.Close()
	var evs struct {
		Items []struct {
			Namespace string `json:"namespace"`
			Name      string `json:"event"`
			AckID     *int64 `json:"ackId"`
		} `json:"items"`
	}
	_ = json.NewDecoder(r.Body).Decode(&evs)
	ok := false
	for _, e := range evs.Items {
		if e.Name == "hello" && e.Namespace == "" {
			ok = true
			break
		}
	}
	if !ok {
		t.Fatalf("expected hello event without namespace")
	}
}

func TestSessionsFilterByQuery(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// create a session
	c, _, _ := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
	_ = c.WriteMessage(websocket.TextMessage, []byte("ok"))
	_ = c.Close()
	time.Sleep(100 * time.Millisecond)

	// query by part of target host
	u, _ := url.Parse(echoWS)
	host := u.Host
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000&q=" + url.QueryEscape(host[:3]))
	defer resp.Body.Close()
	var list struct {
		Items []struct{ Target string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	found := false
	for _, it := range list.Items {
		if strings.Contains(it.Target, host) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("q filter did not match expected target")
	}
}

func TestCORS_Methods(t *testing.T) {
	t.Parallel()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()
	// OPTIONS already checked; ensure headers cover Authorization/Cookie as allowed
	req, _ := http.NewRequest(http.MethodOptions, appSrv.URL+"/api/sessions", nil)
	req.Header.Set("Origin", "http://localhost")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := appSrv.Client().Do(req)
	if err != nil {
		t.Fatalf("options: %v", err)
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatalf("Authorization not allowed in CORS headers")
	}
	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Headers"), "Cookie") {
		t.Fatalf("Cookie not allowed in CORS headers")
	}
}

// Sustained realtime connection test: keeps a network-debugger session open and exchanges frames over time
func TestSustainedRealtimeSession(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// connect monitor to observe lifecycle and frames/events
	mon, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/api/monitor/ws"), nil)
	if err != nil {
		t.Fatalf("monitor dial: %v", err)
	}
	defer mon.Close()

	// start proxy session
	c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
	if err != nil {
		t.Fatalf("proxy dial: %v", err)
	}
	defer c.Close()

	// send a sequence of frames over time, including SIO event
	totalText := 0
	for i := 0; i < 5; i++ {
		msg := "hello-" + strconv.Itoa(i)
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			t.Fatalf("write: %v", err)
		}
		totalText++
		time.Sleep(50 * time.Millisecond)
	}
	// one Socket.IO payload
	_ = c.WriteMessage(websocket.TextMessage, []byte("42/chat,[\"probe\",{\"n\":1}]"))
	totalText++

	// read back from echo to ensure the upstream path works
	readBack := 0
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && readBack < totalText {
		_ = c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if _, _, err := c.ReadMessage(); err == nil {
			readBack++
		}
	}
	if readBack < totalText {
		t.Fatalf("expected to read back %d frames, got %d", totalText, readBack)
	}

	// close client to end session
	_ = c.Close()

	// fetch latest session and validate counters and existence of at least one event
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct {
			ID     string
			Frames struct{ Total, Text int } `json:"frames"`
			Events struct{ Total int }       `json:"events"`
		} `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions found")
	}
	// pick the last inserted
	sess := list.Items[len(list.Items)-1]
	if sess.Frames.Total < totalText || sess.Frames.Text < totalText {
		t.Fatalf("unexpected frame counters: %+v need >=%d text", sess.Frames, totalText)
	}
	if sess.Events.Total < 1 {
		t.Fatalf("expected at least one parsed SIO event")
	}

	// monitor should have received lifecycle events
	hasStarted := false
	hasEnded := false
	monDeadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(monDeadline) && !(hasStarted && hasEnded) {
		mon.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		if _, data, err := mon.ReadMessage(); err == nil {
			var ev monitorEvent
			_ = json.Unmarshal(data, &ev)
			if ev.Type == "session_started" {
				hasStarted = true
			}
			if ev.Type == "session_ended" {
				hasEnded = true
			}
		}
	}
	if !hasStarted || !hasEnded {
		t.Fatalf("monitor did not observe full lifecycle: started=%v ended=%v", hasStarted, hasEnded)
	}
}

// High-load concurrent test: multiple sessions, validate monitor counts and redaction
func TestHighLoadConcurrentSessions(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// monitor
	mon, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/api/monitor/ws"), nil)
	if err != nil {
		t.Fatalf("monitor dial: %v", err)
	}
	defer mon.Close()
	var monFrames int32
	var monEvents int32
	monCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() {
		for {
			select {
			case <-monCtx.Done():
				return
			default:
			}
			_ = mon.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			if _, data, err := mon.ReadMessage(); err == nil {
				var ev monitorEvent
				_ = json.Unmarshal(data, &ev)
				if ev.Type == "frame_added" {
					atomic.AddInt32(&monFrames, 1)
				}
				if ev.Type == "event_added" {
					atomic.AddInt32(&monEvents, 1)
				}
			}
		}
	}()

	// spawn multiple sessions
	sessions := 3
	var wg sync.WaitGroup
	wg.Add(sessions)
	for s := 0; s < sessions; s++ {
		go func(idx int) {
			defer wg.Done()
			c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
			if err != nil {
				t.Errorf("dial: %v", err)
				return
			}
			defer c.Close()
			// send several text JSON with sensitive key, one binary, one SIO event
			for i := 0; i < 5; i++ {
				payload := `{"access_token":"secret` + strconv.Itoa(idx) + `","i":` + strconv.Itoa(i) + `}`
				_ = c.WriteMessage(websocket.TextMessage, []byte(payload))
			}
			_ = c.WriteMessage(websocket.BinaryMessage, []byte{0xAA, 0xBB, 0xCC})
			_ = c.WriteMessage(websocket.TextMessage, []byte("42/admin,17[\"cmd\",{}]"))
			time.Sleep(100 * time.Millisecond)
		}(s)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	cancel()

	// list sessions filtered by target
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000&target=" + url.QueryEscape(echoWS))
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions returned for target")
	}

	// pick last session and validate redaction + events present
	sid := list.Items[len(list.Items)-1].ID
	rf, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/frames?limit=200")
	defer rf.Body.Close()
	var frames struct {
		Items []struct{ Preview string } `json:"items"`
	}
	_ = json.NewDecoder(rf.Body).Decode(&frames)
	redacted := false
	for _, f := range frames.Items {
		if strings.Contains(f.Preview, "access_token") && strings.Contains(f.Preview, "***") {
			redacted = true
			break
		}
	}
	if !redacted {
		t.Fatalf("expected redacted access_token in preview")
	}

	re, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/events?limit=200")
	defer re.Body.Close()
	var events struct {
		Items []struct {
			Namespace string
			Name      string
		} `json:"items"`
	}
	_ = json.NewDecoder(re.Body).Decode(&events)
	hasAdmin := false
	for _, e := range events.Items {
		if e.Namespace == "/admin" || e.Name == "cmd" {
			hasAdmin = true
			break
		}
	}
	if !hasAdmin {
		t.Fatalf("expected admin/cmd event parsed")
	}

	if atomic.LoadInt32(&monFrames) == 0 || atomic.LoadInt32(&monEvents) == 0 {
		t.Fatalf("monitor counters must be > 0; frames=%d events=%d", monFrames, monEvents)
	}
}

func TestRichServerManyEvents(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// monitor to observe event_added/frame_added
	mon, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/api/monitor/ws"), nil)
	if err != nil {
		t.Fatalf("monitor dial: %v", err)
	}
	defer mon.Close()
	hasEvent := false
	hasFrame := false
	go func() {
		deadline := time.Now().Add(1 * time.Second)
		for time.Now().Before(deadline) {
			_ = mon.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			_, data, err := mon.ReadMessage()
			if err != nil {
				break
			}
			var ev monitorEvent
			_ = json.Unmarshal(data, &ev)
			if ev.Type == "event_added" {
				hasEvent = true
			}
			if ev.Type == "frame_added" {
				hasFrame = true
			}
		}
	}()

	// connect to rich server (server-initiated events/ping/binary)
	richTarget := echoWS + "?server=rich"
	c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(richTarget), nil)
	if err != nil {
		t.Fatalf("proxy dial: %v", err)
	}

	// send a couple of client frames too
	_ = c.WriteMessage(websocket.TextMessage, []byte("client-hello"))
	_ = c.WriteMessage(websocket.TextMessage, []byte("42/chat,[\"cli_event\",{}]"))
	time.Sleep(600 * time.Millisecond)
	_ = c.Close()

	// validate via REST
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid := list.Items[len(list.Items)-1].ID

	rf, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/frames?limit=1000")
	defer rf.Body.Close()
	var frames struct {
		Items []struct {
			Opcode  string `json:"opcode"`
			Preview string `json:"preview"`
		} `json:"items"`
	}
	_ = json.NewDecoder(rf.Body).Decode(&frames)
	hasBinary := false
	hasRedacted := false
	hasSrvEvent := false
	for _, f := range frames.Items {
		if strings.HasPrefix(f.Preview, "[binary ") {
			hasBinary = true
		}
		if strings.Contains(f.Preview, "access_token") && strings.Contains(f.Preview, "***") {
			hasRedacted = true
		}
		if strings.Contains(f.Preview, "\"srv_event\"") {
			hasSrvEvent = true
		}
	}
	if !hasBinary {
		t.Fatalf("expected at least one binary frame from server")
	}
	if !hasRedacted {
		t.Fatalf("expected redacted sensitive field in server welcome json")
	}
	if !hasSrvEvent {
		t.Fatalf("expected srv_event frame preview")
	}

	re, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/events?limit=1000")
	defer re.Body.Close()
	var evs struct {
		Items []struct {
			Namespace string `json:"namespace"`
			Name      string `json:"event"`
			AckID     *int64 `json:"ackId"`
		} `json:"items"`
	}
	_ = json.NewDecoder(re.Body).Decode(&evs)
	foundSrv := false
	for _, e := range evs.Items {
		if e.Namespace == "/chat" && e.Name == "srv_event" {
			foundSrv = true
			break
		}
	}
	if !foundSrv {
		t.Fatalf("expected parsed srv_event in events list")
	}

	if !(hasEvent && hasFrame) {
		t.Fatalf("monitor did not signal event/frame: event=%v frame=%v", hasEvent, hasFrame)
	}
}

func TestDeleteSessionAnd404(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// create session
	c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_ = c.WriteMessage(websocket.TextMessage, []byte("hi"))
	time.Sleep(100 * time.Millisecond)
	_ = c.Close()

	// find session id
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid := list.Items[len(list.Items)-1].ID

	// delete
	req, _ := http.NewRequest(http.MethodDelete, appSrv.URL+"/api/sessions/"+sid, nil)
	resp2, err := appSrv.Client().Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status: %d", resp2.StatusCode)
	}

	// get should 404
	resp3, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp3.StatusCode)
	}
}

func TestUpstreamDialFailureSetsClosedError(t *testing.T) {
	t.Parallel()
	appSrv, deps := startAppServer(t)
	defer appSrv.Close()

	badTarget := "ws://127.0.0.1:9/ws" // порт discard, не слушает
	// create session attempt
	_, _, _ = websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(badTarget), nil)
	// подождём обработку
	time.Sleep(200 * time.Millisecond)

	// проверим, что есть сессия с ошибкой/закрытием
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp.Body.Close()
	var list struct {
		Items []struct {
			ID       string
			Error    *string    `json:"error"`
			ClosedAt *time.Time `json:"closedAt"`
		} `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	// допускаем, что могла создаться и закрыться
	hasErr := false
	for _, it := range list.Items {
		if it.Error != nil || it.ClosedAt != nil {
			hasErr = true
			break
		}
	}
	if !hasErr {
		t.Fatalf("expected at least one closed/error session")
	}
	_ = deps // silence linter unused in future extensions
}

func TestListFiltersAndRedaction(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()
	appSrv, _ := startAppServer(t)
	defer appSrv.Close()

	// create a few sessions with different targets and payload containing token
	for i := 0; i < 2; i++ {
		c, _, err := websocket.DefaultDialer.Dial(wsURLFromHTTP(appSrv.URL, "/wsproxy")+"?target="+url.QueryEscape(echoWS), nil)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		// отправим SIO-event и чистый JSON, чтобы preview прошёл через JSON компактор и редактирование
		_ = c.WriteMessage(websocket.TextMessage, []byte("42/chat,[\"payload\",{\"access_token\":\"secret\"}]"))
		_ = c.WriteMessage(websocket.TextMessage, []byte("{\"access_token\":\"secret\",\"x\":1}"))
		time.Sleep(50 * time.Millisecond)
		_ = c.Close()
	}

	// list with target filter
	resp, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000&target=" + url.QueryEscape(echoWS))
	defer resp.Body.Close()
	var list struct {
		Items []struct{ Target string } `json:"items"`
		Total int                       `json:"total"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if list.Total == 0 {
		t.Fatalf("filter by target returned 0")
	}

	// frames and redaction check
	// pick latest session
	resp2, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000")
	defer resp2.Body.Close()
	var list2 struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp2.Body).Decode(&list2)
	sid := list2.Items[len(list2.Items)-1].ID
	r3, _ := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/frames?limit=100")
	defer r3.Body.Close()
	var frames struct {
		Items []struct{ Preview string } `json:"items"`
	}
	_ = json.NewDecoder(r3.Body).Decode(&frames)
	redacted := false
	for _, f := range frames.Items {
		if strings.Contains(f.Preview, "access_token") && strings.Contains(f.Preview, "***") {
			redacted = true
			break
		}
	}
	if !redacted {
		t.Fatalf("sensitive fields not redacted in preview")
	}
}
