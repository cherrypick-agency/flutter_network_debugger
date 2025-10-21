package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	mem "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
	"strconv"
	"strings"
	"time"
)

func (d *Deps) handleListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		if err := d.Svc.ClearAll(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "SESSIONS_CLEAR_FAILED", err.Error(), nil)
			return
		}
		// also close live WS sessions to prevent further events
		if d.Live != nil {
			d.Live.CloseAll()
		}
		// and broadcast a synthetic event so frontends can refresh
		if d.Monitor != nil {
			d.Monitor.Broadcast(MonitorEvent{Type: "sessions_cleared", ID: "*"})
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	q := r.URL.Query().Get("q")
	target := r.URL.Query().Get("_target")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	f := usecase.SessionFilter{Q: q, Target: target, Limit: limit, Offset: offset}
	items, total, err := d.Svc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSIONS_LIST_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items, "total": total})
}

func (d *Deps) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	// path: /api/sessions/{id}[/(frames|events)]
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	id := parts[0]
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
		return
	}
	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			_ = d.Svc.Delete(r.Context(), id)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		sess, ok, err := d.Svc.Get(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_GET_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found", map[string]any{"id": id})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sess)
		return
	}
	switch parts[1] {
	case "frames":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		from := r.URL.Query().Get("from")
		frames, next, err := d.Svc.ListFrames(r.Context(), id, from, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "FRAMES_LIST_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": frames, "next": next})
	case "events":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		from := r.URL.Query().Get("from")
		events, next, err := d.Svc.ListEvents(r.Context(), id, from, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "EVENTS_LIST_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		// Backward-compat: include both "event" and alias "name" fields
		type evView struct {
			ID          string    `json:"id"`
			Ts          time.Time `json:"ts"`
			Namespace   string    `json:"namespace"`
			Event       string    `json:"event"`
			Name        string    `json:"name"`
			AckID       *int64    `json:"ackId,omitempty"`
			ArgsPreview string    `json:"argsPreview"`
		}
		out := make([]evView, 0, len(events))
		for _, e := range events {
			out = append(out, evView{
				ID:          e.ID,
				Ts:          e.Ts,
				Namespace:   e.Namespace,
				Event:       e.Name,
				Name:        e.Name,
				AckID:       e.AckID,
				ArgsPreview: e.ArgsPreview,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": out, "next": next})
	case "http":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		from := r.URL.Query().Get("from")
		txs, next, err := d.Svc.ListHTTPTransactions(r.Context(), id, from, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "HTTP_LIST_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": txs, "next": next})
	case "har":
		// Export HAR 1.2 for this session (HTTP transactions only)
		exportHARForSession(w, r, d, id)
		return
	case "export":
		// aggregate full session with all frames and events
		sess, ok, err := d.Svc.Get(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_GET_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found", map[string]any{"id": id})
			return
		}
		// collect frames
		allFrames := make([]any, 0, 1024)
		from := ""
		for {
			frames, next, err := d.Svc.ListFrames(r.Context(), id, from, 1000)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "FRAMES_LIST_FAILED", err.Error(), map[string]any{"id": id})
				return
			}
			for _, f := range frames {
				allFrames = append(allFrames, f)
			}
			if next == "" {
				break
			}
			from = next
		}
		// collect events
		allEvents := make([]any, 0, 256)
		from = ""
		for {
			events, next, err := d.Svc.ListEvents(r.Context(), id, from, 1000)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "EVENTS_LIST_FAILED", err.Error(), map[string]any{"id": id})
				return
			}
			for _, e := range events {
				allEvents = append(allEvents, e)
			}
			if next == "" {
				break
			}
			from = next
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=network-debugger_session_"+id+".json")
		_ = json.NewEncoder(w).Encode(map[string]any{"session": sess, "frames": allFrames, "events": allEvents})
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
	}
}

// ============================
// V1 handlers with cursor API
// ============================

