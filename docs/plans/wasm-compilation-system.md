# WASM In-App Compilation System

## Overview

Система in-app компиляции для скриптов: пользователь пишет код в Monaco editor → нажимает "Compile" → код компилируется в WASM → скрипт готов к выполнению.

## Requirements

**User Answers**:
- ✅ Компиляция на **Backend** (Go вызывает rustc/tinygo/asc)
- ✅ Хранение **Source + WASM** (можно редактировать)
- ✅ Языки: **JavaScript/TypeScript, Go, Rust, Python** (все 4)
- ✅ Поддержка **полных зависимостей** (Cargo.toml, package.json, go.mod)

## Architecture

```
User writes code → Frontend Editor → API → Compilation Service → Compiler (Rust/Go/JS) → WASM Binary → Storage
```

---

## Фаза 1: Backend - Database & Domain (1-2 часа)

### 1.1 Database Migration

**File**: `migrations/0006_script_compilation.sql`

```sql
-- Add compilation-related fields
ALTER TABLE scripts ADD COLUMN source_code TEXT;
ALTER TABLE scripts ADD COLUMN compilation_status TEXT DEFAULT 'not_compiled'
    CHECK(compilation_status IN ('not_compiled', 'pending', 'compiling', 'success', 'error'));
ALTER TABLE scripts ADD COLUMN compilation_error TEXT;
ALTER TABLE scripts ADD COLUMN dependencies TEXT; -- JSON: {"filename": "content"}
ALTER TABLE scripts ADD COLUMN last_compiled_at TIMESTAMP;

-- Index for filtering by compilation status
CREATE INDEX idx_scripts_compilation_status ON scripts(compilation_status);
```

### 1.2 Domain Updates

**File**: `internal/features/scripting/domain/script.go`

Add fields:
```go
type Script struct {
    // ... existing fields

    // Compilation fields
    SourceCode         string            `json:"sourceCode,omitempty"`
    Dependencies       map[string]string `json:"dependencies,omitempty"` // filename → content
    CompilationStatus  CompilationStatus `json:"compilationStatus"`
    CompilationError   string            `json:"compilationError,omitempty"`
    LastCompiledAt     *time.Time        `json:"lastCompiledAt,omitempty"`
}

type CompilationStatus string

const (
    CompilationNotCompiled CompilationStatus = "not_compiled"
    CompilationPending     CompilationStatus = "pending"
    CompilationCompiling   CompilationStatus = "compiling"
    CompilationSuccess     CompilationStatus = "success"
    CompilationError       CompilationStatus = "error"
)
```

Add methods:
```go
func (s *Script) NeedsRecompilation() bool {
    return s.CompilationStatus != CompilationSuccess ||
           s.LastCompiledAt == nil ||
           s.Code == nil
}

func (s *Script) MarkCompiling() {
    s.CompilationStatus = CompilationCompiling
}

func (s *Script) MarkCompilationSuccess(wasm []byte) {
    s.Code = wasm
    s.CompilationStatus = CompilationSuccess
    now := time.Now()
    s.LastCompiledAt = &now
    s.CompilationError = ""
}

func (s *Script) MarkCompilationError(err error) {
    s.CompilationStatus = CompilationError
    s.CompilationError = err.Error()
}
```

---

## Фаза 2: Backend - Compilation Domain (4-6 часов)

### 2.1 Compiler Interface

**File**: `internal/features/scripting/domain/compiler.go`

