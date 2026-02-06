package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"

	interceptdomain "network-debugger/internal/features/intercept/domain"
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

// isLoopback detects loopback address for simple auth decision
func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsUnspecified())
}

func (d *Deps) handleInterceptRules(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if r.Method == http.MethodGet {
		rules := d.InterceptSvc.ListRules()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rules)
		return
	}
	if r.Method == http.MethodPost {
		var rules []interceptdomain.InterceptRule
		if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		if err := d.InterceptSvc.UpdateRules(r.Context(), rules); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION", err.Error(), nil)
			return
		}
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
	if r.Method == http.MethodGet {
		cfg := d.InterceptSvc.Manager().Config()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
		return
	}
	if r.Method == http.MethodPost {
		var cfg interceptdomain.InterceptConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		cfg.SetDefaults()
		if err := d.InterceptSvc.UpdateConfig(r.Context(), cfg); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION", err.Error(), nil)
			return
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
	items := d.InterceptSvc.Manager().ListPending(limit)
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

	mgr := d.InterceptSvc.Manager()

	if r.Method == http.MethodGet && action == "" {
		if it, ok := mgr.PeekItem(id); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(it)
			return
		}
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no such pending item", nil)
		return
	}

	if r.Method == http.MethodPost && action == "continue" {
		it, ok := mgr.PeekItem(id)
		if !ok {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "no such pending item", nil)
			return
		}
		if it.Direction == interceptdomain.DirRequest {
			var payload struct {
				Action  string              `json:"action"`
				Method  string              `json:"method"`
				URL     string              `json:"url"`
				Headers map[string][]string `json:"headers"`
				BodyB64 *string             `json:"bodyBase64"`
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
			ok := mgr.ContinueRequest(id, &interceptdomain.HTTPRequestDecision{Action: payload.Action, Method: payload.Method, URL: payload.URL, Headers: payload.Headers, Body: body})
			if !ok {
				writeError(w, http.StatusConflict, "CONFLICT", "already finalized", nil)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// response
		var payload struct {
			Action  string              `json:"action"`
			Status  int                 `json:"status"`
			Headers map[string][]string `json:"headers"`
			BodyB64 *string             `json:"bodyBase64"`
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
		ok = mgr.ContinueResponse(id, &interceptdomain.HTTPResponseDecision{Action: payload.Action, Status: payload.Status, Headers: payload.Headers, Body: body})
		if !ok {
			writeError(w, http.StatusConflict, "CONFLICT", "already finalized", nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodPost && action == "cancel" {
		if ok := mgr.Cancel(id); !ok {
			writeError(w, http.StatusConflict, "CONFLICT", "already finalized", nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

// toRequestMatchInput converts *http.Request into a pure domain.RequestMatchInput
func toRequestMatchInput(r *http.Request) interceptdomain.RequestMatchInput {
	scheme := ""
	host := ""
	port := ""
	path := ""
	if r.URL != nil {
		scheme = r.URL.Scheme
		host = r.URL.Hostname()
		port = r.URL.Port()
		path = r.URL.EscapedPath()
	}
	return interceptdomain.RequestMatchInput{
		Method:      r.Method,
		Scheme:      scheme,
		Host:        host,
		Port:        port,
		Path:        path,
		ContentType: strings.ToLower(r.Header.Get("Content-Type")),
		Headers:     map[string][]string(r.Header),
		BodyPreview: "",
	}
}

// toResponseMatchInput converts *http.Response into a pure domain.ResponseMatchInput
func toResponseMatchInput(resp *http.Response) interceptdomain.ResponseMatchInput {
	return interceptdomain.ResponseMatchInput{
		StatusCode:  resp.StatusCode,
		ContentType: strings.ToLower(resp.Header.Get("Content-Type")),
		Headers:     map[string][]string(resp.Header),
		BodyPreview: "",
	}
}
