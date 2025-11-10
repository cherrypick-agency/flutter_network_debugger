package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

type mockScriptRepository struct {
	saveFunc          func(ctx context.Context, script *domain.Script) error
	getFunc           func(ctx context.Context, id string) (*domain.Script, error)
	listFunc          func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error)
	deleteFunc        func(ctx context.Context, id string) error
	updateEnabledFunc func(ctx context.Context, id string, enabled bool) error
}

func (m *mockScriptRepository) Save(ctx context.Context, script *domain.Script) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, script)
	}
	return nil
}

func (m *mockScriptRepository) Get(ctx context.Context, id string) (*domain.Script, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return &domain.Script{ID: id}, nil
}

func (m *mockScriptRepository) List(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*domain.Script{}, nil
}

func (m *mockScriptRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockScriptRepository) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	if m.updateEnabledFunc != nil {
		return m.updateEnabledFunc(ctx, id, enabled)
	}
	return nil
}

type mockScriptExecutor struct {
	validateFunc func(ctx context.Context, script domain.Script) error
	executeFunc  func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error)
	runtime      domain.ScriptRuntime
}

func (m *mockScriptExecutor) Validate(ctx context.Context, script domain.Script) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, script)
	}
	return nil
}

func (m *mockScriptExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return domain.ExecutionResult{}, nil
}

func (m *mockScriptExecutor) Runtime() domain.ScriptRuntime {
	if m.runtime != "" {
		return m.runtime
	}
	return domain.RuntimeExtism
}

func (m *mockScriptExecutor) Close() error {
	return nil
}

type mockCompiler struct {
	language    string
	isAvailable bool
	compileFunc func(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error)
}

func (m *mockCompiler) Language() string {
	return m.language
}

func (m *mockCompiler) IsAvailable() bool {
	return m.isAvailable
}

func (m *mockCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	if m.compileFunc != nil {
		return m.compileFunc(ctx, req)
	}
	return &domain.CompileResult{
		WASMBinary: []byte("wasm"),
		WASMSize:   100,
		Duration:   time.Second,
		Logs:       []string{"Compiled"},
	}, nil
}

func (m *mockCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	return nil
}

func (m *mockCompiler) ValidateDependencies(deps map[string]string) error {
	return nil
}

func setupScriptHandlers(repo *mockScriptRepository, compilationService *usecase.CompilationService) *ScriptHandlers {
	executor := &mockScriptExecutor{runtime: domain.RuntimeExtism}
	service := usecase.NewScriptService(repo)
	service.RegisterExecutor(executor)
	handlers := NewScriptHandlers(service)
	if compilationService != nil {
		handlers.SetCompilationService(compilationService)
	}
	return handlers
}

