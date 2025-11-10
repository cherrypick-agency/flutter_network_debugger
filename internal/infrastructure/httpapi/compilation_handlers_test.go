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

// Composer 1.
func TestCompilationHandlers_CompileScript_Success(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		Name:       "Test Script",
		SourceCode: "fn main() {}",
		Language:   "rust",
		Runtime:    domain.RuntimeExtism,
	}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
		saveFunc: func(ctx context.Context, s *domain.Script) error {
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
	handlers := NewCompilationHandlers(compilationService)

	reqBody := map[string]interface{}{
		"optimize": true,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	// Act
	handlers.CompileScript(w, req)

	// Assert
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

func TestCompilationHandlers_CompileScript_NotFound(t *testing.T) {
	// Arrange
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return nil, errors.New("not found")
		},
	}
	compilationService := usecase.NewCompilationService(repo)
	handlers := NewCompilationHandlers(compilationService)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/test-id/compile", nil)
	req.SetPathValue("id", "test-id")
	w := httptest.NewRecorder()

	// Act
	handlers.CompileScript(w, req)

	// Assert
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

func TestCompilationHandlers_CompileScript_CompilationError(t *testing.T) {
	// Arrange
	scriptID := "test-id"
	script := &domain.Script{
		ID:         scriptID,
		SourceCode: "invalid code",
		Language:   "rust",
	}
	repo := &mockScriptRepo{
		getFunc: func(ctx context.Context, id string) (*domain.Script, error) {
			return script, nil
		},
		saveFunc: func(ctx context.Context, s *domain.Script) error {
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
	handlers := NewCompilationHandlers(compilationService)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/"+scriptID+"/compile", nil)
	req.SetPathValue("id", scriptID)
	w := httptest.NewRecorder()

	// Act
	handlers.CompileScript(w, req)

	// Assert
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

func TestCompilationHandlers_ValidateSyntax_Success(t *testing.T) {
	// Arrange
	compiler := &mockCompiler{
		language:    "rust",
		isAvailable: true,
		validateFunc: func(ctx context.Context, req domain.CompileRequest) error {
			return nil
		},
	}
	repo := &mockScriptRepo{}
	compilationService := usecase.NewCompilationService(repo)
	compilationService.RegisterCompiler(compiler)
	handlers := NewCompilationHandlers(compilationService)

	reqBody := map[string]interface{}{
		"sourceCode": "fn main() {}",
		"language":   "rust",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.ValidateSyntax(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response ValidateSyntaxResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Valid {
		t.Error("Expected valid to be true")
	}
}

func TestCompilationHandlers_ValidateSyntax_Invalid(t *testing.T) {
	// Arrange
	compiler := &mockCompiler{
		language:    "rust",
		isAvailable: true,
		validateFunc: func(ctx context.Context, req domain.CompileRequest) error {
			return errors.New("syntax error")
		},
	}
	repo := &mockScriptRepo{}
	compilationService := usecase.NewCompilationService(repo)
	compilationService.RegisterCompiler(compiler)
	handlers := NewCompilationHandlers(compilationService)

	reqBody := map[string]interface{}{
		"sourceCode": "invalid code",
		"language":   "rust",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/scripts/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handlers.ValidateSyntax(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ValidateSyntaxResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Valid {
		t.Error("Expected valid to be false")
	}
	if response.Error == "" {
		t.Error("Expected error message")
	}
}

func TestCompilationHandlers_ListCompilers_Success(t *testing.T) {
	// Arrange
	compiler1 := &mockCompiler{language: "rust", isAvailable: true}
	compiler2 := &mockCompiler{language: "go", isAvailable: false}
	repo := &mockScriptRepo{}
	compilationService := usecase.NewCompilationService(repo)
	compilationService.RegisterCompiler(compiler1)
	compilationService.RegisterCompiler(compiler2)
	handlers := NewCompilationHandlers(compilationService)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/scripts/compilers", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.ListCompilers(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompilersResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Compilers) == 0 {
		t.Error("Expected at least one available compiler")
	}
	if response.All["rust"] != true {
		t.Error("Expected rust to be available")
	}
	if response.All["go"] != false {
		t.Error("Expected go to be unavailable")
	}
}
