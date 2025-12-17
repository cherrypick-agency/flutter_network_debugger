package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"network-debugger/internal/domain"
	uc "network-debugger/internal/usecase"

	"github.com/rs/zerolog"
)

// mockRepoForImport implements SessionRepository for import testing
type mockRepoForImport struct {
	sessions     map[string]domain.Session
	transactions map[string][]domain.HTTPTransaction
	frames       map[string][]domain.Frame
	clearAllErr  error
	deleteImpErr error
}

func newMockRepoForImport() *mockRepoForImport {
	return &mockRepoForImport{
		sessions:     make(map[string]domain.Session),
		transactions: make(map[string][]domain.HTTPTransaction),
		frames:       make(map[string][]domain.Frame),
	}
}

func (m *mockRepoForImport) CreateSession(ctx context.Context, s domain.Session) error {
	m.sessions[s.ID] = s
	return nil
}

func (m *mockRepoForImport) GetSession(ctx context.Context, id string) (domain.Session, bool, error) {
	s, ok := m.sessions[id]
	return s, ok, nil
}

func (m *mockRepoForImport) DeleteSession(ctx context.Context, id string) error {
	delete(m.sessions, id)
	delete(m.transactions, id)
	delete(m.frames, id)
	return nil
}

func (m *mockRepoForImport) ListSessions(ctx context.Context, filter uc.SessionFilter) ([]domain.Session, int, error) {
	sessions := make([]domain.Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions, len(sessions), nil
}

func (m *mockRepoForImport) IncrementCounters(ctx context.Context, sessionID string, frame domain.Frame) error {
	return nil
}

func (m *mockRepoForImport) SetClosed(ctx context.Context, sessionID string, closedAt time.Time, errMsg *string) error {
	if s, ok := m.sessions[sessionID]; ok {
		s.ClosedAt = &closedAt
		m.sessions[sessionID] = s
	}
	return nil
}

func (m *mockRepoForImport) ClearAllSessions(ctx context.Context) error {
	if m.clearAllErr != nil {
		return m.clearAllErr
	}
	m.sessions = make(map[string]domain.Session)
	m.transactions = make(map[string][]domain.HTTPTransaction)
	m.frames = make(map[string][]domain.Frame)
	return nil
}

func (m *mockRepoForImport) DeleteImportedSessions(ctx context.Context) error {
	if m.deleteImpErr != nil {
		return m.deleteImpErr
	}
	// Delete sessions where ClientAddr == "imported"
	for id, s := range m.sessions {
		if s.ClientAddr == "imported" {
			delete(m.sessions, id)
			delete(m.transactions, id)
			delete(m.frames, id)
		}
	}
	return nil
}

// FrameRepository methods
func (m *mockRepoForImport) AppendFrame(ctx context.Context, sessionID string, frame domain.Frame) error {
	m.frames[sessionID] = append(m.frames[sessionID], frame)
	return nil
}

func (m *mockRepoForImport) ListFrames(ctx context.Context, sessionID, from string, limit int) ([]domain.Frame, string, error) {
	return m.frames[sessionID], "", nil
}

func (m *mockRepoForImport) GetFrameByID(ctx context.Context, sessionID, frameID string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
}

func (m *mockRepoForImport) UpdateFrameBodyFile(ctx context.Context, sessionID, frameID, bodyFile string) error {
	return nil
}

// EventRepository methods
func (m *mockRepoForImport) AppendEvent(ctx context.Context, sessionID string, event domain.Event) error {
	return nil
}

func (m *mockRepoForImport) ListEvents(ctx context.Context, sessionID, from string, limit int) ([]domain.Event, string, error) {
	return nil, "", nil
}

// HTTPTransactionRepository methods
func (m *mockRepoForImport) AppendHTTPTransaction(ctx context.Context, tx domain.HTTPTransaction) error {
	m.transactions[tx.SessionID] = append(m.transactions[tx.SessionID], tx)
	return nil
}

func (m *mockRepoForImport) ListHTTPTransactions(ctx context.Context, sessionID, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	return m.transactions[sessionID], "", nil
}

// Helper methods
func (m *mockRepoForImport) countSessions() int {
	return len(m.sessions)
}

func (m *mockRepoForImport) countImportedSessions() int {
	count := 0
	for _, s := range m.sessions {
		if s.ClientAddr == "imported" {
			count++
		}
	}
	return count
}

func (m *mockRepoForImport) countProxySessions() int {
	count := 0
	for _, s := range m.sessions {
		if s.ClientAddr != "imported" {
			count++
		}
	}
	return count
}

// createTestDeps creates a Deps instance for testing
func createTestDeps(repo *mockRepoForImport) *Deps {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	svc := uc.NewSessionService(repo, repo, repo)
	return &Deps{
		Svc:    svc,
		Logger: &logger,
	}
}

// loadTestHAR loads a HAR test file from testdata/har/
func loadTestHAR(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", "testdata", "har", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test HAR file %s: %v", filename, err)
	}
	return data
}

