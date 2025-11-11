package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	tagsdomain "network-debugger/internal/features/tags/domain"
	tagsuc "network-debugger/internal/features/tags/usecase"
	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

type mockTagsRepoForSessions struct {
	findSessionIDsByTagsFunc func(ctx context.Context, tagNames []string) ([]string, error)
}

func (m *mockTagsRepoForSessions) ListPredefinedTags(ctx context.Context) ([]tagsdomain.PredefinedTag, error) {
	return nil, nil
}

func (m *mockTagsRepoForSessions) CreatePredefinedTag(ctx context.Context, tag tagsdomain.PredefinedTag) error {
	return nil
}

func (m *mockTagsRepoForSessions) DeletePredefinedTag(ctx context.Context, id string) error {
	return nil
}

func (m *mockTagsRepoForSessions) GetSessionTags(ctx context.Context, sessionID string) ([]tagsdomain.SessionTag, error) {
	return nil, nil
}

func (m *mockTagsRepoForSessions) AddSessionTag(ctx context.Context, tag tagsdomain.SessionTag) error {
	return nil
}

func (m *mockTagsRepoForSessions) RemoveSessionTag(ctx context.Context, sessionID, tagName string) error {
	return nil
}

func (m *mockTagsRepoForSessions) BulkAddSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	return nil
}

func (m *mockTagsRepoForSessions) BulkRemoveSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	return nil
}

func (m *mockTagsRepoForSessions) DeleteAllSessionTags(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockTagsRepoForSessions) FindSessionIDsByTags(ctx context.Context, tagNames []string) ([]string, error) {
	if m.findSessionIDsByTagsFunc != nil {
		return m.findSessionIDsByTagsFunc(ctx, tagNames)
	}
	return []string{}, nil
}

func (m *mockTagsRepoForSessions) GetSessionAnnotations(ctx context.Context, sessionID string) ([]tagsdomain.SessionAnnotation, error) {
	return nil, nil
}

func (m *mockTagsRepoForSessions) UpsertSessionAnnotation(ctx context.Context, annotation tagsdomain.SessionAnnotation) error {
	return nil
}

func (m *mockTagsRepoForSessions) DeleteSessionAnnotation(ctx context.Context, sessionID, key string) error {
	return nil
}

func (m *mockTagsRepoForSessions) DeleteAllSessionAnnotations(ctx context.Context, sessionID string) error {
	return nil
}

func setupSessionsHandlerDeps(t *testing.T) *Deps {
	t.Helper()
	logger := zerolog.New(io.Discard)
	m := obs.NewMetrics()
	cfg := config.Config{CORSAllowOrigin: "*"}
	d := &Deps{
		Cfg:     cfg,
		Logger:  &logger,
		Metrics: m,
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
	}
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)
	return d
}

func TestHandleV1ListSessions_GET_DefaultParams(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithQuery(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?q=test&limit=50", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithTypes(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?types=http,ws&status=200,400", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithLimit(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name  string
		limit string
	}{
		{"default limit", ""},
		{"custom limit", "50"},
		{"too large limit", "2000"}, // Should cap at 1000
		{"zero limit", "0"},         // Should use default 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/_api/v1/sessions"
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			d.handleV1ListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleV1ListSessions_GET_WithOffset(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?offset=10", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithCaptureID(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name      string
		captureID string
	}{
		{"current capture", "current"},
		{"specific capture", "5"},
		{"invalid capture", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captureId="+tt.captureID, nil)
			w := httptest.NewRecorder()

			d.handleV1ListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleV1ListSessions_GET_WithIncludeUnassigned(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name  string
		value string
	}{
		{"true", "true"},
		{"1", "1"},
		{"false", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?includeUnassigned="+tt.value, nil)
			w := httptest.NewRecorder()

			d.handleV1ListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleV1ListSessions_GET_WithCapturesAll(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captures=all", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithScanGraphQL(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?scan=graphql", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithTarget(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?_target=example.com", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_DELETE(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleV1ListSessions_DELETE_WithMonitor(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	d.Monitor = NewMonitorHub()

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleListSessions_GET_WithTags(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	mockTagsRepo := &mockTagsRepoForSessions{
		findSessionIDsByTagsFunc: func(ctx context.Context, tagNames []string) ([]string, error) {
			return []string{"session-1", "session-2"}, nil
		},
	}
	d.TagsSvc = tagsuc.NewService(mockTagsRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?tags=important,bug", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleListSessions_GET_WithTags_EmptyTags(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	mockTagsRepo := &mockTagsRepoForSessions{}
	d.TagsSvc = tagsuc.NewService(mockTagsRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?tags=%2C+%2C", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleListSessions_GET_WithTags_Error(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	mockTagsRepo := &mockTagsRepoForSessions{
		findSessionIDsByTagsFunc: func(ctx context.Context, tagNames []string) ([]string, error) {
			return nil, errors.New("tags error")
		},
	}
	d.TagsSvc = tagsuc.NewService(mockTagsRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?tags=important", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d (should continue without tags filter)", w.Code, http.StatusOK)
	}
}

func TestHandleListSessions_GET_WithTags_NoTagsSvc(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	d.TagsSvc = nil

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?tags=important", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleListSessions_DELETE_WithError(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleListSessions_GET_WithLimit(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name  string
		limit string
		want  int
	}{
		{"default limit", "", 50},
		{"custom limit", "25", 25},
		{"zero limit", "0", 50},
		{"negative limit", "-5", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/sessions"
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			d.handleListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleListSessions_GET_WithOffset(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?offset=10", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleListSessions_GET_WithQueryAndTarget(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?q=test&_target=example.com", nil)
	w := httptest.NewRecorder()

	d.handleListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_DELETE_WithError(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleV1SessionByID_GET_NotFound(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleV1SessionByID_GET_EmptyID(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/", nil)
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleV1SessionByID_DELETE_WithTagsSvc(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	mockTagsRepo := &mockTagsRepoForSessions{}
	d.TagsSvc = tagsuc.NewService(mockTagsRepo)

	ctx := context.Background()
	sessionID := "test-session-1"
	sess := domain.Session{ID: sessionID, Target: "ws://example.com"}
	_ = d.Svc.Create(ctx, sess)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions/"+sessionID, nil)
	req.SetPathValue("id", sessionID)
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleV1SessionByID_GET_Frames(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	ctx := context.Background()
	sessionID := "test-session-1"
	sess := domain.Session{ID: sessionID, Target: "ws://example.com"}
	_ = d.Svc.Create(ctx, sess)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/"+sessionID+"/frames", nil)
	req.SetPathValue("id", sessionID)
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1SessionByID_GET_Events(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	ctx := context.Background()
	sessionID := "test-session-1"
	sess := domain.Session{ID: sessionID, Target: "ws://example.com"}
	_ = d.Svc.Create(ctx, sess)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/"+sessionID+"/events", nil)
	req.SetPathValue("id", sessionID)
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1SessionByID_GET_Body(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	ctx := context.Background()
	sessionID := "test-session-1"
	sess := domain.Session{ID: sessionID, Target: "ws://example.com"}
	_ = d.Svc.Create(ctx, sess)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/"+sessionID+"/body", nil)
	req.SetPathValue("id", sessionID)
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleV1SessionByID_GET_UnknownSubresource(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	ctx := context.Background()
	sessionID := "test-session-1"
	sess := domain.Session{ID: sessionID, Target: "ws://example.com"}
	_ = d.Svc.Create(ctx, sess)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/"+sessionID+"/unknown", nil)
	req.SetPathValue("id", sessionID)
	w := httptest.NewRecorder()

	d.handleV1SessionByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
