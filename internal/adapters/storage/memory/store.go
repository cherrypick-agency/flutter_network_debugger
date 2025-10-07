package memory

import (
	"context"
	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
	"strings"
	"sync"
	"time"
)

type sessionEntry struct {
	session   domain.Session
	frames    []domain.Frame
	events    []domain.Event
	httpTxs   []domain.HTTPTransaction
	createdAt time.Time
}

type Store struct {
	mu sync.RWMutex
	// ring by insertion order of session ids
	order []string
	items map[string]*sessionEntry

	maxSessions         int
	maxFramesPerSession int
	ttl                 time.Duration

	// capture state (MVP, process-local)
	currentCapture int
	recording      bool
}

func NewStore(maxSessions, maxFrames int, ttl time.Duration) *Store {
	return &Store{
		order:               make([]string, 0, maxSessions),
		items:               make(map[string]*sessionEntry, maxSessions),
		maxSessions:         maxSessions,
		maxFramesPerSession: maxFrames,
		ttl:                 ttl,
		currentCapture:      0,
		recording:           true,
	}
}

// CaptureControlRepository (MVP)
func (s *Store) RecordingState() (bool, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.recording, s.currentCapture
}

func (s *Store) StartCapture() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentCapture++
	s.recording = true
	return s.currentCapture
}

func (s *Store) StopCapture() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recording = false
	return s.currentCapture
}

// SessionRepository
func (s *Store) CreateSession(ctx context.Context, sess domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// evict by ttl
	s.evictExpiredLocked()
	// evict by capacity
	if len(s.items) >= s.maxSessions {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.items, oldest)
	}
	// Assign capture id if recording
	if s.recording {
		cid := s.currentCapture
		sess.CaptureID = &cid
	}
	s.items[sess.ID] = &sessionEntry{session: sess, frames: make([]domain.Frame, 0, 64), events: make([]domain.Event, 0, 16), httpTxs: make([]domain.HTTPTransaction, 0, 32), createdAt: time.Now()}
	s.order = append(s.order, sess.ID)
	return nil
}

func (s *Store) GetSession(ctx context.Context, id string) (domain.Session, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.items[id]; ok {
		return e.session, true, nil
	}
	return domain.Session{}, false, nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; ok {
		delete(s.items, id)
		// remove from order
		for i, sid := range s.order {
			if sid == id {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
	return nil
}

// ClearAllSessions removes all sessions and associated data
func (s *Store) ClearAllSessions(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// reinitialize map; maps do not support cap(), so we can optionally hint with current len
	s.items = make(map[string]*sessionEntry, len(s.items))
	s.order = s.order[:0]
	// keep capture state as-is; not resetting currentCapture to preserve history
	return nil
}

func (s *Store) ListSessions(ctx context.Context, f usecase.SessionFilter) ([]domain.Session, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// naive scan + filter for MVP
	results := make([]domain.Session, 0, len(s.items))
	for _, id := range s.order { // preserve insertion order
		e := s.items[id]
		if e == nil {
			continue
		}
		// capture filter
		if f.CaptureID != nil {
			if *f.CaptureID >= 0 {
				// exact capture id
				if e.session.CaptureID == nil || *e.session.CaptureID != *f.CaptureID {
					continue
				}
			} else {
				// -1 treated as current
				if e.session.CaptureID == nil || *e.session.CaptureID != s.currentCapture {
					continue
				}
			}
		} else {
			// no specific capture; honor IncludeUnassigned flag
			if !f.IncludeUnassigned {
				// keep only assigned captures (exclude paused/unassigned)
				if e.session.CaptureID == nil {
					continue
				}
			}
		}
		// target filter: allow substring (case-insensitive) to match domain/URL parts
		if f.Target != "" && !containsFold(e.session.Target, f.Target) {
			continue
		}
		// text search best-effort in target
		if f.Q != "" && !strings.Contains(strings.ToLower(e.session.Target), strings.ToLower(f.Q)) {
			continue
		}
		results = append(results, e.session)
	}
	total := len(results)
	start := f.Offset
	if start > total {
		start = total
	}
	end := start + f.Limit
	if f.Limit <= 0 || end > total {
		end = total
	}
	return results[start:end], total, nil
}

func (s *Store) IncrementCounters(ctx context.Context, id string, frame domain.Frame) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[id]; ok {
		e.session.Frames.Total++
		switch frame.Opcode {
		case domain.OpcodeText:
			e.session.Frames.Text++
		case domain.OpcodeBinary:
			e.session.Frames.Binary++
		default:
			e.session.Frames.Control++
		}
	}
	return nil
}

func (s *Store) SetClosed(ctx context.Context, id string, ts time.Time, errMsg *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[id]; ok {
		e.session.ClosedAt = &ts
		e.session.Error = errMsg
	}
	return nil
}

// FrameRepository
func (s *Store) AppendFrame(ctx context.Context, sessionID string, f domain.Frame) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[sessionID]; ok {
		if len(e.frames) >= s.maxFramesPerSession {
			// drop-from-head policy
			e.frames = e.frames[1:]
		}
		e.frames = append(e.frames, f)
	}
	return nil
}

func (s *Store) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[sessionID]
	if !ok {
		return nil, "", nil
	}
	start := 0
	if from != "" {
		// naive linear search for the id position
		for i := range e.frames {
			if e.frames[i].ID == from {
				start = i + 1
				break
			}
		}
	}
	end := start + limit
	if limit <= 0 || end > len(e.frames) {
		end = len(e.frames)
	}
	next := ""
	if end < len(e.frames) {
		next = e.frames[end-1].ID
	}
	out := make([]domain.Frame, end-start)
	copy(out, e.frames[start:end])
	return out, next, nil
}

