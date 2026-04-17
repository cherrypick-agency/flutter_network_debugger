package httpapi

import (
	"strings"
	"sync"
	"time"

	"network-debugger/internal/domain"

	"github.com/zishang520/socket.io/v2/socket"
)

type sessionRealtimeClient interface {
	ID() string
	Emit(event string, payload any)
	JoinSession(sessionID string)
	LeaveSession(sessionID string)
}

type sessionRoomEmitter interface {
	EmitToSessionRoom(sessionID, event string, payload any)
}

type socketRealtimeClient struct {
	socket *socket.Socket
}

func (c *socketRealtimeClient) ID() string { return string(c.socket.Id()) }

func (c *socketRealtimeClient) Emit(event string, payload any) {
	c.socket.Emit(event, payload)
}

func (c *socketRealtimeClient) JoinSession(sessionID string) {
	c.socket.Join(sessionRoomName(sessionID))
}

func (c *socketRealtimeClient) LeaveSession(sessionID string) {
	c.socket.Leave(sessionRoomName(sessionID))
}

type socketSessionRoomEmitter struct {
	io *socket.Server
}

func (e *socketSessionRoomEmitter) EmitToSessionRoom(sessionID, event string, payload any) {
	e.io.To(sessionRoomName(sessionID)).Emit(event, payload)
}

type sessionRealtimeSubscription struct {
	client  sessionRealtimeClient
	filters sioFilters
}

type sessionRealtimeService struct {
	d         *Deps
	rooms     sessionRoomEmitter
	mu        sync.RWMutex
	subs      map[string]*sessionRealtimeSubscription
	aggMu     sync.Mutex
	aggTimers map[string]*time.Timer
}

func newSessionRealtimeService(d *Deps, rooms sessionRoomEmitter) *sessionRealtimeService {
	svc := &sessionRealtimeService{
		d:         d,
		rooms:     rooms,
		subs:      make(map[string]*sessionRealtimeSubscription),
		aggTimers: make(map[string]*time.Timer),
	}
	if d.Monitor == nil {
		d.Logger.Error().Msg("[sessionRealtime] monitor is nil, realtime fanout disabled")
	}
	return svc
}

func (s *sessionRealtimeService) addSubscription(client sessionRealtimeClient, filters sioFilters) {
	s.mu.Lock()
	s.subs[client.ID()] = &sessionRealtimeSubscription{client: client, filters: filters}
	s.mu.Unlock()
}

func (s *sessionRealtimeService) removeClient(client sessionRealtimeClient) {
	clientID := client.ID()
	s.mu.Lock()
	delete(s.subs, clientID)
	s.mu.Unlock()

	s.aggMu.Lock()
	if timer, ok := s.aggTimers[clientID]; ok {
		timer.Stop()
		delete(s.aggTimers, clientID)
	}
	s.aggMu.Unlock()
}

func (s *sessionRealtimeService) sendInit(client sessionRealtimeClient, filters sioFilters) {
	ctx := contextWithNoCancel()
	listFilter := projectionListFilterFromSIO(filters)
	if len(filters.Tags) > 0 && s.d.TagsSvc != nil {
		sessionIDs, err := s.d.TagsSvc.FindSessionIDsByTags(ctx, filters.Tags)
		if err == nil {
			listFilter.SessionIDs = sessionIDs
		} else {
			s.d.Logger.Warn().Err(err).Strs("tags", filters.Tags).Msg("Failed to find sessions by tags")
		}
	}

	projector := newSessionProjector(s.d)
	views, total, _ := projector.listViews(ctx, listFilter, projectionFilterFromSIO(filters))
	groups := projector.aggregateViews(views, filters.GroupBy)

	s.d.Logger.Info().
		Str("conn_id", client.ID()).
		Int("filtered_sessions", len(views)).
		Int("total_sessions", total).
		Int("aggregates", len(groups)).
		Msg("sessionRealtime: emitting sessions:init")
	client.Emit("sessions:init", map[string]any{"items": views, "aggregate": groups})
}

func (s *sessionRealtimeService) subscribeSession(client sessionRealtimeClient, sessionID string) {
	if strings.TrimSpace(sessionID) == "" {
		return
	}
	client.JoinSession(sessionID)
	s.emitSessionCatchUp(client, sessionID)
}

func (s *sessionRealtimeService) unsubscribeSession(client sessionRealtimeClient, sessionID string) {
	if strings.TrimSpace(sessionID) == "" {
		return
	}
	client.LeaveSession(sessionID)
}