func TestScriptHandlers_CreateScript_WithBase64Code(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.Name != "Test Script" {
				t.Errorf("Expected name 'Test Script', got '%s'", script.Name)
			}
			if len(script.Code) == 0 {
				t.Error("Expected code to be set")
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	codeBytes := []byte("test wasm code")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"name":        "Test Script",
		"code":        codeBase64,
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
		"priority":    10,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateScript(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestScriptHandlers_CreateScript_InvalidBase64(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"name":     "Test Script",
		"code":     "invalid base64!!!",
		"language": "rust",
		"runtime":  "extism",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_CreateScript_WithSourceCode(t *testing.T) {
	scripts := make(map[string]*domain.Script)
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			scripts[script.ID] = script
			return nil
		},
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			if script, ok := scripts[id]; ok {
				return script, nil
			}
			return nil, errors.New("not found")
		},
		deleteFunc: func(ctx context.Context, id string) error {
			delete(scripts, id)
			return nil
		},
	}
	compiler := &mockCompiler{
		language:    "rust",
		isAvailable: true,
		compileFunc: func(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
			return &domain.CompileResult{
				WASMBinary: []byte("wasm"),
				WASMSize:   100,
				Duration:   time.Second,
				Logs:       []string{"Compiled"},
			}, nil
		},
	}
	compilationService := usecase.NewCompilationService(repo)
	compilationService.RegisterCompiler(compiler)
	handlers := setupScriptHandlers(repo, compilationService)

	reqBody := map[string]interface{}{
		"name":        "Test Script",
		"sourceCode":  "fn main() {}",
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
		"priority":    10,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateScript(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestScriptHandlers_CreateScript_CompilationFailure(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			return nil
		},
		deleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}
	compiler := &mockCompiler{
		language:    "rust",
		isAvailable: true,
		compileFunc: func(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
			return nil, errors.New("compilation failed")
		},
	}
	compilationService := usecase.NewCompilationService(repo)
	compilationService.RegisterCompiler(compiler)
	handlers := setupScriptHandlers(repo, compilationService)

	reqBody := map[string]interface{}{
		"name":       "Test Script",
		"sourceCode": "invalid rust code",
		"language":   "rust",
		"runtime":    "extism",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_CreateScript_WithMatchRules(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if len(script.MatchRules.Methods) != 1 || script.MatchRules.Methods[0] != "GET" {
				t.Errorf("Expected match rule method GET, got %v", script.MatchRules.Methods)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	codeBytes := []byte("test wasm code")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"name":        "Test Script",
		"code":        codeBase64,
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
		"priority":    10,
		"matchRules": map[string]interface{}{
			"methods":     []string{"GET"},
			"hostPattern": "example.com",
			"pathPattern": "/api/*",
			"patternType": "wildcard",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateScript(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestScriptHandlers_CreateScript_InvalidRegexPattern(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	codeBytes := []byte("test wasm code")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"name":        "Test Script",
		"code":        codeBase64,
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
		"priority":    10,
		"matchRules": map[string]interface{}{
			"hostPattern": "[invalid regex",
			"patternType": "regex",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ListScripts_Success(t *testing.T) {
	scripts := []*domain.Script{
		{ID: "1", Name: "Script 1"},
		{ID: "2", Name: "Script 2"},
	}
	repo := &mockScriptRepository{
		listFunc: func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
			return scripts, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts", nil)
	w := httptest.NewRecorder()

	handlers.ListScripts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []*domain.Script
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(response))
	}
}

func TestScriptHandlers_ListScripts_ServiceError(t *testing.T) {
	repo := &mockScriptRepository{
		listFunc: func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
			return nil, errors.New("service error")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts", nil)
	w := httptest.NewRecorder()

	handlers.ListScripts(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_GetScript_Success(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:   scriptID,
		Name: "Test Script",
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			if id != scriptID {
				t.Errorf("Expected id '%s', got '%s'", scriptID, id)
			}
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID, nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.GetScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response domain.Script
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != scriptID {
		t.Errorf("Expected id '%s', got '%s'", scriptID, response.ID)
	}
}

func TestScriptHandlers_GetScript_NotFound(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.GetScript(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_GetScript_MissingID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/", nil)
	w := httptest.NewRecorder()

	handlers.GetScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_Success(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Old Name",
		Runtime:     domain.RuntimeExtism,
		Code:        []byte("wasm"),
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.Name != "New Name" {
				t.Errorf("Expected name 'New Name', got '%s'", script.Name)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	reqBody := map[string]interface{}{
		"name": "New Name",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/_api/v1/scripts/"+scriptID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UpdateScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_NotFound(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	reqBody := map[string]interface{}{
		"name": "New Name",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/_api/v1/scripts/test-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.UpdateScript(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_WithSourceCode(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		Code:        []byte("wasm"),
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.SourceCode != "new code" {
				t.Errorf("Expected sourceCode 'new code', got '%s'", script.SourceCode)
			}
			if len(script.Code) != 0 {
				t.Error("Expected code to be cleared when sourceCode is updated")
			}
			if script.Enabled {
				t.Error("Expected script to be disabled when sourceCode is updated")
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	reqBody := map[string]interface{}{
		"sourceCode": "new code",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/_api/v1/scripts/"+scriptID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UpdateScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_DeleteScript_Success(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		deleteFunc: func(ctx context.Context, id string) error {
			if id != scriptID {
				t.Errorf("Expected id '%s', got '%s'", scriptID, id)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/scripts/"+scriptID, nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.DeleteScript(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestScriptHandlers_DeleteScript_ServiceError(t *testing.T) {
	repo := &mockScriptRepository{
		deleteFunc: func(ctx context.Context, id string) error {
			return errors.New("delete failed")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/scripts/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.DeleteScript(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_Success(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		updateEnabledFunc: func(ctx context.Context, id string, enabled bool) error {
			if id != scriptID {
				t.Errorf("Expected id '%s', got '%s'", scriptID, id)
			}
			if !enabled {
				t.Error("Expected enabled = true")
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	reqBody := map[string]interface{}{
		"enabled": true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/_api/v1/scripts/"+scriptID+"/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.ToggleScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_ServiceError(t *testing.T) {
	repo := &mockScriptRepository{
		updateEnabledFunc: func(ctx context.Context, id string, enabled bool) error {
			return errors.New("toggle failed")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	reqBody := map[string]interface{}{
		"enabled": true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/_api/v1/scripts/test-id/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.ToggleScript(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_TestScript_Success(t *testing.T) {
	repo := &mockScriptRepository{}
	executor := &mockScriptExecutor{
		runtime: domain.RuntimeExtism,
		executeFunc: func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
			return domain.ExecutionResult{}, nil
		},
	}
	service := usecase.NewScriptService(repo)
	service.RegisterExecutor(executor)
	handlers := NewScriptHandlers(service)

	codeBytes := []byte("test wasm code")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"script": map[string]interface{}{
			"name":        "Test Script",
			"runtime":     "extism",
			"code":        codeBase64,
			"language":    "rust",
			"triggerType": "request",
			"priority":    10,
		},
		"testRequest": map[string]interface{}{
			"method": "GET",
			"url":    "https://example.com",
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.TestScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_TestScript_InvalidBase64(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"script": map[string]interface{}{
			"name":     "Test Script",
			"runtime":  "extism",
			"code":     "invalid base64!!!",
			"language": "rust",
		},
		"testRequest": map[string]interface{}{
			"method": "GET",
			"url":    "https://example.com",
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.TestScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ListExamples_Success(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/examples", nil)
	w := httptest.NewRecorder()

	handlers.ListExamples(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []ScriptExample
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) == 0 {
		t.Error("Expected at least one example")
	}
}

func TestScriptHandlers_UploadProject_Success(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Test Script",
		Language:    "rust",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte("wasm"),
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if len(script.Dependencies) == 0 {
				t.Error("Expected dependencies to be set")
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	fw, _ := zipWriter.Create("main.rs")
	fw.Write([]byte("fn main() {}"))
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("project", "project.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_UploadProject_NotZIP(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("project", "project.txt")
	part.Write([]byte("not a zip"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UploadProject_NoFiles(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("project", "project.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_DownloadProject_Success(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID+"/download-project", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.DownloadProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/zip" {
		t.Errorf("Expected Content-Type 'application/zip', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestScriptHandlers_ListProjectFiles_Success(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Language:   "rust",
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID+"/files", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.ListProjectFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if fileCount, ok := response["fileCount"].(float64); !ok || fileCount != 2 {
		t.Errorf("Expected fileCount = 2, got %v", response["fileCount"])
	}
}

func TestScriptHandlers_ExportScriptAsZip_Success(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
		Code:       []byte("wasm"),
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID+"/export-zip", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.ExportScriptAsZip(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/zip" {
		t.Errorf("Expected Content-Type 'application/zip', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestScriptHandlers_ImportScriptFromZip_Success(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.Name != "Imported Script" {
				t.Errorf("Expected name 'Imported Script', got '%s'", script.Name)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"name":        "Imported Script",
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
		"sourceCode":  "fn main() {}",
	}
	metadataFile, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metadataFile).Encode(metadata)

	mainFile, _ := zipWriter.Create("src/lib.rs")
	mainFile.Write([]byte("fn main() {}"))

	wasmFile, _ := zipWriter.Create("output.wasm")
	wasmFile.Write([]byte("wasm"))
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.ImportScriptFromZip(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestScriptHandlers_ImportScriptFromZip_MissingMetadata(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	mainFile, _ := zipWriter.Create("main.rs")
	mainFile.Write([]byte("fn main() {}"))
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.ImportScriptFromZip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ImportScriptFromZip_InvalidZIP(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write([]byte("not a zip"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.ImportScriptFromZip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
