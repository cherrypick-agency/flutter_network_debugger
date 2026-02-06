package usecase

import (
	"context"
	"fmt"
	"log"

	"network-debugger/internal/features/intercept/domain"
	"network-debugger/pkg/shared/id"
)

const maxInterceptRules = 1000

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
	log.Printf("[InterceptService] UpdateRules: %d rules", len(rules))
	if len(rules) > maxInterceptRules {
		return fmt.Errorf("too many rules: %d (max %d)", len(rules), maxInterceptRules)
	}
	for i := range rules {
		rules[i].SetDefaults()
		if rules[i].ID == "" {
			rules[i].ID = id.New()
		}
		if err := rules[i].Validate(); err != nil {
			log.Printf("[InterceptService] UpdateRules: rule[%d] validation error: %v", i, err)
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
	log.Printf("[InterceptService] UpdateConfig: enabled=%v requests=%v responses=%v timeoutMs=%d queueMax=%d overflow=%s",
		cfg.Enabled, cfg.Requests, cfg.Responses, cfg.TimeoutMs, cfg.QueueMax, cfg.Overflow)
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
	log.Printf("[InterceptService] LoadAndApplyFromDB: config loaded, enabled=%v", cfg.Enabled)
	s.manager.UpdateConfig(cfg)

	rules, err := s.ruleRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("load rules: %w", err)
	}
	log.Printf("[InterceptService] LoadAndApplyFromDB: %d rules loaded", len(rules))
	s.manager.UpdateRules(rules)
	return nil
}
