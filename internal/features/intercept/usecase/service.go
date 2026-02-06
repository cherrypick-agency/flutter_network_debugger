package usecase

import (
	"context"
	"fmt"

	"network-debugger/internal/features/intercept/domain"
	"network-debugger/pkg/shared/id"
)

// InterceptService orchestrates intercept feature use cases
type InterceptService struct {
	manager  *InterceptorManager
	ruleRepo domain.RuleRepository
	cfgRepo  domain.ConfigRepository
}

// NewInterceptService creates a new InterceptService
func NewInterceptService(manager *InterceptorManager, ruleRepo domain.RuleRepository, cfgRepo domain.ConfigRepository) *InterceptService {
	return &InterceptService{
		manager:  manager,
		ruleRepo: ruleRepo,
		cfgRepo:  cfgRepo,
	}
}

// Manager exposes the underlying InterceptorManager for the proxy layer
func (s *InterceptService) Manager() *InterceptorManager {
	return s.manager
}

// UpdateRules validates, persists and applies rules
func (s *InterceptService) UpdateRules(ctx context.Context, rules []domain.InterceptRule) error {
	for i := range rules {
		rules[i].SetDefaults()
		if rules[i].ID == "" {
			rules[i].ID = id.New()
		}
		if err := rules[i].Validate(); err != nil {
			return fmt.Errorf("rule[%d]: %w", i, err)
		}
	}
	if s.ruleRepo != nil {
		if err := s.ruleRepo.SaveAll(ctx, rules); err != nil {
			return fmt.Errorf("persist rules: %w", err)
		}
	}
	s.manager.UpdateRules(rules)
	return nil
}

// ListRules returns current rules from the manager (in-memory)
func (s *InterceptService) ListRules() []domain.InterceptRule {
	return s.manager.ListRules()
}

// LoadConfig loads configuration from the repository
func (s *InterceptService) LoadConfig(ctx context.Context) (domain.InterceptConfig, error) {
	if s.cfgRepo == nil {
		return s.manager.Config(), nil
	}
	return s.cfgRepo.Load(ctx)
}

// UpdateConfig validates, persists and applies configuration
func (s *InterceptService) UpdateConfig(ctx context.Context, cfg domain.InterceptConfig) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	if s.cfgRepo != nil {
		if err := s.cfgRepo.Save(ctx, cfg); err != nil {
			return fmt.Errorf("persist config: %w", err)
		}
	}
	s.manager.UpdateConfig(cfg)
	return nil
}

// LoadAndApplyFromDB loads rules and config from DB and applies to manager (called at startup)
func (s *InterceptService) LoadAndApplyFromDB(ctx context.Context) error {
	if s.cfgRepo == nil || s.ruleRepo == nil {
		return nil
	}
	cfg, err := s.cfgRepo.Load(ctx)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	s.manager.UpdateConfig(cfg)

	rules, err := s.ruleRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("load rules: %w", err)
	}
	s.manager.UpdateRules(rules)
	return nil
}