```go
package domain

import (
    "context"
    "time"
)

// Compiler interface (PORT in Clean Architecture)
type Compiler interface {
    // Language returns the supported language (e.g., "rust", "go", "javascript")
    Language() string

    // Compile compiles source code to WASM binary
    Compile(ctx context.Context, req CompileRequest) (*CompileResult, error)

    // ValidateSyntax validates source code syntax without full compilation
    ValidateSyntax(ctx context.Context, req CompileRequest) error

    // ValidateDependencies validates dependency manifest (Cargo.toml, go.mod, etc.)
    ValidateDependencies(deps map[string]string) error

    // IsAvailable checks if compiler toolchain is installed
    IsAvailable() bool
}

// CompileRequest contains all data needed for compilation
type CompileRequest struct {
    SourceCode   string            // Main source code
    Dependencies map[string]string // filename → content (Cargo.toml, go.mod, package.json)
    Language     string            // rust, go, javascript, typescript, python
    Optimize     bool              // Enable optimizations
    Debug        bool              // Include debug symbols
}

// CompileResult contains compilation output
type CompileResult struct {
    WASMBinary []byte        // Compiled WASM binary
    Logs       []string      // Compilation logs
    Warnings   []string      // Compilation warnings
    Duration   time.Duration // Compilation time
    WASMSize   int64         // Binary size in bytes
}

// CompilationError represents compilation failure
type CompilationError struct {
    Message   string // Error message
    Line      int    // Line number (if available)
    Column    int    // Column number (if available)
    Code      string // Error code (e.g., "E0308")
    Severity  string // error, warning, note
    Traceback string // Full traceback
}

func (e *CompilationError) Error() string {
    if e.Line > 0 {
        return fmt.Sprintf("%s:%d:%d: %s", e.Severity, e.Line, e.Column, e.Message)
    }
    return fmt.Sprintf("%s: %s", e.Severity, e.Message)
}
```

### 2.2 Compilation Repository Interface

**File**: `internal/features/scripting/domain/compilation_repository.go`

```go
package domain

import "context"

// CompilationRepository manages compilation artifacts (optional - for caching)
type CompilationRepository interface {
    // SaveArtifact caches compiled WASM for faster recompilation
    SaveArtifact(ctx context.Context, scriptID string, artifact CompileResult) error

    // GetArtifact retrieves cached artifact
    GetArtifact(ctx context.Context, scriptID string) (*CompileResult, error)

    // InvalidateArtifact removes cached artifact
    InvalidateArtifact(ctx context.Context, scriptID string) error
}
```

---

## Фаза 3: Backend - Compilation Service (UseCase Layer)

### 3.1 Compilation Service

**File**: `internal/features/scripting/usecase/compilation_service.go`

```go
package usecase

import (
    "context"
    "fmt"
    "time"

    "go-proxy/internal/features/scripting/domain"
)

// CompilationService orchestrates script compilation (USE CASE)
type CompilationService struct {
    compilers map[string]domain.Compiler
    repo      domain.ScriptRepository
    // Optional: artifactRepo domain.CompilationRepository for caching
}

func NewCompilationService(repo domain.ScriptRepository) *CompilationService {
    return &CompilationService{
        compilers: make(map[string]domain.Compiler),
        repo:      repo,
    }
}

// RegisterCompiler registers a compiler implementation (ADAPTER)
func (s *CompilationService) RegisterCompiler(compiler domain.Compiler) {
    s.compilers[compiler.Language()] = compiler
}

// CompileScript compiles a script and updates it in repository
func (s *CompilationService) CompileScript(ctx context.Context, scriptID string, optimize bool) (*domain.CompileResult, error) {
    // Load script
    script, err := s.repo.Get(ctx, scriptID)
    if err != nil {
        return nil, fmt.Errorf("load script: %w", err)
    }

    // Validate
    if script.SourceCode == "" {
        return nil, fmt.Errorf("script has no source code")
    }

    // Get compiler
    compiler, ok := s.compilers[script.Language]
    if !ok {
        return nil, fmt.Errorf("unsupported language: %s", script.Language)
    }

    if !compiler.IsAvailable() {
        return nil, fmt.Errorf("compiler for %s is not available", script.Language)
    }

    // Mark as compiling
    script.MarkCompiling()
    if err := s.repo.Save(ctx, script); err != nil {
        return nil, fmt.Errorf("update status: %w", err)
    }

    // Compile
    req := domain.CompileRequest{
        SourceCode:   script.SourceCode,
        Dependencies: script.Dependencies,
        Language:     script.Language,
        Optimize:     optimize,
    }

    result, err := compiler.Compile(ctx, req)
    if err != nil {
        script.MarkCompilationError(err)
        s.repo.Save(ctx, script)
        return nil, fmt.Errorf("compilation failed: %w", err)
    }

    // Update script with compiled WASM
    script.MarkCompilationSuccess(result.WASMBinary)
    if err := s.repo.Save(ctx, script); err != nil {
        return nil, fmt.Errorf("save compiled script: %w", err)
    }

    return result, nil
}

// ValidateSyntax validates script syntax without compiling
func (s *CompilationService) ValidateSyntax(ctx context.Context, script *domain.Script) error {
    compiler, ok := s.compilers[script.Language]
    if !ok {
        return fmt.Errorf("unsupported language: %s", script.Language)
    }

    req := domain.CompileRequest{
        SourceCode:   script.SourceCode,
        Dependencies: script.Dependencies,
        Language:     script.Language,
    }

    return compiler.ValidateSyntax(ctx, req)
}

// GetAvailableCompilers returns list of available compilers
func (s *CompilationService) GetAvailableCompilers() []string {
    var available []string
    for lang, compiler := range s.compilers {
        if compiler.IsAvailable() {
            available = append(available, lang)
        }
    }
    return available
}
```

