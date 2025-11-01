package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mdomain "network-debugger/internal/features/mapping/domain"
	mappingrt "network-debugger/internal/features/mapping/runtime"
	mappinguc "network-debugger/internal/features/mapping/usecase"
	"network-debugger/internal/infrastructure/config"
)

// Mock mapping repository
type mockMappingRepo struct {
	listFunc    func(ctx context.Context) ([]mdomain.MapRule, error)
	upsertFunc  func(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error)
	deleteFunc  func(ctx context.Context, id string) error
	reorderFunc func(ctx context.Context, ids []string) error
}

func (m *mockMappingRepo) List(ctx context.Context) ([]mdomain.MapRule, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []mdomain.MapRule{}, nil
}

func (m *mockMappingRepo) Upsert(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, r)
	}
	return r, nil
}

func (m *mockMappingRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockMappingRepo) Reorder(ctx context.Context, ids []string) error {
	if m.reorderFunc != nil {
		return m.reorderFunc(ctx, ids)
	}
	return nil
}

func setupMappingDeps() *Deps {
	repo := &mockMappingRepo{}
	return &Deps{
		Cfg: config.Config{
			BodySpoolDir: "/tmp/test-spool",
		},
		Mapping: mappinguc.NewService(repo),
		MapRt:   mappingrt.New(),
	}
}

func setupMappingDepsWithRepo(repo *mockMappingRepo) *Deps {
	return &Deps{
		Cfg: config.Config{
			BodySpoolDir: "/tmp/test-spool",
		},
		Mapping: mappinguc.NewService(repo),
		MapRt:   mappingrt.New(),
	}
}

func TestHandleMappingConfig_GET(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/mapping/config", nil)
	w := httptest.NewRecorder()

	deps.handleMappingConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp mappingConfigDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Enabled {
		t.Error("Expected Enabled = true")
	}
	if resp.UploadMaxMB != 20 {
		t.Errorf("UploadMaxMB = %d, want 20", resp.UploadMaxMB)
	}
}

func TestHandleMappingConfig_POST(t *testing.T) {
	deps := setupMappingDeps()

	input := mappingConfigDTO{Enabled: false, UploadMaxMB: 50}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/config", bytes.NewReader(body))
	w := httptest.NewRecorder()

	deps.handleMappingConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp mappingConfigDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Enabled != input.Enabled {
		t.Errorf("Enabled = %v, want %v", resp.Enabled, input.Enabled)
	}
	if resp.UploadMaxMB != input.UploadMaxMB {
		t.Errorf("UploadMaxMB = %d, want %d", resp.UploadMaxMB, input.UploadMaxMB)
	}
}

func TestHandleMappingConfig_POST_BadJSON(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/config", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	deps.handleMappingConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleMappingConfig_InvalidMethod(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/mapping/config", nil)
	w := httptest.NewRecorder()

	deps.handleMappingConfig(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleMappingRules_GET(t *testing.T) {
	repo := &mockMappingRepo{
		listFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
			return []mdomain.MapRule{
				{ID: "rule1", Enabled: true, Priority: 1},
				{ID: "rule2", Enabled: false, Priority: 2},
			}, nil
		},
	}
	deps := setupMappingDepsWithRepo(repo)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/mapping/rules", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []mapRuleDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Got %d rules, want 2", len(resp))
	}
	if resp[0].ID != "rule1" {
		t.Errorf("First rule ID = %q, want rule1", resp[0].ID)
	}
}

func TestHandleMappingRules_GET_NoService(t *testing.T) {
	deps := setupMappingDeps()
	deps.Mapping = nil

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/mapping/rules", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleMappingRules_GET_ListError(t *testing.T) {
	repo := &mockMappingRepo{
		listFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
			return nil, errors.New("list failed")
		},
	}
	deps := setupMappingDepsWithRepo(repo)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/mapping/rules", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleMappingRules_POST_Upsert(t *testing.T) {
	upsertCalled := false
	repo := &mockMappingRepo{
		upsertFunc: func(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
			upsertCalled = true
			return r, nil
		},
		listFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
			return []mdomain.MapRule{}, nil
		},
	}
	deps := setupMappingDepsWithRepo(repo)

	input := mapRuleDTO{
		ID:       "test-rule",
		Enabled:  true,
		Priority: 10,
		Kind:     "local",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/rules", bytes.NewReader(body))
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if !upsertCalled {
		t.Error("Upsert was not called")
	}

	var resp mapRuleDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.ID != input.ID {
		t.Errorf("ID = %q, want %q", resp.ID, input.ID)
	}
}

