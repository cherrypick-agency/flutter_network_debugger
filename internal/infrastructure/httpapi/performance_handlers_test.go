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
	performanceuc "network-debugger/internal/features/performance/usecase"
	"network-debugger/internal/usecase"
)

type mockSessionRepoForPerformance struct {
	listFunc func(ctx context.Context, filter usecase.SessionFilter) ([]domain.Session, int, error)
}

func (m *mockSessionRepoForPerformance) CreateSession(ctx context.Context, s domain.Session) error {
	return nil
}

func (m *mockSessionRepoForPerformance) GetSession(ctx context.Context, id string) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (m *mockSessionRepoForPerformance) DeleteSession(ctx context.Context, id string) error {
	return nil
}

func (m *mockSessionRepoForPerformance) ListSessions(ctx context.Context, filter usecase.SessionFilter) ([]domain.Session, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []domain.Session{}, 0, nil
}

func (m *mockSessionRepoForPerformance) IncrementCounters(ctx context.Context, id string, frame domain.Frame) error {
	return nil
}

func (m *mockSessionRepoForPerformance) SetClosed(ctx context.Context, id string, closedAt time.Time, errMsg *string) error {
	return nil
}

func (m *mockSessionRepoForPerformance) ClearAllSessions(ctx context.Context) error {
	return nil
}

func (m *mockSessionRepoForPerformance) DeleteImportedSessions(ctx context.Context) error {
	return nil
}

func (m *mockSessionRepoForPerformance) AddSpoolFile(ctx context.Context, sessionID string, path string) {
}

type mockFrameRepoForPerformance struct{}

func (m *mockFrameRepoForPerformance) AppendFrame(ctx context.Context, sessionID string, f domain.Frame) error {
	return nil
}

func (m *mockFrameRepoForPerformance) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	return nil, "", nil
}

func (m *mockFrameRepoForPerformance) GetFrameByID(ctx context.Context, sessionID string, frameID string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
}

type mockEventRepoForPerformance struct{}

func (m *mockEventRepoForPerformance) AppendEvent(ctx context.Context, sessionID string, e domain.Event) error {
	return nil
}

func (m *mockEventRepoForPerformance) ListEvents(ctx context.Context, sessionID string, from string, limit int) ([]domain.Event, string, error) {
	return nil, "", nil
}

func setupPerformanceDeps(service *performanceuc.Service) *Deps {
	return &Deps{
		PerformanceSvc: service,
	}
}

func TestHandlePerformanceOverview_GET_Success(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	mockRepo := &mockSessionRepoForPerformance{
		listFunc: func(ctx context.Context, filter usecase.SessionFilter) ([]domain.Session, int, error) {
			return []domain.Session{}, 0, nil
		},
	}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	fromStr := from.Format(time.RFC3339)
	toStr := to.Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview?from="+fromStr+"&to="+toStr, nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		return
	}

	var resp performanceOverviewDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.TimeRange.From.IsZero() {
		t.Error("TimeRange.From should not be zero")
	}
	if resp.TimeRange.To.IsZero() {
		t.Error("TimeRange.To should not be zero")
	}
	if resp.ResponseTimeStats.Count != 0 {
		t.Errorf("ResponseTimeStats.Count = %d, want 0", resp.ResponseTimeStats.Count)
	}
}

func TestHandlePerformanceOverview_GET_DefaultTimeRange(t *testing.T) {
	mockRepo := &mockSessionRepoForPerformance{
		listFunc: func(ctx context.Context, filter usecase.SessionFilter) ([]domain.Session, int, error) {
			return []domain.Session{}, 0, nil
		},
	}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview", nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandlePerformanceOverview_GET_InvalidFromFormat(t *testing.T) {
	mockRepo := &mockSessionRepoForPerformance{}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview?from=invalid-date", nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePerformanceOverview_GET_InvalidToFormat(t *testing.T) {
	from := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	mockRepo := &mockSessionRepoForPerformance{}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview?from="+from+"&to=invalid-date", nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePerformanceOverview_GET_InvalidTimeRange(t *testing.T) {
	from := time.Now().Format(time.RFC3339)
	to := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	mockRepo := &mockSessionRepoForPerformance{}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview?from="+from+"&to="+to, nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePerformanceOverview_GET_ServiceUnavailable(t *testing.T) {
	deps := &Deps{
		PerformanceSvc: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview", nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandlePerformanceOverview_GET_ServiceError(t *testing.T) {
	mockRepo := &mockSessionRepoForPerformance{
		listFunc: func(ctx context.Context, filter usecase.SessionFilter) ([]domain.Session, int, error) {
			return nil, 0, errors.New("database error")
		},
	}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview", nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandlePerformanceOverview_InvalidMethod(t *testing.T) {
	mockRepo := &mockSessionRepoForPerformance{}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/performance/overview", nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlePerformanceOverview_GET_WithCustomTimeRange(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	mockRepo := &mockSessionRepoForPerformance{
		listFunc: func(ctx context.Context, filter usecase.SessionFilter) ([]domain.Session, int, error) {
			if filter.TimeFrom != nil && !filter.TimeFrom.Equal(from) {
				t.Errorf("From time = %v, want %v", filter.TimeFrom, from)
			}
			if filter.TimeTo != nil && !filter.TimeTo.Equal(to) {
				t.Errorf("To time = %v, want %v", filter.TimeTo, to)
			}
			return []domain.Session{}, 0, nil
		},
	}
	sessionSvc := usecase.NewSessionService(mockRepo, &mockFrameRepoForPerformance{}, &mockEventRepoForPerformance{})
	service := performanceuc.NewService(sessionSvc)
	deps := setupPerformanceDeps(service)

	fromStr := from.Format(time.RFC3339)
	toStr := to.Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/performance/overview?from="+fromStr+"&to="+toStr, nil)
	w := httptest.NewRecorder()

	deps.handlePerformanceOverview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}
