package usecase

import (
	"context"
	"time"

	sdomain "network-debugger/internal/features/settings/domain"
	"network-debugger/internal/infrastructure/config"
)

type Service struct {
	settingsRepo sdomain.SettingsRepository
	profilesRepo sdomain.ThrottleProfilesRepository
}

func NewService(sr sdomain.SettingsRepository, pr sdomain.ThrottleProfilesRepository) *Service {
	return &Service{settingsRepo: sr, profilesRepo: pr}
}

func (s *Service) Load(ctx context.Context) (sdomain.RuntimeSettings, error) {
	return s.settingsRepo.Load(ctx)
}

func (s *Service) SaveRuntime(ctx context.Context, rs sdomain.RuntimeSettings) (sdomain.RuntimeSettings, error) {
	if rs.UpdatedAt.IsZero() {
		rs.UpdatedAt = time.Now().UTC()
	}
	if err := s.settingsRepo.Save(ctx, rs); err != nil {
		return sdomain.RuntimeSettings{}, err
	}
	return s.settingsRepo.Load(ctx)
}

func (s *Service) ListProfiles(ctx context.Context) ([]sdomain.ThrottleProfile, error) {
	return s.profilesRepo.List(ctx)
}

func (s *Service) UpsertProfile(ctx context.Context, p sdomain.ThrottleProfile) (sdomain.ThrottleProfile, error) {
	return s.profilesRepo.Upsert(ctx, p)
}

func (s *Service) DeleteProfile(ctx context.Context, id string) error {
	return s.profilesRepo.Delete(ctx, id)
}

// ApplyOverlay applies settings from DB to the runtime config.
func ApplyOverlay(cfg *config.Config, rs sdomain.RuntimeSettings) {
	if cfg == nil {
		return
	}
	// response delay
	if rs.ResponseDelayMinMs > 0 && rs.ResponseDelayMaxMs > 0 {
		cfg.ResponseDelayMinMs = rs.ResponseDelayMinMs
		cfg.ResponseDelayMaxMs = rs.ResponseDelayMaxMs
		cfg.ResponseDelayMs = 0
	} else if rs.ResponseDelayMs >= 0 { // 0 — valid for disabling
		cfg.ResponseDelayMs = rs.ResponseDelayMs
		cfg.ResponseDelayMinMs = 0
		cfg.ResponseDelayMaxMs = 0
	}
	// throttle
	cfg.ThrottleEnabled = rs.ThrottleEnabled
	cfg.ThrottleDownKbps = rs.ThrottleDownKbps
	cfg.ThrottleUpKbps = rs.ThrottleUpKbps
	cfg.ThrottlePacketLoss = rs.ThrottlePacketLoss
	cfg.ThrottleOffline = rs.ThrottleOffline

	// mapping
	cfg.MappingEnabled = rs.MappingEnabled
	if rs.MappingUploadMaxMB > 0 {
		cfg.MappingUploadMaxMB = rs.MappingUploadMaxMB
	}
}
