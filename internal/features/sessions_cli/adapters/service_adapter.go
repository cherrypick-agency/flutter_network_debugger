package adapters

import (
	"context"
	memory "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	"network-debugger/internal/features/sessions_cli/ports"
	"network-debugger/internal/usecase"
)

// ServiceQueries — адаптер поверх usecase.SessionService.
type ServiceQueries struct {
	svc *usecase.SessionService
}

func NewServiceQueries(s *usecase.SessionService) *ServiceQueries {
	return &ServiceQueries{svc: s}
}

// Get реализует ports.SessionQuery.
func (s *ServiceQueries) Get(ctx context.Context, id string) (domain.Session, bool, error) {
	return s.svc.Get(ctx, id)
}

// ListHTTPTransactions реализует ports.HTTPTxQuery.
func (s *ServiceQueries) ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	return s.svc.ListHTTPTransactions(ctx, sessionID, from, limit)
}

// ListFrames реализует ports.FramesQuery.
func (s *ServiceQueries) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	return s.svc.ListFrames(ctx, sessionID, from, limit)
}

// Гарантия интерфейсов
var _ ports.SessionQuery = (*ServiceQueries)(nil)
var _ ports.HTTPTxQuery = (*ServiceQueries)(nil)
var _ ports.FramesQuery = (*ServiceQueries)(nil)

// FindHTTPTransaction — оптимизированный доступ к транзакции по ID для in-memory хранилища.
func (s *ServiceQueries) FindHTTPTransaction(ctx context.Context, sessionID string, txID string) (domain.HTTPTransaction, bool) {
	if repo := s.svc.SessionsRepoUnsafe(); repo != nil {
		if ms, ok := repo.(*memory.Store); ok {
			return ms.FindHTTPTransaction(sessionID, txID)
		}
	}
	return domain.HTTPTransaction{}, false
}