---

## Фаза 4: Backend - Compiler Implementations (Infrastructure Layer)

### 4.1 Common Workspace Manager

**File**: `internal/features/scripting/infrastructure/compilers/workspace.go`

```go
package compilers

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

// Workspace manages temporary compilation workspace
type Workspace struct {
    Path string
}

func NewWorkspace(scriptID string) (*Workspace, error) {
    tmpDir := filepath.Join(os.TempDir(), "go-proxy-compile", scriptID)
    if err := os.MkdirAll(tmpDir, 0755); err != nil {
        return nil, fmt.Errorf("create workspace: %w", err)
    }

    return &Workspace{Path: tmpDir}, nil
}

func (w *Workspace) WriteFile(filename string, content []byte) error {
    path := filepath.Join(w.Path, filename)
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    return os.WriteFile(path, content, 0644)
}

func (w *Workspace) ReadFile(filename string) ([]byte, error) {
    return os.ReadFile(filepath.Join(w.Path, filename))
}

func (w *Workspace) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, name, args...)
    cmd.Dir = w.Path
    return cmd.CombinedOutput()
}

func (w *Workspace) Cleanup() error {
    return os.RemoveAll(w.Path)
}
```

### 4.2 AssemblyScript Compiler (PRIORITY 1)

**File**: `internal/features/scripting/infrastructure/compilers/assemblyscript.go`

```go
package compilers

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "go-proxy/internal/features/scripting/domain"
)

type AssemblyScriptCompiler struct {
    ascPath string // Path to asc binary
}

func NewAssemblyScriptCompiler() *AssemblyScriptCompiler {
    return &AssemblyScriptCompiler{
        ascPath: "asc", // Assume in PATH
    }
}

func (c *AssemblyScriptCompiler) Language() string {
    return "assemblyscript"
}

func (c *AssemblyScriptCompiler) IsAvailable() bool {
    cmd := exec.Command(c.ascPath, "--version")
    return cmd.Run() == nil
}

func (c *AssemblyScriptCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
    start := time.Now()

    // Create workspace
    ws, err := NewWorkspace(fmt.Sprintf("as-%d", time.Now().UnixNano()))
    if err != nil {
        return nil, err
    }
    defer ws.Cleanup()

    // Write source code
    if err := ws.WriteFile("index.ts", []byte(req.SourceCode)); err != nil {
        return nil, fmt.Errorf("write source: %w", err)
    }

    // Write package.json if exists
    if packageJSON, ok := req.Dependencies["package.json"]; ok {
        if err := ws.WriteFile("package.json", []byte(packageJSON)); err != nil {
            return nil, fmt.Errorf("write package.json: %w", err)
        }

        // Install dependencies
        output, err := ws.ExecuteCommand(ctx, "npm", "install")
        if err != nil {
            return nil, fmt.Errorf("npm install failed: %s", output)
        }
    }

    // Compile
    args := []string{"index.ts", "-o", "output.wasm"}
    if req.Optimize {
        args = append(args, "--optimize")
    }

    output, err := ws.ExecuteCommand(ctx, c.ascPath, args...)
    if err != nil {
        return nil, c.parseCompilationError(string(output))
    }

    // Read compiled WASM
    wasm, err := ws.ReadFile("output.wasm")
    if err != nil {
        return nil, fmt.Errorf("read wasm: %w", err)
    }

    return &domain.CompileResult{
        WASMBinary: wasm,
        Logs:       strings.Split(string(output), "\n"),
        Duration:   time.Since(start),
        WASMSize:   int64(len(wasm)),
    }, nil
}

func (c *AssemblyScriptCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
    // AssemblyScript doesn't have separate syntax check, run full compile
    _, err := c.Compile(ctx, req)
    return err
}

func (c *AssemblyScriptCompiler) ValidateDependencies(deps map[string]string) error {
    packageJSON, ok := deps["package.json"]
    if !ok {
        return nil // No dependencies
    }

    var pkg map[string]interface{}
    if err := json.Unmarshal([]byte(packageJSON), &pkg); err != nil {
        return fmt.Errorf("invalid package.json: %w", err)
    }

    return nil
}

func (c *AssemblyScriptCompiler) parseCompilationError(output string) error {
    // TODO: Parse AssemblyScript error format
    return fmt.Errorf("compilation failed: %s", output)
}
```