func (s *sessionRealtimeService) handleMonitorEvent(ev domain.MonitorEvent) {
	switch ev.Type {
	case "session_started", "session_ended", "frame_added", "event_added", "http_tx_added", "mapping_applied", "session_error", "sessions_cleared":
	default:
		return
	}

	s.mu.RLock()
	subs := make([]*sessionRealtimeSubscription, 0, len(s.subs))
	for _, sub := range s.subs {
		subs = append(subs, sub)
	}
	s.mu.RUnlock()

	for _, sub := range subs {
		s.applyEventToSubscription(ev.Type, ev.ID, sub)
	}
	s.emitToSessionRoom(ev.Type, ev.ID)
}

func (s *sessionRealtimeService) applyEventToSubscription(evType, sessionID string, sub *sessionRealtimeSubscription) {
	current, ok := s.currentSubscription(sub.client.ID())
	if !ok {
		return
	}
	if evType == "sessions_cleared" {
		current.client.Emit("sessions:init", map[string]any{"items": []any{}, "aggregate": []any{}})
		return
	}

	ctx := contextWithNoCancel()
	projector := newSessionProjector(s.d)
	view, ok, err := projector.viewByID(ctx, sessionID, projectionFilterFromSIO(current.filters))
	if err == nil && ok && view != nil {
		current.client.Emit("sessions:upsert", *view)
		s.scheduleAggregate(current)
		return
	}
	current.client.Emit("sessions:remove", map[string]any{"id": sessionID})
	s.scheduleAggregate(current)
}

func (s *sessionRealtimeService) scheduleAggregate(sub *sessionRealtimeSubscription) {
	clientID := sub.client.ID()

	s.aggMu.Lock()
	defer s.aggMu.Unlock()

	if timer, ok := s.aggTimers[clientID]; ok {
		timer.Reset(1 * time.Second)
		return
	}

	s.aggTimers[clientID] = time.AfterFunc(1*time.Second, func() {
		current, ok := s.currentSubscription(clientID)
		if !ok {
			return
		}
		s.emitAggregate(current)
	})
}

func (s *sessionRealtimeService) emitAggregate(sub *sessionRealtimeSubscription) {
	ctx := contextWithNoCancel()
	projector := newSessionProjector(s.d)
	views, _, _ := projector.listViews(ctx, projectionListFilterFromSIO(sub.filters), projectionFilterFromSIO(sub.filters))
	groups := projector.aggregateViews(views, sub.filters.GroupBy)
	sub.client.Emit("aggregate:update", groups)
}

func (s *sessionRealtimeService) emitToSessionRoom(evType, sessionID string) {
	if sessionID == "" {
		return
	}

	ctx := contextWithNoCancel()
	switch evType {
	case "frame_added":
		if frames, _, _ := s.d.Svc.ListFrames(ctx, sessionID, "", 1<<30); len(frames) > 0 {
			s.rooms.EmitToSessionRoom(sessionID, "session:frames", frames[len(frames)-1:])
		}
	case "event_added", "sio_probe":
		if evs, _, _ := s.d.Svc.ListEvents(ctx, sessionID, "", 1<<30); len(evs) > 0 {
			s.rooms.EmitToSessionRoom(sessionID, "session:events", evs[len(evs)-1:])
		}
	case "http_tx_added":
		if txs, _, _ := s.d.Svc.ListHTTPTransactions(ctx, sessionID, "", 1<<30); len(txs) > 0 {
			s.rooms.EmitToSessionRoom(sessionID, "session:http", txs[len(txs)-1:])
		}
	case "session_ended":
		s.rooms.EmitToSessionRoom(sessionID, "session:ended", map[string]any{"id": sessionID})
	case "session_started":
		s.rooms.EmitToSessionRoom(sessionID, "session:started", map[string]any{"id": sessionID})
	}
}

func (s *sessionRealtimeService) emitSessionCatchUp(client sessionRealtimeClient, sessionID string) {
	ctx := contextWithNoCancel()
	if frames, _, _ := s.d.Svc.ListFrames(ctx, sessionID, "", 1000); len(frames) > 0 {
		client.Emit("session:frames", frames)
	}
	if events, _, _ := s.d.Svc.ListEvents(ctx, sessionID, "", 1000); len(events) > 0 {
		client.Emit("session:events", events)
	}
	if txs, _, _ := s.d.Svc.ListHTTPTransactions(ctx, sessionID, "", 1000); len(txs) > 0 {
		client.Emit("session:http", txs)
	}
}

func (s *sessionRealtimeService) currentSubscription(clientID string) (*sessionRealtimeSubscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sub, ok := s.subs[clientID]
	return sub, ok
}

func sessionRoomName(sessionID string) socket.Room {
	return socket.Room("session:" + sessionID)
}