// TestImportHAR_ModeMerge tests merge mode (default behavior)
func TestImportHAR_ModeMerge(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Add some existing proxy sessions
	repo.sessions["proxy1"] = domain.Session{ID: "proxy1", ClientAddr: "127.0.0.1:1234"}
	repo.sessions["proxy2"] = domain.Session{ID: "proxy2", ClientAddr: "192.168.1.1:5678"}

	// Load test HAR
	harData := loadTestHAR(t, "simple_http.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har?mode=merge", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["imported"].(float64) != 1 {
		t.Errorf("Expected 1 imported, got %v", result["imported"])
	}

	if result["failed"].(float64) != 0 {
		t.Errorf("Expected 0 failed, got %v", result["failed"])
	}

	// Verify sessions: should have proxy + imported
	totalSessions := repo.countSessions()
	if totalSessions != 3 {
		t.Errorf("Expected 3 total sessions (2 proxy + 1 imported), got %d", totalSessions)
	}

	proxySessions := repo.countProxySessions()
	if proxySessions != 2 {
		t.Errorf("Expected 2 proxy sessions, got %d", proxySessions)
	}

	importedSessions := repo.countImportedSessions()
	if importedSessions != 1 {
		t.Errorf("Expected 1 imported session, got %d", importedSessions)
	}
}

// TestImportHAR_ModeReplaceAll tests replace_all mode
func TestImportHAR_ModeReplaceAll(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Add some existing proxy sessions
	repo.sessions["proxy1"] = domain.Session{ID: "proxy1", ClientAddr: "127.0.0.1:1234"}
	repo.sessions["imported1"] = domain.Session{ID: "imported1", ClientAddr: "imported"}

	// Load test HAR
	harData := loadTestHAR(t, "simple_http.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har?mode=replace_all", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["imported"].(float64) != 1 {
		t.Errorf("Expected 1 imported, got %v", result["imported"])
	}

	// Verify sessions: all old sessions should be deleted
	totalSessions := repo.countSessions()
	if totalSessions != 1 {
		t.Errorf("Expected 1 total session (only new import), got %d", totalSessions)
	}

	proxySessions := repo.countProxySessions()
	if proxySessions != 0 {
		t.Errorf("Expected 0 proxy sessions (deleted), got %d", proxySessions)
	}

	importedSessions := repo.countImportedSessions()
	if importedSessions != 1 {
		t.Errorf("Expected 1 imported session, got %d", importedSessions)
	}
}

// TestImportHAR_ModeReplaceImported tests replace_imported mode
func TestImportHAR_ModeReplaceImported(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Add some existing sessions (mix of proxy and imported)
	repo.sessions["proxy1"] = domain.Session{ID: "proxy1", ClientAddr: "127.0.0.1:1234"}
	repo.sessions["proxy2"] = domain.Session{ID: "proxy2", ClientAddr: "192.168.1.1:5678"}
	repo.sessions["imported1"] = domain.Session{ID: "imported1", ClientAddr: "imported"}
	repo.sessions["imported2"] = domain.Session{ID: "imported2", ClientAddr: "imported"}

	// Load test HAR
	harData := loadTestHAR(t, "simple_http.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har?mode=replace_imported", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["imported"].(float64) != 1 {
		t.Errorf("Expected 1 imported, got %v", result["imported"])
	}

	// Verify sessions: proxy sessions should remain, old imported deleted
	totalSessions := repo.countSessions()
	if totalSessions != 3 {
		t.Errorf("Expected 3 total sessions (2 proxy + 1 new imported), got %d", totalSessions)
	}

	proxySessions := repo.countProxySessions()
	if proxySessions != 2 {
		t.Errorf("Expected 2 proxy sessions (preserved), got %d", proxySessions)
	}

	importedSessions := repo.countImportedSessions()
	if importedSessions != 1 {
		t.Errorf("Expected 1 imported session (old deleted, new added), got %d", importedSessions)
	}
}

// TestImportHAR_InvalidMode tests invalid mode parameter
func TestImportHAR_InvalidMode(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Load test HAR
	harData := loadTestHAR(t, "simple_http.har")

	// Create request with invalid mode
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har?mode=invalid_mode", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatalf("Expected error object in response, got %v", result)
	}

	if errorObj["code"] != "INVALID_MODE" {
		t.Errorf("Expected error code INVALID_MODE, got %v", errorObj["code"])
	}
}

// TestImportHAR_SimpleHTTP tests importing simple HTTP request
func TestImportHAR_SimpleHTTP(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Load test HAR
	harData := loadTestHAR(t, "simple_http.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["imported"].(float64) != 1 {
		t.Errorf("Expected 1 imported, got %v", result["imported"])
	}

	if result["total"].(float64) != 1 {
		t.Errorf("Expected 1 total, got %v", result["total"])
	}

	// Verify session created
	if len(repo.sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(repo.sessions))
	}

	// Verify session details
	var session domain.Session
	for _, s := range repo.sessions {
		session = s
	}

	if session.ClientAddr != "imported" {
		t.Errorf("Expected ClientAddr 'imported', got %s", session.ClientAddr)
	}

	if session.Kind != "http" {
		t.Errorf("Expected Kind 'http', got %s", session.Kind)
	}

	if session.Target != "https://api.example.com/users/123" {
		t.Errorf("Expected Target URL, got %s", session.Target)
	}

	// Verify HTTP transaction created
	txs := repo.transactions[session.ID]
	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	tx := txs[0]
	if tx.Method != "GET" {
		t.Errorf("Expected Method GET, got %s", tx.Method)
	}

	if tx.Status != 200 {
		t.Errorf("Expected Status 200, got %d", tx.Status)
	}

	if tx.ContentType != "application/json" {
		t.Errorf("Expected ContentType application/json, got %s", tx.ContentType)
	}
}

// TestImportHAR_MultipleEntries tests importing multiple HTTP entries
func TestImportHAR_MultipleEntries(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Load test HAR with 5 entries
	harData := loadTestHAR(t, "multiple_entries.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["imported"].(float64) != 5 {
		t.Errorf("Expected 5 imported, got %v", result["imported"])
	}

	if result["total"].(float64) != 5 {
		t.Errorf("Expected 5 total, got %v", result["total"])
	}

	// Verify sessions created
	if len(repo.sessions) != 5 {
		t.Errorf("Expected 5 sessions, got %d", len(repo.sessions))
	}

	// Verify all sessions are HTTP type
	for _, s := range repo.sessions {
		if s.Kind != "http" {
			t.Errorf("Expected Kind 'http', got %s for session %s", s.Kind, s.ID)
		}
		if s.ClientAddr != "imported" {
			t.Errorf("Expected ClientAddr 'imported', got %s for session %s", s.ClientAddr, s.ID)
		}
	}

	// Verify each session has a transaction
	for sessionID := range repo.sessions {
		txs := repo.transactions[sessionID]
		if len(txs) != 1 {
			t.Errorf("Expected 1 transaction for session %s, got %d", sessionID, len(txs))
		}
	}
}

// TestImportHAR_WebSocket tests importing WebSocket session with messages
func TestImportHAR_WebSocket(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Load test HAR with WebSocket messages
	harData := loadTestHAR(t, "websocket.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["imported"].(float64) != 1 {
		t.Errorf("Expected 1 imported, got %v", result["imported"])
	}

	// Verify session created with kind "ws"
	if len(repo.sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(repo.sessions))
	}

	var session domain.Session
	for _, s := range repo.sessions {
		session = s
	}

	if session.Kind != "ws" {
		t.Errorf("Expected Kind 'ws', got %s", session.Kind)
	}

	if session.Target != "wss://ws.example.com/socket" {
		t.Errorf("Expected WebSocket URL, got %s", session.Target)
	}

	// Verify WebSocket frames created
	frames := repo.frames[session.ID]
	if len(frames) != 4 {
		t.Errorf("Expected 4 WebSocket frames, got %d", len(frames))
	}

	// Verify frame directions
	expectedDirections := []domain.Direction{
		domain.DirectionClientToUpstream, // send
		domain.DirectionUpstreamToClient, // receive
		domain.DirectionClientToUpstream, // send
		domain.DirectionUpstreamToClient, // receive
	}

	for i, frame := range frames {
		if frame.Direction != expectedDirections[i] {
			t.Errorf("Frame %d: expected direction %v, got %v", i, expectedDirections[i], frame.Direction)
		}

		if frame.Opcode != domain.OpcodeText {
			t.Errorf("Frame %d: expected Opcode Text, got %v", i, frame.Opcode)
		}
	}

	// Verify frame data
	if frames[0].Preview != "Hello Server" {
		t.Errorf("Frame 0: expected 'Hello Server', got %s", frames[0].Preview)
	}

	if frames[1].Preview != "Hello Client" {
		t.Errorf("Frame 1: expected 'Hello Client', got %s", frames[1].Preview)
	}
}

