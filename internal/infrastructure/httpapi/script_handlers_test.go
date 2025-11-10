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
	"reflect"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

// Composer 1.
// Mock ScriptRepository для тестирования handlers
type mockScriptRepo struct {
	scripts           map[string]*domain.Script
	saveFunc          func(ctx context.Context, script *domain.Script) error
	getFunc           func(ctx context.Context, id string) (*domain.Script, error)
	listFunc          func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error)
	deleteFunc        func(ctx context.Context, id string) error
	updateEnabledFunc func(ctx context.Context, id string, enabled bool) error
}

func (m *mockScriptRepo) Save(ctx context.Context, script *domain.Script) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, script)
	}
	if m.scripts == nil {
		m.scripts = make(map[string]*domain.Script)
	}
	m.scripts[script.ID] = script
	return nil
}

func (m *mockScriptRepo) Get(ctx context.Context, id string) (*domain.Script, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	if m.scripts == nil {
		return nil, errors.New("not found")
	}
	script, ok := m.scripts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return script, nil
}

func (m *mockScriptRepo) List(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	if m.scripts == nil {
		return []*domain.Script{}, nil
	}
	result := make([]*domain.Script, 0, len(m.scripts))
	for _, script := range m.scripts {
		result = append(result, script)
	}
	return result, nil
}

func (m *mockScriptRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	if m.scripts != nil {
		delete(m.scripts, id)
	}
	return nil
}

func (m *mockScriptRepo) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	if m.updateEnabledFunc != nil {
		return m.updateEnabledFunc(ctx, id, enabled)
	}
	if m.scripts != nil && m.scripts[id] != nil {
		m.scripts[id].Enabled = enabled
	}
	return nil
}

// Mock Compiler для тестирования compilation handlers
type mockCompiler struct {
	language     string
	isAvailable  bool
	compileFunc  func(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error)
	validateFunc func(ctx context.Context, req domain.CompileRequest) error
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
		WASMBinary: []byte("wasm binary"),
		WASMSize:   100,
		Duration:   time.Second,
		Logs:       []string{"Compiled"},
	}, nil
}

func (m *mockCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, req)
	}
	return nil
}

func (m *mockCompiler) ValidateDependencies(deps map[string]string) error {
	return nil
}

func setupScriptHandlers(repo domain.ScriptRepository, compilationService *usecase.CompilationService) *ScriptHandlers {
	scriptService := usecase.NewScriptService(repo)
	// Регистрируем mock executor для тестирования TestScript
	mockExecutor := &mockExecutor{runtime: domain.RuntimeExtism}
	scriptService.RegisterExecutor(mockExecutor)

	handlers := NewScriptHandlers(scriptService)
	if compilationService != nil {
		handlers.SetCompilationService(compilationService)
	}
	return handlers
}

// Mock executor для тестирования
type mockExecutor struct {
	runtime domain.ScriptRuntime
}

func (m *mockExecutor) Runtime() domain.ScriptRuntime {
	return m.runtime
}

func (m *mockExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	result := domain.ScriptResult{
		Modified: true,
		ModifiedRequest: &domain.HTTPRequest{
			Method: "GET",
			URL:    "https://example.com",
		},
	}
	output, _ := json.Marshal(result)
	return domain.ExecutionResult{
		Output: output,
		Logs:   []string{"Test log"},
	}, nil
}

func (m *mockExecutor) Validate(ctx context.Context, script domain.Script) error {
	return nil
}

func (m *mockExecutor) Close() error {
	return nil
}

// Используем рефлексию для установки приватного поля service (только для тестов)
func setPrivateField(obj interface{}, fieldName string, value interface{}) {
	v := reflect.ValueOf(obj).Elem()
	f := v.FieldByName(fieldName)
	if f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(value))
	}
}

