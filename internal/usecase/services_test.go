package usecase

import (
	"context"
	"network-debugger/internal/domain"
	"testing"
	"time"
)

// Mock repository implementations
type mockSessionRepo struct {
	sessions map[string]domain.Session
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, s domain.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, id string) (domain.Session, bool, error) {
	s, ok := m.sessions[id]
	return s, ok, nil
}

func (m *mockSessionRepo) DeleteSession(ctx context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *mockSessionRepo) ListSessions(ctx context.Context, f SessionFilter) ([]domain.Session, int, error) {
	var result []domain.Session
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result, len(result), nil
}

func (m *mockSessionRepo) IncrementCounters(ctx context.Context, id string, frame domain.Frame) error {
	return nil
}

func (m *mockSessionRepo) SetClosed(ctx context.Context, id string, closedAt time.Time, errMsg *string) error {
	if s, ok := m.sessions[id]; ok {
		s.ClosedAt = &closedAt
		s.Error = errMsg
		m.sessions[id] = s
	}
	return nil
}

func (m *mockSessionRepo) ClearAllSessions(ctx context.Context) error {
	m.sessions = make(map[string]domain.Session)
	return nil
}

func (m *mockSessionRepo) AddSpoolFile(ctx context.Context, sessionID string, path string) {}

type mockFrameRepo struct {
	frames map[string][]domain.Frame
}

func (m *mockFrameRepo) AppendFrame(ctx context.Context, sessionID string, f domain.Frame) error {
	m.frames[sessionID] = append(m.frames[sessionID], f)
	return nil
}

func (m *mockFrameRepo) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	frames := m.frames[sessionID]
	if limit <= 0 || limit > len(frames) {
		limit = len(frames)
	}
	return frames[:limit], "", nil
}

type mockEventRepo struct {
	events map[string][]domain.Event
}

func (m *mockEventRepo) AppendEvent(ctx context.Context, sessionID string, e domain.Event) error {
	m.events[sessionID] = append(m.events[sessionID], e)
	return nil
}

func (m *mockEventRepo) ListEvents(ctx context.Context, sessionID string, from string, limit int) ([]domain.Event, string, error) {
	events := m.events[sessionID]
	if limit <= 0 || limit > len(events) {
		limit = len(events)
	}
	return events[:limit], "", nil
}

type mockHTTPTxRepo struct {
	txs map[string][]domain.HTTPTransaction
}

func (m *mockHTTPTxRepo) AppendHTTPTransaction(ctx context.Context, tx domain.HTTPTransaction) error {
	m.txs[tx.SessionID] = append(m.txs[tx.SessionID], tx)
	return nil
}

func (m *mockHTTPTxRepo) ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	txs := m.txs[sessionID]
	if limit <= 0 || limit > len(txs) {
		limit = len(txs)
	}
	return txs[:limit], "", nil
}

func TestNewSessionService(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	frameRepo := &mockFrameRepo{frames: make(map[string][]domain.Frame)}
	eventRepo := &mockEventRepo{events: make(map[string][]domain.Event)}

	svc := NewSessionService(sessRepo, frameRepo, eventRepo)

	if svc == nil {
		t.Fatal("NewSessionService returned nil")
	}
	if svc.sessions == nil {
		t.Error("sessions repo not set")
	}
	if svc.frames == nil {
		t.Error("frames repo not set")
	}
	if svc.events == nil {
		t.Error("events repo not set")
	}
}

func TestSessionService_Create(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	sess := domain.Session{ID: "test1", Target: "ws://example.com"}
	err := svc.Create(ctx, sess)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if len(sessRepo.sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessRepo.sessions))
	}
}

func TestSessionService_Get(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: map[string]domain.Session{
		"test1": {ID: "test1", Target: "ws://example.com"},
	}}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	sess, ok, err := svc.Get(ctx, "test1")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !ok {
		t.Error("session not found")
	}
	if sess.ID != "test1" {
		t.Errorf("session ID = %s, want test1", sess.ID)
	}
}

func TestSessionService_List(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: map[string]domain.Session{
		"test1": {ID: "test1"},
		"test2": {ID: "test2"},
	}}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	sessions, total, err := svc.List(ctx, SessionFilter{})

	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestSessionService_Delete(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: map[string]domain.Session{
		"test1": {ID: "test1"},
	}}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	err := svc.Delete(ctx, "test1")

	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if len(sessRepo.sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessRepo.sessions))
	}
}

func TestSessionService_ClearAll(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: map[string]domain.Session{
		"test1": {ID: "test1"},
		"test2": {ID: "test2"},
	}}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	err := svc.ClearAll(ctx)

	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}
	if len(sessRepo.sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessRepo.sessions))
	}
}

func TestSessionService_AddFrame(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	frameRepo := &mockFrameRepo{frames: make(map[string][]domain.Frame)}
	svc := NewSessionService(sessRepo, frameRepo, &mockEventRepo{})
	ctx := context.Background()

	frame := domain.Frame{ID: "frame1", Opcode: domain.OpcodeText}
	err := svc.AddFrame(ctx, "test1", frame)

	if err != nil {
		t.Fatalf("AddFrame failed: %v", err)
	}
	if len(frameRepo.frames["test1"]) != 1 {
		t.Errorf("expected 1 frame, got %d", len(frameRepo.frames["test1"]))
	}
}

func TestSessionService_AddEvent(t *testing.T) {
	eventRepo := &mockEventRepo{events: make(map[string][]domain.Event)}
	svc := NewSessionService(&mockSessionRepo{}, &mockFrameRepo{}, eventRepo)
	ctx := context.Background()

	event := domain.Event{ID: "event1", Namespace: "/chat"}
	err := svc.AddEvent(ctx, "test1", event)

	if err != nil {
		t.Fatalf("AddEvent failed: %v", err)
	}
	if len(eventRepo.events["test1"]) != 1 {
		t.Errorf("expected 1 event, got %d", len(eventRepo.events["test1"]))
	}
}

