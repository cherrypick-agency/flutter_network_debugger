package domain

import "context"

// MapRulesRepository — repository port for rules
type MapRulesRepository interface {
	List(ctx context.Context) ([]MapRule, error)
	Upsert(ctx context.Context, r MapRule) (MapRule, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) error
}
