package httpapi

import (
	"encoding/json"
	"net/http"
	sdomain "network-debugger/internal/features/settings/domain"
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

func (d *Deps) handleV1Throttle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		curEnabled := d.Cfg.ThrottleEnabled
		curDown := d.Cfg.ThrottleDownKbps
		curUp := d.Cfg.ThrottleUpKbps
		curLoss := d.Cfg.ThrottlePacketLoss
		curLatency := d.Cfg.ThrottleLatencyMs
		curJitter := d.Cfg.ThrottleLatencyJitter
		curOffline := d.Cfg.ThrottleOffline
		if d.Settings != nil {
			if rs, err := d.Settings.Load(contextWithNoCancel()); err == nil {
				curEnabled = rs.ThrottleEnabled
				curDown = rs.ThrottleDownKbps
				curUp = rs.ThrottleUpKbps
				curLoss = rs.ThrottlePacketLoss
				curLatency = rs.ThrottleLatencyMs
				curJitter = rs.ThrottleLatencyJitter
				curOffline = rs.ThrottleOffline
			}
		}
		out := throttleDTO{
			Enabled:       curEnabled,
			DownKbps:      curDown,
			UpKbps:        curUp,
			PacketLossPct: curLoss,
			LatencyMs:     curLatency,
			LatencyJitter: curJitter,
			Offline:       curOffline,
			Presets: []any{
				map[string]any{"name": "No throttling", "downKbps": 0, "upKbps": 0, "packetLossPct": 0, "latencyMs": 0, "latencyJitter": 0},
				map[string]any{"name": "3G", "downKbps": 400, "upKbps": 400, "packetLossPct": 0, "latencyMs": 100, "latencyJitter": 20},
				map[string]any{"name": "Slow 4G", "downKbps": 1500, "upKbps": 1500, "packetLossPct": 0, "latencyMs": 50, "latencyJitter": 10},
				map[string]any{"name": "Fast 4G", "downKbps": 3000, "upKbps": 3000, "packetLossPct": 0, "latencyMs": 30, "latencyJitter": 5},
				map[string]any{"name": "Satellite", "downKbps": 2000, "upKbps": 1000, "packetLossPct": 2, "latencyMs": 600, "latencyJitter": 50},
				map[string]any{"name": "Offline", "offline": true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	case http.MethodPost:
		var in throttleDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		d.Cfg.ThrottleEnabled = in.Enabled
		d.Cfg.ThrottleDownKbps = in.DownKbps
		d.Cfg.ThrottleUpKbps = in.UpKbps
		d.Cfg.ThrottlePacketLoss = in.PacketLossPct
		d.Cfg.ThrottleLatencyMs = in.LatencyMs
		d.Cfg.ThrottleLatencyJitter = in.LatencyJitter
		d.Cfg.ThrottleOffline = in.Offline
		if d.Settings != nil {
			cur, _ := d.Settings.Load(contextWithNoCancel())
			cur.ThrottleEnabled = d.Cfg.ThrottleEnabled
			cur.ThrottleDownKbps = d.Cfg.ThrottleDownKbps
			cur.ThrottleUpKbps = d.Cfg.ThrottleUpKbps
			cur.ThrottlePacketLoss = d.Cfg.ThrottlePacketLoss
			cur.ThrottleLatencyMs = d.Cfg.ThrottleLatencyMs
			cur.ThrottleLatencyJitter = d.Cfg.ThrottleLatencyJitter
			cur.ThrottleOffline = d.Cfg.ThrottleOffline
			cur.ResponseDelayMs = d.Cfg.ResponseDelayMs
			cur.ResponseDelayMinMs = d.Cfg.ResponseDelayMinMs
			cur.ResponseDelayMaxMs = d.Cfg.ResponseDelayMaxMs
			_, _ = d.Settings.SaveRuntime(contextWithNoCancel(), cur)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(throttleDTO{
			Enabled:       d.Cfg.ThrottleEnabled,
			DownKbps:      d.Cfg.ThrottleDownKbps,
			UpKbps:        d.Cfg.ThrottleUpKbps,
			PacketLossPct: d.Cfg.ThrottlePacketLoss,
			LatencyMs:     d.Cfg.ThrottleLatencyMs,
			LatencyJitter: d.Cfg.ThrottleLatencyJitter,
			Offline:       d.Cfg.ThrottleOffline,
		})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// handleV1ThrottleProfiles: GET list, POST upsert, DELETE /{id}
func (d *Deps) handleV1ThrottleProfiles(w http.ResponseWriter, r *http.Request) {
	if d.Settings == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{})
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, _ := d.Settings.ListProfiles(contextWithNoCancel())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
		return
	case http.MethodPost:
		var in struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			DownKbps      int    `json:"downKbps"`
			UpKbps        int    `json:"upKbps"`
			PacketLossPct int    `json:"packetLossPct"`
			LatencyMs     int    `json:"latencyMs"`
			LatencyJitter int    `json:"latencyJitter"`
			Offline       bool   `json:"offline"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		p, err := d.Settings.UpsertProfile(contextWithNoCancel(), sdomain.ThrottleProfile{
			ID: in.ID, Name: in.Name, DownKbps: in.DownKbps, UpKbps: in.UpKbps, PacketLossPct: in.PacketLossPct, LatencyMs: in.LatencyMs, LatencyJitter: in.LatencyJitter, Offline: in.Offline,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, "UPSERT_FAILED", err.Error(), nil)
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
		if err := d.Settings.DeleteProfile(contextWithNoCancel(), id); err != nil {
			writeError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error(), nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
