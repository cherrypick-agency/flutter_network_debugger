package httpapi

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"network-debugger/internal/usecase"

	"github.com/zishang520/socket.io/v2/socket"
)

type sioFilters struct {
	Q                 string
	Target            string
	Types             []string
	StatusGroups      []string
	Tags              []string // filter by tags
	CaptureScope      string   // "current" | "all" | ""
	CaptureIDExplicit *int
	IncludeUnassigned bool
	Limit             int
	GroupBy           string
}

type sioSubscription struct {
	socket  *socket.Socket
	filters sioFilters
}

type sioHub struct {
	d         *Deps
	io        *socket.Server
	mu        sync.RWMutex
	subs      map[*socket.Socket]*sioSubscription
	aggMu     sync.Mutex
	aggTimers map[*socket.Socket]*time.Timer
}

func newSioHub(d *Deps, io *socket.Server) *sioHub {
	h := &sioHub{d: d, io: io, subs: make(map[*socket.Socket]*sioSubscription), aggTimers: make(map[*socket.Socket]*time.Timer)}
	// Subscribe to internal event bus
	if d.Monitor == nil {
		d.Logger.Error().Msg("[sioHub] ERROR: d.Monitor is NIL! Cannot subscribe to events!")
	} else {
		d.Logger.Info().Msg("[sioHub] d.Monitor exists, starting runBroadcast() goroutine")
	}
	go h.runBroadcast()
	return h
}

func (h *sioHub) addSub(s *socket.Socket, f sioFilters) {
	h.mu.Lock()
	h.subs[s] = &sioSubscription{socket: s, filters: f}
	h.mu.Unlock()
}

func (h *sioHub) removeSocket(s *socket.Socket) {
	h.mu.Lock()
	delete(h.subs, s)
	h.mu.Unlock()
	h.aggMu.Lock()
	if t, ok := h.aggTimers[s]; ok {
		t.Stop()
		delete(h.aggTimers, s)
	}
	h.aggMu.Unlock()
}

func (h *sioHub) runBroadcast() {
	h.d.Logger.Info().Msg("[sioHub] runBroadcast() goroutine STARTED")
	sub := h.d.Monitor.Subscribe()
	h.d.Logger.Info().Msg("[sioHub] Monitor.Subscribe() completed, waiting for events...")
	defer h.d.Monitor.Unsubscribe(sub)
	for ev := range sub {
		h.d.Logger.Info().Str("event_type", ev.Type).Str("session_id", ev.ID).Msg("[sioHub] Received event from Monitor")
		switch ev.Type {
		case "session_started", "session_ended", "frame_added", "event_added", "http_tx_added", "mapping_applied", "session_error", "sessions_cleared":
			h.mu.RLock()
			h.d.Logger.Info().Str("event_type", ev.Type).Str("session_id", ev.ID).Int("subscribers", len(h.subs)).Msg("[sioHub] Broadcasting to Socket.IO clients")
			for _, s := range h.subs {
				h.applyEventToSub(ev.Type, ev.ID, s)
			}
			h.mu.RUnlock()

			// Emit to session-specific rooms for detail pages
			h.emitToSessionRoom(ev.Type, ev.ID)
		}
	}
	h.d.Logger.Info().Msg("[sioHub] runBroadcast() goroutine EXITED (channel closed)")
}

func (h *sioHub) applyEventToSub(evType, sessionID string, s *sioSubscription) {
	// use the current subscription for the connection at processing time
	h.mu.RLock()
	if cur, ok := h.subs[s.socket]; ok {
		s = cur
	}
	h.mu.RUnlock()
	if evType == "sessions_cleared" {
		s.socket.Emit("sessions:init", map[string]any{"items": []any{}, "aggregate": []any{}})
		return
	}
	ctx := contextWithNoCancel()
	sess, ok, _ := h.d.Svc.Get(ctx, sessionID)
	h.d.Logger.Info().Str("session_id", sessionID).Bool("found", ok).Msg("[sioHub] applyEventToSub: session lookup")
	if ok {
		view := sessionV1{Session: sess}
		meta, sz := h.d.enrichWithHTTPMeta(ctx, sess)
		if meta != nil {
			view.HttpMeta = meta
		}
		if sz != nil {
			view.Sizes = sz
		}
		passFilters := h.passQuickFilters(view, s.filters)
		// Additional check by tags (if specified in filter)
		if passFilters && len(s.filters.Tags) > 0 && h.d.TagsSvc != nil {
			if !h.sessionHasAnyTag(sessionID, s.filters.Tags) {
				passFilters = false
			}
		}
		passCapture := h.passCapture(view, s.filters)
		h.d.Logger.Info().Str("session_id", sessionID).Bool("pass_filters", passFilters).Bool("pass_capture", passCapture).Msg("[sioHub] applyEventToSub: filter check")
		if !passFilters || !passCapture {
			s.socket.Emit("sessions:remove", map[string]any{"id": sessionID})
			return
		}
		h.d.Logger.Info().Str("session_id", sessionID).Str("socket_id", string(s.socket.Id())).Msg("[sioHub] Emitting sessions:upsert")
		s.socket.Emit("sessions:upsert", view)
		h.scheduleAggregate(s)
		return
	}
	s.socket.Emit("sessions:remove", map[string]any{"id": sessionID})
	h.scheduleAggregate(s)
}

