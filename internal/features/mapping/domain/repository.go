package domain

import "context"

// MapRulesRepository — порт репозитория правил
type MapRulesRepository interface {
	List(ctx context.Context) ([]MapRule, error)
	Upsert(ctx context.Context, r MapRule) (MapRule, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) error
}
