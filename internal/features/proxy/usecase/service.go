package usecase

import (
	"context"
	"time"

	pdomain "network-debugger/internal/features/proxy/domain"
	"network-debugger/internal/infrastructure/config"
)

type Service struct{ repo pdomain.ProxyConfigRepository }

func NewService(r pdomain.ProxyConfigRepository) *Service { return &Service{repo: r} }

func (s *Service) Load(ctx context.Context) (pdomain.ProxyConfig, error) { return s.repo.Load(ctx) }

func (s *Service) Save(ctx context.Context, c pdomain.ProxyConfig) (pdomain.ProxyConfig, error) {
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now().UTC()
	}
	if err := s.repo.Save(ctx, c); err != nil {
		return pdomain.ProxyConfig{}, err
	}
	return s.repo.Load(ctx)
}

// ApplyOverlay transfers values to runtime-config for backward compatibility.
func ApplyOverlay(cfg *config.Config, pc pdomain.ProxyConfig) {
	// Currently cfg has no fields for proxy ports - leaving empty for compatibility.
	_ = cfg // for future use
	_ = pc
}
