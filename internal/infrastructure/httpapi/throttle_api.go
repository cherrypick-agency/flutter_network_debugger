package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type throttleDTO struct {
	Enabled       bool  `json:"enabled"`
	DownKbps      int   `json:"downKbps"`
	UpKbps        int   `json:"upKbps"`
	PacketLossPct int   `json:"packetLossPct"`
	LatencyMs     int   `json:"latencyMs"`     // Base latency (RTT/ping simulation)
	LatencyJitter int   `json:"latencyJitter"` // Random jitter ± ms
	Offline       bool  `json:"offline"`
	Presets       []any `json:"presets,omitempty"`
}

type throttleProfileDTO struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DownKbps      int    `json:"downKbps"`
	UpKbps        int    `json:"upKbps"`
	PacketLossPct int    `json:"packetLossPct"`
	LatencyMs     int    `json:"latencyMs"`
	LatencyJitter int    `json:"latencyJitter"`
	Offline       bool   `json:"offline"`
}

func (d *Deps) handleV1Throttle(w http.ResponseWriter, r *http.Request) {
	service := newThrottleAdminService(d)
	switch r.Method {
	case http.MethodGet:
		out := service.load(contextWithNoCancel())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	case http.MethodPost:
		var in throttleDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		out := service.apply(contextWithNoCancel(), in)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// handleV1ThrottleProfiles: GET list, POST upsert, DELETE /{id}
func (d *Deps) handleV1ThrottleProfiles(w http.ResponseWriter, r *http.Request) {
	service := newThrottleAdminService(d)
	switch r.Method {
	case http.MethodGet:
		items, apiErr := service.listProfiles(contextWithNoCancel())
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
		return
	case http.MethodPost:
		var in throttleProfileDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		p, apiErr := service.upsertProfile(contextWithNoCancel(), in)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p)
		return
	case http.MethodDelete:
		id := strings.TrimPrefix(r.URL.Path, "/_api/v1/throttle/profiles/")
		if id == "" || id == "/" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if apiErr := service.deleteProfile(contextWithNoCancel(), id); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
