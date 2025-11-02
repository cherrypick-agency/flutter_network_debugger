package ports

import (
	"context"
	"network-debugger/internal/domain"
)

// SessionQuery — чтение сессий.
type SessionQuery interface {
	Get(ctx context.Context, id string) (domain.Session, bool, error)
}

// HTTPTxQuery — чтение HTTP‑транзакций.
type HTTPTxQuery interface {
	ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error)
}

// FramesQuery — чтение фреймов (для предпросмотра request/response).
type FramesQuery interface {
	ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error)
}
