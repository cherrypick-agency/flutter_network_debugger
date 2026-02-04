package integration

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestForwardProxy_WSSIOEvents(t *testing.T) {
	t.Parallel()

	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, deps := startAppServer(t)
	defer appSrv.Close()

	// monitor not strictly required here; will check via REST

	// forward via http proxy
	proxyHost := ensureForwardProxyAddr(t, appSrv, deps)
	proxyURL := &url.URL{Scheme: "http", Host: proxyHost}
	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)

	// rich server emits SIO-like frames
	u := echoWS + "?server=rich"
	c, _, err := d.Dial(u, nil)
	if err != nil {
		t.Fatalf("forward dial: %v", err)
	}
	defer c.Close()
	// also send one client SIO event
	_ = c.WriteMessage(websocket.TextMessage, []byte("42/chat,[\"cli_event\",{}]"))

	// wait for frames/events to be recorded
	time.Sleep(400 * time.Millisecond)
	_ = c.Close()

	// check via REST that events appeared (filter by target host via q)
	uu, _ := url.Parse(echoWS)
	resp, err := appSrv.Client().Get(appSrv.URL + "/api/sessions?limit=1000&q=" + url.QueryEscape(uu.Host))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var list struct {
		Items []struct{ ID string } `json:"items"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	sid := list.Items[len(list.Items)-1].ID
	// Polling events: expect at least one recognized SIO event
	ok := pollUntil(2*time.Second, 50*time.Millisecond, 200*time.Millisecond, func() bool {
		r2, err := appSrv.Client().Get(appSrv.URL + "/api/sessions/" + sid + "/events?limit=1000")
		if err != nil {
			return false
		}
		var evs struct {
			Items []struct{ Name string } `json:"items"`
		}
		_ = json.NewDecoder(r2.Body).Decode(&evs)
		r2.Body.Close()
		return len(evs.Items) > 0
	})
	if !ok {
		t.Fatalf("no Socket.IO events observed via forward proxy")
	}
}
