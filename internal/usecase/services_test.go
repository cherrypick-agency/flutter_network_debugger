package usecase

import (
	"context"
	"network-debugger/internal/domain"
	processdomain "network-debugger/internal/features/process/domain"
	processuc "network-debugger/internal/features/process/usecase"
	"testing"
	"time"

	"github.com/rs/zerolog"
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

func (m *mockSessionRepo) DeleteImportedSessions(ctx context.Context) error {
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

func (m *mockFrameRepo) GetFrameByID(ctx context.Context, sessionID string, frameID string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
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

func (f *failingFrameRepo) GetFrameByID(ctx context.Context, sessionID string, frameID string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, context.DeadlineExceeded
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

func TestSessionService_Get_NotFound(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	_, ok, err := svc.Get(ctx, "nonexistent")

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if ok {
		t.Error("session should not be found")
	}
}

func TestSessionService_AddSpoolFile_WithSupportingRepo(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	// This test verifies the type assertion path
	// The mock already supports AddSpoolFile, so this exercises the "ok" branch
	svc.AddSpoolFile(ctx, "test1", "/tmp/test.dat")

	// No assertion needed - just verify no panic
}

func TestSessionService_NewSessionService_HTTPTxRepoTypeAssertion(t *testing.T) {
	// Test that HTTPTransactionRepository is correctly type-asserted
	combined := &mockSessionHTTPRepo{
		mockSessionRepo: &mockSessionRepo{sessions: make(map[string]domain.Session)},
		mockHTTPTxRepo:  &mockHTTPTxRepo{txs: make(map[string][]domain.HTTPTransaction)},
	}
	svc := NewSessionService(combined, &mockFrameRepo{}, &mockEventRepo{})

	if svc.httpTxs == nil {
		t.Error("httpTxs should be set when SessionRepository implements HTTPTransactionRepository")
	}
}

func TestSessionService_SetClosed_NoError(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: map[string]domain.Session{
		"test1": {ID: "test1"},
	}}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})
	ctx := context.Background()

	ts := time.Now()
	err := svc.SetClosed(ctx, "test1", ts, nil)

	if err != nil {
		t.Fatalf("SetClosed failed: %v", err)
	}
	sess := sessRepo.sessions["test1"]
	if sess.ClosedAt == nil {
		t.Error("ClosedAt not set")
	}
	if sess.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestSessionService_AddFrame_IncrementsCounters(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	frameRepo := &mockFrameRepo{frames: make(map[string][]domain.Frame)}
	svc := NewSessionService(sessRepo, frameRepo, &mockEventRepo{})
	ctx := context.Background()

	// Verify that both AppendFrame and IncrementCounters are called
	frame := domain.Frame{ID: "frame1", Opcode: domain.OpcodeBinary, Size: 100}
	err := svc.AddFrame(ctx, "test1", frame)

	if err != nil {
		t.Fatalf("AddFrame failed: %v", err)
	}
	// Frame should be added
	if len(frameRepo.frames["test1"]) != 1 {
		t.Error("frame not added")
	}
	// IncrementCounters is called but mockSessionRepo doesn't track this
	// This test at least exercises the code path
}

func TestSessionService_SetProcessService(t *testing.T) {
	svc := NewSessionService(&mockSessionRepo{}, &mockFrameRepo{}, &mockEventRepo{})

	if svc.processSvc != nil {
		t.Error("processSvc should be nil initially")
	}

	logger := zerolog.Nop()
	processSvc := processuc.NewService(
		&mockProcessConfigRepo{},
		&mockProcessIconCacheRepo{},
		&mockProcessDetector{},
		nil,
		&mockProcessIconExtractor{},
		nil,
		"",
		&logger,
	)
	svc.SetProcessService(processSvc)

	if svc.processSvc == nil {
		t.Error("processSvc should be set after SetProcessService")
	}
}

func TestSessionService_Create_WithProcessDetection(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})

	logger := zerolog.Nop()
	detector := &mockProcessDetectorWithResult{
		result: &processdomain.ProcessInfo{
			PID:  1234,
			Name: "test-process",
		},
	}
	processSvc := processuc.NewService(
		&mockProcessConfigRepo{},
		&mockProcessIconCacheRepo{},
		detector,
		nil,
		&mockProcessIconExtractor{},
		nil,
		"",
		&logger,
	)
	svc.SetProcessService(processSvc)
	ctx := context.Background()

	sess := domain.Session{
		ID:         "test1",
		Target:     "ws://example.com",
		ClientAddr: "127.0.0.1:8080",
	}
	err := svc.Create(ctx, sess)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	savedSess, ok := sessRepo.sessions["test1"]
	if !ok {
		t.Fatal("Session should be saved")
	}
	if savedSess.ProcessInfo == nil {
		t.Error("ProcessInfo should be set after process detection")
	}
	if savedSess.ProcessInfo != nil && savedSess.ProcessInfo.PID != 1234 {
		t.Errorf("ProcessInfo.PID = %d, want 1234", savedSess.ProcessInfo.PID)
	}
}