### 4.3 TinyGo Compiler (PRIORITY 2)

**File**: `internal/features/scripting/infrastructure/compilers/tinygo.go`

```go
package compilers

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "go-proxy/internal/features/scripting/domain"
)

type TinyGoCompiler struct {
    tinygoPath string
}

func NewTinyGoCompiler() *TinyGoCompiler {
    return &TinyGoCompiler{
        tinygoPath: "tinygo",
    }
}

func (c *TinyGoCompiler) Language() string {
    return "go"
}

func (c *TinyGoCompiler) IsAvailable() bool {
    cmd := exec.Command(c.tinygoPath, "version")
    return cmd.Run() == nil
}

func (c *TinyGoCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
    start := time.Now()

    ws, err := NewWorkspace(fmt.Sprintf("go-%d", time.Now().UnixNano()))
    if err != nil {
        return nil, err
    }
    defer ws.Cleanup()

    // Write main.go
    if err := ws.WriteFile("main.go", []byte(req.SourceCode)); err != nil {
        return nil, err
    }

    // Write go.mod if exists
    if goMod, ok := req.Dependencies["go.mod"]; ok {
        if err := ws.WriteFile("go.mod", []byte(goMod)); err != nil {
            return nil, err
        }

        // Download dependencies
        output, err := ws.ExecuteCommand(ctx, "go", "mod", "download")
        if err != nil {
            return nil, fmt.Errorf("go mod download failed: %s", output)
        }
    }

    // Compile
    args := []string{"build", "-target=wasi", "-o", "output.wasm"}
    if req.Optimize {
        args = append(args, "-opt=2")
    }
    args = append(args, "main.go")

    output, err := ws.ExecuteCommand(ctx, c.tinygoPath, args...)
    if err != nil {
        return nil, fmt.Errorf("compilation failed: %s", output)
    }

    wasm, err := ws.ReadFile("output.wasm")
    if err != nil {
        return nil, err
    }

    return &domain.CompileResult{
        WASMBinary: wasm,
        Logs:       strings.Split(string(output), "\n"),
        Duration:   time.Since(start),
        WASMSize:   int64(len(wasm)),
    }, nil
}

func (c *TinyGoCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
    _, err := c.Compile(ctx, req)
    return err
}

func (c *TinyGoCompiler) ValidateDependencies(deps map[string]string) error {
    // TODO: Parse go.mod
    return nil
}
```

### 4.4 Rust Compiler (PRIORITY 3)

**File**: `internal/features/scripting/infrastructure/compilers/rust.go`

```go
package compilers

import (
    "context"
    "fmt"
    "strings"
    "time"

    "go-proxy/internal/features/scripting/domain"
)

type RustCompiler struct {
    cargoPath string
}

func NewRustCompiler() *RustCompiler {
    return &RustCompiler{
        cargoPath: "cargo",
    }
}

func (c *RustCompiler) Language() string {
    return "rust"
}

func (c *RustCompiler) IsAvailable() bool {
    cmd := exec.Command(c.cargoPath, "--version")
    return cmd.Run() == nil
}

func (c *RustCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
    start := time.Now()

    ws, err := NewWorkspace(fmt.Sprintf("rust-%d", time.Now().UnixNano()))
    if err != nil {
        return nil, err
    }
    defer ws.Cleanup()

    // Write Cargo.toml
    cargoToml, ok := req.Dependencies["Cargo.toml"]
    if !ok {
        // Generate minimal Cargo.toml
        cargoToml = c.generateMinimalCargoToml()
    }
    if err := ws.WriteFile("Cargo.toml", []byte(cargoToml)); err != nil {
        return nil, err
    }

    // Write src/lib.rs
    if err := ws.WriteFile("src/lib.rs", []byte(req.SourceCode)); err != nil {
        return nil, err
    }

    // Compile
    args := []string{"build", "--target", "wasm32-unknown-unknown"}
    if req.Optimize {
        args = append(args, "--release")
    }

    output, err := ws.ExecuteCommand(ctx, c.cargoPath, args...)
    if err != nil {
        return nil, c.parseRustError(string(output))
    }

    // Read WASM
    wasmPath := "target/wasm32-unknown-unknown/debug/plugin.wasm"
    if req.Optimize {
        wasmPath = "target/wasm32-unknown-unknown/release/plugin.wasm"
    }

    wasm, err := ws.ReadFile(wasmPath)
    if err != nil {
        return nil, err
    }

    return &domain.CompileResult{
        WASMBinary: wasm,
        Logs:       strings.Split(string(output), "\n"),
        Duration:   time.Since(start),
        WASMSize:   int64(len(wasm)),
    }, nil
}

func (c *RustCompiler) generateMinimalCargoToml() string {
    return `[package]
