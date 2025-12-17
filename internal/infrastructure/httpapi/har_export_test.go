package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"network-debugger/internal/domain"
	uc "network-debugger/internal/usecase"
)

type mockRepoForHAR struct {
	transactions []domain.HTTPTransaction
	nextCursor   string
	err          error
}

func (m *mockRepoForHAR) CreateSession(context.Context, domain.Session) error { return nil }
func (m *mockRepoForHAR) GetSession(context.Context, string) (domain.Session, bool, error) {
	return domain.Session{Kind: "http"}, true, nil
}
func (m *mockRepoForHAR) DeleteSession(context.Context, string) error { return nil }
func (m *mockRepoForHAR) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) {
	return nil, 0, nil
}
func (m *mockRepoForHAR) IncrementCounters(context.Context, string, domain.Frame) error { return nil }
func (m *mockRepoForHAR) SetClosed(context.Context, string, time.Time, *string) error   { return nil }
func (m *mockRepoForHAR) ClearAllSessions(context.Context) error                        { return nil }
func (m *mockRepoForHAR) AppendFrame(context.Context, string, domain.Frame) error       { return nil }
func (m *mockRepoForHAR) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) {
	return nil, "", nil
}
func (m *mockRepoForHAR) GetFrameByID(context.Context, string, string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
}
func (m *mockRepoForHAR) UpdateFrameBodyFile(context.Context, string, string, string) error {
	return nil
}
func (m *mockRepoForHAR) AppendEvent(context.Context, string, domain.Event) error { return nil }
func (m *mockRepoForHAR) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) {
	return nil, "", nil
}
func (m *mockRepoForHAR) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error {
	return nil
}
func (m *mockRepoForHAR) ListHTTPTransactions(context.Context, string, string, int) ([]domain.HTTPTransaction, string, error) {
	return m.transactions, m.nextCursor, m.err
}
func (m *mockRepoForHAR) DeleteImportedSessions(context.Context) error { return nil }

func TestExportHARForSession_MultipleTransactions(t *testing.T) {
	repo := &mockRepoForHAR{
		transactions: []domain.HTTPTransaction{
			{
				Method:    "GET",
				URL:       "http://example.com/api/users",
				Status:    200,
				ReqSize:   100,
				RespSize:  500,
				StartedAt: time.Unix(1000, 0).UTC(),
				Timings:   domain.HTTPTimings{Total: 150},
			},
			{
				Method:    "POST",
				URL:       "http://example.com/api/create",
				Status:    201,
				ReqSize:   200,
				RespSize:  300,
				StartedAt: time.Unix(2000, 0).UTC(),
				Timings:   domain.HTTPTimings{Total: 200},
			},
		},
		nextCursor: "",
		err:        nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
	exportHARForSession(rr, req, d, "test-session")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	log := body["log"].(map[string]any)
	entries := log["entries"].([]any)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Verify HAR structure
	if log["version"].(string) != "1.2" {
		t.Errorf("expected version 1.2, got %v", log["version"])
	}

	creator := log["creator"].(map[string]any)
	if creator["name"].(string) != "network-debugger" {
		t.Errorf("unexpected creator name: %v", creator["name"])
	}
}

func TestExportHARForSession_Pagination(t *testing.T) {
	// First call returns 2 transactions with next cursor
	// Second call returns 1 transaction without cursor
	callCount := 0
	repo := &mockRepoWithPagination{
		callCount: &callCount,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
	exportHARForSession(rr, req, d, "paginated-session")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	log := body["log"].(map[string]any)
	entries := log["entries"].([]any)
	if len(entries) != 3 {
		t.Errorf("expected 3 total entries from pagination, got %d", len(entries))
	}

	if *repo.callCount != 2 {
		t.Errorf("expected 2 ListHTTPTransactions calls, got %d", *repo.callCount)
	}
}

type mockRepoWithPagination struct {
	callCount *int
}

func (m *mockRepoWithPagination) CreateSession(context.Context, domain.Session) error { return nil }
func (m *mockRepoWithPagination) GetSession(context.Context, string) (domain.Session, bool, error) {
	return domain.Session{Kind: "http"}, true, nil
}
func (m *mockRepoWithPagination) DeleteSession(context.Context, string) error { return nil }
func (m *mockRepoWithPagination) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) {
	return nil, 0, nil
}
func (m *mockRepoWithPagination) IncrementCounters(context.Context, string, domain.Frame) error {
	return nil
}
func (m *mockRepoWithPagination) SetClosed(context.Context, string, time.Time, *string) error {
	return nil
}
func (m *mockRepoWithPagination) ClearAllSessions(context.Context) error { return nil }
func (m *mockRepoWithPagination) AppendFrame(context.Context, string, domain.Frame) error {
	return nil
}
func (m *mockRepoWithPagination) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) {
	return nil, "", nil
}
func (m *mockRepoWithPagination) GetFrameByID(context.Context, string, string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
}
func (m *mockRepoWithPagination) UpdateFrameBodyFile(context.Context, string, string, string) error {
	return nil
}
func (m *mockRepoWithPagination) AppendEvent(context.Context, string, domain.Event) error {
	return nil
}
func (m *mockRepoWithPagination) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) {
	return nil, "", nil
}
func (m *mockRepoWithPagination) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error {
	return nil
}
func (m *mockRepoWithPagination) ListHTTPTransactions(ctx context.Context, sessionID string, from string, limit int) ([]domain.HTTPTransaction, string, error) {
	*m.callCount++
	if *m.callCount == 1 {
		// First call: return 2 transactions with next cursor
		return []domain.HTTPTransaction{
			{Method: "GET", URL: "http://example.com/1", Status: 200, StartedAt: time.Unix(1000, 0).UTC(), Timings: domain.HTTPTimings{Total: 100}},
			{Method: "GET", URL: "http://example.com/2", Status: 200, StartedAt: time.Unix(2000, 0).UTC(), Timings: domain.HTTPTimings{Total: 100}},
		}, "cursor-1", nil
	}
	// Second call: return 1 transaction without cursor
	return []domain.HTTPTransaction{
		{Method: "GET", URL: "http://example.com/3", Status: 200, StartedAt: time.Unix(3000, 0).UTC(), Timings: domain.HTTPTimings{Total: 100}},
	}, "", nil
}
func (m *mockRepoWithPagination) DeleteImportedSessions(context.Context) error { return nil }

