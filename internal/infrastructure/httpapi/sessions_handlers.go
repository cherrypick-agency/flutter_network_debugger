package httpapi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	mem "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
	"os"
	"strconv"
	"strings"
	"sync"
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
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sessions_cleared", ID: "*"})
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

	// Parse tags filter (comma-separated)
	var sessionIDs []string
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam != "" && d.TagsSvc != nil {
		tagNames := strings.Split(tagsParam, ",")
		// Trim whitespace from tag names
		for i := range tagNames {
			tagNames[i] = strings.TrimSpace(tagNames[i])
		}
		// Filter out empty tags
		filtered := make([]string, 0, len(tagNames))
		for _, t := range tagNames {
			if t != "" {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) > 0 {
			ids, err := d.TagsSvc.FindSessionIDsByTags(r.Context(), filtered)
			if err == nil {
				sessionIDs = ids
			}
		}
	}

	f := usecase.SessionFilter{
		Q:          q,
		Target:     target,
		Limit:      limit,
		Offset:     offset,
		SessionIDs: sessionIDs,
	}
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
		// Check if this is /frames/{frameId}/body
		if len(parts) >= 4 && parts[3] == "body" {
			d.handleFrameBody(w, r, id, parts[2])
			return
		}
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
		// streaming export to reduce memory footprint
		sess, ok, err := d.Svc.Get(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_GET_FAILED", err.Error(), map[string]any{"id": id})
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "session not found", map[string]any{"id": id})
			return
		}
		// decide gzip by flag or Accept-Encoding
		useGzip := false
		if g := r.URL.Query().Get("gzip"); g == "1" || strings.ToLower(g) == "true" {
			useGzip = true
		} else if strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip") {
			useGzip = true
		}
		var outWriter interface{ Write([]byte) (int, error) }
		var gz *gzip.Writer
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=network-debugger_session_"+id+".json")
		if useGzip {
			w.Header().Set("Content-Encoding", "gzip")
			gz = gzip.NewWriter(w)
			outWriter = gz
		} else {
			outWriter = w
		}
		flush := func() {
			if gz != nil {
				_ = gz.Flush()
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		write := func(b []byte) { _, _ = outWriter.Write(b) }
		b, _ := json.Marshal(sess)
		write([]byte("{\"session\":"))
		write(b)
		write([]byte(",\"frames\":["))
		from := ""
		first := true
		for {
			frames, next, err := d.Svc.ListFrames(r.Context(), id, from, 1000)
			if err != nil {
				break
			}
			for _, fr := range frames {
				fb, _ := json.Marshal(fr)
				if !first {
					write([]byte(","))
				} else {
					first = false
				}
				write(fb)
			}
			flush()
			if next == "" {
				break
			}
			from = next
		}
		write([]byte("],\"events\":["))
		from = ""
		first = true
		for {
			ev, next, err := d.Svc.ListEvents(r.Context(), id, from, 1000)
			if err != nil {
				break
			}
			for _, e := range ev {
				eb, _ := json.Marshal(e)
				if !first {
					write([]byte(","))
				} else {
					first = false
				}
				write(eb)
			}
			flush()
			if next == "" {
				break
			}
			from = next
		}
		write([]byte("]}"))
		if gz != nil {
			_ = gz.Close()
		}
		return
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
		// Close active WS sessions and notify frontends so new events don't arrive in old sessions
		if d.Live != nil {
			d.Live.CloseAll()
		}
		if d.Monitor != nil {
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sessions_cleared", ID: "*"})
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	q := r.URL.Query().Get("q")
	target := r.URL.Query().Get("_target")
	rawTypes := r.URL.Query().Get("types")
	rawStatus := r.URL.Query().Get("status")
	types := splitCSV(rawTypes)
	statusGroups := splitCSV(rawStatus)
	// Tags filter (comma-separated)
	var sessionIDs []string
	if raw := strings.TrimSpace(r.URL.Query().Get("tags")); raw != "" && d.TagsSvc != nil {
		names := strings.Split(raw, ",")
		for i := range names {
			names[i] = strings.TrimSpace(names[i])
		}
		filtered := make([]string, 0, len(names))
		for _, t := range names {
			if t != "" {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) > 0 {
			if ids, err := d.TagsSvc.FindSessionIDsByTags(r.Context(), filtered); err == nil {
				sessionIDs = ids
			}
		}
	}
	// enable deep GraphQL body check only if explicitly requested
	rawScan := r.URL.Query().Get("scan")
	scan := splitCSV(rawScan)
	scanGraphQL := false
	for _, s := range scan {
		if s == "graphql" {
			scanGraphQL = true
			break
		}
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	// For MVP we reuse offset-based List and synthesize a cursor as last id.
	// A real cursor would be a stable token (e.g., startedAt+id).
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	f := usecase.SessionFilter{Q: q, Target: target, Limit: limit, Offset: offset, SessionIDs: sessionIDs}
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
	// Enrich with httpMeta/sizes best-effort and apply quick filters
	views := make([]sessionV1, 0, len(items))
	for _, s := range items {
		view := sessionV1{Session: s}
		meta, sz := d.enrichWithHTTPMeta(r.Context(), s)
		if meta != nil {
			view.HttpMeta = meta
		}
		if sz != nil {
			view.Sizes = sz
		}
		// Quick filters by types (with optional deep-scan for GraphQL)
		if len(types) > 0 {
			tags := getBaseTags(view)
			// If graphql is requested but not in base tags — at client's request do deep body check
			needsGraphQL := false
			for _, t := range types {
				if t == "graphql" {
					needsGraphQL = true
					break
				}
			}
			if needsGraphQL && scanGraphQL {
				if _, ok := tags["graphql"]; !ok {
					if ok2 := detectGraphQLByBody(r.Context(), d, s.ID); ok2 {
						tags["graphql"] = struct{}{}
					}
				}
			}
			if !hasAnyTag(types, tags) {
				continue
			}
		}
		// Quick filters by status groups
		if len(statusGroups) > 0 {
			if !matchesAnyStatusGroup(statusGroups, view.HttpMeta) {
				continue
			}
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
			// CASCADE DELETE: Remove tags and annotations before deleting session
			if d.TagsSvc != nil {
				// Best effort deletion - ignore errors
				_ = d.TagsSvc.DeleteAllSessionTags(r.Context(), id)
				_ = d.TagsSvc.DeleteAllAnnotations(r.Context(), id)
			}

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
			info := classifyProxyErrorString(*sess.Error)
			meta = &httpMetaV1{
				Method:           "",
				Status:           0,
				Mime:             "",
				DurationMs:       0,
				Streaming:        false,
				Headers:          map[string]string{},
				ErrorCategory:    info.Category,
				ErrorCode:        info.Code,
				ErrorUserMessage: info.UserMessage,
				ErrorMessage:     info.UserMessage,
			}
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
		// Check if this is /frames/{frameId}/body
		if len(parts) >= 4 && parts[3] == "body" {
			d.handleFrameBody(w, r, id, parts[2])
			return
		}
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
	case "tags":
		d.handleSessionTags(w, r)
	case "annotations":
		d.handleSessionAnnotations(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
	}
}

// ===== helpers for quick filters =====

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hasAnyTag(want []string, tags map[string]struct{}) bool {
	if len(want) == 0 {
		return true
	}
	for _, w := range want {
		if _, ok := tags[w]; ok {
			return true
		}
	}
	return false
}

// tagsForSession builds a compact set of tags based on scheme/kind and mime
// --- base tags cache (without deep body analysis)
var baseTagsCache sync.Map // map[string]baseTagsEntry

type baseTagsEntry struct {
	tags   map[string]struct{}
	kind   string
	target string
	mime   string
}

func getBaseTags(v sessionV1) map[string]struct{} {
	mime := ""
	if v.HttpMeta != nil {
		mime = strings.ToLower(v.HttpMeta.Mime)
	}
	key := v.ID
	sigKind := strings.ToLower(v.Kind)
	sigTarget := v.Target
	if val, ok := baseTagsCache.Load(key); ok {
		e := val.(baseTagsEntry)
		if e.kind == sigKind && e.target == sigTarget && e.mime == mime {
			return e.tags
		}
	}
	tags := computeBaseTags(v)
	baseTagsCache.Store(key, baseTagsEntry{tags: tags, kind: sigKind, target: sigTarget, mime: mime})
	return tags
}

func computeBaseTags(v sessionV1) map[string]struct{} {
	tags := map[string]struct{}{}
	// scheme/kind
	if u, err := (&urlParser{}).parse(v.Target); err == nil {
		scheme := strings.ToLower(u.Scheme)
		switch scheme {
		case "https":
			tags["https"] = struct{}{}
		case "http":
			tags["http"] = struct{}{}
		case "ws", "wss":
			tags["ws"] = struct{}{}
		}
		// GraphQL heuristic by URL path
		p := strings.ToLower(u.Path)
		if strings.Contains(p, "graphql") {
			tags["graphql"] = struct{}{}
		}
	}
	if strings.ToLower(v.Kind) == "ws" {
		tags["ws"] = struct{}{}
	}
	if strings.ToLower(v.Kind) == firebaseSessionKind {
		tags["firebase"] = struct{}{}
		tags["rtdb"] = struct{}{}
		tags["firebase_database"] = struct{}{}
	}
	if u, err := (&urlParser{}).parse(v.Target); err == nil {
		host := strings.ToLower(u.Host)
		if i := strings.Index(host, ":"); i > 0 {
			host = host[:i]
		}
		if strings.HasSuffix(host, "firebaseio.com") || strings.HasSuffix(host, "firebasedatabase.app") {
			tags["firebase"] = struct{}{}
		}
	}
	// mime categories
	if v.HttpMeta != nil {
		mime := strings.ToLower(v.HttpMeta.Mime)
		if mime != "" {
			if strings.Contains(mime, "json") {
				tags["json"] = struct{}{}
			}
			if strings.Contains(mime, "x-www-form-urlencoded") || strings.Contains(mime, "multipart/form-data") {
				tags["form"] = struct{}{}
			}
			if strings.Contains(mime, "xml") {
				tags["xml"] = struct{}{}
			}
			if strings.Contains(mime, "javascript") || strings.HasSuffix(mime, "/js") {
				tags["js"] = struct{}{}
			}
			if strings.Contains(mime, "css") {
				tags["css"] = struct{}{}
			}
			if strings.Contains(mime, "graphql") {
				tags["graphql"] = struct{}{}
			}
			if strings.HasPrefix(mime, "image/") || strings.HasPrefix(mime, "audio/") || strings.HasPrefix(mime, "video/") || strings.HasPrefix(mime, "font/") {
				tags["media"] = struct{}{}
			}
			if strings.Contains(mime, "html") || strings.Contains(mime, "pdf") || strings.Contains(mime, "rtf") || strings.HasPrefix(mime, "text/plain") {
				tags["document"] = struct{}{}
			}
		}
	}
	if len(tags) == 0 {
		tags["other"] = struct{}{}
	}
	return tags
}

func matchesAnyStatusGroup(groups []string, meta *httpMetaV1) bool {
	if len(groups) == 0 {
		return true
	}
	if meta == nil {
		return false
	}
	st := meta.Status
	for _, g := range groups {
		switch g {
		case "1xx":
			if st >= 100 && st <= 199 {
				return true
			}
		case "2xx":
			if st >= 200 && st <= 299 {
				return true
			}
		case "3xx":
			if st >= 300 && st <= 399 {
				return true
			}
		case "4xx":
			if st >= 400 && st <= 499 {
				return true
			}
		case "5xx":
			if st >= 500 && st <= 599 {
				return true
			}
		}
	}
	return false
}

// Minimal URL parser wrapper to keep import surface small
type urlParser struct{}

func (*urlParser) parse(raw string) (*urlURL, error) {
	// we reuse net/url without exporting to file header to avoid changing imports here
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return (*urlURL)(u), nil
}

type urlURL url.URL

// --- GraphQL deep detection cache (bound to frame count)
var gqlBodyCache sync.Map // map[string]gqlBodyEntry

type gqlBodyEntry struct {
	has        bool
	framesSeen int
}

func detectGraphQLByBody(ctx context.Context, d *Deps, sessionID string) bool {
	if frames, _, err := d.Svc.ListFrames(ctx, sessionID, "", 1000); err == nil {
		// cache check by frames count
		if v, ok := gqlBodyCache.Load(sessionID); ok {
			e := v.(gqlBodyEntry)
			if e.framesSeen == len(frames) {
				return e.has
			}
		}
		// scan from the end to find latest http_request
		has := false
		for i := len(frames) - 1; i >= 0; i-- {
			var prev map[string]any
			if err := json.Unmarshal([]byte(frames[i].Preview), &prev); err != nil {
				continue
			}
			if t, _ := prev["type"].(string); t != "http_request" {
				continue
			}
			// quick header check
			if h, ok := prev["headers"].(map[string]any); ok {
				for k, v := range h {
					lk := strings.ToLower(k)
					if lk == "content-type" {
						if strings.Contains(strings.ToLower(v.(string)), "graphql") {
							has = true
							break
						}
					}
				}
			}
			if has {
				break
			}
			// inspect body (compacted JSON string if recognized)
			body, _ := prev["body"].(string)
			if body != "" {
				var obj map[string]any
				if json.Unmarshal([]byte(body), &obj) == nil {
					if _, ok := obj["query"]; ok {
						has = true
					} else if _, ok := obj["operationName"]; ok {
						has = true
					}
				}
			}
			break // only latest request
		}
		gqlBodyCache.Store(sessionID, gqlBodyEntry{has: has, framesSeen: len(frames)})
		return has
	}
	return false
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
		key := normalizeHost(s.Target)
		if key == "" {
			key = "unknown"
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
		d.Logger.Info().Bool("recording", rec).Int("current", cur).Msg("capture GET")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp{Recording: rec, Current: cur})
	case http.MethodPost:
		var body struct {
			Action string `json:"action"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		act := strings.ToLower(body.Action)
		d.Logger.Info().Str("action", act).Msg("capture POST")
		switch act {
		case "start":
			beforeRec, beforeCur := mem.RecordingState()
			cur := mem.StartCapture()
			afterRec, afterCur := mem.RecordingState()
			d.Logger.Info().
				Bool("before_recording", beforeRec).Int("before_current", beforeCur).
				Bool("after_recording", afterRec).Int("after_current", afterCur).
				Msg("capture START applied")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp{Recording: true, Current: cur})
		case "stop":
			beforeRec, beforeCur := mem.RecordingState()
			cur := mem.StopCapture()
			afterRec, afterCur := mem.RecordingState()
			d.Logger.Info().
				Bool("before_recording", beforeRec).Int("before_current", beforeCur).
				Bool("after_recording", afterRec).Int("after_current", afterCur).
				Msg("capture STOP applied")
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

// handleV1CaptureReset clears all sessions and starts a new capture
func (d *Deps) handleV1CaptureReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "use POST", nil)
		return
	}

	// Get memory store
	repo := sessionsRepoOf(d.Svc)
	mem, ok := repo.(*mem.Store)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "CAPTURE_UNAVAILABLE", "capture unsupported", nil)
		return
	}

	// Clear all sessions
	if err := d.Svc.ClearAll(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "CLEAR_FAILED", err.Error(), nil)
		return
	}

	// Close live connections
	if d.Live != nil {
		d.Live.CloseAll()
	}

	// Broadcast clear event
	if d.Monitor != nil {
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sessions_cleared", ID: "*"})
	}

	// Start new capture
	newCapture := mem.StartCapture()

	d.Logger.Info().Int("capture", newCapture).Msg("capture reset")

	// Return new capture state
	type resp struct {
		Recording bool `json:"recording"`
		Current   int  `json:"current"`
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp{Recording: true, Current: newCapture})
}

// resetCaptureBeforeRequest resets capture when _resetCapture=true query param is present.
// Called from handleHTTPProxy before proxying the first request.
func (d *Deps) resetCaptureBeforeRequest() {
	repo := sessionsRepoOf(d.Svc)
	mem, ok := repo.(*mem.Store)
	if !ok {
		return
	}

	// Clear all sessions
	_ = d.Svc.ClearAll(context.Background())

	// Close live connections
	if d.Live != nil {
		d.Live.CloseAll()
	}

	// Broadcast clear event
	if d.Monitor != nil {
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sessions_cleared", ID: "*"})
	}

	// Start new capture
	newCapture := mem.StartCapture()
	d.Logger.Info().Int("capture", newCapture).Msg("capture reset via _resetCapture param")
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
	Method           string            `json:"method"`
	Status           int               `json:"status"`
	Mime             string            `json:"mime"`
	DurationMs       int64             `json:"durationMs"`
	Streaming        bool              `json:"streaming"`
	Headers          map[string]string `json:"headers"`
	Cache            *cacheMetaV1      `json:"cache,omitempty"`
	CORS             *corsMetaV1       `json:"cors,omitempty"`
	Preflight        *preflightLinkV1  `json:"preflight,omitempty"`
	ErrorCategory    string            `json:"errorCategory,omitempty"`
	ErrorCode        string            `json:"errorCode,omitempty"`
	ErrorUserMessage string            `json:"errorUserMessage,omitempty"`
	ErrorMessage     string            `json:"errorMessage,omitempty"`
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

// Rough classification of network errors for UI
func classifyNetError(msg string) string {
	return classifyProxyErrorString(msg).Code
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
			if strings.EqualFold(a, r) {
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
	for k, v := range h {
		if strings.EqualFold(k, key) {
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

// computeQuickHTTPMetaFromFrames — quick best-effort: extract method/status/mime from frame previews,
// so frontend sees base meta before transaction is recorded.
func (d *Deps) computeQuickHTTPMetaFromFrames(ctx context.Context, sessionID string) *httpMetaV1 {
	frames, _, _ := d.Svc.ListFrames(ctx, sessionID, "", 1000)
	if len(frames) == 0 {
		return nil
	}
	var method string
	var status int
	var mime string
	var reqTs, respTs time.Time
	// from end to beginning search for http_response with status and type
	for i := len(frames) - 1; i >= 0; i-- {
		var m map[string]any
		if err := json.Unmarshal([]byte(frames[i].Preview), &m); err != nil {
			continue
		}
		if t, _ := m["type"].(string); t == "http_response" {
			respTs = frames[i].Ts
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
	// forward search for first http_request with method
	for i := 0; i < len(frames); i++ {
		var m map[string]any
		if err := json.Unmarshal([]byte(frames[i].Preview), &m); err != nil {
			continue
		}
		if t, _ := m["type"].(string); t == "http_request" {
			reqTs = frames[i].Ts
			if meth, ok := m["method"].(string); ok {
				method = meth
			}
			break
		}
	}
	if method == "" && status == 0 && mime == "" {
		return nil
	}
	// Calculate duration from frame timestamps
	var durationMs int64
	if !reqTs.IsZero() && !respTs.IsZero() {
		dur := respTs.Sub(reqTs)
		durationMs = dur.Milliseconds()
		if durationMs < 0 {
			durationMs = 0
		}
		// If duration > 0 but < 1ms, round to 1ms (like in durationMs function)
		if durationMs == 0 && dur > 0 {
			durationMs = 1
		}
	}
	out := &httpMetaV1{Method: method, Status: status, Mime: mime, DurationMs: durationMs}
	return out
}

// headerGetCaseInsensitive returns header value regardless of key case.
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
	// fast variants for exact key match
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

// enrichWithHTTPMeta — common enrichment logic for DRY (used in REST API and WebSocket).
// Tries to get httpMeta from HTTPTransaction, if failed — fallback to extracting from frames.
func (d *Deps) enrichWithHTTPMeta(ctx context.Context, sess domain.Session) (*httpMetaV1, *sizeInfoV1) {
	// 1. Try to get from HTTPTransaction
	meta, sz := d.computeHTTPMeta(ctx, sess.ID)

	// 2. If failed and there's an error — return error meta
	if meta == nil && sess.Error != nil {
		info := classifyProxyErrorString(*sess.Error)
		meta = &httpMetaV1{
			Method:           "",
			Status:           0,
			Mime:             "",
			DurationMs:       0,
			Streaming:        false,
			Headers:          map[string]string{},
			ErrorCategory:    info.Category,
			ErrorCode:        info.Code,
			ErrorUserMessage: info.UserMessage,
			ErrorMessage:     info.UserMessage,
		}
		return meta, sz
	}

	// 3. If failed and no error — fallback to extracting from frames
	if meta == nil {
		meta = d.computeQuickHTTPMetaFromFrames(ctx, sess.ID)
	}

	return meta, sz
}

// handleFrameBody serves the raw body content for a specific frame
func (d *Deps) handleFrameBody(w http.ResponseWriter, r *http.Request, sessionID string, frameID string) {
	// Use direct GetFrameByID for O(1) lookup instead of O(n) ListFrames
	frame, found, err := d.Svc.GetFrameByID(r.Context(), sessionID, frameID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "FRAME_GET_FAILED", err.Error(), map[string]any{"sessionId": sessionID, "frameId": frameID})
		return
	}

	if !found {
		writeError(w, http.StatusNotFound, "FRAME_NOT_FOUND", "frame not found", map[string]any{"frameId": frameID})
		return
	}

	targetFrame := &frame

	// Check if BodyFile exists
	if targetFrame.BodyFile == "" {
		// No body file - return preview as fallback
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Body-Source", "preview")
		if _, err := w.Write([]byte(targetFrame.Preview)); err != nil {
			// Log error but response already started
			log.Printf("Error writing preview response: %v", err)
		}
		return
	}

	// Check file size before reading to prevent OOM
	fileInfo, err := os.Stat(targetFrame.BodyFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File was cleaned up by GC or DeleteSession
			writeError(w, http.StatusGone, "BODY_FILE_EXPIRED", "Body file no longer available (cleaned up)", map[string]any{"frameId": frameID})
			return
		}
		writeError(w, http.StatusInternalServerError, "BODY_STAT_FAILED", err.Error(), map[string]any{"frameId": frameID, "bodyFile": targetFrame.BodyFile})
		return
	}

	const maxBodySize = 100 * 1024 * 1024 // 100 MB limit
	if fileInfo.Size() >= maxBodySize {
		writeError(w, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE",
			fmt.Sprintf("body file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileInfo.Size(), maxBodySize),
			map[string]any{"frameId": frameID, "fileSize": fileInfo.Size(), "maxSize": maxBodySize})
		return
	}

	// Read from BodyFile
	bodyData, err := os.ReadFile(targetFrame.BodyFile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BODY_READ_FAILED", err.Error(), map[string]any{"frameId": frameID, "bodyFile": targetFrame.BodyFile})
		return
	}

	// Return raw bytes
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Body-Source", "file")
	w.Header().Set("X-Frame-Id", frameID)
	w.Header().Set("Content-Length", strconv.Itoa(len(bodyData)))
	if _, err := w.Write(bodyData); err != nil {
		// Log error but response already started
		log.Printf("Error writing body response: %v", err)
	}
}