name = "plugin"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
extism-pdk = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
`
}

func (c *RustCompiler) parseRustError(output string) error {
    // TODO: Parse Rust error format (error[E0308]: ...)
    return fmt.Errorf("rust compilation failed: %s", output)
}

func (c *RustCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
    // Use cargo check
    ws, err := NewWorkspace(fmt.Sprintf("rust-check-%d", time.Now().UnixNano()))
    if err != nil {
        return err
    }
    defer ws.Cleanup()

    // ... similar setup
    output, err := ws.ExecuteCommand(ctx, c.cargoPath, "check")
    if err != nil {
        return fmt.Errorf("syntax check failed: %s", output)
    }
    return nil
}

func (c *RustCompiler) ValidateDependencies(deps map[string]string) error {
    // TODO: Parse Cargo.toml
    return nil
}
```

---

## Фаза 5: Backend - HTTP API

### 5.1 Compilation Handlers

**File**: `internal/infrastructure/httpapi/compilation_handlers.go`

```go
package httpapi

import (
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"

    "go-proxy/internal/features/scripting/usecase"
)

type CompilationHandlers struct {
    service *usecase.CompilationService
}

func NewCompilationHandlers(service *usecase.CompilationService) *CompilationHandlers {
    return &CompilationHandlers{service: service}
}

// POST /_api/v1/scripts/:id/compile
func (h *CompilationHandlers) CompileScript(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    scriptID := vars["id"]

    var req struct {
        Optimize bool `json:"optimize"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    result, err := h.service.CompileScript(r.Context(), scriptID, req.Optimize)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":          "success",
        "wasmSize":        result.WASMSize,
        "compilationTime": result.Duration.String(),
        "logs":            result.Logs,
    })
}

// POST /_api/v1/scripts/validate
func (h *CompilationHandlers) ValidateSyntax(w http.ResponseWriter, r *http.Request) {
    var script domain.Script
    if err := json.NewDecoder(r.Body).Decode(&script); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := h.service.ValidateSyntax(r.Context(), &script); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "valid": false,
            "error": err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "valid": true,
    })
}

// GET /_api/v1/scripts/compilers
func (h *CompilationHandlers) ListCompilers(w http.ResponseWriter, r *http.Request) {
    compilers := h.service.GetAvailableCompilers()
    json.NewEncoder(w).Encode(map[string]interface{}{
        "compilers": compilers,
    })
}
```

### 5.2 Router Registration

**File**: `internal/infrastructure/httpapi/router.go` (update)

```go
// In setupRoutes() function, add:
if d.ScriptSvc != nil && d.CompilationSvc != nil {
    compilationHandlers := NewCompilationHandlers(d.CompilationSvc)
    router.HandleFunc("POST /_api/v1/scripts/{id}/compile", compilationHandlers.CompileScript)
    router.HandleFunc("POST /_api/v1/scripts/validate", compilationHandlers.ValidateSyntax)
    router.HandleFunc("GET /_api/v1/scripts/compilers", compilationHandlers.ListCompilers)
}
```

---

## Фаза 6: Frontend

### 6.1 Domain Updates

**File**: `frontend/lib/features/scripts/domain/entities/script.dart`

```dart
@freezed
sealed class Script with _$Script {
  const Script._();