func TestSessionService_ListFrames(t *testing.T) {
	frameRepo := &mockFrameRepo{frames: map[string][]domain.Frame{
		"test1": {{ID: "frame1"}, {ID: "frame2"}},
	}}
	svc := NewSessionService(&mockSessionRepo{}, frameRepo, &mockEventRepo{})
	ctx := context.Background()

	frames, _, err := svc.ListFrames(ctx, "test1", "", 10)

	if err != nil {
		t.Fatalf("ListFrames failed: %v", err)
	}
	if len(frames) != 2 {
		t.Errorf("expected 2 frames, got %d", len(frames))
	}
}

func TestSessionService_ListEvents(t *testing.T) {
	eventRepo := &mockEventRepo{events: map[string][]domain.Event{
		"test1": {{ID: "event1"}, {ID: "event2"}},
	}}
	svc := NewSessionService(&mockSessionRepo{}, &mockFrameRepo{}, eventRepo)
	ctx := context.Background()

	events, _, err := svc.ListEvents(ctx, "test1", "", 10)

	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestSessionService_SetClosed(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: map[string]domain.Session{
		"test1": {ID: "test1"},
	}}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	ts := time.Now()
	errMsg := "connection closed"
	err := svc.SetClosed(ctx, "test1", ts, &errMsg)

	if err != nil {
		t.Fatalf("SetClosed failed: %v", err)
	}
	sess := sessRepo.sessions["test1"]
	if sess.ClosedAt == nil {
		t.Error("ClosedAt not set")
	}
	if sess.Error == nil || *sess.Error != errMsg {
		t.Error("Error not set correctly")
	}
}

func TestSessionService_AddHTTPTransaction_NoRepo(t *testing.T) {
	// Test when httpTxs is nil
	svc := NewSessionService(&mockSessionRepo{}, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	tx := domain.HTTPTransaction{ID: "tx1", SessionID: "test1"}
	err := svc.AddHTTPTransaction(ctx, tx)

	if err != nil {
		t.Errorf("AddHTTPTransaction should not error when httpTxs is nil: %v", err)
	}
}

func TestSessionService_ListHTTPTransactions_NoRepo(t *testing.T) {
	// Test when httpTxs is nil
	svc := NewSessionService(&mockSessionRepo{}, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	txs, next, err := svc.ListHTTPTransactions(ctx, "test1", "", 10)

	if err != nil {
		t.Errorf("ListHTTPTransactions should not error when httpTxs is nil: %v", err)
	}
	if txs != nil {
		t.Error("txs should be nil when httpTxs is nil")
	}
	if next != "" {
		t.Error("next should be empty when httpTxs is nil")
	}
}

func TestSessionService_AddSpoolFile(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	// Should not panic even if repo doesn't support AddSpoolFile
	svc.AddSpoolFile(ctx, "test1", "/tmp/spool")
}

func TestSessionService_SessionsRepoUnsafe(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})

	repo := svc.SessionsRepoUnsafe()
	if repo == nil {
		t.Error("SessionsRepoUnsafe returned nil")
	}
}

// Failing mocks for error path testing
type failingFrameRepo struct{}

func (f *failingFrameRepo) AppendFrame(ctx context.Context, sessionID string, frame domain.Frame) error {
	return context.DeadlineExceeded // Use standard error
}

func (f *failingFrameRepo) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	return nil, "", context.DeadlineExceeded
}

func TestSessionService_AddFrame_ErrorPath(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	frameRepo := &failingFrameRepo{}
	svc := NewSessionService(sessRepo, frameRepo, &mockEventRepo{})
	ctx := context.Background()

	frame := domain.Frame{ID: "frame1", Opcode: domain.OpcodeText}
	err := svc.AddFrame(ctx, "test1", frame)

	if err == nil {
		t.Error("AddFrame should return error when AppendFrame fails")
	}
}

// Combined session+HTTP repo mock for testing HTTP functionality
type mockSessionHTTPRepo struct {
	*mockSessionRepo
	*mockHTTPTxRepo
}

func TestSessionService_AddHTTPTransaction_WithRepo(t *testing.T) {
	// Create a repo that implements both SessionRepository and HTTPTransactionRepository
	combined := &mockSessionHTTPRepo{
		mockSessionRepo: &mockSessionRepo{sessions: make(map[string]domain.Session)},
		mockHTTPTxRepo:  &mockHTTPTxRepo{txs: make(map[string][]domain.HTTPTransaction)},
	}
	svc := NewSessionService(combined, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	tx := domain.HTTPTransaction{ID: "tx1", SessionID: "test1"}
	err := svc.AddHTTPTransaction(ctx, tx)

	if err != nil {
		t.Fatalf("AddHTTPTransaction failed: %v", err)
	}
	if len(combined.mockHTTPTxRepo.txs["test1"]) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(combined.mockHTTPTxRepo.txs["test1"]))
	}
}

func TestSessionService_ListHTTPTransactions_WithRepo(t *testing.T) {
	combined := &mockSessionHTTPRepo{
		mockSessionRepo: &mockSessionRepo{sessions: make(map[string]domain.Session)},
		mockHTTPTxRepo: &mockHTTPTxRepo{txs: map[string][]domain.HTTPTransaction{
			"test1": {{ID: "tx1"}, {ID: "tx2"}},
		}},
	}
	svc := NewSessionService(combined, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	txs, _, err := svc.ListHTTPTransactions(ctx, "test1", "", 10)

	if err != nil {
		t.Fatalf("ListHTTPTransactions failed: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}
}
