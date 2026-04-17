package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tagsdomain "network-debugger/internal/features/tags/domain"
	tagsuc "network-debugger/internal/features/tags/usecase"
)

// Composer 1.
// Mock tags repository for testing handlers
type mockTagsRepo struct {
	listPredefinedTagsFunc      func(ctx context.Context) ([]tagsdomain.PredefinedTag, error)
	createPredefinedTagFunc     func(ctx context.Context, tag tagsdomain.PredefinedTag) error
	deletePredefinedTagFunc     func(ctx context.Context, id string) error
	getSessionTagsFunc          func(ctx context.Context, sessionID string) ([]tagsdomain.SessionTag, error)
	addSessionTagFunc           func(ctx context.Context, tag tagsdomain.SessionTag) error
	removeSessionTagFunc        func(ctx context.Context, sessionID, tagName string) error
	bulkAddSessionTagsFunc      func(ctx context.Context, sessionIDs []string, tagNames []string) error
	bulkRemoveSessionTagsFunc   func(ctx context.Context, sessionIDs []string, tagNames []string) error
	getSessionAnnotationsFunc   func(ctx context.Context, sessionID string) ([]tagsdomain.SessionAnnotation, error)
	upsertSessionAnnotationFunc func(ctx context.Context, annotation tagsdomain.SessionAnnotation) error
	deleteSessionAnnotationFunc func(ctx context.Context, sessionID, key string) error
}

func (m *mockTagsRepo) ListPredefinedTags(ctx context.Context) ([]tagsdomain.PredefinedTag, error) {
	if m.listPredefinedTagsFunc != nil {
		return m.listPredefinedTagsFunc(ctx)
	}
	return []tagsdomain.PredefinedTag{}, nil
}

func (m *mockTagsRepo) CreatePredefinedTag(ctx context.Context, tag tagsdomain.PredefinedTag) error {
	if m.createPredefinedTagFunc != nil {
		return m.createPredefinedTagFunc(ctx, tag)
	}
	return nil
}

func (m *mockTagsRepo) DeletePredefinedTag(ctx context.Context, id string) error {
	if m.deletePredefinedTagFunc != nil {
		return m.deletePredefinedTagFunc(ctx, id)
	}
	return nil
}

func (m *mockTagsRepo) GetSessionTags(ctx context.Context, sessionID string) ([]tagsdomain.SessionTag, error) {
	if m.getSessionTagsFunc != nil {
		return m.getSessionTagsFunc(ctx, sessionID)
	}
	return []tagsdomain.SessionTag{}, nil
}

func (m *mockTagsRepo) AddSessionTag(ctx context.Context, tag tagsdomain.SessionTag) error {
	if m.addSessionTagFunc != nil {
		return m.addSessionTagFunc(ctx, tag)
	}
	return nil
}

func (m *mockTagsRepo) RemoveSessionTag(ctx context.Context, sessionID, tagName string) error {
	if m.removeSessionTagFunc != nil {
		return m.removeSessionTagFunc(ctx, sessionID, tagName)
	}
	return nil
}

func (m *mockTagsRepo) BulkAddSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	if m.bulkAddSessionTagsFunc != nil {
		return m.bulkAddSessionTagsFunc(ctx, sessionIDs, tagNames)
	}
	return nil
}

func (m *mockTagsRepo) BulkRemoveSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	if m.bulkRemoveSessionTagsFunc != nil {
		return m.bulkRemoveSessionTagsFunc(ctx, sessionIDs, tagNames)
	}
	return nil
}

func (m *mockTagsRepo) DeleteAllSessionTags(ctx context.Context, sessionID string) error {
	return nil
}

func (m *mockTagsRepo) FindSessionIDsByTags(ctx context.Context, tagNames []string) ([]string, error) {
	return []string{}, nil
}

func (m *mockTagsRepo) GetSessionAnnotations(ctx context.Context, sessionID string) ([]tagsdomain.SessionAnnotation, error) {
	if m.getSessionAnnotationsFunc != nil {
		return m.getSessionAnnotationsFunc(ctx, sessionID)
	}
	return []tagsdomain.SessionAnnotation{}, nil
}

