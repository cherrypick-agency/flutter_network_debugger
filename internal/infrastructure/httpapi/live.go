package httpapi

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

// LiveSessions stores active WS sessions for message injection from API.
type LiveSessions struct {
	mu sync.RWMutex
	m  map[string]*liveWS
}

type liveWS struct {
	client   *websocket.Conn
	upstream *websocket.Conn
	// single writer in gorilla/websocket
	writeMu sync.Mutex
}

func NewLiveSessions() *LiveSessions {
	return &LiveSessions{m: make(map[string]*liveWS)}
}

func (ls *LiveSessions) Register(sessionID string, client, upstream *websocket.Conn) {
	if sessionID == "" {
		return
	}
	ls.mu.Lock()
	ls.m[sessionID] = &liveWS{client: client, upstream: upstream}
	ls.mu.Unlock()
}

func (ls *LiveSessions) Unregister(sessionID string) {
	if sessionID == "" {
		return
	}
	ls.mu.Lock()
	delete(ls.m, sessionID)
	ls.mu.Unlock()
}

// CloseAll closes all active WS sessions (client and upstream), clearing the map.
func (ls *LiveSessions) CloseAll() {
	ls.mu.Lock()
	for id, w := range ls.m {
		if w.client != nil {
			_ = w.client.Close()
		}
		if w.upstream != nil {
			_ = w.upstream.Close()
		}
		delete(ls.m, id)
	}
	ls.mu.Unlock()
}

// SendText sends a text frame in the specified direction.
// direction: "client->upstream" or "upstream->client".
func (ls *LiveSessions) SendText(sessionID string, direction string, payload string) error {
	ls.mu.RLock()
	w := ls.m[sessionID]
	ls.mu.RUnlock()
	if w == nil {
		return errors.New("session not found or closed")
	}
	switch direction {
	case "client->upstream":
		if w.upstream == nil {
			return errors.New("upstream not available")
		}
		w.writeMu.Lock()
		defer w.writeMu.Unlock()
		return w.upstream.WriteMessage(websocket.TextMessage, []byte(payload))
	case "upstream->client":
		if w.client == nil {
			return errors.New("client not available")
		}
		w.writeMu.Lock()
		defer w.writeMu.Unlock()
		return w.client.WriteMessage(websocket.TextMessage, []byte(payload))
	default:
		return errors.New("invalid direction")
	}
}
