package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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
	compilationService := usecase.NewCompilationService(repo, nil)
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
	compilationService := usecase.NewCompilationService(repo, nil)
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

func TestScriptHandlers_CreateScript_WithDependencies(t *testing.T) {
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
	compilationService := usecase.NewCompilationService(repo, nil)
	compilationService.RegisterCompiler(compiler)
	handlers := setupScriptHandlers(repo, compilationService)

	reqBody := map[string]interface{}{
		"name": "Test Script",
		"dependencies": map[string]string{
			"main.rs":    "fn main() {}",
			"Cargo.toml": "[package]\nname = \"test\"",
		},
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

func TestScriptHandlers_CreateScript_NoLanguageWithSourceCode(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"name":       "Test Script",
		"sourceCode": "fn main() {}",
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

func TestScriptHandlers_CreateScript_NoCompilationService(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

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

	handlers.CreateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_CreateScript_NoCodeSourceOrDependencies(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"name":     "Test Script",
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

func TestScriptHandlers_CreateScript_WithConfig(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			if script.Config.TimeoutMs != 10000 {
				t.Errorf("Expected TimeoutMs 10000, got %d", script.Config.TimeoutMs)
			}
			if script.Config.MemoryLimitMB != 20 {
				t.Errorf("Expected MemoryLimitMB 20, got %d", script.Config.MemoryLimitMB)
			}
			if len(script.Config.AllowedHosts) != 1 || script.Config.AllowedHosts[0] != "example.com" {
				t.Errorf("Expected AllowedHosts [example.com], got %v", script.Config.AllowedHosts)
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
		"config": map[string]interface{}{
			"timeoutMs":     10000,
			"memoryLimitMB": 20,
			"allowedHosts":  []string{"example.com"},
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

func TestScriptHandlers_CreateScript_InvalidRegexHostPattern(t *testing.T) {
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

func TestScriptHandlers_CreateScript_InvalidRegexPathPattern(t *testing.T) {
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
			"pathPattern": "[invalid regex",
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

func TestScriptHandlers_CreateScript_ServiceError(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			return errors.New("service error")
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

func TestScriptHandlers_UpdateScript_EmptyID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"name": "New Name",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/_api/v1/scripts/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handlers.UpdateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_InvalidJSON(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodPut, "/_api/v1/scripts/"+scriptID, bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UpdateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_WithConfig(t *testing.T) {
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
			if script.Config.TimeoutMs != 15000 {
				t.Errorf("Expected TimeoutMs 15000, got %d", script.Config.TimeoutMs)
			}
			if script.Config.MemoryLimitMB != 30 {
				t.Errorf("Expected MemoryLimitMB 30, got %d", script.Config.MemoryLimitMB)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	reqBody := map[string]interface{}{
		"config": map[string]interface{}{
			"timeoutMs":     15000,
			"memoryLimitMB": 30,
			"allowedHosts":  []string{"example.com"},
		},
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

func TestScriptHandlers_UpdateScript_ServiceError(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		Code:        []byte("wasm"),
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			return errors.New("update failed")
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

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UpdateScript_WithAllFields(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:          scriptID,
		Name:        "Old Name",
		Description: "Old Description",
		Runtime:     domain.RuntimeExtism,
		Code:        []byte("wasm"),
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    5,
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
			if script.Description != "New Description" {
				t.Errorf("Expected description 'New Description', got '%s'", script.Description)
			}
			if script.Language != "go" {
				t.Errorf("Expected language 'go', got '%s'", script.Language)
			}
			if script.TriggerType != domain.TriggerResponse {
				t.Errorf("Expected triggerType response, got %v", script.TriggerType)
			}
			if script.Priority != 20 {
				t.Errorf("Expected priority 20, got %d", script.Priority)
			}
			return nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	codeBytes := []byte("new wasm")
	codeBase64 := base64.StdEncoding.EncodeToString(codeBytes)

	reqBody := map[string]interface{}{
		"name":        "New Name",
		"description": "New Description",
		"code":        codeBase64,
		"language":    "go",
		"triggerType": "response",
		"priority":    20,
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
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:   id,
				Code: []byte("wasm"),
			}, nil
		},
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
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:   id,
				Code: []byte("wasm"),
			}, nil
		},
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

func TestScriptHandlers_ToggleScript_DartWithSourceCode(t *testing.T) {
	scriptID := "test-dart"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:         id,
				Runtime:    domain.RuntimeDart,
				SourceCode: "void main() {}",
				// Code is empty -- Dart scripts use SourceCode
			}, nil
		},
		updateEnabledFunc: func(ctx context.Context, id string, enabled bool) error {
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
		t.Errorf("Expected status %d for Dart script with SourceCode, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_NoExecutableCode(t *testing.T) {
	scriptID := "test-empty"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:      id,
				Runtime: domain.RuntimeExtism,
				// No Code, no SourceCode
			}, nil
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

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for script without executable code, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_DartNoCode(t *testing.T) {
	scriptID := "test-dart-empty"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:      id,
				Runtime: domain.RuntimeDart,
				// No Code, no SourceCode
			}, nil
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

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for Dart script without any code, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_EmptyID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	reqBody := map[string]interface{}{
		"enabled": true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/_api/v1/scripts//toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handlers.ToggleScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ToggleScript_InvalidJSON(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodPatch, "/_api/v1/scripts/test-id/toggle", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.ToggleScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
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

func TestScriptHandlers_TestScript_InvalidJSON(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.TestScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_TestScript_ExecutionError(t *testing.T) {
	repo := &mockScriptRepository{}
	executor := &mockScriptExecutor{
		runtime: domain.RuntimeExtism,
		executeFunc: func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
			return domain.ExecutionResult{}, errors.New("execution failed")
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

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success=false, got %v", response["success"])
	}
}

func TestScriptHandlers_TestScript_WithMatchRulesAndConfig(t *testing.T) {
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
			"matchRules": map[string]interface{}{
				"methods":     []string{"GET", "POST"},
				"hostPattern": "example.com",
				"pathPattern": "/api/*",
				"patternType": "wildcard",
			},
			"config": map[string]interface{}{
				"timeoutMs":     10000,
				"memoryLimitMB": 20,
				"allowedHosts":  []string{"example.com"},
			},
		},
		"testRequest": map[string]interface{}{
			"method": "GET",
			"url":    "https://example.com/api/test",
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

	handlers.TestScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestScriptHandlers_TestScript_WithModifiedRequest(t *testing.T) {
	repo := &mockScriptRepository{}
	executor := &mockScriptExecutor{
		runtime: domain.RuntimeExtism,
		executeFunc: func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
			scriptResult := domain.ScriptResult{
				Modified: true,
				ModifiedRequest: &domain.HTTPRequest{
					Method: "POST",
					URL:    "https://example.com/modified",
					Headers: map[string][]string{
						"X-Modified": {"true"},
					},
					Body: []byte("modified body"),
				},
			}
			output, _ := json.Marshal(scriptResult)
			result := domain.ExecutionResult{
				Output: output,
			}
			return result, nil
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

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["modifiedRequest"]; !ok {
		t.Error("Expected modifiedRequest in response")
	}
}

func TestScriptHandlers_TestScript_WithModifiedResponse(t *testing.T) {
	repo := &mockScriptRepository{}
	executor := &mockScriptExecutor{
		runtime: domain.RuntimeExtism,
		executeFunc: func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
			scriptResult := domain.ScriptResult{
				Modified: true,
				ModifiedResponse: &domain.HTTPResponse{
					Status: 201,
					Headers: map[string][]string{
						"X-Modified": {"true"},
					},
					Body: []byte("modified response"),
				},
			}
			output, _ := json.Marshal(scriptResult)
			result := domain.ExecutionResult{
				Output: output,
			}
			return result, nil
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
			"triggerType": "response",
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

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := response["modifiedResponse"]; !ok {
		t.Error("Expected modifiedResponse in response")
	}
}

func TestScriptHandlers_TestScript_WithResultError(t *testing.T) {
	repo := &mockScriptRepository{}
	executor := &mockScriptExecutor{
		runtime: domain.RuntimeExtism,
		executeFunc: func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
			result := domain.ExecutionResult{
				Output: []byte(`{"error":"script execution failed"}`),
			}
			return result, nil
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

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || success {
		t.Errorf("Expected success=false, got %v", response["success"])
	}
	if _, ok := response["error"]; !ok {
		t.Error("Expected error in response")
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
	part, _ := writer.CreateFormFile("file", "project.zip")
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
	part, _ := writer.CreateFormFile("file", "project.txt")
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
	part, _ := writer.CreateFormFile("file", "project.zip")
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

func TestScriptHandlers_UploadProject_EmptyID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts//upload-project", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_UploadProject_ScriptNotFound(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test-id/upload-project", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_UploadProject_NoFile(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_DownloadProject_EmptyID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts//download-project", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handlers.DownloadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_DownloadProject_ScriptNotFound(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id/download-project", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.DownloadProject(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_ListProjectFiles_EmptyID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts//files", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handlers.ListProjectFiles(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ExportScriptAsZip_EmptyID(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts//export-zip", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	handlers.ExportScriptAsZip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ExportScriptAsZip_ScriptNotFound(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id/export-zip", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.ExportScriptAsZip(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetMainFilename(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"rust", "src/lib.rs"},
		{"go", "main.go"},
		{"javascript", "index.ts"},
		{"typescript", "index.ts"},
		{"dart", "main.dart"},
		{"python", "main.py"},
		{"zig", "main.zig"},
		{"kotlin", "Main.kt"},
		{"swift", "main.swift"},
		{"c", "main.c"},
		{"cpp", "main.cpp"},
		{"c++", "main.cpp"},
		{"unknown", "main.txt"},
		{"RUST", "src/lib.rs"},
		{"Go", "main.go"},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result := getMainFilename(tt.language)
			if result != tt.expected {
				t.Errorf("getMainFilename(%q) = %q, want %q", tt.language, result, tt.expected)
			}
		})
	}
}

func TestGetDependencyType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"Cargo.toml", "config"},
		{"package.json", "config"},
		{"go.mod", "config"},
		{"go.sum", "config"},
		{"main.rs", "source"},
		{"lib.go", "source"},
		{"index.ts", "source"},
		{"app.js", "source"},
		{"main.dart", "source"},
		{"script.py", "source"},
		{"main.zig", "source"},
		{"Main.kt", "source"},
		{"app.swift", "source"},
		{"main.c", "source"},
		{"app.cpp", "source"},
		{"header.h", "source"},
		{"header.hpp", "source"},
		{"README.md", "other"},
		{"LICENSE", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getDependencyType(tt.filename)
			if result != tt.expected {
				t.Errorf("getDependencyType(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestValidateScriptData(t *testing.T) {
	tests := []struct {
		name    string
		script  *domain.Script
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid script",
			script: &domain.Script{
				Name:     "Test Script",
				Language: "rust",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			script: &domain.Script{
				Language: "rust",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing language",
			script: &domain.Script{
				Name: "Test Script",
			},
			wantErr: true,
			errMsg:  "language is required",
		},
		{
			name: "invalid language",
			script: &domain.Script{
				Name:     "Test Script",
				Language: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid language",
		},
		{
			name: "source code too large",
			script: &domain.Script{
				Name:       "Test Script",
				Language:   "rust",
				SourceCode: string(make([]byte, 500_001)),
			},
			wantErr: true,
			errMsg:  "source code too large",
		},
		{
			name: "dependency file too large",
			script: &domain.Script{
				Name:     "Test Script",
				Language: "rust",
				Dependencies: map[string]string{
					"file.rs": string(make([]byte, 500_001)),
				},
			},
			wantErr: true,
			errMsg:  "dependency file too large",
		},
		{
			name: "total project size too large",
			script: &domain.Script{
				Name:       "Test Script",
				Language:   "rust",
				SourceCode: string(make([]byte, 400_000)),
				Dependencies: map[string]string{
					"file1.rs":  string(make([]byte, 400_000)),
					"file2.rs":  string(make([]byte, 400_000)),
					"file3.rs":  string(make([]byte, 400_000)),
					"file4.rs":  string(make([]byte, 400_000)),
					"file5.rs":  string(make([]byte, 400_000)),
					"file6.rs":  string(make([]byte, 400_000)),
					"file7.rs":  string(make([]byte, 400_000)),
					"file8.rs":  string(make([]byte, 400_000)),
					"file9.rs":  string(make([]byte, 400_000)),
					"file10.rs": string(make([]byte, 400_000)),
					"file11.rs": string(make([]byte, 400_000)),
					"file12.rs": string(make([]byte, 400_000)),
					"file13.rs": string(make([]byte, 400_000)),
				},
			},
			wantErr: true,
			errMsg:  "total project size too large",
		},
		{
			name: "valid language case insensitive",
			script: &domain.Script{
				Name:     "Test Script",
				Language: "RUST",
			},
			wantErr: false,
		},
		{
			name: "valid language go",
			script: &domain.Script{
				Name:     "Test Script",
				Language: "go",
			},
			wantErr: false,
		},
		{
			name: "valid language javascript",
			script: &domain.Script{
				Name:     "Test Script",
				Language: "javascript",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScriptData(tt.script)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateScriptData() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateScriptData() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateScriptData() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestScriptHandlers_ImportScriptFromZip_ValidationError(t *testing.T) {
	repo := &mockScriptRepository{}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"name":        "",
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
	}
	metadataFile, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metadataFile).Encode(metadata)
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

func TestScriptHandlers_ImportScriptFromZip_InvalidLanguage(t *testing.T) {
	repo := &mockScriptRepository{}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"name":        "Test Script",
		"language":    "invalid",
		"runtime":     "extism",
		"triggerType": "request",
	}
	metadataFile, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metadataFile).Encode(metadata)
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

func TestScriptHandlers_ImportScriptFromZip_NoFile(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", nil)
	w := httptest.NewRecorder()

	handlers.ImportScriptFromZip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ImportScriptFromZip_ServiceError(t *testing.T) {
	repo := &mockScriptRepository{
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			return errors.New("service error")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"name":        "Test Script",
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
		"sourceCode":  "fn main() {}",
	}
	metadataFile, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metadataFile).Encode(metadata)
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

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_UploadProject_UpdateScriptError(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{ID: scriptID}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			return errors.New("update failed")
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
	part, _ := writer.CreateFormFile("file", "project.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestScriptHandlers_UploadProject_FileTooLarge(t *testing.T) {
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
	fw, _ := zipWriter.Create("large.rs")
	largeContent := make([]byte, 501*1024)
	fw.Write(largeContent)
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "project.zip")
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

func TestScriptHandlers_UploadProject_TotalSizeExceedsLimit(t *testing.T) {
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
	for i := 0; i < 20; i++ {
		fw, _ := zipWriter.Create(fmt.Sprintf("file%d.rs", i))
		content := make([]byte, 300*1024)
		fw.Write(content)
	}
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "project.zip")
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

func TestScriptHandlers_ImportScriptFromZip_InvalidMetadataJSON(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	metadataFile, _ := zipWriter.Create("metadata.json")
	metadataFile.Write([]byte("{invalid json}"))
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

func TestScriptHandlers_ImportScriptFromZip_MissingName(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"language":    "rust",
		"runtime":     "extism",
		"triggerType": "request",
	}
	metadataFile, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metadataFile).Encode(metadata)
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

func TestScriptHandlers_ImportScriptFromZip_MissingLanguage(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	metadata := map[string]interface{}{
		"name":        "Test Script",
		"runtime":     "extism",
		"triggerType": "request",
	}
	metadataFile, _ := zipWriter.Create("metadata.json")
	json.NewEncoder(metadataFile).Encode(metadata)
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

func TestScriptHandlers_UploadProject_PathTraversal(t *testing.T) {
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
	}
	handlers := setupScriptHandlers(repo, nil)

	// Pre-create a script with all required fields for validation
	testScript := &domain.Script{
		ID:          "test-id",
		Name:        "Test Script",
		Language:    "rust",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
	}
	scripts["test-id"] = testScript
	_ = testScript

	// Create ZIP with path traversal
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	fw, _ := zw.Create("../../etc/passwd")
	fw.Write([]byte("malicious content"))
	zw.Close()

	// Build multipart request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "project.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test-id/upload-project", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for path traversal, got %d", http.StatusBadRequest, w.Code)
	}

	// Verify error message mentions path traversal
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if !strings.Contains(resp["error"], "path traversal") {
		t.Errorf("Expected error about path traversal, got: %s", resp["error"])
	}
}

func TestScriptHandlers_UploadProject_SkipsHiddenFiles(t *testing.T) {
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
	}
	handlers := setupScriptHandlers(repo, nil)

	// Pre-create a script with all required fields for validation
	testScript := &domain.Script{
		ID:          "test-id",
		Name:        "Test Script",
		Language:    "rust",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
	}
	scripts["test-id"] = testScript
	_ = testScript

	// Create ZIP with hidden files and valid files
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	// Hidden file
	fw, _ := zw.Create(".gitignore")
	fw.Write([]byte("target/"))
	// __MACOSX file
	fw2, _ := zw.Create("__MACOSX/._main.rs")
	fw2.Write([]byte("macosx metadata"))
	// Valid source file
	fw3, _ := zw.Create("project/main.rs")
	fw3.Write([]byte("fn main() {}"))
	zw.Close()

	// Build multipart request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "project.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test-id/upload-project", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify only valid file was imported
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	fileCount, _ := resp["fileCount"].(float64)
	if fileCount != 1 {
		t.Errorf("Expected 1 file (only main.rs), got %v", fileCount)
	}
}

func TestScriptHandlers_ImportScriptFromZip_PathTraversal(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	// Create ZIP with metadata.json and a path traversal file
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	// Valid metadata
	metadata := map[string]interface{}{
		"name":          "Malicious Script",
		"language":      "rust",
		"sourceCode":    "fn main() {}",
		"exportVersion": "1.0",
	}
	metaBytes, _ := json.Marshal(metadata)
	fw, _ := zw.Create("metadata.json")
	fw.Write(metaBytes)

	// Path traversal file
	fw2, _ := zw.Create("../../../etc/shadow")
	fw2.Write([]byte("malicious content"))

	zw.Close()

	// Build multipart request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write(zipBuf.Bytes())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.ImportScriptFromZip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for path traversal, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestScriptHandlers_ListProjectFiles_ScriptNotFound(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("script not found")
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/nonexistent/files", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handlers.ListProjectFiles(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestScriptHandlers_ListProjectFiles_NoDependencies(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:           id,
				Name:         "Test Script",
				Language:     "rust",
				Dependencies: nil,
			}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id/files", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.ListProjectFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	fileCount, _ := resp["fileCount"].(float64)
	if fileCount != 0 {
		t.Errorf("Expected 0 files, got %v", fileCount)
	}
}

func TestScriptHandlers_ListProjectFiles_SortOrder(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:       id,
				Name:     "Test Script",
				Language: "rust",
				Dependencies: map[string]string{
					"zebra.rs":    "// zebra",
					"alpha.rs":    "// alpha",
					"middle.toml": "[package]",
				},
			}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id/files", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.ListProjectFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	files, ok := resp["files"].([]interface{})
	if !ok {
		t.Fatal("Expected files array in response")
	}
	if len(files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(files))
	}

	// Verify sort order: alpha.rs, middle.toml, zebra.rs
	expectedOrder := []string{"alpha.rs", "middle.toml", "zebra.rs"}
	for i, expected := range expectedOrder {
		fileMap, ok := files[i].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected file map at index %d", i)
		}
		filename, _ := fileMap["filename"].(string)
		if filename != expected {
			t.Errorf("Expected file[%d] = %s, got %s", i, expected, filename)
		}
	}
}

func TestScriptHandlers_ExportScriptAsZip_VerifyZipContents(t *testing.T) {
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:         id,
				Name:       "Export Test",
				Language:   "rust",
				SourceCode: "fn main() {}",
				Dependencies: map[string]string{
					"Cargo.toml": "[package]\nname = \"test\"",
				},
				Code: []byte("wasm binary data"),
			}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/test-id/export-zip", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	handlers.ExportScriptAsZip(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/zip" {
		t.Errorf("Expected Content-Type application/zip, got %s", contentType)
	}

	// Parse the ZIP response
	zipReader, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("Failed to read ZIP response: %v", err)
	}

	// Collect file names
	fileNames := make(map[string]bool)
	for _, f := range zipReader.File {
		fileNames[f.Name] = true
	}

	// Verify expected files exist
	expectedFiles := []string{"metadata.json", "src/lib.rs", "Cargo.toml", "output.wasm"}
	for _, expected := range expectedFiles {
		if !fileNames[expected] {
			t.Errorf("Expected file %s in ZIP, not found. Files: %v", expected, fileNames)
		}
	}
}

// === Security tests for new protections ===

func TestScriptHandlers_UploadProject_TooManyFiles(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{ID: scriptID}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	for i := 0; i < 1002; i++ {
		fw, _ := zipWriter.Create(fmt.Sprintf("file%d.rs", i))
		fw.Write([]byte("fn main() {}"))
	}
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "project.zip")
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
	if !strings.Contains(w.Body.String(), "too many files") {
		t.Errorf("Expected error about too many files, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_UploadProject_DuplicateFiles(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{ID: scriptID}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	fw1, _ := zipWriter.Create("main.rs")
	fw1.Write([]byte("fn main() { // version 1 }"))
	fw2, _ := zipWriter.Create("main.rs")
	fw2.Write([]byte("fn main() { // version 2 }"))
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "project.zip")
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
	if !strings.Contains(w.Body.String(), "duplicate file") {
		t.Errorf("Expected error about duplicate file, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_UploadProject_InvalidMagicBytes(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{ID: scriptID}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "project.zip")
	part.Write([]byte("this is not a zip file at all"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/upload-project", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UploadProject(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "bad signature") {
		t.Errorf("Expected error about bad signature, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_UploadProject_NullByteFilename(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{ID: scriptID}, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	fw, _ := zipWriter.Create("main\x00.rs")
	fw.Write([]byte("fn main() {}"))
	zipWriter.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "project.zip")
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
	if !strings.Contains(w.Body.String(), "null byte") {
		t.Errorf("Expected error about null byte, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_ImportScriptFromZip_TooManyFiles(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	metadataFile, _ := zipWriter.Create("metadata.json")
	metadata := map[string]interface{}{
		"name":     "Test",
		"language": "rust",
	}
	json.NewEncoder(metadataFile).Encode(metadata)
	for i := 0; i < 1001; i++ {
		fw, _ := zipWriter.Create(fmt.Sprintf("file%d.rs", i))
		fw.Write([]byte("fn main() {}"))
	}
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
	if !strings.Contains(w.Body.String(), "too many files") {
		t.Errorf("Expected error about too many files, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_UpdateScript_DependencySizeLimit(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:       scriptID,
		Name:     "Test Script",
		Language: "rust",
		Runtime:  domain.RuntimeExtism,
		Code:     []byte("wasm"),
	}
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
	}
	handlers := setupScriptHandlers(repo, nil)

	// Create a dependency that exceeds maxDependencyFileSize (500KB)
	largeDep := strings.Repeat("x", 501*1024)
	reqBody := map[string]interface{}{
		"dependencies": map[string]string{
			"large.rs": largeDep,
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/_api/v1/scripts/"+scriptID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.UpdateScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "too large") {
		t.Errorf("Expected error about dependency too large, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_UpdateScript_SourceCodeResetsStatus(t *testing.T) {
	scriptID := "test-id"
	existingScript := &domain.Script{
		ID:                scriptID,
		Name:              "Test Script",
		Runtime:           domain.RuntimeExtism,
		Code:              []byte("wasm"),
		Language:          "rust",
		TriggerType:       domain.TriggerRequest,
		SourceCode:        "old code",
		CompilationStatus: domain.CompilationSuccess,
		CompilationError:  "",
		ValidationStatus:  domain.ValidationValid,
		ValidationError:   "",
		Enabled:           true,
	}

	var savedScript *domain.Script
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return existingScript, nil
		},
		saveFunc: func(ctx context.Context, script *domain.Script) error {
			savedScript = script
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
	if savedScript == nil {
		t.Fatal("Expected script to be saved")
	}
	if savedScript.CompilationStatus != domain.CompilationNotCompiled {
		t.Errorf("Expected CompilationStatus to be reset to %s, got %s", domain.CompilationNotCompiled, savedScript.CompilationStatus)
	}
	if savedScript.ValidationStatus != domain.ValidationNotValidated {
		t.Errorf("Expected ValidationStatus to be reset to %s, got %s", domain.ValidationNotValidated, savedScript.ValidationStatus)
	}
	if len(savedScript.Code) != 0 {
		t.Error("Expected WASM code to be cleared")
	}
	if savedScript.Enabled {
		t.Error("Expected script to be disabled")
	}
}

func TestScriptHandlers_ToggleScript_NoWasmCode(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepository{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return &domain.Script{
				ID:         id,
				Code:       []byte{},
				SourceCode: "fn main() {}",
			}, nil
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

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "executable code") {
		t.Errorf("Expected error about executable code, got: %s", w.Body.String())
	}
}

func TestScriptHandlers_ImportScriptFromZip_InvalidMagicBytes(t *testing.T) {
	handlers := setupScriptHandlers(nil, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "script.zip")
	part.Write([]byte("not a zip file"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/import-zip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handlers.ImportScriptFromZip(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "bad signature") {
		t.Errorf("Expected error about bad signature, got: %s", w.Body.String())
	}
}