  const factory Script({
    required String id,
    required String name,
    String? description,
    required ScriptRuntime runtime,
    required String code,  // WASM binary (base64) or empty if only source
    String? sourceCode,    // NEW: Source code
    @Default({}) Map<String, String> dependencies,  // NEW: filename → content
    @Default(CompilationStatus.notCompiled) CompilationStatus compilationStatus,  // NEW
    String? compilationError,  // NEW
    DateTime? lastCompiledAt,  // NEW
    required String language,
    required TriggerType triggerType,
    @Default(10) int priority,
    @Default(true) bool enabled,
    MatchRules? matchRules,
    ScriptConfig? config,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _Script;

  factory Script.fromJson(Map<String, dynamic> json) => _$ScriptFromJson(json);
}

enum CompilationStatus {
  @JsonValue('not_compiled') notCompiled,
  @JsonValue('pending') pending,
  @JsonValue('compiling') compiling,
  @JsonValue('success') success,
  @JsonValue('error') error,
}

extension ScriptCompilationX on Script {
  bool get needsCompilation => compilationStatus != CompilationStatus.success;
  bool get isCompiling => compilationStatus == CompilationStatus.compiling;
  bool get hasCompilationError => compilationStatus == CompilationStatus.error;

  String get compilationStatusDisplay {
    switch (compilationStatus) {
      case CompilationStatus.notCompiled:
        return 'Not Compiled';
      case CompilationStatus.pending:
        return 'Pending';
      case CompilationStatus.compiling:
        return 'Compiling...';
      case CompilationStatus.success:
        return 'Compiled';
      case CompilationStatus.error:
        return 'Error';
    }
  }

  IconData get compilationStatusIcon {
    switch (compilationStatus) {
      case CompilationStatus.notCompiled:
        return Icons.circle_outlined;
      case CompilationStatus.pending:
      case CompilationStatus.compiling:
        return Icons.hourglass_empty;
      case CompilationStatus.success:
        return Icons.check_circle;
      case CompilationStatus.error:
        return Icons.error;
    }
  }

  Color get compilationStatusColor {
    switch (compilationStatus) {
      case CompilationStatus.notCompiled:
        return Colors.grey;
      case CompilationStatus.pending:
      case CompilationStatus.compiling:
        return Colors.orange;
      case CompilationStatus.success:
        return Colors.green;
      case CompilationStatus.error:
        return Colors.red;
    }
  }
}
```

### 6.2 API Service

**File**: `frontend/lib/features/scripts/data/services/scripts_api_service.dart` (update)

```dart
class ScriptsApiService {
  // ... existing methods

  Future<CompilationResult> compileScript(String id, {bool optimize = true}) async {
    final response = await _dio.post(
      '$_baseUrl/scripts/$id/compile',
      data: {'optimize': optimize},
    );
    return CompilationResult.fromJson(response.data);
  }

  Future<ValidationResult> validateSyntax(Script script) async {
    final response = await _dio.post(
      '$_baseUrl/scripts/validate',
      data: script.toJson(),
    );
    return ValidationResult.fromJson(response.data);
  }

  Future<List<String>> getAvailableCompilers() async {
    final response = await _dio.get('$_baseUrl/scripts/compilers');
    return List<String>.from(response.data['compilers']);
  }
}

@freezed
sealed class CompilationResult with _$CompilationResult {
  const factory CompilationResult({
    required String status,
    required int wasmSize,
    required String compilationTime,
    required List<String> logs,
  }) = _CompilationResult;

  factory CompilationResult.fromJson(Map<String, dynamic> json) =>
      _$CompilationResultFromJson(json);
}
```

### 6.3 Compilation Store

**File**: `frontend/lib/features/scripts/application/stores/compilation_store.dart`

```dart
class CompilationStore = _CompilationStore with _$CompilationStore;

abstract class _CompilationStore with Store {
  final ScriptsRepository _repository;

  _CompilationStore(this._repository);

  @observable
  ObservableMap<String, CompilationProgress> compilations = ObservableMap();

  @observable
  ObservableList<String> logs = ObservableList();

  @action
  Future<void> compileScript(String scriptId, {bool optimize = true}) async {
    compilations[scriptId] = CompilationProgress(status: 'compiling');
    logs.clear();

    try {
      final result = await _repository.compileScript(scriptId, optimize: optimize);

      compilations[scriptId] = CompilationProgress(
        status: 'success',
        result: result,
      );

      logs.addAll(result.logs);
    } catch (e) {
      compilations[scriptId] = CompilationProgress(
        status: 'error',
        error: e.toString(),
      );
    }
  }

  @action
  Future<bool> validateSyntax(Script script) async {
    try {
      final result = await _repository.validateSyntax(script);
      return result.valid;
    } catch (e) {
      return false;
    }
  }
}

class CompilationProgress {
  final String status;
  final CompilationResult? result;
  final String? error;

  CompilationProgress({
    required this.status,
    this.result,
    this.error,
  });
}
```

### 6.4 UI: Compile Button

**File**: `frontend/lib/features/scripts/presentation/widgets/script_editor_dialog.dart` (update Code Tab)

```dart
// Add to Code Tab toolbar
Row(
  children: [
    // Compile button
    ElevatedButton.icon(
      icon: Observer(
        builder: (_) => compilationStore.isCompiling(script.id)
            ? SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : Icon(Icons.build),
      ),
      label: Text('Compile'),
      onPressed: () async {
        await compilationStore.compileScript(script.id, optimize: true);
        // Reload script to get updated WASM
        await scriptsStore.loadScripts();
      },
    ),

    SizedBox(width: 8),

    // Validate button
    TextButton.icon(
      icon: Icon(Icons.check),
      label: Text('Validate'),
      onPressed: () async {
        final valid = await compilationStore.validateSyntax(script);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(valid ? 'Syntax is valid' : 'Syntax errors found'),
            backgroundColor: valid ? Colors.green : Colors.red,
          ),
        );
      },
    ),

    Spacer(),

    // Compilation status
    Observer(
      builder: (_) => Chip(
        avatar: Icon(
          script.compilationStatusIcon,
          size: 16,
          color: script.compilationStatusColor,
        ),
        label: Text(script.compilationStatusDisplay),
      ),
    ),
  ],
)
```

### 6.5 UI: Dependencies Tab

**File**: `frontend/lib/features/scripts/presentation/widgets/dependencies_tab.dart`

```dart
class DependenciesTab extends StatelessWidget {
  final Script script;
  final Function(Map<String, String>) onDependenciesChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Language-specific dependency editor
        if (script.language == 'rust') _buildCargoTomlEditor(),
        if (script.language == 'go') _buildGoModEditor(),
        if (script.language == 'javascript' || script.language == 'typescript')
          _buildPackageJsonEditor(),
      ],
    );
  }

  Widget _buildCargoTomlEditor() {
    return MonacoCodeEditor(
      language: 'toml',
      value: script.dependencies['Cargo.toml'] ?? _defaultCargoToml,
      onChanged: (value) {
        final deps = Map<String, String>.from(script.dependencies);
        deps['Cargo.toml'] = value;
        onDependenciesChanged(deps);
      },
    );
  }

  String get _defaultCargoToml => '''
[package]
name = "plugin"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
extism-pdk = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
''';
}
```

### 6.6 UI: Compilation Logs Panel

**File**: `frontend/lib/features/scripts/presentation/widgets/compilation_logs_panel.dart`

```dart
class CompilationLogsPanel extends StatelessWidget {
  final List<String> logs;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 200,
      decoration: BoxDecoration(
        color: Colors.black87,
        border: Border.all(color: Colors.grey),
      ),
      child: ListView.builder(
        itemCount: logs.length,
        itemBuilder: (context, index) {
          final log = logs[index];
          return Padding(
            padding: EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            child: Text(
              log,
              style: TextStyle(
                fontFamily: 'monospace',
                fontSize: 12,
                color: _getLogColor(log),
              ),
            ),
          );
        },
      ),
    );
  }

  Color _getLogColor(String log) {
    if (log.contains('error')) return Colors.red;
    if (log.contains('warning')) return Colors.orange;
    if (log.contains('success') || log.contains('Finished')) return Colors.green;
    return Colors.grey[300]!;
  }
}
```

---

## Фаза 7: Testing

### 7.1 E2E Tests

**File**: `internal/e2e/compilation_test.go`

```go
package e2e