func TestHandleMappingRules_POST_Reorder(t *testing.T) {
	reorderCalled := false
	var receivedIDs []string
	repo := &mockMappingRepo{
		reorderFunc: func(ctx context.Context, ids []string) error {
			reorderCalled = true
			receivedIDs = ids
			return nil
		},
		listFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
			return []mdomain.MapRule{}, nil
		},
	}
	deps := setupMappingDepsWithRepo(repo)

	ids := []string{"id1", "id2", "id3"}
	body, _ := json.Marshal(ids)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/rules/reorder", bytes.NewReader(body))
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if !reorderCalled {
		t.Error("Reorder was not called")
	}

	if len(receivedIDs) != 3 || receivedIDs[0] != "id1" {
		t.Errorf("receivedIDs = %v, want %v", receivedIDs, ids)
	}
}

func TestHandleMappingRules_POST_ReorderBadJSON(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/rules/reorder", strings.NewReader("bad json"))
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleMappingRules_POST_UpsertBadJSON(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/rules", strings.NewReader("{invalid}"))
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleMappingRules_InvalidMethod(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodPut, "/_api/v1/mapping/rules", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRules(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleMappingRuleByID_DELETE(t *testing.T) {
	deletedID := ""
	repo := &mockMappingRepo{
		deleteFunc: func(ctx context.Context, id string) error {
			deletedID = id
			return nil
		},
		listFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
			return []mdomain.MapRule{}, nil
		},
	}
	deps := setupMappingDepsWithRepo(repo)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/mapping/rules/test-id-123", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRuleByID(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}

	if deletedID != "test-id-123" {
		t.Errorf("deletedID = %q, want test-id-123", deletedID)
	}
}

func TestHandleMappingRuleByID_NoService(t *testing.T) {
	deps := setupMappingDeps()
	deps.Mapping = nil

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/mapping/rules/test-id", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRuleByID(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleMappingRuleByID_InvalidMethod(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/mapping/rules/test-id", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRuleByID(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleMappingRuleByID_MissingID(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/mapping/rules/", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRuleByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleMappingRuleByID_DeleteError(t *testing.T) {
	repo := &mockMappingRepo{
		deleteFunc: func(ctx context.Context, id string) error {
			return errors.New("delete failed")
		},
	}
	deps := setupMappingDepsWithRepo(repo)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/mapping/rules/test-id", nil)
	w := httptest.NewRecorder()

	deps.handleMappingRuleByID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleMappingUpload_Success(t *testing.T) {
	deps := setupMappingDeps()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.json")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, _ = io.WriteString(part, `{"test": "data"}`)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	deps.handleMappingUpload(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["fileName"] != "test.json" {
		t.Errorf("fileName = %v, want test.json", resp["fileName"])
	}
	if resp["blobPath"] == nil || resp["blobPath"] == "" {
		t.Error("Expected non-empty blobPath")
	}
	// ContentType detection may vary, just check it exists
	if resp["contentType"] == nil || resp["contentType"] == "" {
		t.Error("Expected non-empty contentType")
	}
}

func TestHandleMappingUpload_InvalidMethod(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/mapping/upload", nil)
	w := httptest.NewRecorder()

	deps.handleMappingUpload(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleMappingUpload_BadMultipart(t *testing.T) {
	deps := setupMappingDeps()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/upload", strings.NewReader("not multipart"))
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	deps.handleMappingUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleMappingUpload_NoFile(t *testing.T) {
	deps := setupMappingDeps()

	// Create multipart form without file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("other", "value")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mapping/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	deps.handleMappingUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
