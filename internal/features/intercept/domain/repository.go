package domain

import "context"

// RuleRepository persists intercept rules
type RuleRepository interface {
	SaveAll(ctx context.Context, rules []InterceptRule) error
	List(ctx context.Context) ([]InterceptRule, error)
}

// ConfigRepository persists intercept configuration
type ConfigRepository interface {
	Load(ctx context.Context) (InterceptConfig, error)
	Save(ctx context.Context, cfg InterceptConfig) error
}