func TestScriptHandlers_CreateScript_WithBase64Code(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
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
		"description": "Test Description",
		"runtime":     "extism",
		"code":        codeBase64,
		"language":    "rust",
		"triggerType": "request",
		"priority":    10,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.CreateScript(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response domain.Script
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID == "" {
		t.Error("Expected response to contain 'id' field")
	}
	if response.Name != "Test Script" {
		t.Errorf("Expected name 'Test Script', got '%s'", response.Name)
	}
}

func TestScriptHandlers_CreateScript_InvalidBase64(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"name":     "Test Script",
		"code":     "invalid base64!!!",
		"language": "rust",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.CreateScript(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_CreateScript_WithSourceCode(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.SourceCode != "fn main() {}" {
				t.Errorf("Expected sourceCode 'fn main() {}', got '%s'", script.SourceCode)
			}
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
		"name":       "Test Script",
		"sourceCode": "fn main() {}",
		"language":   "rust",
		"runtime":    "extism",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.CreateScript(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["id"] == "" {
		t.Error("Expected response to contain 'id' field")
	}
}

func TestScriptHandlers_CreateScript_CompilationFailure(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
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

	// Act
	handlers.CreateScript(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_CreateScript_WithMatchRules(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if len(script.MatchRules.Methods) != 1 || script.MatchRules.Methods[0] != "GET" {
				t.Errorf("Expected match rule method 'GET', got %v", script.MatchRules.Methods)
			}
			if script.MatchRules.HostPattern != "example.com" {
				t.Errorf("Expected hostPattern 'example.com', got '%s'", script.MatchRules.HostPattern)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	codeBytes := []byte("test wasm code")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"name":     "Test Script",
		"code":     codeBase64,
		"language": "rust",
		"runtime":  "extism",
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

	// Act
	handlers.CreateScript(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestScriptHandlers_CreateScript_InvalidRegexPattern(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	codeBytes := []byte("test wasm code")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"name":     "Test Script",
		"code":     codeBase64,
		"language": "rust",
		"runtime":  "extism",
		"matchRules": map[string]interface{}{
			"hostPattern": "[invalid regex",
			"patternType": "regex",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.CreateScript(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ListScripts_Success(t *testing.T) {
	// Arrange
	scripts := []*domain.Script{
		{ID: "1", Name: "Script 1"},
		{ID: "2", Name: "Script 2"},
	}
	repo := &mockScriptRepo{
		listFunc: func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
			return scripts, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.ListScripts(w, req)

	// Assert
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
	// Arrange
	repo := &mockScriptRepo{
		listFunc: func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
			return nil, errors.New("database error")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.ListScripts(w, req)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_GetScript_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	script := &domain.Script{
		ID:   scriptID,
		Name: "Test Script",
	}
	repo := &mockScriptRepo{
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

	// Act
	handlers.GetScript(w, req)

	// Assert
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
	// Arrange
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	// Act
	handlers.GetScript(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_GetScript_MissingID(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.GetScript(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Old Name",
		Runtime:     domain.RuntimeExtism,
		Code:        []byte("wasm"),
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
		Enabled:     true,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
	}
	repo := &mockScriptRepo{
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

	// Act
	handlers.UpdateScript(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_NotFound(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
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

	// Act
	handlers.UpdateScript(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_WithSourceCode(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Test Script",
		SourceCode:  "old code",
		Code:        []byte("old wasm"),
		Enabled:     true,
		Runtime:     domain.RuntimeExtism,
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
	}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.SourceCode != "new code" {
				t.Errorf("Expected sourceCode 'new code', got '%s'", script.SourceCode)
			}
			if len(script.Code) != 0 {
				t.Error("Expected code to be cleared after sourceCode update")
			}
			if script.Enabled {
				t.Error("Expected script to be disabled after sourceCode update")
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

	// Act
	handlers.UpdateScript(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_DeleteScript_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{ID: id, Runtime: domain.RuntimeExtism}, nil
		},
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

	// Act
	handlers.DeleteScript(w, req)

	// Assert
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestScriptHandlers_DeleteScript_ServiceError(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{ID: id, Runtime: domain.RuntimeExtism}, nil
		},
		deleteFunc: func(ctx context.Context, id string) error {
			return errors.New("delete failed")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/scripts/test-id", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	// Act
	handlers.DeleteScript(w, req)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	repo := &mockScriptRepo{
		updateEnabledFunc: func(ctx context.Context, id string, enabled bool) error {
			if id != scriptID {
				t.Errorf("Expected id '%s', got '%s'", scriptID, id)
			}
			if !enabled {
				t.Error("Expected enabled to be true")
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

	// Act
	handlers.ToggleScript(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_InvalidBody(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodPatch, "/_api/v1/scripts/test-id/toggle", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	// Act
	handlers.ToggleScript(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_TestScript_Success(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{}
	handlers := setupScriptHandlers(repo, nil)

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
			"headers": map[string][]string{
				"Content-Type": {"application/json"},
			},
			"body": "test body",
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.TestScript(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}
}

func TestScriptHandlers_TestScript_InvalidBase64(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"script": map[string]interface{}{
			"code": "invalid base64!!!",
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

	// Act
	handlers.TestScript(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ListExamples_Success(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/examples", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.ListExamples(w, req)

	// Assert
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
	// Arrange
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
	}
	repo := &mockScriptRepo{
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

	// Create ZIP file in memory
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	fw, _ := zipWriter.Create("main.rs")
	fw.Write([]byte("fn main() {}"))
	fw2, _ := zipWriter.Create("Cargo.toml")
	fw2.Write([]byte("[package]\nname = \"test\""))
	zipWriter.Close()

	// Create multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("project", "project.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	// Act
	handlers.UploadProject(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}
}

func TestScriptHandlers_UploadProject_NotZIP(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepo{
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

	// Act
	handlers.UploadProject(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UploadProject_NoFiles(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	// Create empty ZIP
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

	// Act
	handlers.UploadProject(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_DownloadProject_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Name:       "Test Script",
		SourceCode: "fn main() {}",
		Language:   "rust",
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID+"/download-project", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	// Act
	handlers.DownloadProject(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/zip" {
		t.Errorf("Expected Content-Type 'application/zip', got '%s'", w.Header().Get("Content-Type"))
	}

	// Verify ZIP contents
	zipData := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	if len(zipReader.File) == 0 {
		t.Error("Expected at least one file in ZIP")
	}
}

func TestScriptHandlers_ListProjectFiles_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		SourceCode: "fn main() {}",
		Language:   "rust",
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID+"/files", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	// Act
	handlers.ListProjectFiles(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if fileCount, ok := response["fileCount"].(float64); !ok || fileCount < 1 {
		t.Error("Expected at least one file")
	}
}

func TestScriptHandlers_ExportScriptAsZip_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Name:       "Test Script",
		SourceCode: "fn main() {}",
		Language:   "rust",
		Code:       []byte("wasm binary"),
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/"+scriptID+"/export-zip", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	// Act
	handlers.ExportScriptAsZip(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/zip" {
		t.Errorf("Expected Content-Type 'application/zip', got '%s'", w.Header().Get("Content-Type"))
	}

	// Verify ZIP contains metadata.json
	zipData := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read ZIP: %v", err)
	}

	foundMetadata := false
	for _, file := range zipReader.File {
		if file.Name == "metadata.json" {
			foundMetadata = true
			break
		}
	}
	if !foundMetadata {
		t.Error("Expected metadata.json in ZIP")
	}
}

func TestScriptHandlers_ImportScriptFromZip_Success(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.Name != "Imported Script" {
				t.Errorf("Expected name 'Imported Script', got '%s'", script.Name)
			}
			return nil
		},
	}
	// Регистрируем executor для валидации
	scriptService := usecase.NewScriptService(repo)
	mockExecutor := &mockExecutor{runtime: domain.RuntimeExtism}
	scriptService.RegisterExecutor(mockExecutor)
	handlers := NewScriptHandlers(scriptService)

	// Create ZIP with metadata and files
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"name":        "Imported Script",
		"description": "Test Description",
		"language":    "rust",
		"runtime":     "extism",
		"sourceCode":  "fn main() {}",
		"dependencies": map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	metadataJSON, _ := json.Marshal(metadata)
	metadataFile, _ := zipWriter.Create("metadata.json")
	metadataFile.Write(metadataJSON)

	wasmFile, _ := zipWriter.Create("output.wasm")
	wasmFile.Write([]byte("wasm binary"))
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Act
	handlers.ImportScriptFromZip(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestScriptHandlers_ImportScriptFromZip_MissingMetadata(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	// Create ZIP without metadata.json
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	zipWriter.Create("main.rs")
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Act
	handlers.ImportScriptFromZip(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ImportScriptFromZip_InvalidZIP(t *testing.T) {
	// Arrange
	handlers := setupScriptHandlers(nil, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write([]byte("not a zip"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Act
	handlers.ImportScriptFromZip(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
