package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

type mockScriptRepositoryForCompilation struct {
	getFunc  func(ctx context.Context, id string) (*domain.Script, error)
	saveFunc func(ctx context.Context, script *domain.Script) error
}

func (m *mockScriptRepositoryForCompilation) Save(ctx context.Context, script *domain.Script) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, script)
	}
	return nil
}

func (m *mockScriptRepositoryForCompilation) Get(ctx context.Context, id string) (*domain.Script, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockScriptRepositoryForCompilation) List(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
	return nil, nil
}

func (m *mockScriptRepositoryForCompilation) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockScriptRepositoryForCompilation) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	return nil
}

type mockCompilerForCompilation struct {
	language     string
	isAvailable  bool
	compileFunc  func(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error)
	validateFunc func(ctx context.Context, req domain.CompileRequest) error
}

func (m *mockCompilerForCompilation) Language() string {
	return m.language
}

func (m *mockCompilerForCompilation) IsAvailable() bool {
	return m.isAvailable
}

func (m *mockCompilerForCompilation) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
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

func (m *mockCompilerForCompilation) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, req)
	}
	return nil
}

func (m *mockCompilerForCompilation) ValidateDependencies(deps map[string]string) error {
	return nil
}

func setupCompilationHandlers(repo *mockScriptRepositoryForCompilation) *CompilationHandlers {
	service := usecase.NewCompilationService(repo, nil)
	handlers := NewCompilationHandlers(service)
	return handlers
}

func TestCompilationHandlers_CompileScript_Success(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
	}
	repo := &mockScriptRepositoryForCompilation{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			if id != scriptID {
				t.Errorf("Expected id '%s', got '%s'", scriptID, id)
			}
			return script, nil
		},
		saveFunc: func(ctx context.Context, s *domain.Script) error {
			return nil
		},
	}
	compiler := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
	}
	handlers := setupCompilationHandlers(repo)
	handlers.service.RegisterCompiler(compiler)

	reqBody := map[string]interface{}{
		"optimize": true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.CompileScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompileScriptResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}
	if response.WASMSize != 100 {
		t.Errorf("Expected WASMSize 100, got %d", response.WASMSize)
	}
}

func TestCompilationHandlers_CompileScript_MissingID(t *testing.T) {
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts//compile", nil)
	w := httptest.NewRecorder()

	handlers.CompileScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilationHandlers_CompileScript_ScriptNotFound(t *testing.T) {
	scriptID := "test-id"
	repo := &mockScriptRepositoryForCompilation{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	handlers := setupCompilationHandlers(repo)

	reqBody := map[string]interface{}{
		"optimize": true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/compile", bytes.NewReader(body))
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.CompileScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response CompileScriptResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}
}

func TestCompilationHandlers_CompileScript_CompilationFailure(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Language:   "rust",
		SourceCode: "invalid code",
	}
	repo := &mockScriptRepositoryForCompilation{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
		saveFunc: func(ctx context.Context, s *domain.Script) error {
			return nil
		},
	}
	compiler := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
		compileFunc: func(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
			return nil, errors.New("compilation failed")
		},
	}
	handlers := setupCompilationHandlers(repo)
	handlers.service.RegisterCompiler(compiler)

	reqBody := map[string]interface{}{
		"optimize": false,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/compile", bytes.NewReader(body))
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.CompileScript(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response CompileScriptResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", response.Status)
	}
}

func TestCompilationHandlers_CompileScript_DefaultOptimize(t *testing.T) {
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Language:   "rust",
		SourceCode: "fn main() {}",
	}
	repo := &mockScriptRepositoryForCompilation{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
		saveFunc: func(ctx context.Context, s *domain.Script) error {
			return nil
		},
	}
	compiler := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
	}
	handlers := setupCompilationHandlers(repo)
	handlers.service.RegisterCompiler(compiler)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/compile", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	handlers.CompileScript(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompilationHandlers_ValidateSyntax_Success(t *testing.T) {
	compiler := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
		validateFunc: func(ctx context.Context, req domain.CompileRequest) error {
			return nil
		},
	}
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})
	handlers.service.RegisterCompiler(compiler)

	reqBody := map[string]interface{}{
		"sourceCode": "fn main() {}",
		"language":   "rust",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.ValidateSyntax(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response ValidateSyntaxResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Valid {
		t.Error("Expected Valid to be true")
	}
}

func TestCompilationHandlers_ValidateSyntax_InvalidSyntax(t *testing.T) {
	compiler := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
		validateFunc: func(ctx context.Context, req domain.CompileRequest) error {
			return errors.New("syntax error")
		},
	}
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})
	handlers.service.RegisterCompiler(compiler)

	reqBody := map[string]interface{}{
		"sourceCode": "invalid rust code",
		"language":   "rust",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.ValidateSyntax(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ValidateSyntaxResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Valid {
		t.Error("Expected Valid to be false")
	}
	if response.Error == "" {
		t.Error("Expected Error to be set")
	}
}

func TestCompilationHandlers_ValidateSyntax_InvalidBody(t *testing.T) {
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.ValidateSyntax(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilationHandlers_ValidateSyntax_UnsupportedLanguage(t *testing.T) {
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})

	reqBody := map[string]interface{}{
		"sourceCode": "console.log('test');",
		"language":   "javascript",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.ValidateSyntax(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilationHandlers_ValidateSyntax_WithDependencies(t *testing.T) {
	compiler := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
		validateFunc: func(ctx context.Context, req domain.CompileRequest) error {
			if len(req.Dependencies) == 0 {
				t.Error("Expected dependencies to be passed")
			}
			return nil
		},
	}
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})
	handlers.service.RegisterCompiler(compiler)

	reqBody := map[string]interface{}{
		"sourceCode": "fn main() {}",
		"language":   "rust",
		"dependencies": map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.ValidateSyntax(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompilationHandlers_ListCompilers_Success(t *testing.T) {
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})

	compiler1 := &mockCompilerForCompilation{
		language:    "rust",
		isAvailable: true,
	}
	compiler2 := &mockCompilerForCompilation{
		language:    "go",
		isAvailable: false,
	}
	compiler3 := &mockCompilerForCompilation{
		language:    "zig",
		isAvailable: true,
	}

	handlers.service.RegisterCompiler(compiler1)
	handlers.service.RegisterCompiler(compiler2)
	handlers.service.RegisterCompiler(compiler3)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/compilers", nil)
	w := httptest.NewRecorder()

	handlers.ListCompilers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompilersResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Compilers) != 2 {
		t.Errorf("Expected 2 available compilers, got %d", len(response.Compilers))
	}

	if len(response.All) != 3 {
		t.Errorf("Expected 3 total compilers, got %d", len(response.All))
	}

	if !response.All["rust"] {
		t.Error("Expected rust to be available")
	}
	if response.All["go"] {
		t.Error("Expected go to be unavailable")
	}
	if !response.All["zig"] {
		t.Error("Expected zig to be available")
	}
}

func TestCompilationHandlers_ListCompilers_Empty(t *testing.T) {
	handlers := setupCompilationHandlers(&mockScriptRepositoryForCompilation{})

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/compilers", nil)
	w := httptest.NewRecorder()

	handlers.ListCompilers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompilersResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Compilers) != 0 {
		t.Errorf("Expected 0 compilers, got %d", len(response.Compilers))
	}
}