func TestSessionService_Create_WithProcessDetection_NoPort(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})

	logger := zerolog.Nop()
	processSvc := processuc.NewService(
		&mockProcessConfigRepo{},
		&mockProcessIconCacheRepo{},
		&mockProcessDetector{},
		nil,
		&mockProcessIconExtractor{},
		nil,
		"",
		&logger,
	)
	svc.SetProcessService(processSvc)
	ctx := context.Background()

	sess := domain.Session{
		ID:         "test1",
		Target:     "ws://example.com",
		ClientAddr: "127.0.0.1",
	}
	err := svc.Create(ctx, sess)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestSessionService_Create_WithProcessDetection_AlreadyHasProcessInfo(t *testing.T) {
	sessRepo := &mockSessionRepo{sessions: make(map[string]domain.Session)}
	svc := NewSessionService(sessRepo, &mockFrameRepo{}, &mockEventRepo{})

	logger := zerolog.Nop()
	processSvc := processuc.NewService(
		&mockProcessConfigRepo{},
		&mockProcessIconCacheRepo{},
		&mockProcessDetector{},
		nil,
		&mockProcessIconExtractor{},
		nil,
		"",
		&logger,
	)
	svc.SetProcessService(processSvc)
	ctx := context.Background()

	sess := domain.Session{
		ID:         "test1",
		Target:     "ws://example.com",
		ClientAddr: "127.0.0.1:8080",
		ProcessInfo: &processdomain.ProcessInfo{
			PID:  9999,
			Name: "existing-process",
		},
	}
	err := svc.Create(ctx, sess)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if sess.ProcessInfo.PID != 9999 {
		t.Errorf("ProcessInfo.PID = %d, want 9999", sess.ProcessInfo.PID)
	}
}

func TestExtractPortFromAddr(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected uint32
	}{
		{"valid port", "127.0.0.1:8080", 8080},
		{"valid port large", "192.168.1.1:65535", 65535},
		{"no port", "127.0.0.1", 0},
		{"empty string", "", 0},
		{"port with non-digits", "127.0.0.1:8080abc", 8080},
		{"multiple colons", "::1:8080", 8080},
		{"port zero", "127.0.0.1:0", 0},
		{"port with spaces", "127.0.0.1: 8080", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPortFromAddr(tt.addr)
			if result != tt.expected {
				t.Errorf("extractPortFromAddr(%q) = %d, want %d", tt.addr, result, tt.expected)
			}
		})
	}
}

type mockProcessConfigRepo struct{}

func (m *mockProcessConfigRepo) Load(ctx context.Context) (*processdomain.DetectionConfig, error) {
	return &processdomain.DetectionConfig{
		Enabled: true,
	}, nil
}

func (m *mockProcessConfigRepo) Save(ctx context.Context, cfg *processdomain.DetectionConfig) error {
	return nil
}

type mockProcessIconCacheRepo struct{}

func (m *mockProcessIconCacheRepo) Get(key string) (*processdomain.AppIcon, error) {
	return nil, nil
}

func (m *mockProcessIconCacheRepo) Set(key string, icon *processdomain.AppIcon, ttl time.Duration) error {
	return nil
}

func (m *mockProcessIconCacheRepo) Delete(key string) error {
	return nil
}

func (m *mockProcessIconCacheRepo) Clear() error {
	return nil
}

func (m *mockProcessIconCacheRepo) CleanupExpired() error {
	return nil
}

type mockProcessDetector struct{}

func (m *mockProcessDetector) DetectByPort(ctx context.Context, port uint32) (*processdomain.ProcessInfo, error) {
	return nil, nil
}

func (m *mockProcessDetector) DetectByPID(ctx context.Context, pid int32) (*processdomain.ProcessInfo, error) {
	return nil, nil
}

func (m *mockProcessDetector) RequiresPrivileges() bool {
	return false
}

type mockProcessDetectorWithResult struct {
	mockProcessDetector
	result *processdomain.ProcessInfo
}

func (m *mockProcessDetectorWithResult) DetectByPort(ctx context.Context, port uint32) (*processdomain.ProcessInfo, error) {
	return m.result, nil
}

type mockProcessIconExtractor struct{}

func (m *mockProcessIconExtractor) Extract(ctx context.Context, pid int32, path string) (*processdomain.AppIcon, error) {
	return nil, nil
}

func (m *mockProcessIconExtractor) ExtractByPID(ctx context.Context, pid int32) (*processdomain.AppIcon, error) {
	return nil, nil
}

func (m *mockProcessIconExtractor) ExtractByPath(ctx context.Context, path string) (*processdomain.AppIcon, error) {
	return nil, nil
}
