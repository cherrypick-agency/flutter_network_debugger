package httpapi

import (
	"context"
	"encoding/json"
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
	CaptureScope      string // "current" | "all" | ""
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
	// Подписка на внутреннюю шину событий
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
	// используем актуальную подписку для соединения на момент обработки
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
		if meta, sz := h.d.computeHTTPMeta(ctx, sess.ID); meta != nil || sz != nil {
			view.HttpMeta = meta
			view.Sizes = sz
		} else if sess.Error != nil {
			// Если транзакций ещё нет, но уже есть ошибка — передадим ошибочную мету, чтобы UI видел статус сразу
			code := classifyNetError(*sess.Error)
			view.HttpMeta = &httpMetaV1{ErrorCode: code, ErrorMessage: *sess.Error}
		} else {
			if quick := h.computeQuickHTTPMetaFromFrames(ctx, sess.ID); quick != nil {
				view.HttpMeta = quick
			}
		}
		passFilters := h.passQuickFilters(view, s.filters)
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

// computeQuickHTTPMetaFromFrames — быстрый best-effort: вытащим method/status/mime из превью кадров,
// чтобы фронт видел базовую мету до записи транзакции.
func (h *sioHub) computeQuickHTTPMetaFromFrames(ctx context.Context, sessionID string) *httpMetaV1 {
	frames, _, _ := h.d.Svc.ListFrames(ctx, sessionID, "", 1000)
	if len(frames) == 0 {
		return nil
	}
	var method string
	var status int
	var mime string
	// от конца к началу ищем http_response со статусом и типом
	for i := len(frames) - 1; i >= 0; i-- {
		var m map[string]any
		if err := json.Unmarshal([]byte(frames[i].Preview), &m); err != nil {
			continue
		}
		if t, _ := m["type"].(string); t == "http_response" {
			if st, ok := m["status"].(float64); ok {
				status = int(st)
			}
			if hmap, ok := m["headers"].(map[string]any); ok {
				if ct, ok2 := headerGetCaseInsensitive(hmap, "content-type"); ok2 {
					mime = ct
				}
			}
			break
		}
	}
	// вперёд ищем первый http_request с методом
	for i := 0; i < len(frames); i++ {
		var m map[string]any
		if err := json.Unmarshal([]byte(frames[i].Preview), &m); err != nil {
			continue
		}
		if t, _ := m["type"].(string); t == "http_request" {
			if meth, ok := m["method"].(string); ok {
				method = meth
			}
			break
		}
	}
	if method == "" && status == 0 && mime == "" {
		return nil
	}
	out := &httpMetaV1{Method: method, Status: status, Mime: mime}
	return out
}

// headerGetCaseInsensitive возвращает значение заголовка независимо от регистра ключа.
func headerGetCaseInsensitive(h map[string]any, lowerName string) (string, bool) {
	for k, v := range h {
		if strings.EqualFold(k, lowerName) {
			switch vv := v.(type) {
			case string:
				return vv, true
			case []any:
				if len(vv) > 0 {
					if s, ok := vv[0].(string); ok {
						return s, true
					}
				}
			case []string:
				if len(vv) > 0 {
					return vv[0], true
				}
			}
		}
	}
	// быстрые варианты в случае точного совпадения ключа
	if v, ok := h["content-type"]; ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	if v, ok := h["Content-Type"]; ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	return "", false
}

// (раньше игнорировали внутренние пути UI, но по требованиям оставляем полную видимость)

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
		// If includeUnassigned is true, show all sessions including those without CaptureID
		// This allows new sessions (with CaptureID == nil) to appear in real-time
		if f.IncludeUnassigned {
			return true
		}
		return v.CaptureID != nil
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
		// берём актуальные фильтры подписки на момент срабатывания таймера
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
	// Пересчёт агрегатов на основе текущих фильтров подписки (до 1000)
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
	// стабилизируем порядок: по count desc, затем по key asc
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
	// стабилизируем порядок: по count desc, затем по key asc
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

// NewSocketIOServer: Socket.IO сервер + подписка sessions:subscribe (initial + инкременты)
func NewSocketIOServer(d *Deps) http.Handler {
	// Создаём Socket.IO сервер v3/v4 (zishang520/socket.io/v2)
	// Библиотека автоматически настраивает WebSocket транспорт и CORS
	io := socket.NewServer(nil, nil)

	hub := newSioHub(d, io)

	io.On("connection", func(clients ...any) {
		// Type assertion для получения Socket из variadic параметра
		client := clients[0].(*socket.Socket)
		connID := string(client.Id())

		d.Logger.Info().Str("conn_id", connID).Msg("Socket.IO client connected")
		// Автоподписка с дефолтными фильтрами, чтобы клиент сразу получил initial
		// По умолчанию Limit 1000 и текущий capture (passCapture обрабатывает пустой scope как "current")
		// IncludeUnassigned: true позволяет видеть новые сессии с CaptureID == nil
		f := sioFilters{Limit: 1000, GroupBy: "domain", IncludeUnassigned: true}
		hub.addSub(client, f)
		d.Logger.Info().Str("conn_id", connID).Msg("Sending sessions:init")
		hub.sendInit(client, f)
		d.Logger.Info().Str("conn_id", connID).Msg("Sent sessions:init")

		// Регистрируем обработчики событий - теперь variadic any вместо EventPayload
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

		// Подписка на конкретную сессию для детальной страницы
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

			// Отправка начальных данных (catch-up)
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

		// Отписка от конкретной сессии
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
	return f
}

// normalizeHost извлекает и нормализует host из URL (убирает дефолтные порты и userinfo)
func normalizeHost(targetURL string) string {
	// Простая парсинг URL: удаляем схему
	key := targetURL
	scheme := ""
	if i := strings.Index(key, "://"); i >= 0 {
		scheme = strings.ToLower(key[:i])
		key = key[i+3:]
	}

	// Убираем userinfo (если есть)
	if at := strings.IndexByte(key, '@'); at >= 0 {
		key = key[at+1:]
	}

	// Берём часть до первого слеша (host:port или host)
	if j := strings.IndexByte(key, '/'); j >= 0 {
		key = key[:j]
	}

	// Убираем дефолтные порты
	if scheme == "http" && strings.HasSuffix(key, ":80") {
		key = key[:len(key)-3]
	} else if scheme == "https" && strings.HasSuffix(key, ":443") {
		key = key[:len(key)-4]
	}

	return key
}

// groupKeyFor вычисляет ключ группировки по настройке groupBy.
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
	// Берём широкий лимит для сервисного списка, так как далее мы фильтруем по
	// типам/статусам и применяем пользовательский Limit уже ПОСЛЕ фильтров.
	out := usecase.SessionFilter{Q: f.Q, Target: f.Target, Limit: 1000, Offset: 0}
	if f.CaptureIDExplicit != nil {
		out.CaptureID = f.CaptureIDExplicit
	} else {
		switch strings.ToLower(strings.TrimSpace(f.CaptureScope)) {
		case "all":
			out.CaptureID = nil
			out.IncludeUnassigned = true
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