// handleV1ListSessions implements GET /_api/v1/sessions with cursor pagination and sorting.
func (d *Deps) handleV1ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		if err := d.Svc.ClearAll(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "SESSIONS_CLEAR_FAILED", err.Error(), nil)
			return
		}
		// Закрываем активные WS-сессии и уведомляем фронты, чтобы не прилетали новые события в старые сессии
		if d.Live != nil {
			d.Live.CloseAll()
		}
		if d.Monitor != nil {
			d.Monitor.Broadcast(MonitorEvent{Type: "sessions_cleared", ID: "*"})
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	q := r.URL.Query().Get("q")
	target := r.URL.Query().Get("_target")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	// For MVP we reuse offset-based List and synthesize a cursor as last id.
	// A real cursor would be a stable token (e.g., startedAt+id).
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	f := usecase.SessionFilter{Q: q, Target: target, Limit: limit, Offset: offset}
	// capture filters
	capStr := r.URL.Query().Get("captureId")
	if capStr != "" {
		if capStr == "current" {
			v := -1
			f.CaptureID = &v
		} else if n, err := strconv.Atoi(capStr); err == nil {
			f.CaptureID = &n
		}
	}
	if inc := r.URL.Query().Get("includeUnassigned"); inc == "true" || inc == "1" {
		f.IncludeUnassigned = true
	}
	if r.URL.Query().Get("captures") == "all" {
		f.CaptureID = nil
		f.IncludeUnassigned = true
	}
	items, total, err := d.Svc.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSIONS_LIST_FAILED", err.Error(), nil)
		return
	}
	// Enrich with httpMeta/sizes best-effort
	views := make([]sessionV1, 0, len(items))
	for _, s := range items {
		view := sessionV1{Session: s}
		meta, sz := d.computeHTTPMeta(r.Context(), s.ID)
		if meta == nil && s.Error != nil {
			code := classifyNetError(*s.Error)
			meta = &httpMetaV1{Method: "", Status: 0, Mime: "", DurationMs: 0, Streaming: false, Headers: map[string]string{}, ErrorCode: code, ErrorMessage: *s.Error}
		}
		if meta != nil {
			view.HttpMeta = meta
		}
		if sz != nil {
			view.Sizes = sz
		}
		views = append(views, view)
	}
	w.Header().Set("Content-Type", "application/json")
	next := ""
	if offset+limit < total {
		next = strconv.Itoa(offset + limit)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"items": views, "next": next})
}

// handleV1SessionByID dispatches to subresources: frames/events/body/http
func (d *Deps) handleV1SessionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/_api/v1/sessions/")
	parts := strings.Split(path, "/")
	id := parts[0]
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
		return
	}
	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			_ = d.Svc.Delete(r.Context(), id)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		sess, ok, err := d.Svc.Get(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_GET_FAILED", err.Error(), nil)
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found", map[string]any{"id": id})
			return
		}
		view := sessionV1{Session: sess}
		meta, sz := d.computeHTTPMeta(r.Context(), id)
		if meta == nil && sess.Error != nil {
			code := classifyNetError(*sess.Error)
			meta = &httpMetaV1{Method: "", Status: 0, Mime: "", DurationMs: 0, Streaming: false, Headers: map[string]string{}, ErrorCode: code, ErrorMessage: *sess.Error}
		}
		if meta != nil {
			view.HttpMeta = meta
		}
		if sz != nil {
			view.Sizes = sz
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(view)
		return
	}
	switch parts[1] {
	case "frames":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		from := r.URL.Query().Get("from")
		frames, next, err := d.Svc.ListFrames(r.Context(), id, from, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "FRAMES_LIST_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": frames, "next": next})
	case "events":
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		from := r.URL.Query().Get("from")
		events, next, err := d.Svc.ListEvents(r.Context(), id, from, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "EVENTS_LIST_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": events, "next": next})
	case "body":
		// Placeholder: body storage not implemented in memory store => 404 with reason
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"stored": false, "reason": "not_implemented"})
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
	}
}

