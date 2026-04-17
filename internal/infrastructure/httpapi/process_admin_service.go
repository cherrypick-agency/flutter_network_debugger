package httpapi

import (
	"context"
	"net/http"

	"network-debugger/internal/features/process/domain"
)

type processHelperStatusDTO struct {
	Running   bool   `json:"running"`
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
}

type processInstallResponseDTO struct {
	Status string `json:"status"`
}

type processAdminService struct {
	d *Deps
}

func newProcessAdminService(d *Deps) processAdminService {
	return processAdminService{d: d}
}

func (s processAdminService) ensureAvailable() *sessionAPIError {
	if s.d.ProcessSvc == nil {
		return &sessionAPIError{
			Status:  http.StatusServiceUnavailable,
			Code:    "FEATURE_DISABLED",
			Message: "process feature not available",
		}
	}
	return nil
}

func (s processAdminService) loadConfig(ctx context.Context) (processConfigDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return processConfigDTO{}, apiErr
	}
	cfg, err := s.d.ProcessSvc.GetConfig(ctx)
	if err != nil {
		return processConfigDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "CONFIG_LOAD_FAILED",
			Message: err.Error(),
		}
	}
	return processConfigFromDomain(cfg), nil
}

func (s processAdminService) saveConfig(ctx context.Context, dto processConfigDTO) (processConfigDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return processConfigDTO{}, apiErr
	}
	cfg := &domain.DetectionConfig{
		ID:              1,
		Enabled:         dto.Enabled,
		UseHelperTool:   dto.UseHelperTool,
		HelperInstalled: dto.HelperInstalled,
		CacheEnabled:    dto.CacheEnabled,
		CacheTTLSeconds: dto.CacheTTL,
		FallbackEnabled: dto.FallbackEnabled,
	}
	if err := s.d.ProcessSvc.SaveConfig(ctx, cfg); err != nil {
		return processConfigDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "CONFIG_SAVE_FAILED",
			Message: err.Error(),
		}
	}
	return processConfigFromDomain(cfg), nil
}

func (s processAdminService) helperStatus() (processHelperStatusDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return processHelperStatusDTO{}, apiErr
	}
	status := s.d.ProcessSvc.CheckHelperStatus()
	return processHelperStatusDTO{
		Running:   status.Running,
		Installed: status.Installed,
		Version:   status.Version,
	}, nil
}

func (s processAdminService) installHelper(ctx context.Context) (processInstallResponseDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return processInstallResponseDTO{}, apiErr
	}
	if err := s.d.ProcessSvc.InstallHelper(ctx); err != nil {
		return processInstallResponseDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "INSTALL_FAILED",
			Message: err.Error(),
		}
	}
	return processInstallResponseDTO{Status: "installed"}, nil
}

func processConfigFromDomain(cfg *domain.DetectionConfig) processConfigDTO {
	if cfg == nil {
		return processConfigDTO{}
	}
	return processConfigDTO{
		Enabled:         cfg.Enabled,
		UseHelperTool:   cfg.UseHelperTool,
		HelperInstalled: cfg.HelperInstalled,
		CacheEnabled:    cfg.CacheEnabled,
		CacheTTL:        cfg.CacheTTLSeconds,
		FallbackEnabled: cfg.FallbackEnabled,
	}
}