// TestImportHAR_InvalidVersion tests HAR file with unsupported version
func TestImportHAR_InvalidVersion(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Load test HAR with version 2.0 (unsupported)
	harData := loadTestHAR(t, "invalid_version.har")

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response - should fail with UNSUPPORTED_VERSION
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatalf("Expected error object in response, got %v", result)
	}

	if errorObj["code"] != "UNSUPPORTED_VERSION" {
		t.Errorf("Expected error code UNSUPPORTED_VERSION, got %v", errorObj["code"])
	}

	// Verify no sessions created
	if len(repo.sessions) != 0 {
		t.Errorf("Expected 0 sessions (import should fail), got %d", len(repo.sessions))
	}
}

// TestImportHAR_InvalidJSON tests malformed JSON
func TestImportHAR_InvalidJSON(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Create invalid JSON
	invalidJSON := []byte(`{invalid json`)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har", bytes.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatalf("Expected error object in response, got %v", result)
	}

	if errorObj["code"] != "INVALID_HAR" {
		t.Errorf("Expected error code INVALID_HAR, got %v", errorObj["code"])
	}
}

// TestImportHAR_MethodNotAllowed tests non-POST methods
func TestImportHAR_MethodNotAllowed(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Create GET request (not allowed)
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/import/har", nil)
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestImportHAR_DefaultModeIsMerge tests that default mode is merge
func TestImportHAR_DefaultModeIsMerge(t *testing.T) {
	repo := newMockRepoForImport()
	deps := createTestDeps(repo)

	// Add existing session
	repo.sessions["proxy1"] = domain.Session{ID: "proxy1", ClientAddr: "127.0.0.1:1234"}

	// Load test HAR
	harData := loadTestHAR(t, "simple_http.har")

	// Create request WITHOUT mode parameter
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/import/har", bytes.NewReader(harData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	deps.handleImportHAR(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify both sessions exist (merge behavior)
	totalSessions := repo.countSessions()
	if totalSessions != 2 {
		t.Errorf("Expected 2 sessions (merge mode), got %d", totalSessions)
	}

	proxySessions := repo.countProxySessions()
	if proxySessions != 1 {
		t.Errorf("Expected 1 proxy session (preserved), got %d", proxySessions)
	}

	importedSessions := repo.countImportedSessions()
	if importedSessions != 1 {
		t.Errorf("Expected 1 imported session, got %d", importedSessions)
	}
}
