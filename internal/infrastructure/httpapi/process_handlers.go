package httpapi

import (
	"encoding/json"
	"net/http"

	"network-debugger/internal/features/process/domain"
)

// processConfigDTO - DTO для конфигурации детекции процессов
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
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		// Получить текущую конфигурацию
		cfg, err := d.ProcessSvc.GetConfig(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dto := processConfigDTO{
			Enabled:         cfg.Enabled,
			UseHelperTool:   cfg.UseHelperTool,
			HelperInstalled: cfg.HelperInstalled,
			CacheEnabled:    cfg.CacheEnabled,
			CacheTTL:        cfg.CacheTTLSeconds,
			FallbackEnabled: cfg.FallbackEnabled,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto)

	case http.MethodPost:
		// Обновить конфигурацию
		var dto processConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Конвертировать DTO в domain entity
		cfg := &domain.DetectionConfig{
			ID:              1, // singleton
			Enabled:         dto.Enabled,
			UseHelperTool:   dto.UseHelperTool,
			HelperInstalled: dto.HelperInstalled,
			CacheEnabled:    dto.CacheEnabled,
			CacheTTLSeconds: dto.CacheTTL,
			FallbackEnabled: dto.FallbackEnabled,
		}

		if err := d.ProcessSvc.SaveConfig(ctx, cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Вернуть обновленную конфигурацию
		updatedDto := processConfigDTO{
			Enabled:         cfg.Enabled,
			UseHelperTool:   cfg.UseHelperTool,
			HelperInstalled: cfg.HelperInstalled,
			CacheEnabled:    cfg.CacheEnabled,
			CacheTTL:        cfg.CacheTTLSeconds,
			FallbackEnabled: cfg.FallbackEnabled,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updatedDto)

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

	status := d.ProcessSvc.CheckHelperStatus()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"running":   status.Running,
		"installed": status.Installed,
		"version":   status.Version,
	})
}

// handleV1ProcessHelperInstall - POST /_api/v1/process/helper/install
func (d *Deps) handleV1ProcessHelperInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	if err := d.ProcessSvc.InstallHelper(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "installed",
	})
}
