package usecase

import (
	"context"
	"network-debugger/internal/domain"
	"time"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, s domain.Session) error
	GetSession(ctx context.Context, id string) (domain.Session, bool, error)
	DeleteSession(ctx context.Context, id string) error
	ListSessions(ctx context.Context, f SessionFilter) ([]domain.Session, int, error)
	IncrementCounters(ctx context.Context, id string, frame domain.Frame) error
	SetClosed(ctx context.Context, id string, closedAt time.Time, errMsg *string) error
	ClearAllSessions(ctx context.Context) error
}

type FrameRepository interface {
	AppendFrame(ctx context.Context, sessionID string, f domain.Frame) error
	ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error)
}

type EventRepository interface {
	AppendEvent(ctx context.Context, sessionID string, e domain.Event) error
	ListEvents(ctx context.Context, sessionID string, from string, limit int) ([]domain.Event, string, error)
}

type HTTPTransactionRepository interface {
	AppendHTTPTransaction(ctx context.Context, tx domain.HTTPTransaction) error
	ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error)
}

// Optional repository for capture control (in-memory MVP)
type CaptureControlRepository interface {
	RecordingState() (bool, int)
	StartCapture() int
	StopCapture() int
}

type SessionFilter struct {
	Q                 string
	Target            string
	Direction         *domain.Direction
	Opcode            *domain.Opcode
	Limit             int
	Offset            int
	CaptureID         *int // nil: any; -1 means current; otherwise exact id
	IncludeUnassigned bool // include sessions with CaptureID==nil
}