import (
    "testing"
    "time"
)

func TestE2E_Compilation_AssemblyScript(t *testing.T) {
    // Start server
    cleanup := startTestServer(t)
    defer cleanup()

    // Create script with source code
    sourceCode := `
export function process(input: ArrayBuffer): ArrayBuffer {
  // Simple passthrough
  return input;
}
`

    scriptID := createScriptWithSource(t, baseURL, sourceCode, "assemblyscript")

    // Compile
    result := compileScript(t, baseURL, scriptID, true)
    assert.Equal(t, "success", result.Status)
    assert.Greater(t, result.WASMSize, 0)

    // Verify script updated
    script := getScript(t, baseURL, scriptID)
    assert.Equal(t, "success", script.CompilationStatus)
    assert.NotEmpty(t, script.Code) // WASM binary
}

func TestE2E_Compilation_TinyGo(t *testing.T) {
    sourceCode := `
package main

import "github.com/extism/go-pdk"

//export process
func process() int32 {
    input := pdk.Input()
    pdk.OutputMemory(input)
    return 0
}

func main() {}
`

    scriptID := createScriptWithSource(t, baseURL, sourceCode, "go")
    result := compileScript(t, baseURL, scriptID, true)
    assert.Equal(t, "success", result.Status)
}

func TestE2E_Compilation_Rust_WithDependencies(t *testing.T) {
    sourceCode := loadTestSource(t, "rust/add_header.rs")
    cargoToml := loadTestSource(t, "rust/Cargo.toml")

    dependencies := map[string]string{
        "Cargo.toml": cargoToml,
    }

    scriptID := createScriptWithSourceAndDeps(t, baseURL, sourceCode, "rust", dependencies)

    result := compileScript(t, baseURL, scriptID, true)
    assert.Equal(t, "success", result.Status)

    // Test execution
    resp := makeProxiedRequest(t, "GET", "/test")
    assert.Equal(t, "Rust", resp.Header.Get("X-Script-Processed"))
}

