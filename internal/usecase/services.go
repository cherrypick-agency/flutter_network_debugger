package usecase

import (
	"context"
	"network-debugger/internal/domain"
	processuc "network-debugger/internal/features/process/usecase"
	"time"
)

type SessionService struct {
	sessions   SessionRepository
	frames     FrameRepository
	events     EventRepository
	httpTxs    HTTPTransactionRepository
	processSvc *processuc.Service
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

// SetProcessService sets the process detection service (optional)
func (s *SessionService) SetProcessService(p *processuc.Service) {
	s.processSvc = p
}

// Temporary unsafe accessor for underlying sessions repo (for in-memory capture MVP)
func (s *SessionService) SessionsRepoUnsafe() any { return s.sessions }

func (s *SessionService) Create(ctx context.Context, sess domain.Session) error {
	// Try to detect process if ProcessService is available
	if s.processSvc != nil && sess.ProcessInfo == nil {
		// Extract port from ClientAddr (format: "ip:port")
		if port := extractPortFromAddr(sess.ClientAddr); port > 0 {
			processInfo, err := s.processSvc.DetectForConnection(ctx, port)
			if err == nil && processInfo != nil {
				sess.ProcessInfo = processInfo
			}
			// Ignore detection errors - this is not critical
		}
	}

	return s.sessions.CreateSession(ctx, sess)
}

// extractPortFromAddr extracts port from address in "ip:port" format
func extractPortFromAddr(addr string) uint32 {
	// Simple parsing of the last part after ":"
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			portStr := addr[i+1:]
			var port uint32
			for _, c := range portStr {
				if c >= '0' && c <= '9' {
					port = port*10 + uint32(c-'0')
				}
			}
			return port
		}
	}
	return 0
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

func (s *SessionService) DeleteImported(ctx context.Context) error {
	return s.sessions.DeleteImportedSessions(ctx)
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

func (s *SessionService) GetFrameByID(ctx context.Context, sessionID string, frameID string) (domain.Frame, bool, error) {
	return s.frames.GetFrameByID(ctx, sessionID, frameID)
}

func (s *SessionService) UpdateFrameBodyFile(ctx context.Context, sessionID string, frameID string, bodyFile string) error {
	return s.frames.UpdateFrameBodyFile(ctx, sessionID, frameID, bodyFile)
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

// AddSpoolFile tracks spool file path for a session if underlying repo supports it.
func (s *SessionService) AddSpoolFile(ctx context.Context, sessionID string, path string) {
	if repo := s.SessionsRepoUnsafe(); repo != nil {
		if ms, ok := repo.(interface {
			AddSpoolFile(context.Context, string, string)
		}); ok {
			ms.AddSpoolFile(ctx, sessionID, path)
		}
	}
}

func (s *SessionService) ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	if s.httpTxs == nil {
		return nil, "", nil
	}
	return s.httpTxs.ListHTTPTransactions(ctx, sessionID, from, limit)
}
