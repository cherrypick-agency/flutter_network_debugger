package usecase

import (
	"context"
	mdomain "network-debugger/internal/features/mapping/domain"
	"time"
)

// Service инкапсулирует бизнес‑операции над правилами маппинга
type Service struct {
	repo mdomain.MapRulesRepository
}

func NewService(r mdomain.MapRulesRepository) *Service { return &Service{repo: r} }

func (s *Service) List(ctx context.Context) ([]mdomain.MapRule, error) {
	return s.repo.List(ctx)
}

func (s *Service) Upsert(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
	// нормализация дат
	if r.CreatedAt.IsZero() {
		now := time.Now().UTC()
		r.CreatedAt = now
		r.UpdatedAt = now
	} else if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now().UTC()
	}
	return s.repo.Upsert(ctx, r)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Reorder пересчитывает приоритеты согласно переданному порядку идентификаторов
func (s *Service) Reorder(ctx context.Context, ids []string) error {
	return s.repo.Reorder(ctx, ids)
}
