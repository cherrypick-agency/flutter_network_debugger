package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	interceptdomain "network-debugger/internal/features/intercept/domain"
)

const interceptMaxBodySize = 2 * 1024 * 1024 // 2MB for admin JSON payloads

func (d *Deps) interceptAuthOK(r *http.Request) bool {
	// If AdminToken is empty and server listens on loopback — allow
	if d.Cfg.AdminToken == "" {
		// best-effort: trust local development
		return isLoopback(r.RemoteAddr)
	}
	tok := r.Header.Get("X-Admin-Token")
	return tok != "" && subtle.ConstantTimeCompare([]byte(tok), []byte(d.Cfg.AdminToken)) == 1
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
	admin := newInterceptAdminService(d)
	if r.Method == http.MethodGet {
		rules, apiErr := admin.listRules()
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(rules); err != nil {
			log.Printf("[InterceptHandler] encode error: %v", err)
		}
		return
	}
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, interceptMaxBodySize)
		var rules []interceptdomain.InterceptRule
		if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		if apiErr := admin.saveRules(r.Context(), rules); apiErr != nil {
			writeSessionAPIError(w, apiErr)
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
	admin := newInterceptAdminService(d)
	if r.Method == http.MethodGet {
		cfg, apiErr := admin.loadConfig()
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cfg); err != nil {
			log.Printf("[InterceptHandler] encode error: %v", err)
		}
		return
	}
	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, interceptMaxBodySize)
		var cfg interceptdomain.InterceptConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		if apiErr := admin.saveConfig(r.Context(), cfg); apiErr != nil {
			writeSessionAPIError(w, apiErr)
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
	admin := newInterceptAdminService(d)
	limit, apiErr := admin.parsePendingLimit(r.URL.Query().Get("limit"))
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
	items, apiErr := admin.listPending(limit)
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("[InterceptHandler] handleInterceptPending: encode error: %v", err)
	}
}

func (d *Deps) handleInterceptItem(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	admin := newInterceptAdminService(d)
	itemPath, apiErr := admin.parseItemPath(r.URL.Path)
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
	if r.Method == http.MethodGet && itemPath.Action == "" {
		item, apiErr := admin.pendingItem(itemPath.ID)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(item); err != nil {
			log.Printf("[InterceptHandler] handleInterceptItem: encode error: %v", err)
		}
		return
	}

	if r.Method == http.MethodPost && itemPath.Action == "continue" {
		item, apiErr := admin.pendingItem(itemPath.ID)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		if item.Direction == interceptdomain.DirRequest {
			var payload interceptContinueRequestPayload
			r.Body = http.MaxBytesReader(w, r.Body, interceptMaxBodySize)
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
				return
			}
			if apiErr := admin.continueRequest(itemPath.ID, payload); apiErr != nil {
				writeSessionAPIError(w, apiErr)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var payload interceptContinueResponsePayload
		r.Body = http.MaxBytesReader(w, r.Body, interceptMaxBodySize)
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
			return
		}
		if apiErr := admin.continueResponse(itemPath.ID, payload); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodPost && itemPath.Action == "cancel" {
		if apiErr := admin.cancel(itemPath.ID); apiErr != nil {
			writeSessionAPIError(w, apiErr)
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