func (m *mockTagsRepo) UpsertSessionAnnotation(ctx context.Context, annotation tagsdomain.SessionAnnotation) error {
	if m.upsertSessionAnnotationFunc != nil {
		return m.upsertSessionAnnotationFunc(ctx, annotation)
	}
	return nil
}

func (m *mockTagsRepo) DeleteSessionAnnotation(ctx context.Context, sessionID, key string) error {
	if m.deleteSessionAnnotationFunc != nil {
		return m.deleteSessionAnnotationFunc(ctx, sessionID, key)
	}
	return nil
}

func (m *mockTagsRepo) DeleteAllSessionAnnotations(ctx context.Context, sessionID string) error {
	return nil
}

func setupTagsDeps(repo *mockTagsRepo) *Deps {
	return &Deps{
		TagsSvc: tagsuc.NewService(repo),
	}
}

// Composer 1.
// Tests for handlePredefinedTags

func TestHandlePredefinedTags_GET_Success(t *testing.T) {
	repo := &mockTagsRepo{
		listPredefinedTagsFunc: func(ctx context.Context) ([]tagsdomain.PredefinedTag, error) {
			return []tagsdomain.PredefinedTag{
				{ID: "tag1", Name: "Important", Color: "#ff0000", Category: "priority", DisplayOrder: 1, CreatedAt: time.Now()},
				{ID: "tag2", Name: "Bug", Color: "#00ff00", Category: "type", DisplayOrder: 2, CreatedAt: time.Now()},
			}, nil
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/tags/predefined", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("Response should contain 'items' array")
	}

	if len(items) != 2 {
		t.Errorf("Got %d items, want 2", len(items))
	}
}

func TestHandlePredefinedTags_GET_ServiceError(t *testing.T) {
	repo := &mockTagsRepo{
		listPredefinedTagsFunc: func(ctx context.Context) ([]tagsdomain.PredefinedTag, error) {
			return nil, errors.New("database error")
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/tags/predefined", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandlePredefinedTags_POST_Success(t *testing.T) {
	var createdTag tagsdomain.PredefinedTag
	repo := &mockTagsRepo{
		createPredefinedTagFunc: func(ctx context.Context, tag tagsdomain.PredefinedTag) error {
			createdTag = tag
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	body := map[string]any{
		"name":         "New Tag",
		"color":        "#0000ff",
		"category":     "custom",
		"displayOrder": 10,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/predefined", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	if createdTag.Name != "New Tag" {
		t.Errorf("Tag name = %q, want 'New Tag'", createdTag.Name)
	}
	if createdTag.Color != "#0000ff" {
		t.Errorf("Tag color = %q, want '#0000ff'", createdTag.Color)
	}
}

func TestHandlePredefinedTags_POST_DefaultValues(t *testing.T) {
	var createdTag tagsdomain.PredefinedTag
	repo := &mockTagsRepo{
		createPredefinedTagFunc: func(ctx context.Context, tag tagsdomain.PredefinedTag) error {
			createdTag = tag
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	body := map[string]any{
		"name": "Tag Without Color",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/predefined", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	if createdTag.Color != "#808080" {
		t.Errorf("Default color = %q, want '#808080'", createdTag.Color)
	}
	if createdTag.Category != "general" {
		t.Errorf("Default category = %q, want 'general'", createdTag.Category)
	}
}

func TestHandlePredefinedTags_POST_MissingName(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	body := map[string]any{
		"color": "#ff0000",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/predefined", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePredefinedTags_POST_NameTooLong(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	longName := strings.Repeat("a", 101)
	body := map[string]any{
		"name": longName,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/predefined", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePredefinedTags_POST_InvalidJSON(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/predefined", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePredefinedTags_InvalidMethod(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/tags/predefined", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlePredefinedTags_FeatureDisabled(t *testing.T) {
	deps := &Deps{TagsSvc: nil}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/tags/predefined", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTags(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

// Composer 1.
// Tests for handlePredefinedTagByID

func TestHandlePredefinedTagByID_DELETE_Success(t *testing.T) {
	deletedID := ""
	repo := &mockTagsRepo{
		deletePredefinedTagFunc: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/tags/predefined/tag-123", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTagByID(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if deletedID != "tag-123" {
		t.Errorf("deletedID = %q, want 'tag-123'", deletedID)
	}
}

func TestHandlePredefinedTagByID_DELETE_MissingID(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/tags/predefined/", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTagByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandlePredefinedTagByID_DELETE_ServiceError(t *testing.T) {
	repo := &mockTagsRepo{
		deletePredefinedTagFunc: func(ctx context.Context, id string) error {
			return errors.New("delete failed")
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/tags/predefined/tag-123", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTagByID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandlePredefinedTagByID_InvalidMethod(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/tags/predefined/tag-123", nil)
	w := httptest.NewRecorder()

	deps.handlePredefinedTagByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// Composer 1.
// Tests for handleSessionTags

func TestHandleSessionTags_GET_Success(t *testing.T) {
	repo := &mockTagsRepo{
		getSessionTagsFunc: func(ctx context.Context, sessionID string) ([]tagsdomain.SessionTag, error) {
			return []tagsdomain.SessionTag{
				{ID: "st1", SessionID: "session-123", TagName: "important", CreatedAt: time.Now()},
				{ID: "st2", SessionID: "session-123", TagName: "bug", CreatedAt: time.Now()},
			}, nil
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/session-123/tags", nil)
	w := httptest.NewRecorder()

	deps.handleSessionTags(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("Response should contain 'items' array")
	}

	if len(items) != 2 {
		t.Errorf("Got %d items, want 2", len(items))
	}
}

func TestHandleSessionTags_POST_Success(t *testing.T) {
	var addedTag tagsdomain.SessionTag
	repo := &mockTagsRepo{
		addSessionTagFunc: func(ctx context.Context, tag tagsdomain.SessionTag) error {
			addedTag = tag
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	body := map[string]any{
		"tagName": "new-tag",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionTags(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	if addedTag.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want 'session-123'", addedTag.SessionID)
	}
	if addedTag.TagName != "new-tag" {
		t.Errorf("TagName = %q, want 'new-tag'", addedTag.TagName)
	}
}

func TestHandleSessionTags_POST_MissingTagName(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	body := map[string]any{}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleSessionTags_POST_TagNameTooLong(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	longTagName := strings.Repeat("a", 101)
	body := map[string]any{
		"tagName": longTagName,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/tags", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleSessionTags_DELETE_Success(t *testing.T) {
	var deletedSessionID, deletedTagName string
	repo := &mockTagsRepo{
		removeSessionTagFunc: func(ctx context.Context, sessionID, tagName string) error {
			deletedSessionID = sessionID
			deletedTagName = tagName
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions/session-123/tags/important", nil)
	w := httptest.NewRecorder()

	deps.handleSessionTags(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if deletedSessionID != "session-123" {
		t.Errorf("SessionID = %q, want 'session-123'", deletedSessionID)
	}
	if deletedTagName != "important" {
		t.Errorf("TagName = %q, want 'important'", deletedTagName)
	}
}

func TestHandleSessionTags_InvalidPath(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/session-123/invalid", nil)
	w := httptest.NewRecorder()

	deps.handleSessionTags(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// Composer 1.
// Tests for handleBulkTags

func TestHandleBulkTags_POST_Add_Success(t *testing.T) {
	var receivedSessionIDs, receivedTagNames []string
	repo := &mockTagsRepo{
		bulkAddSessionTagsFunc: func(ctx context.Context, sessionIDs []string, tagNames []string) error {
			receivedSessionIDs = sessionIDs
			receivedTagNames = tagNames
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	body := map[string]any{
		"operation":  "add",
		"sessionIds": []string{"session-1", "session-2"},
		"tagNames":   []string{"tag1", "tag2"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if len(receivedSessionIDs) != 2 {
		t.Errorf("Got %d sessionIDs, want 2", len(receivedSessionIDs))
	}
	if len(receivedTagNames) != 2 {
		t.Errorf("Got %d tagNames, want 2", len(receivedTagNames))
	}
}

func TestHandleBulkTags_POST_Remove_Success(t *testing.T) {
	repo := &mockTagsRepo{
		bulkRemoveSessionTagsFunc: func(ctx context.Context, sessionIDs []string, tagNames []string) error {
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	body := map[string]any{
		"operation":  "remove",
		"sessionIds": []string{"session-1"},
		"tagNames":   []string{"tag1"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleBulkTags_POST_TooManySessions(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	sessionIDs := make([]string, 101)
	for i := range sessionIDs {
		sessionIDs[i] = "session-" + string(rune(i))
	}

	body := map[string]any{
		"operation":  "add",
		"sessionIds": sessionIDs,
		"tagNames":   []string{"tag1"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleBulkTags_POST_TooManyTags(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	tagNames := make([]string, 51)
	for i := range tagNames {
		tagNames[i] = "tag-" + string(rune(i))
	}

	body := map[string]any{
		"operation":  "add",
		"sessionIds": []string{"session-1"},
		"tagNames":   tagNames,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleBulkTags_POST_InvalidOperation(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	body := map[string]any{
		"operation":  "invalid",
		"sessionIds": []string{"session-1"},
		"tagNames":   []string{"tag1"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleBulkTags_POST_EmptyArrays(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	body := map[string]any{
		"operation":  "add",
		"sessionIds": []string{},
		"tagNames":   []string{},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleBulkTags_InvalidMethod(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/tags/bulk", nil)
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleBulkTags_FeatureDisabled(t *testing.T) {
	deps := &Deps{TagsSvc: nil}

	body := map[string]any{
		"operation":  "add",
		"sessionIds": []string{"session-1"},
		"tagNames":   []string{"tag1"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/tags/bulk", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleBulkTags(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

// Composer 1.
// Tests for handleSessionAnnotations

func TestHandleSessionAnnotations_GET_Success(t *testing.T) {
	repo := &mockTagsRepo{
		getSessionAnnotationsFunc: func(ctx context.Context, sessionID string) ([]tagsdomain.SessionAnnotation, error) {
			return []tagsdomain.SessionAnnotation{
				{ID: "ann1", SessionID: "session-123", Key: "note", Value: "Important session", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, nil
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/session-123/annotations", nil)
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	items, ok := resp["items"].([]any)
	if !ok {
		t.Fatal("Response should contain 'items' array")
	}

	if len(items) != 1 {
		t.Errorf("Got %d items, want 1", len(items))
	}
}

func TestHandleSessionAnnotations_POST_Success(t *testing.T) {
	var addedAnnotation tagsdomain.SessionAnnotation
	repo := &mockTagsRepo{
		upsertSessionAnnotationFunc: func(ctx context.Context, annotation tagsdomain.SessionAnnotation) error {
			addedAnnotation = annotation
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	body := map[string]any{
		"key":   "note",
		"value": "Test annotation",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/annotations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusCreated)
	}

	if addedAnnotation.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want 'session-123'", addedAnnotation.SessionID)
	}
	if addedAnnotation.Key != "note" {
		t.Errorf("Key = %q, want 'note'", addedAnnotation.Key)
	}
}

func TestHandleSessionAnnotations_POST_MissingKey(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	body := map[string]any{
		"value": "Test annotation",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/annotations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleSessionAnnotations_POST_KeyTooLong(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	longKey := strings.Repeat("a", 256)
	body := map[string]any{
		"key":   longKey,
		"value": "Test",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/annotations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleSessionAnnotations_POST_ValueTooLong(t *testing.T) {
	deps := setupTagsDeps(&mockTagsRepo{})

	longValue := strings.Repeat("a", 10001)
	body := map[string]any{
		"key":   "note",
		"value": longValue,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/sessions/session-123/annotations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleSessionAnnotations_DELETE_Success(t *testing.T) {
	var deletedSessionID, deletedKey string
	repo := &mockTagsRepo{
		deleteSessionAnnotationFunc: func(ctx context.Context, sessionID, key string) error {
			deletedSessionID = sessionID
			deletedKey = key
			return nil
		},
	}
	deps := setupTagsDeps(repo)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions/session-123/annotations/note", nil)
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if deletedSessionID != "session-123" {
		t.Errorf("SessionID = %q, want 'session-123'", deletedSessionID)
	}
	if deletedKey != "note" {
		t.Errorf("Key = %q, want 'note'", deletedKey)
	}
}

func TestHandleSessionAnnotations_FeatureDisabled(t *testing.T) {
	deps := &Deps{TagsSvc: nil}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/session-123/annotations", nil)
	w := httptest.NewRecorder()

	deps.handleSessionAnnotations(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}
