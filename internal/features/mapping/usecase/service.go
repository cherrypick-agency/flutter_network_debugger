package usecase

import (
	"context"
	mdomain "network-debugger/internal/features/mapping/domain"
	"time"
)

// Service encapsulates business operations on mapping rules
type Service struct {
	repo mdomain.MapRulesRepository
}

func NewService(r mdomain.MapRulesRepository) *Service { return &Service{repo: r} }

func (s *Service) List(ctx context.Context) ([]mdomain.MapRule, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*mdomain.MapRule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Upsert(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
	// date normalization
	now := time.Now().UTC()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	return s.repo.Upsert(ctx, r)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Reorder recalculates priorities according to the given order of identifiers
func (s *Service) Reorder(ctx context.Context, ids []string) error {
	return s.repo.Reorder(ctx, ids)
}