// sessionHasAnyTag returns true if the session has at least one of the specified tags
func (h *sioHub) sessionHasAnyTag(sessionID string, want []string) bool {
	ctx := contextWithNoCancel()
	tags, err := h.d.TagsSvc.GetSessionTags(ctx, sessionID)
	if err != nil || len(tags) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		set[strings.ToLower(strings.TrimSpace(t.TagName))] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[strings.ToLower(strings.TrimSpace(w))]; ok {
			return true
		}
	}
	return false
}

// (previously ignored internal UI paths, but per requirements we keep full visibility)

func (h *sioHub) passCapture(v sessionV1, f sioFilters) bool {
	if f.CaptureIDExplicit != nil {
		return v.CaptureID != nil && *v.CaptureID == *f.CaptureIDExplicit
	}
	switch strings.ToLower(strings.TrimSpace(f.CaptureScope)) {
	case "all":
		if !f.IncludeUnassigned && v.CaptureID == nil {
			return false
		}
		return true
	case "current", "":
		// If "show paused" is enabled, pass all (including those without CaptureID)
		if f.IncludeUnassigned {
			return true
		}
		// For "current" we pass ONLY sessions of the current capture
		cur := -1
		if repo := sessionsRepoOf(h.d.Svc); repo != nil {
			if rs, ok := repo.(interface{ RecordingState() (bool, int) }); ok {
				_, cur = rs.RecordingState()
			}
		}
		return v.CaptureID != nil && *v.CaptureID == cur
	default:
		return true
	}
}

func (h *sioHub) passQuickFilters(v sessionV1, f sioFilters) bool {
	if len(f.Types) > 0 {
		tags := getBaseTags(v)
		if !hasAnyTag(f.Types, tags) {
			return false
		}
	}
	if len(f.StatusGroups) > 0 {
		if !matchesAnyStatusGroup(f.StatusGroups, v.HttpMeta) {
			return false
		}
	}
	return true
}

func (h *sioHub) scheduleAggregate(s *sioSubscription) {
	h.aggMu.Lock()
	if t, ok := h.aggTimers[s.socket]; ok {
		t.Reset(1 * time.Second)
		h.aggMu.Unlock()
		return
	}
	t := time.AfterFunc(1*time.Second, func() {
		// get current subscription filters at timer firing time
		h.mu.RLock()
		cur, ok := h.subs[s.socket]
		h.mu.RUnlock()
		if !ok {
			return
		}
		h.emitAggregate(cur)
	})
	h.aggTimers[s.socket] = t
	h.aggMu.Unlock()
}

func (h *sioHub) emitAggregate(s *sioSubscription) {
	// Recalculate aggregates based on current subscription filters (up to 1000)
	ctx := contextWithNoCancel()
	uf := toUsecaseFilter(s.filters)
	list, _, _ := h.d.Svc.List(ctx, uf)
	sort.Slice(list, func(i, j int) bool { return list[i].StartedAt.Before(list[j].StartedAt) })
	views := make([]sessionV1, 0, len(list))
	for _, ss := range list {
		v := sessionV1{Session: ss}
		if meta, sz := h.d.computeHTTPMeta(ctx, ss.ID); meta != nil || sz != nil {
			v.HttpMeta = meta
			v.Sizes = sz
		}
		if !h.passQuickFilters(v, s.filters) || !h.passCapture(v, s.filters) {
			continue
		}
		views = append(views, v)
		if s.filters.Limit > 0 && len(views) >= s.filters.Limit {
			break
		}
	}
	agg := make(map[string]int)
	for _, v := range views {
		key := groupKeyFor(v, s.filters.GroupBy)
		agg[key]++
	}
	groups := make([]map[string]any, 0, len(agg))
	for k, cnt := range agg {
		groups = append(groups, map[string]any{"key": k, "count": cnt})
	}
	// stabilize order: by count desc, then by key asc
	sort.Slice(groups, func(i, j int) bool {
		ci := groups[i]["count"].(int)
		cj := groups[j]["count"].(int)
		if ci != cj {
			return ci > cj
		}
		ki := groups[i]["key"].(string)
		kj := groups[j]["key"].(string)
		return ki < kj
	})
	s.socket.Emit("aggregate:update", groups)
}