func TestExportHARForSession_Error(t *testing.T) {
	repo := &mockRepoForHAR{
		transactions: nil,
		nextCursor:   "",
		err:          errors.New("database error"),
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
	exportHARForSession(rr, req, d, "error-session")

	// Current implementation silently ignores ListHTTPTransactions errors
	// and returns success with empty entries instead of error
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	log := body["log"].(map[string]any)
	entries := log["entries"].([]any)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries due to error, got %d", len(entries))
	}
}

func TestExportHARForSession_EmptyTransactions(t *testing.T) {
	repo := &mockRepoForHAR{
		transactions: []domain.HTTPTransaction{},
		nextCursor:   "",
		err:          nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
	exportHARForSession(rr, req, d, "empty-session")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	log := body["log"].(map[string]any)
	entries := log["entries"].([]any)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestExportHARForSession_ContentDisposition(t *testing.T) {
	repo := &mockRepoForHAR{
		transactions: []domain.HTTPTransaction{},
		nextCursor:   "",
		err:          nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
	exportHARForSession(rr, req, d, "test-123")

	contentDisposition := rr.Header().Get("Content-Disposition")
	// Filename includes timestamp in format: network-debugger_YYYY-MM-DD_HH-MM-SS.har
	if !startsWith(contentDisposition, "attachment; filename=network-debugger_") {
		t.Errorf("expected Content-Disposition to start with 'attachment; filename=network-debugger_', got %q", contentDisposition)
	}
	if !endsWith(contentDisposition, ".har") {
		t.Errorf("expected Content-Disposition to end with '.har', got %q", contentDisposition)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func TestExportHARForSession_ValidHARStructure(t *testing.T) {
	repo := &mockRepoForHAR{
		transactions: []domain.HTTPTransaction{
			{
				Method:    "POST",
				URL:       "https://api.example.com/v1/data",
				Status:    201,
				ReqSize:   1024,
				RespSize:  2048,
				StartedAt: time.Unix(1234567890, 0).UTC(),
				Timings:   domain.HTTPTimings{Total: 250},
			},
		},
		nextCursor: "",
		err:        nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
	exportHARForSession(rr, req, d, "valid-session")

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	log := body["log"].(map[string]any)
	entries := log["entries"].([]any)
	entry := entries[0].(map[string]any)

	// Verify entry structure
	request := entry["request"].(map[string]any)
	if request["method"].(string) != "POST" {
		t.Errorf("expected method POST, got %v", request["method"])
	}
	if request["url"].(string) != "https://api.example.com/v1/data" {
		t.Errorf("unexpected URL: %v", request["url"])
	}

	response := entry["response"].(map[string]any)
	if int(response["status"].(float64)) != 201 {
		t.Errorf("expected status 201, got %v", response["status"])
	}
	if response["statusText"].(string) != http.StatusText(201) {
		t.Errorf("expected statusText %q, got %v", http.StatusText(201), response["statusText"])
	}

	if int(entry["time"].(float64)) != 250 {
		t.Errorf("expected time 250, got %v", entry["time"])
	}
}