// EventRepository
func (s *Store) AppendEvent(ctx context.Context, sessionID string, ev domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[sessionID]; ok {
		e.events = append(e.events, ev)
		e.session.Events.Total++
		e.session.Events.SIO++
	}
	return nil
}

func (s *Store) ListEvents(ctx context.Context, sessionID string, from string, limit int) ([]domain.Event, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[sessionID]
	if !ok {
		return nil, "", nil
	}
	start := 0
	if from != "" {
		for i := range e.events {
			if e.events[i].ID == from {
				start = i + 1
				break
			}
		}
	}
	end := start + limit
	if limit <= 0 || end > len(e.events) {
		end = len(e.events)
	}
	next := ""
	if end < len(e.events) {
		next = e.events[end-1].ID
	}
	out := make([]domain.Event, end-start)
	copy(out, e.events[start:end])
	return out, next, nil
}

// HTTPTransactionRepository
func (s *Store) AppendHTTPTransaction(ctx context.Context, tx domain.HTTPTransaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.items[tx.SessionID]; ok {
		e.httpTxs = append(e.httpTxs, tx)
	}
	return nil
}

func (s *Store) ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[sessionID]
	if !ok {
		return nil, "", nil
	}
	start := 0
	if from != "" {
		for i := range e.httpTxs {
			if e.httpTxs[i].ID == from {
				start = i + 1
				break
			}
		}
	}
	end := start + limit
	if limit <= 0 || end > len(e.httpTxs) {
		end = len(e.httpTxs)
	}
	next := ""
	if end < len(e.httpTxs) {
		next = e.httpTxs[end-1].ID
	}
	out := make([]domain.HTTPTransaction, end-start)
	copy(out, e.httpTxs[start:end])
	return out, next, nil
}

func (s *Store) evictExpiredLocked() {
	if s.ttl <= 0 {
		return
	}
	now := time.Now()
	i := 0
	for i < len(s.order) {
		id := s.order[i]
		e := s.items[id]
		if e == nil || now.Sub(e.createdAt) > s.ttl {
			delete(s.items, id)
			s.order = append(s.order[:i], s.order[i+1:]...)
			continue
		}
		i++
	}
}

// helpers
func containsFold(s, substr string) bool {
	// naive case-insensitive contains for MVP
	// avoid strings.EqualFold across slices; convert both to lower
	// to keep dependency surface minimal in MVP
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c = c + 32
		}
		b = append(b, c)
	}
	bl := string(b)
	bb := make([]byte, 0, len(substr))
	for i := 0; i < len(substr); i++ {
		c := substr[i]
		if 'A' <= c && c <= 'Z' {
			c = c + 32
		}
		bb = append(bb, c)
	}
	sub := string(bb)
	return indexOf(bl, sub) >= 0
}

func indexOf(s, sub string) int {
	// very small, naive search to avoid importing strings
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