// emitToSessionRoom emits session-specific events to rooms for detail page subscriptions
func (h *sioHub) emitToSessionRoom(evType, sessionID string) {
	if sessionID == "" {
		return
	}
	ctx := contextWithNoCancel()
	roomName := socket.Room("session:" + sessionID)

	switch evType {
	case "frame_added":
		if frames, _, _ := h.d.Svc.ListFrames(ctx, sessionID, "", 1<<30); len(frames) > 0 {
			last := frames[len(frames)-1:]
			h.io.To(roomName).Emit("session:frames", last)
		}
	case "event_added", "sio_probe":
		if evs, _, _ := h.d.Svc.ListEvents(ctx, sessionID, "", 1<<30); len(evs) > 0 {
			last := evs[len(evs)-1:]
			h.io.To(roomName).Emit("session:events", last)
		}
	case "http_tx_added":
		if txs, _, _ := h.d.Svc.ListHTTPTransactions(ctx, sessionID, "", 1<<30); len(txs) > 0 {
			last := txs[len(txs)-1:]
			h.io.To(roomName).Emit("session:http", last)
		}
	case "session_ended":
		h.io.To(roomName).Emit("session:ended", map[string]any{"id": sessionID})
	case "session_started":
		h.io.To(roomName).Emit("session:started", map[string]any{"id": sessionID})
	}
}

// sendInit composes initial sessions list and aggregate according to filters and emits to the connection
func (h *sioHub) sendInit(s *socket.Socket, f sioFilters) {
	ctx := contextWithNoCancel()
	uf := toUsecaseFilter(f)

	// Apply tags filter if specified
	if len(f.Tags) > 0 && h.d.TagsSvc != nil {
		sessionIDs, err := h.d.TagsSvc.FindSessionIDsByTags(ctx, f.Tags)
		if err == nil {
			uf.SessionIDs = sessionIDs
		} else {
			h.d.Logger.Warn().Err(err).Strs("tags", f.Tags).Msg("Failed to find sessions by tags")
		}
	}

	list, _, _ := h.d.Svc.List(ctx, uf)
	h.d.Logger.Info().Str("conn_id", string(s.Id())).Int("total_sessions", len(list)).Msg("sendInit: fetched sessions")
	sort.Slice(list, func(i, j int) bool { return list[i].StartedAt.Before(list[j].StartedAt) })
	views := make([]sessionV1, 0, len(list))
	for _, s := range list {
		v := sessionV1{Session: s}
		if meta, sz := h.d.computeHTTPMeta(ctx, s.ID); meta != nil || sz != nil {
			v.HttpMeta = meta
			v.Sizes = sz
		}
		if !h.passQuickFilters(v, f) || !h.passCapture(v, f) {
			continue
		}
		views = append(views, v)
		if f.Limit > 0 && len(views) >= f.Limit {
			break
		}
	}
	agg := make(map[string]int)
	for _, v := range views {
		key := groupKeyFor(v, f.GroupBy)
		agg[key]++
	}
	groups := make([]map[string]any, 0, len(agg))
	for k, cnt := range agg {
		groups = append(groups, map[string]any{"key": k, "count": cnt})
	}
	// stabilize order: by count desc, then by key asc
	sort.Slice(groups, func(i, j int) bool {
		ci := groups[i]["count"].(int)
		cj := groups[j]["count"].(int)
		if ci != cj {
			return ci > cj
		}
		ki := groups[i]["key"].(string)
		kj := groups[j]["key"].(string)
		return ki < kj
	})
	h.d.Logger.Info().
		Str("conn_id", string(s.Id())).
		Int("filtered_sessions", len(views)).
		Int("aggregates", len(groups)).
		Msg("sendInit: emitting sessions:init")
	s.Emit("sessions:init", map[string]any{"items": views, "aggregate": groups})
}