func TestE2E_Compilation_SyntaxError(t *testing.T) {
    sourceCode := `
export function process(input: ArrayBuffer {  // Missing closing paren
  return input;
}
`

    scriptID := createScriptWithSource(t, baseURL, sourceCode, "assemblyscript")

    _, err := compileScript(t, baseURL, scriptID, false)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "compilation failed")
}
```

---

## Implementation Sequence (RECOMMENDED)

### Sprint 1: MVP - AssemblyScript Only (1-2 days)
1. ✅ DB migration (`0006_script_compilation.sql`)
2. ✅ Domain updates (`Script` entity + `Compiler` interface)
3. ✅ `CompilationService` (use case)
4. ✅ `AssemblyScriptCompiler` implementation
5. ✅ API endpoints (`/compile`, `/validate`)
6. ✅ Frontend: Domain entities + API service
7. ✅ Frontend: Compile button + status indicator
8. ✅ E2E test for AssemblyScript

### Sprint 2: TinyGo Support (1 day)
1. ✅ `TinyGoCompiler` implementation
2. ✅ Frontend: Dependencies tab (go.mod editor)
3. ✅ E2E tests

### Sprint 3: Rust Support (2 days)
1. ✅ `RustCompiler` implementation
2. ✅ Cargo.toml parser/generator
3. ✅ Frontend: Cargo.toml editor
4. ✅ E2E tests with dependencies

### Sprint 4: Polish (1 day)
1. ✅ Compilation logs panel (UI)
2. ✅ Real-time compilation status
3. ✅ Error highlighting in Monaco
4. ✅ Auto-compile on save (settings)
5. ✅ Compiler availability detection

---

## Technical Considerations

### Security
1. **Sandbox compilation**: Run compilers in restricted environment
2. **Resource limits**: Timeout (5 min), disk space (1GB temp)
3. **Dependency whitelist**: Optional whitelist for packages
4. **Code review**: Audit before enabling

### Performance
1. **Cargo cache**: Share `~/.cargo` between compilations
2. **npm cache**: Share `node_modules`
3. **Parallel builds**: Queue system for multiple compilations
4. **Background jobs**: Don't block API requests

### Error Handling
1. **Compiler not found**: Show install instructions
2. **Syntax errors**: Parse and highlight in editor
3. **Timeout**: Suggest code simplification
4. **Dependencies**: Show which package failed

---

## Estimated Time
- Backend: 10-12 hours
- Frontend: 5-6 hours
- Testing: 3-4 hours
- **TOTAL: 18-22 hours (2-3 days)**

---

## Success Criteria
- ✅ User writes AssemblyScript in Monaco editor
- ✅ Clicks "Compile" → WASM generated
- ✅ Script executes successfully
- ✅ Compilation logs visible
- ✅ Errors displayed with line numbers
- ✅ TinyGo and Rust work similarly
- ✅ Dependencies (Cargo.toml, package.json) supported
