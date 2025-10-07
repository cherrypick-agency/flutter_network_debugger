package usecase

import (
	"context"
	"network-debugger/internal/domain"
	"time"
)

type SessionService struct {
	sessions SessionRepository
	frames   FrameRepository
	events   EventRepository
	httpTxs  HTTPTransactionRepository
}

func NewSessionService(s SessionRepository, f FrameRepository, e EventRepository) *SessionService {
	// Backward-compat: when memory store implements HTTPTransactionRepository as well,
	// we can type-assert and attach; otherwise http features are disabled.
	var h HTTPTransactionRepository
	if v, ok := any(s).(HTTPTransactionRepository); ok {
		h = v
	}
	return &SessionService{sessions: s, frames: f, events: e, httpTxs: h}
}

// Temporary unsafe accessor for underlying sessions repo (for in-memory capture MVP)
func (s *SessionService) SessionsRepoUnsafe() any { return s.sessions }

func (s *SessionService) Create(ctx context.Context, sess domain.Session) error {
	return s.sessions.CreateSession(ctx, sess)
}

func (s *SessionService) Get(ctx context.Context, id string) (domain.Session, bool, error) {
	return s.sessions.GetSession(ctx, id)
}

func (s *SessionService) List(ctx context.Context, f SessionFilter) ([]domain.Session, int, error) {
	return s.sessions.ListSessions(ctx, f)
}

func (s *SessionService) Delete(ctx context.Context, id string) error {
	return s.sessions.DeleteSession(ctx, id)
}

func (s *SessionService) ClearAll(ctx context.Context) error {
	return s.sessions.ClearAllSessions(ctx)
}

func (s *SessionService) AddFrame(ctx context.Context, sessionID string, frame domain.Frame) error {
	if err := s.frames.AppendFrame(ctx, sessionID, frame); err != nil {
		return err
	}
	return s.sessions.IncrementCounters(ctx, sessionID, frame)
}

func (s *SessionService) AddEvent(ctx context.Context, sessionID string, event domain.Event) error {
	return s.events.AppendEvent(ctx, sessionID, event)
}

func (s *SessionService) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	return s.frames.ListFrames(ctx, sessionID, from, limit)
}

func (s *SessionService) ListEvents(ctx context.Context, sessionID string, from string, limit int) ([]domain.Event, string, error) {
	return s.events.ListEvents(ctx, sessionID, from, limit)
}

func (s *SessionService) SetClosed(ctx context.Context, id string, closedAt time.Time, errMsg *string) error {
	return s.sessions.SetClosed(ctx, id, closedAt, errMsg)
}

func (s *SessionService) AddHTTPTransaction(ctx context.Context, tx domain.HTTPTransaction) error {
	if s.httpTxs == nil {
		return nil
	}
	return s.httpTxs.AppendHTTPTransaction(ctx, tx)
}

func (s *SessionService) ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	if s.httpTxs == nil {
		return nil, "", nil
	}
	return s.httpTxs.ListHTTPTransactions(ctx, sessionID, from, limit)
}
