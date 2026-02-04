package ports

import (
	"context"
	"network-debugger/internal/domain"
)

// SessionQuery — session reading.
type SessionQuery interface {
	Get(ctx context.Context, id string) (domain.Session, bool, error)
}

// HTTPTxQuery — reading HTTP transactions.
type HTTPTxQuery interface {
	ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error)
}

// FramesQuery — reading frames (for request/response preview).
type FramesQuery interface {
	ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error)
}