// handleV1SessionsAggregate implements GET /_api/v1/sessions/aggregate
func (d *Deps) handleV1SessionsAggregate(w http.ResponseWriter, r *http.Request) {
	// MVP: group by domain derived from Session.Target; compute count only.
	// Advanced stats (avgDuration, p95, statusClass) require richer storage of httpMeta; skipped for MVP.
	items, _, err := d.Svc.List(r.Context(), usecase.SessionFilter{Limit: 1000, Offset: 0})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSIONS_LIST_FAILED", err.Error(), nil)
		return
	}
	agg := map[string]int{}
	for _, s := range items {
		key := s.Target
		if i := strings.Index(key, "://"); i >= 0 {
			key = key[i+3:]
		}
		if j := strings.IndexByte(key, '/'); j >= 0 {
			key = key[:j]
		}
		agg[key]++
	}
	type group struct {
		Key   string `json:"key"`
		Count int    `json:"count"`
	}
	out := struct {
		Groups []group `json:"groups"`
	}{Groups: make([]group, 0, len(agg))}
	for k, v := range agg {
		out.Groups = append(out.Groups, group{Key: k, Count: v})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// --- Capture controls (MVP, memory-backed) ---
func (d *Deps) handleV1Capture(w http.ResponseWriter, r *http.Request) {
	// Use memory store if available
	// For MVP we rely on memory.Store being the concrete repository
	repo := sessionsRepoOf(d.Svc)
	mem, ok := repo.(*mem.Store)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "CAPTURE_UNAVAILABLE", "capture unsupported", nil)
		return
	}
	type resp struct {
		Recording bool `json:"recording"`
		Current   int  `json:"current"`
	}
	switch r.Method {
	case http.MethodGet:
		rec, cur := mem.RecordingState()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp{Recording: rec, Current: cur})
	case http.MethodPost:
		var body struct {
			Action string `json:"action"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		act := strings.ToLower(body.Action)
		switch act {
		case "start":
			cur := mem.StartCapture()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp{Recording: true, Current: cur})
		case "stop":
			cur := mem.StopCapture()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp{Recording: false, Current: cur})
		default:
			writeError(w, http.StatusBadRequest, "BAD_ACTION", "action must be start|stop", nil)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use GET/POST", nil)
	}
}

func (d *Deps) handleV1Captures(w http.ResponseWriter, r *http.Request) {
	// Build list of distinct capture ids from sessions
	items, _, _ := d.Svc.List(r.Context(), usecase.SessionFilter{Limit: 100000, Offset: 0})
	used := map[int]struct{}{}
	for _, s := range items {
		if s.CaptureID != nil {
			used[*s.CaptureID] = struct{}{}
		}
	}
	// Always include current capture id
	if repo := sessionsRepoOf(d.Svc); repo != nil {
		if ms, ok := repo.(interface{ RecordingState() (bool, int) }); ok {
			_, cur := ms.RecordingState()
			used[cur] = struct{}{}
		}
	}
	out := make([]map[string]any, 0, len(used))
	for id := range used {
		out = append(out, map[string]any{"id": id})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": out})
}

// helper to get underlying session repository (MVP, not ideal)
func sessionsRepoOf(svc *usecase.SessionService) any {
	// access unexported field via known struct; in real project expose via interface
	return any(svc).(*usecase.SessionService).SessionsRepoUnsafe()
}

// ---- V1 view models ----
type sessionV1 struct {
	domain.Session
	HttpMeta *httpMetaV1 `json:"httpMeta,omitempty"`
	Sizes    *sizeInfoV1 `json:"sizes,omitempty"`
}

// augmentations
type httpMetaV1 struct {
	Method       string            `json:"method"`
	Status       int               `json:"status"`
	Mime         string            `json:"mime"`
	DurationMs   int64             `json:"durationMs"`
	Streaming    bool              `json:"streaming"`
	Headers      map[string]string `json:"headers"`
	Cache        *cacheMetaV1      `json:"cache,omitempty"`
	CORS         *corsMetaV1       `json:"cors,omitempty"`
	Preflight    *preflightLinkV1  `json:"preflight,omitempty"`
	ErrorCode    string            `json:"errorCode,omitempty"`
	ErrorMessage string            `json:"errorMessage,omitempty"`
}

type sizeInfoV1 struct {
	RequestBytes  int `json:"requestBytes"`
	ResponseBytes int `json:"responseBytes"`
}

// cache/cors/preflight view models
type cacheMetaV1 struct {
	Status     string            `json:"status"` // HIT/MISS/REVALIDATED/UNKNOWN
	Directives map[string]string `json:"directives,omitempty"`
	ETag       string            `json:"etag,omitempty"`
	Age        int               `json:"age,omitempty"`
}

type corsMetaV1 struct {
	Ok             bool     `json:"ok"`
	Reason         string   `json:"reason,omitempty"`
	AllowedOrigin  string   `json:"allowedOrigin,omitempty"`
	AllowedMethods []string `json:"allowedMethods,omitempty"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
	Vary           string   `json:"vary,omitempty"`
}

type preflightLinkV1 struct {
	IsPreflight   bool   `json:"isPreflight"`
	MainSessionId string `json:"mainSessionId,omitempty"`
}

