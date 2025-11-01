package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestLiveSessions_RegisterUnregister(t *testing.T) {
	ls := NewLiveSessions()

	// Register with empty sessionID should be ignored
	ls.Register("", nil, nil)
	if len(ls.m) != 0 {
		t.Error("empty sessionID should be ignored")
	}

	// Register normal session
	ls.Register("session-1", nil, nil)
	if len(ls.m) != 1 {
		t.Error("session should be registered")
	}

	// Unregister with empty sessionID should be ignored
	ls.Unregister("")
	if len(ls.m) != 1 {
		t.Error("session should still be registered")
	}

	// Unregister session
	ls.Unregister("session-1")
	if len(ls.m) != 0 {
		t.Error("session should be unregistered")
	}
}

func TestLiveSessions_CloseAll(t *testing.T) {
	// Create test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Keep connection open until closed
		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Create client connections
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client1: %v", err)
	}

	client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client2: %v", err)
	}

	ls := NewLiveSessions()
	ls.Register("session-1", client1, nil)
	ls.Register("session-2", client2, nil)

	if len(ls.m) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(ls.m))
	}

	// CloseAll should close all connections and clear map
	ls.CloseAll()

	if len(ls.m) != 0 {
		t.Errorf("Expected 0 sessions after CloseAll, got %d", len(ls.m))
	}
}

func TestLiveSessions_CloseAll_WithNilConnections(t *testing.T) {
	ls := NewLiveSessions()
	ls.Register("session-1", nil, nil)
	ls.Register("session-2", nil, nil)

	// Should not panic with nil connections
	ls.CloseAll()

	if len(ls.m) != 0 {
		t.Error("All sessions should be removed")
	}
}

func TestLiveSessions_SendText_Success(t *testing.T) {
	// Create test WebSocket server for client connection
	clientServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Read one message and verify it
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
			return
		}
		if msgType != websocket.TextMessage {
			t.Errorf("Expected TextMessage, got %d", msgType)
		}
		if string(msg) != "hello client" {
			t.Errorf("Expected 'hello client', got %q", string(msg))
		}
	}))
	defer clientServer.Close()

	// Create test WebSocket server for upstream connection
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Read one message and verify it
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
			return
		}
		if msgType != websocket.TextMessage {
			t.Errorf("Expected TextMessage, got %d", msgType)
		}
		if string(msg) != "hello upstream" {
			t.Errorf("Expected 'hello upstream', got %q", string(msg))
		}
	}))
	defer upstreamServer.Close()

	// Connect to both servers
	clientWSURL := "ws" + strings.TrimPrefix(clientServer.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(clientWSURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to client server: %v", err)
	}
	defer clientConn.Close()

	upstreamWSURL := "ws" + strings.TrimPrefix(upstreamServer.URL, "http")
	upstreamConn, _, err := websocket.DefaultDialer.Dial(upstreamWSURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to upstream server: %v", err)
	}
	defer upstreamConn.Close()

	ls := NewLiveSessions()
	ls.Register("session-1", clientConn, upstreamConn)

	// Send to upstream
	err = ls.SendText("session-1", "client->upstream", "hello upstream")
	if err != nil {
		t.Errorf("SendText to upstream failed: %v", err)
	}

	// Send to client
	err = ls.SendText("session-1", "upstream->client", "hello client")
	if err != nil {
		t.Errorf("SendText to client failed: %v", err)
	}
}