// NewSocketIOServer: Socket.IO server + sessions:subscribe subscription (initial + increments)
func NewSocketIOServer(d *Deps) http.Handler {
	// Create Socket.IO server v3/v4 (zishang520/socket.io/v2)
	// Library automatically configures WebSocket transport and CORS
	io := socket.NewServer(nil, nil)

	hub := newSioHub(d, io)

	io.On("connection", func(clients ...any) {
		// Type assertion to get Socket from variadic parameter
		client := clients[0].(*socket.Socket)
		connID := string(client.Id())

		d.Logger.Info().Str("conn_id", connID).Msg("Socket.IO client connected")
		// Auto-subscribe with default filters so client immediately receives initial
		// Default Limit 1000 and current capture (passCapture treats empty scope as "current")
		// IncludeUnassigned: true allows seeing new sessions with CaptureID == nil
		f := sioFilters{Limit: 1000, GroupBy: "domain", IncludeUnassigned: true}
		hub.addSub(client, f)
		d.Logger.Info().Str("conn_id", connID).Msg("Sending sessions:init")
		hub.sendInit(client, f)
		d.Logger.Info().Str("conn_id", connID).Msg("Sent sessions:init")

		// Register event handlers - now variadic any instead of EventPayload
		client.On("sessions:subscribe", func(datas ...any) {
			var payload map[string]any
			if len(datas) > 0 {
				if m, ok := datas[0].(map[string]any); ok {
					payload = m
				}
			}
			f := parseSioFilters(payload)
			if f.Limit <= 0 || f.Limit > 1000 {
				f.Limit = 1000
			}
			hub.addSub(client, f)
			hub.sendInit(client, f)
		})

		// Subscribe to specific session for detail page
		client.On("session:subscribe", func(datas ...any) {
			var payload map[string]any
			if len(datas) > 0 {
				if m, ok := datas[0].(map[string]any); ok {
					payload = m
				}
			}
			sessionID, _ := payload["sessionId"].(string)
			if sessionID == "" {
				d.Logger.Warn().Str("conn_id", connID).Msg("session:subscribe without sessionId")
				return
			}
			roomName := socket.Room("session:" + sessionID)
			client.Join(roomName)
			d.Logger.Info().Str("conn_id", connID).Str("session_id", sessionID).Str("room", string(roomName)).Msg("Client joined session room")

			// Send initial data (catch-up)
			ctx := contextWithNoCancel()
			if frames, _, _ := d.Svc.ListFrames(ctx, sessionID, "", 1000); len(frames) > 0 {
				client.Emit("session:frames", frames)
			}
			if evs, _, _ := d.Svc.ListEvents(ctx, sessionID, "", 1000); len(evs) > 0 {
				client.Emit("session:events", evs)
			}
			if txs, _, _ := d.Svc.ListHTTPTransactions(ctx, sessionID, "", 1000); len(txs) > 0 {
				client.Emit("session:http", txs)
			}
		})

		// Unsubscribe from specific session
		client.On("session:unsubscribe", func(datas ...any) {
			var payload map[string]any
			if len(datas) > 0 {
				if m, ok := datas[0].(map[string]any); ok {
					payload = m
				}
			}
			sessionID, _ := payload["sessionId"].(string)
			if sessionID == "" {
				return
			}
			roomName := socket.Room("session:" + sessionID)
			client.Leave(roomName)
			d.Logger.Info().Str("conn_id", connID).Str("session_id", sessionID).Str("room", string(roomName)).Msg("Client left session room")
		})

		client.On("error", func(datas ...any) {
			var err error
			if len(datas) > 0 {
				if e, ok := datas[0].(error); ok {
					err = e
				}
			}
			connID := string(client.Id())
			d.Logger.Error().Err(err).Str("conn_id", connID).Msg("Socket.IO connection error")
		})

		client.On("disconnect", func(datas ...any) {
			var reason string
			if len(datas) > 0 {
				if r, ok := datas[0].(string); ok {
					reason = r
				}
			}
			connID := string(client.Id())
			d.Logger.Info().Str("conn_id", connID).Str("reason", reason).Msg("Socket.IO client disconnected")
			hub.removeSocket(client)
		})
	})

	return io.ServeHandler(nil)
}