// computeHTTPMeta derives httpMeta/sizes from stored HTTP transactions; best-effort.
func (d *Deps) computeHTTPMeta(ctx context.Context, sessionID string) (*httpMetaV1, *sizeInfoV1) {
	txs, _, err := d.Svc.ListHTTPTransactions(ctx, sessionID, "", 1000000)
	if err != nil || len(txs) == 0 {
		return nil, nil
	}
	tx := txs[len(txs)-1]
	meta := &httpMetaV1{
		Method:     tx.Method,
		Status:     tx.Status,
		Mime:       tx.ContentType,
		DurationMs: tx.Timings.Total,
		Streaming:  false,
		Headers:    map[string]string{},
	}
	sizes := &sizeInfoV1{RequestBytes: tx.ReqSize, ResponseBytes: tx.RespSize}

	// Extract headers from latest response preview and request preview
	var reqHeaders map[string]string
	var respHeaders map[string]string
	if frames, _, _ := d.Svc.ListFrames(ctx, sessionID, "", 1000); len(frames) > 0 {
		for i := len(frames) - 1; i >= 0; i-- {
			var prev map[string]any
			if err := json.Unmarshal([]byte(frames[i].Preview), &prev); err != nil {
				continue
			}
			if t, _ := prev["type"].(string); t == "http_response" && respHeaders == nil {
				if h, ok := prev["headers"].(map[string]any); ok {
					respHeaders = mapToStringMap(h)
				}
			}
			if t, _ := prev["type"].(string); t == "http_request" && reqHeaders == nil {
				if h, ok := prev["headers"].(map[string]any); ok {
					reqHeaders = mapToStringMap(h)
				}
			}
			if reqHeaders != nil && respHeaders != nil {
				break
			}
		}
	}
	if respHeaders != nil {
		meta.Headers = respHeaders
	}

	// Cache meta
	meta.Cache = computeCacheMeta(tx.Status, respHeaders)
	// CORS meta
	isPreflight := strings.ToUpper(tx.Method) == http.MethodOptions && hasHeaderFold(reqHeaders, "Access-Control-Request-Method")
	meta.CORS = computeCORSMeta(strings.ToUpper(tx.Method), reqHeaders, respHeaders, isPreflight)
	// Preflight link (best-effort only marks preflight in this session)
	meta.Preflight = &preflightLinkV1{IsPreflight: isPreflight}

	return meta, sizes
}

func mapToStringMap(h map[string]any) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		out[k] = toString(v)
	}
	return out
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func computeCacheMeta(status int, hdr map[string]string) *cacheMetaV1 {
	if hdr == nil {
		return &cacheMetaV1{Status: "UNKNOWN"}
	}
	cc := getFold(hdr, "Cache-Control")
	etag := getFold(hdr, "ETag")
	ageStr := getFold(hdr, "Age")
	age := 0
	if ageStr != "" {
		if n, err := strconv.Atoi(ageStr); err == nil {
			age = n
		}
	}
	directives := parseCacheControl(cc)
	st := "MISS"
	if status == http.StatusNotModified {
		st = "REVALIDATED"
	} else if age > 0 {
		st = "HIT"
	}
	return &cacheMetaV1{Status: st, Directives: directives, ETag: etag, Age: age}
}

func parseCacheControl(s string) map[string]string {
	if s == "" {
		return nil
	}
	res := map[string]string{}
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if i := strings.IndexByte(p, '='); i >= 0 {
			k := strings.TrimSpace(p[:i])
			v := strings.TrimSpace(p[i+1:])
			res[strings.ToLower(k)] = strings.Trim(v, "\"")
		} else {
			res[strings.ToLower(p)] = "true"
		}
	}
	return res
}

func computeCORSMeta(method string, req, resp map[string]string, isPreflight bool) *corsMetaV1 {
	if req == nil || resp == nil {
		return &corsMetaV1{Ok: false, Reason: "missing headers"}
	}
	origin := getFold(req, "Origin")
	if origin == "" {
		return &corsMetaV1{Ok: true, Reason: "no origin"}
	}
	allowOrigin := getFold(resp, "Access-Control-Allow-Origin")
	allowMethods := csvToSlice(getFold(resp, "Access-Control-Allow-Methods"))
	allowHeaders := csvToSlice(getFold(resp, "Access-Control-Allow-Headers"))
	vary := getFold(resp, "Vary")

	ok := false
	reason := ""
	if isPreflight {
		reqMethod := strings.ToUpper(getFold(req, "Access-Control-Request-Method"))
		reqHeaders := csvToSlice(getFold(req, "Access-Control-Request-Headers"))
		originOk := (allowOrigin == "*" || allowOrigin == origin)
		methodOk := containsFoldSlice(allowMethods, reqMethod)
		headersOk := allAllowedFold(allowHeaders, reqHeaders)
		ok = originOk && methodOk && headersOk
		if !originOk {
			reason = "origin"
		} else if !methodOk {
			reason = "method"
		} else if !headersOk {
			reason = "headers"
		}
	} else {
		originOk := (allowOrigin == "*" || allowOrigin == origin)
		methodOk := containsFoldSlice(allowMethods, method)
		ok = originOk && (len(allowMethods) == 0 || methodOk)
		if !originOk {
			reason = "origin"
		} else if len(allowMethods) > 0 && !methodOk {
			reason = "method"
		}
	}
	return &corsMetaV1{Ok: ok, Reason: reason, AllowedOrigin: allowOrigin, AllowedMethods: allowMethods, AllowedHeaders: allowHeaders, Vary: vary}
}

