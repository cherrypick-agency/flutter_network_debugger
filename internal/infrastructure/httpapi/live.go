package httpapi

import (
    "errors"
    "sync"
    "github.com/gorilla/websocket"
)

// LiveSessions хранит активные WS-сессии для инжекта сообщений из API.
type LiveSessions struct {
    mu sync.RWMutex
    m  map[string]*liveWS
}

type liveWS struct {
    client   *websocket.Conn
    upstream *websocket.Conn
    // один writer в gorilla/websocket
    writeMu  sync.Mutex
}

func NewLiveSessions() *LiveSessions {
    return &LiveSessions{m: make(map[string]*liveWS)}
}

func (ls *LiveSessions) Register(sessionID string, client, upstream *websocket.Conn) {
    if sessionID == "" { return }
    ls.mu.Lock()
    ls.m[sessionID] = &liveWS{client: client, upstream: upstream}
    ls.mu.Unlock()
}

func (ls *LiveSessions) Unregister(sessionID string) {
    if sessionID == "" { return }
    ls.mu.Lock()
    delete(ls.m, sessionID)
    ls.mu.Unlock()
}

// SendText отправляет текстовый фрейм в заданном направлении.
// direction: "client->upstream" или "upstream->client".
func (ls *LiveSessions) SendText(sessionID string, direction string, payload string) error {
    ls.mu.RLock()
    w := ls.m[sessionID]
    ls.mu.RUnlock()
    if w == nil { return errors.New("session not found or closed") }
    switch direction {
    case "client->upstream":
        if w.upstream == nil { return errors.New("upstream not available") }
        w.writeMu.Lock()
        defer w.writeMu.Unlock()
        return w.upstream.WriteMessage(websocket.TextMessage, []byte(payload))
    case "upstream->client":
        if w.client == nil { return errors.New("client not available") }
        w.writeMu.Lock()
        defer w.writeMu.Unlock()
        return w.client.WriteMessage(websocket.TextMessage, []byte(payload))
    default:
        return errors.New("invalid direction")
    }
}