func parseSioFilters(m map[string]any) sioFilters {
	f := sioFilters{Limit: 1000, GroupBy: "domain"}
	if v, ok := m["q"].(string); ok {
		f.Q = v
	}
	if v, ok := m["target"].(string); ok {
		f.Target = v
	}
	if v, ok := m["types"].([]any); ok {
		f.Types = make([]string, 0, len(v))
		for _, x := range v {
			if s, ok2 := x.(string); ok2 {
				f.Types = append(f.Types, strings.ToLower(strings.TrimSpace(s)))
			}
		}
	}
	if v, ok := m["status"].([]any); ok {
		f.StatusGroups = make([]string, 0, len(v))
		for _, x := range v {
			if s, ok2 := x.(string); ok2 {
				f.StatusGroups = append(f.StatusGroups, strings.ToLower(strings.TrimSpace(s)))
			}
		}
	}
	if v, ok := m["captureScope"].(string); ok {
		f.CaptureScope = v
	}
	if v, ok := m["includeUnassigned"].(bool); ok {
		f.IncludeUnassigned = v
	}
	if v, ok := m["limit"].(float64); ok {
		f.Limit = int(v)
	} else if s, ok2 := m["limit"].(string); ok2 {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			f.Limit = n
		}
	} else if iv, ok3 := m["limit"].(int); ok3 {
		f.Limit = iv
	}
	if v, ok := m["groupBy"].(string); ok && strings.TrimSpace(v) != "" {
		f.GroupBy = strings.ToLower(strings.TrimSpace(v))
	}
	if v, ok := m["captureId"].(float64); ok {
		n := int(v)
		f.CaptureIDExplicit = &n
	}
	if v, ok := m["tags"].([]any); ok {
		f.Tags = make([]string, 0, len(v))
		for _, x := range v {
			if s, ok2 := x.(string); ok2 {
				tagName := strings.TrimSpace(s)
				if tagName != "" {
					f.Tags = append(f.Tags, tagName)
				}
			}
		}
	}
	return f
}

// normalizeHost extracts and normalizes host from URL (removes default ports and userinfo)
func normalizeHost(targetURL string) string {
	// Simple URL parsing: remove scheme
	key := targetURL
	scheme := ""
	if i := strings.Index(key, "://"); i >= 0 {
		scheme = strings.ToLower(key[:i])
		key = key[i+3:]
	}

	// Remove userinfo (if present)
	if at := strings.IndexByte(key, '@'); at >= 0 {
		key = key[at+1:]
	}

	// Take part up to first slash (host:port or host)
	if j := strings.IndexByte(key, '/'); j >= 0 {
		key = key[:j]
	}

	// Remove default ports
	if scheme == "http" && strings.HasSuffix(key, ":80") {
		key = key[:len(key)-3]
	} else if scheme == "https" && strings.HasSuffix(key, ":443") {
		key = key[:len(key)-4]
	}

	return key
}

// groupKeyFor calculates the grouping key based on groupBy setting.
func groupKeyFor(v sessionV1, groupBy string) string {
	gb := strings.ToLower(strings.TrimSpace(groupBy))
	switch gb {
	case "method":
		if v.HttpMeta != nil && strings.TrimSpace(v.HttpMeta.Method) != "" {
			return strings.ToUpper(v.HttpMeta.Method)
		}
		return "unknown"
	case "mime", "contenttype":
		if v.HttpMeta != nil && strings.TrimSpace(v.HttpMeta.Mime) != "" {
			mime := strings.ToLower(v.HttpMeta.Mime)
			if i := strings.IndexByte(mime, ';'); i >= 0 {
				mime = strings.TrimSpace(mime[:i])
			}
			return mime
		}
		return "unknown"
	case "status", "statusgroup":
		if v.HttpMeta != nil && v.HttpMeta.Status > 0 {
			s := v.HttpMeta.Status / 100
			if s >= 1 && s <= 5 {
				return strconv.Itoa(s) + "xx"
			}
			return strconv.Itoa(v.HttpMeta.Status)
		}
		return "unknown"
	case "host", "domain", "target_host":
		fallthrough
	default:
		key := normalizeHost(v.Target)
		if key == "" {
			return "unknown"
		}
		return key
	}
}

func toUsecaseFilter(f sioFilters) usecase.SessionFilter {
	// Take wide limit for service list, since later we filter by
	// types/statuses and apply user Limit AFTER filters.
	out := usecase.SessionFilter{Q: f.Q, Target: f.Target, Limit: 1000, Offset: 0}
	if f.CaptureIDExplicit != nil {
		out.CaptureID = f.CaptureIDExplicit
	} else {
		switch strings.ToLower(strings.TrimSpace(f.CaptureScope)) {
		case "all":
			out.CaptureID = nil
			// Respect includeUnassigned from filter (don't force to true)
			if f.IncludeUnassigned {
				out.IncludeUnassigned = true
			}
		case "current", "":
			v := -1
			out.CaptureID = &v
		}
	}
	if f.IncludeUnassigned {
		out.IncludeUnassigned = true
	}
	return out
}