// Грубая классификация сетевых ошибок для UI
func classifyNetError(msg string) string {
	m := strings.ToLower(msg)
	switch {
	case strings.Contains(m, "context deadline exceeded") || strings.Contains(m, "timeout"):
		return "TIMEOUT"
	case strings.Contains(m, "no such host") || strings.Contains(m, "server misbehaving"):
		return "DNS"
	case strings.Contains(m, "x509") || strings.Contains(m, "certificate") || strings.Contains(m, "tls"):
		return "TLS"
	case strings.Contains(m, "connection refused") || strings.Contains(m, "cannot assign"):
		return "CONNECT"
	case strings.Contains(m, "connection reset") || strings.Contains(m, "reset by peer"):
		return "RST"
	case strings.Contains(m, "before full header") || strings.Contains(m, "unexpected eof") || strings.Contains(m, "early eof") || strings.Contains(m, "eof"):
		return "EOF"
	case strings.Contains(m, "request canceled") || strings.Contains(m, "client canceled"):
		return "CANCEL"
	default:
		return "ERROR"
	}
}

func allAllowedFold(allowed []string, requested []string) bool {
	if len(requested) == 0 {
		return true
	}
	if len(allowed) == 0 {
		return false
	}
	for _, r := range requested {
		ok := false
		for _, a := range allowed {
			if strings.ToLower(a) == strings.ToLower(r) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func getFold(h map[string]string, key string) string {
	if h == nil {
		return ""
	}
	lk := strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == lk {
			return v
		}
	}
	return ""
}

func hasHeaderFold(h map[string]string, key string) bool { return getFold(h, key) != "" }
func csvToSlice(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
func containsFoldSlice(arr []string, val string) bool {
	lv := strings.ToLower(val)
	for _, a := range arr {
		if strings.ToLower(a) == lv {
			return true
		}
	}
	return false
}

// handleSessionStream provides Server-Sent Events for real-time updates of a session.
// Path: /api/sessions_stream/{id}
func (d *Deps) handleSessionStream(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sessions_stream/")
	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "STREAM_UNSUPPORTED", "stream unsupported", nil)
		return
	}

	// Push-based stream: subscribe to in-process monitor bus and fan-out events for this session
	sub := d.Monitor.Subscribe()
	defer d.Monitor.Unsubscribe(sub)
	enc := json.NewEncoder(w)
	// initial catch-up (optional): send last chunks
	if frames, _, _ := d.Svc.ListFrames(r.Context(), id, "", 1000); len(frames) > 0 {
		_ = writeSSE(w, flusher, "frames", frames, enc)
	}
	if evs, _, _ := d.Svc.ListEvents(r.Context(), id, "", 1000); len(evs) > 0 {
		_ = writeSSE(w, flusher, "events", evs, enc)
	}
	if txs, _, _ := d.Svc.ListHTTPTransactions(r.Context(), id, "", 1000); len(txs) > 0 {
		_ = writeSSE(w, flusher, "http", txs, enc)
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-sub:
			// Filter by session id
			if ev.ID != id {
				continue
			}
			switch ev.Type {
			case "frame_added":
				if frames, _, _ := d.Svc.ListFrames(r.Context(), id, "", 1<<30); len(frames) > 0 {
					last := frames[len(frames)-1:]
					_ = writeSSE(w, flusher, "frames", last, enc)
				}
			case "event_added", "sio_probe":
				if evs, _, _ := d.Svc.ListEvents(r.Context(), id, "", 1<<30); len(evs) > 0 {
					last := evs[len(evs)-1:]
					_ = writeSSE(w, flusher, "events", last, enc)
				}
			case "http_tx_added":
				if txs, _, _ := d.Svc.ListHTTPTransactions(r.Context(), id, "", 1<<30); len(txs) > 0 {
					last := txs[len(txs)-1:]
					_ = writeSSE(w, flusher, "http", last, enc)
				}
			case "session_ended", "session_started":
				_ = writeSSE(w, flusher, ev.Type, ev, enc)
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any, enc *json.Encoder) error {
	_, _ = w.Write([]byte("event: " + event + "\n"))
	// write data: <json> in one line
	_ = enc.Encode(data)
	_, _ = w.Write([]byte("\n"))
	flusher.Flush()
	return nil
}
