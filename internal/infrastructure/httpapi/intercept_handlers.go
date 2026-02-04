package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (d *Deps) interceptAuthOK(r *http.Request) bool {
	// If AdminToken is empty and server listens on loopback — allow
	if d.Cfg.AdminToken == "" {
		// best-effort: trust local development
		return isLoopback(r.RemoteAddr)
	}
	tok := r.Header.Get("X-Admin-Token")
	return tok != "" && tok == d.Cfg.AdminToken
}

func (d *Deps) handleInterceptRules(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if r.Method == http.MethodGet {
		rules := d.Interceptor.ListRules()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rules)
		return
	}
	if r.Method == http.MethodPost {
		var rules []InterceptRule
		if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		d.Interceptor.UpdateRules(rules)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (d *Deps) handleInterceptConfig(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	type cfgDTO struct {
		Enabled      bool   `json:"enabled"`
		Requests     bool   `json:"requests"`
		Responses    bool   `json:"responses"`
		TimeoutMs    int    `json:"timeoutMs"`
		QueueMax     int    `json:"queueMax"`
		BodyMaxBytes int    `json:"bodyMaxBytes"`
		Reencode     bool   `json:"reencode"`
		Overflow     string `json:"overflow"`
	}
	if r.Method == http.MethodGet {
		out := cfgDTO{
			Enabled:      d.Cfg.InterceptEnabled,
			Requests:     d.Cfg.InterceptRequests,
			Responses:    d.Cfg.InterceptResponses,
			TimeoutMs:    d.Cfg.InterceptTimeoutMs,
			QueueMax:     d.Cfg.InterceptQueueMax,
			BodyMaxBytes: d.Cfg.InterceptBodyMaxBytes,
			Reencode:     d.Cfg.InterceptReencode,
			Overflow:     d.Cfg.InterceptOverflow,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	}
	if r.Method == http.MethodPost {
		var in cfgDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		d.Cfg.InterceptEnabled = in.Enabled
		d.Cfg.InterceptRequests = in.Requests
		d.Cfg.InterceptResponses = in.Responses
		if in.TimeoutMs > 0 {
			d.Cfg.InterceptTimeoutMs = in.TimeoutMs
		}
		if in.QueueMax >= 0 {
			d.Cfg.InterceptQueueMax = in.QueueMax
		}
		if in.BodyMaxBytes > 0 {
			d.Cfg.InterceptBodyMaxBytes = in.BodyMaxBytes
		}
		d.Cfg.InterceptReencode = in.Reencode
		if strings.TrimSpace(in.Overflow) != "" {
			d.Cfg.InterceptOverflow = in.Overflow
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (d *Deps) handleInterceptPending(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	limit := 0
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	items := d.Interceptor.ListPending(limit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (d *Deps) handleInterceptItem(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	base := "/_api/v1/intercept/items/"
	if !strings.HasPrefix(r.URL.Path, base) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "", nil)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, base)
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "missing id", nil)
		return
	}
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if r.Method == http.MethodGet && action == "" {
		if it, ok := d.Interceptor.PeekItem(id); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(it)
			return
		}
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no such pending item", nil)
		return
	}

	if r.Method == http.MethodPost && action == "continue" {
		it, ok := d.Interceptor.PeekItem(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "no such pending item", nil)
			return
		}
		if it.Direction == DirRequest {
			var payload struct {
				Action  string      `json:"action"`
				Method  string      `json:"method"`
				URL     string      `json:"url"`
				Headers http.Header `json:"headers"`
				BodyB64 *string     `json:"bodyBase64"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
				return
			}
			var body []byte
			if payload.BodyB64 != nil {
				if b, err := base64.StdEncoding.DecodeString(*payload.BodyB64); err == nil {
					body = b
				}
			}
			ok := d.Interceptor.ContinueRequest(id, &HTTPRequestDecision{Action: payload.Action, Method: payload.Method, URL: payload.URL, Headers: payload.Headers, Body: body})
			if !ok {
				writeError(w, http.StatusConflict, "CONFLICT", "already finalized", nil)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// response
		var payload struct {
			Action  string      `json:"action"`
			Status  int         `json:"status"`
			Headers http.Header `json:"headers"`
			BodyB64 *string     `json:"bodyBase64"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		var body []byte
		if payload.BodyB64 != nil {
			if b, err := base64.StdEncoding.DecodeString(*payload.BodyB64); err == nil {
				body = b
			}
		}
		ok = d.Interceptor.ContinueResponse(id, &HTTPResponseDecision{Action: payload.Action, Status: payload.Status, Headers: payload.Headers, Body: body})
		if !ok {
			writeError(w, http.StatusConflict, "CONFLICT", "already finalized", nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodPost && action == "cancel" {
		if ok := d.Interceptor.Cancel(id); !ok {
			writeError(w, http.StatusConflict, "CONFLICT", "already finalized", nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
