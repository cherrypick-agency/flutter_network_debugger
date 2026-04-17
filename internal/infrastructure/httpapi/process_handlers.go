package httpapi

import (
	"encoding/json"
	"net/http"
)

// processConfigDTO - DTO for process detection configuration
type processConfigDTO struct {
	Enabled         bool `json:"enabled"`
	UseHelperTool   bool `json:"useHelperTool"`
	HelperInstalled bool `json:"helperInstalled"`
	CacheEnabled    bool `json:"cacheEnabled"`
	CacheTTL        int  `json:"cacheTtl"`
	FallbackEnabled bool `json:"fallbackEnabled"`
}

// handleV1ProcessConfig - GET/POST /_api/v1/process/config
func (d *Deps) handleV1ProcessConfig(w http.ResponseWriter, r *http.Request) {
	service := newProcessAdminService(d)
	switch r.Method {
	case http.MethodGet:
		dto, apiErr := service.loadConfig(r.Context())
		if apiErr != nil {
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto)

	case http.MethodPost:
		var dto processConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updated, apiErr := service.saveConfig(r.Context(), dto)
		if apiErr != nil {
			http.Error(w, apiErr.Message, apiErr.Status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleV1ProcessHelperStatus - GET /_api/v1/process/helper/status
func (d *Deps) handleV1ProcessHelperStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status, apiErr := newProcessAdminService(d).helperStatus()
	if apiErr != nil {
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// handleV1ProcessHelperInstall - POST /_api/v1/process/helper/install
func (d *Deps) handleV1ProcessHelperInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, apiErr := newProcessAdminService(d).installHelper(r.Context())
	if apiErr != nil {
		http.Error(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
