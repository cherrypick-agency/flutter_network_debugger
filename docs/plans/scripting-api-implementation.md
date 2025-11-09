# Scripting API Implementation Plan

**Дата**: 2025-11-03
**Статус**: Утверждено
**Оценка времени**: 8 дней разработки
**Сложность**: Средняя

## Executive Summary

Реализация мультиязычной системы скриптинга для Network Debugger, позволяющей пользователям создавать плагины для модификации HTTP request/response в реальном времени.

**Архитектурное решение**: Гибридный подход (Extism + Dart Subprocess)
- **Extism (WASM)**: Для Rust, Go, JavaScript, Python плагинов
- **Dart Subprocess (JSON-RPC)**: Для Dart-специфичных скриптов

**Ключевые преимущества**:
- ✅ Battle-tested компоненты (Extism используется в Zed Editor, JSON-RPC - стандарт LSP)
- ✅ Простая реализация (8 дней)
- ✅ Масштабируемость (легко добавить новые языки)
- ✅ Clean Architecture + DDD + SOLID принципы

---

## 1. Архитектурное Решение

### 1.1 Почему Extism + Dart Subprocess?

#### Сравнение подходов:

| Подход | Языки | Сложность | Производительность | Battle-tested |
|--------|-------|-----------|-------------------|---------------|
| **Extism (WASM)** | Rust, Go, JS, C, C# | Низкая | Высокая | ✅ Zed, moonrepo |
| **Pure Wazero** | Все (WASM) | Средняя | Высокая | ✅ Arcjet, wasmCloud |
| **Dart Subprocess** | Dart | Средняя | Средняя | ✅ LSP, MCP 2024 |
| **dart_eval** | Dart | Высокая | Низкая | ⚠️ Только Flutter |
| **HTTP Webhooks** | Все | Очень низкая | Низкая | ✅ Zapier, n8n |

**Выбор**: Extism (primary) + Dart Subprocess (optional)

**Обоснование**:
1. **Extism** использует **Wazero под капотом** - получаем лучшее из обоих миров
2. **Низкая сложность** - в 5-10x меньше кода чем pure Wazero
3. **Готовые PDK** для 8 языков (не нужно писать binding'и вручную)
4. **Battle-tested** - production use в Zed, Shopify Functions, moonrepo
5. **Dart support** через subprocess - проверенный паттерн (Language Server Protocol, Model Context Protocol)

### 1.2 Когда использовать что?

**Extism (WASM)** - для 90% use cases:
- HTTP request/response трансформация
- JSON обработка
- Header manipulation
- Mock responses
- Производительные фильтры

**Dart Subprocess** - для Dart-специфичных задач:
- Flutter ecosystem интеграция
- package:http анализ
- Dart-native библиотеки
- Complex Dart logic

### 1.3 Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│              Presentation Layer (HTTP API, WebSocket)        │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                   Application/Use Case Layer                 │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         ScriptService (Orchestrator)                 │  │
│  │  - ExecuteForRequest(ctx, req) → modifiedReq         │  │
│  │  - ExecuteForResponse(ctx, resp) → modifiedResp      │  │
│  │  - CreateScript / GetScript / ListScripts            │  │
│  └────────────────┬─────────────────────────────────────┘  │
│                   │                                          │
│  ┌────────────────▼─────────────────────────────────────┐  │
│  │      ScriptExecutor (Port - Interface)               │  │
│  │  Execute(ctx, req) → (result, error)                 │  │
│  │  Runtime() → ScriptRuntime                           │  │
│  │  Validate(script) → error                            │  │
│  └──┬──────────────────────┬──────────────────────────┬─┘  │
└─────┼──────────────────────┼──────────────────────────┼────┘
      │                      │                          │
┌─────▼──────────┐  ┌────────▼──────────┐  ┌───────────▼──────┐
│ ExtismExecutor │  │  DartExecutor     │  │ Future: Python,  │
│ (WASM)         │  │  (Subprocess)     │  │ Node.js, etc.    │
└────────────────┘  └───────────────────┘  └──────────────────┘
       │                     │
       │                     │
┌──────▼─────────┐  ┌────────▼──────────────────────┐
│ WASM Modules   │  │ Dart Scripts via JSON-RPC     │
│ (.wasm files)  │  │ (stdin/stdout communication)  │
└────────────────┘  └───────────────────────────────┘
```

---

## 2. Domain Layer Design (DDD)

### 2.1 Entities

**Script** - Aggregate Root:
```go
type Script struct {
    // Identity
    ID          string
    Name        string
    Description string

    // Script Configuration
    Runtime     ScriptRuntime  // extism | dart
    Code        []byte         // WASM binary или Dart source
    Language    string         // rust | go | javascript | dart

    // Trigger Configuration
    TriggerType TriggerType    // request | response | both
    MatchRules  MatchRules     // Pattern matching (reuse from mapping)
    Priority    int            // Execution order

    // Resource Limits
    Config      ScriptConfig   // timeout, memory limits

    // State
    Enabled     bool

    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Validation (domain logic)
func (s *Script) Validate() error {
    if s.Name == "" {
        return errors.New("script name required")
    }
    if len(s.Code) == 0 {
        return errors.New("script code required")
    }
    if s.Config.TimeoutMs <= 0 {
        s.Config.TimeoutMs = 5000 // default 5s
    }
    return nil
}
```

### 2.2 Value Objects

```go
// ScriptRuntime - какой executor использовать
type ScriptRuntime string

const (
    RuntimeExtism ScriptRuntime = "extism"
    RuntimeDart   ScriptRuntime = "dart"
)

// ScriptConfig - настройки выполнения
type ScriptConfig struct {
    TimeoutMs     int    // Execution timeout
    MemoryLimitMB int    // Memory limit (для WASM)
    AllowedHosts  []string // Allowed hosts для http_fetch
}

// TriggerType - когда выполнять скрипт
type TriggerType string

const (
    TriggerRequest  TriggerType = "request"
    TriggerResponse TriggerType = "response"
    TriggerBoth     TriggerType = "both"
)

// MatchRules - правила сопоставления (переиспользуем из mapping feature)
type MatchRules struct {
    Methods      []string  // ["GET", "POST"]
    HostPattern  string    // "api.example.com"
    PathPattern  string    // "/users/*"
    PatternType  PatternType // exact | prefix | regex
}

// ScriptContext - контекст выполнения (передаётся в скрипт)
type ScriptContext struct {
    Request  *HTTPRequest
    Response *HTTPResponse
    Session  *SessionInfo
    Transaction *TransactionInfo
}

type HTTPRequest struct {
    Method  string
    URL     string
    Headers map[string][]string
    Body    []byte
}

type HTTPResponse struct {
    Status  int
    Headers map[string][]string
    Body    []byte
}

// ScriptResult - результат выполнения
type ScriptResult struct {
    Modified         bool
    ModifiedRequest  *HTTPRequest
    ModifiedResponse *HTTPResponse
    Logs             []string
    Duration         time.Duration
    Error            string
}
```

### 2.3 Ports (Interfaces)

```go
// ScriptExecutor - PORT для выполнения скриптов
type ScriptExecutor interface {
    // Execute выполняет скрипт с заданным input
    Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error)

    // Runtime возвращает тип runtime
    Runtime() ScriptRuntime

    // Validate проверяет синтаксис скрипта
    Validate(ctx context.Context, script Script) error

    // Close освобождает ресурсы
    Close() error
}

// ScriptRepository - PORT для хранения скриптов
type ScriptRepository interface {
    Save(ctx context.Context, script *Script) error
    Get(ctx context.Context, id string) (*Script, error)
    List(ctx context.Context, filter ScriptFilter) ([]*Script, error)
    Delete(ctx context.Context, id string) error
    UpdateEnabled(ctx context.Context, id string, enabled bool) error
}

// ExecutionRequest - запрос на выполнение
type ExecutionRequest struct {
    Script  Script
    Input   []byte // JSON-encoded ScriptContext
    Context map[string]any
}

// ExecutionResult - результат выполнения
type ExecutionResult struct {
    Output   []byte
    Logs     []string
    Duration time.Duration
    Error    string
}
```

### 2.4 Domain Events

```go
type ScriptExecutedEvent struct {
    ScriptID  string
    Duration  time.Duration
    Success   bool
    Error     string
    Timestamp time.Time
}

type ScriptFailedEvent struct {
    ScriptID string
    Error    error
    Timestamp time.Time
}

type ScriptCreatedEvent struct {
    ScriptID string
    Name     string
}
```

---

## 3. Implementation Phases

### Phase 1: Domain Layer (День 1)

**Цель**: Создать чистую бизнес-логику без зависимостей

**Файлы**:
```
internal/features/scripting/domain/
├── executor.go          # ScriptExecutor interface
├── script.go           # Script entity, value objects
├── repository.go       # ScriptRepository interface
└── events.go           # Domain events
```

**Детали**:

**`executor.go`**:
```go
package domain

import (
    "context"
    "time"
)

type ScriptRuntime string

const (
    RuntimeExtism ScriptRuntime = "extism"
    RuntimeDart   ScriptRuntime = "dart"
)

// ScriptExecutor - интерфейс для всех runtime'ов
type ScriptExecutor interface {
    Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error)
    Runtime() ScriptRuntime
    Validate(ctx context.Context, script Script) error
    Close() error
}

type ExecutionRequest struct {
    Script  Script
    Input   []byte
    Context map[string]any
}

type ExecutionResult struct {
    Output   []byte
    Logs     []string
    Duration time.Duration
    Error    string
}
```

**`script.go`**:
```go
package domain

import (
    "errors"
    "time"
)

type Script struct {
    ID          string
    Name        string
    Description string
    Runtime     ScriptRuntime
    Code        []byte
    Language    string
    TriggerType TriggerType
    MatchRules  MatchRules
    Priority    int
    Config      ScriptConfig
    Enabled     bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type TriggerType string

const (
    TriggerRequest  TriggerType = "request"
    TriggerResponse TriggerType = "response"
    TriggerBoth     TriggerType = "both"
)

type MatchRules struct {
    Methods     []string
    HostPattern string
    PathPattern string
    PatternType PatternType
}

type PatternType string

const (
    PatternExact  PatternType = "exact"
    PatternPrefix PatternType = "prefix"
    PatternRegex  PatternType = "regex"
)

type ScriptConfig struct {
    TimeoutMs     int
    MemoryLimitMB int
    AllowedHosts  []string
}

// Validate - domain validation logic
func (s *Script) Validate() error {
    if s.Name == "" {
        return errors.New("script name is required")
    }
    if len(s.Code) == 0 {
        return errors.New("script code is required")
    }
    if s.Config.TimeoutMs <= 0 {
        s.Config.TimeoutMs = 5000
    }
    if s.Config.MemoryLimitMB <= 0 {
        s.Config.MemoryLimitMB = 10
    }
    return nil
}

// Context and result types
type ScriptContext struct {
    Request     *HTTPRequest
    Response    *HTTPResponse
    Session     *SessionInfo
    Transaction *TransactionInfo
}

type HTTPRequest struct {
    Method  string
    URL     string
    Headers map[string][]string
    Body    []byte
}

type HTTPResponse struct {
    Status  int
    Headers map[string][]string
    Body    []byte
}

type SessionInfo struct {
    ID         string
    ClientAddr string
}

type TransactionInfo struct {
    ID       string
    Duration time.Duration
}

type ScriptResult struct {
    Modified         bool
    ModifiedRequest  *HTTPRequest
    ModifiedResponse *HTTPResponse
    Logs             []string
    Error            string
}
```

---

### Phase 2: Database Schema (День 1)

**Миграция** `migrations/0005_scripting.sql`:
```sql
-- Scripting feature
CREATE TABLE IF NOT EXISTS scripts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,

    -- Runtime configuration
    runtime TEXT NOT NULL CHECK(runtime IN ('extism', 'dart')),
    code BLOB NOT NULL,
    language TEXT NOT NULL, -- 'rust', 'javascript', 'go', 'dart', etc.

    -- Trigger configuration
    trigger_type TEXT NOT NULL CHECK(trigger_type IN ('request', 'response', 'both')),
    priority INTEGER DEFAULT 0,

    -- Match rules (JSON)
    match_methods TEXT, -- JSON array: ["GET", "POST"]
    match_host_pattern TEXT,
    match_path_pattern TEXT,
    pattern_type TEXT CHECK(pattern_type IN ('exact', 'prefix', 'regex')),

    -- Resource limits
    timeout_ms INTEGER DEFAULT 5000,
    memory_limit_mb INTEGER DEFAULT 10,
    allowed_hosts TEXT, -- JSON array

    -- State
    enabled BOOLEAN DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scripts_enabled ON scripts(enabled);
CREATE INDEX idx_scripts_runtime ON scripts(runtime);
CREATE INDEX idx_scripts_priority ON scripts(priority DESC);
CREATE INDEX idx_scripts_trigger_type ON scripts(trigger_type);
```

**GORM Models** `internal/features/scripting/infrastructure/persistence/models.go`:
```go
package persistence

import (
    "database/sql/driver"
    "encoding/json"
    "time"

    "network-debugger/internal/features/scripting/domain"
)

type ScriptModel struct {
    ID          string `gorm:"primaryKey"`
    Name        string `gorm:"not null"`
    Description string

    Runtime  string `gorm:"not null"`
    Code     []byte `gorm:"type:blob;not null"`
    Language string `gorm:"not null"`

    TriggerType string `gorm:"not null"`
    Priority    int    `gorm:"default:0"`

    MatchMethods     StringArray `gorm:"type:text"`
    MatchHostPattern string
    MatchPathPattern string
    PatternType      string

    TimeoutMs     int `gorm:"default:5000"`
    MemoryLimitMB int `gorm:"default:10"`
    AllowedHosts  StringArray `gorm:"type:text"`

    Enabled bool `gorm:"default:true"`

    CreatedAt time.Time
    UpdatedAt time.Time
}

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
    if len(s) == 0 {
        return nil, nil
    }
    return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
    if value == nil {
        *s = []string{}
        return nil
    }
    return json.Unmarshal(value.([]byte), s)
}

func (m *ScriptModel) TableName() string {
    return "scripts"
}

// ToDomain converts GORM model to domain entity
func (m *ScriptModel) ToDomain() *domain.Script {
    return &domain.Script{
        ID:          m.ID,
        Name:        m.Name,
        Description: m.Description,
        Runtime:     domain.ScriptRuntime(m.Runtime),
        Code:        m.Code,
        Language:    m.Language,
        TriggerType: domain.TriggerType(m.TriggerType),
        MatchRules: domain.MatchRules{
            Methods:     m.MatchMethods,
            HostPattern: m.MatchHostPattern,
            PathPattern: m.MatchPathPattern,
            PatternType: domain.PatternType(m.PatternType),
        },
        Priority: m.Priority,
        Config: domain.ScriptConfig{
            TimeoutMs:     m.TimeoutMs,
            MemoryLimitMB: m.MemoryLimitMB,
            AllowedHosts:  m.AllowedHosts,
        },
        Enabled:   m.Enabled,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }
}

// FromDomain converts domain entity to GORM model
func FromDomain(s *domain.Script) *ScriptModel {
    return &ScriptModel{
        ID:               s.ID,
        Name:             s.Name,
        Description:      s.Description,
        Runtime:          string(s.Runtime),
        Code:             s.Code,
        Language:         s.Language,
        TriggerType:      string(s.TriggerType),
        Priority:         s.Priority,
        MatchMethods:     s.MatchRules.Methods,
        MatchHostPattern: s.MatchRules.HostPattern,
        MatchPathPattern: s.MatchRules.PathPattern,
        PatternType:      string(s.MatchRules.PatternType),
        TimeoutMs:        s.Config.TimeoutMs,
        MemoryLimitMB:    s.Config.MemoryLimitMB,
        AllowedHosts:     s.Config.AllowedHosts,
        Enabled:          s.Enabled,
        CreatedAt:        s.CreatedAt,
        UpdatedAt:        s.UpdatedAt,
    }
}
```

**Repository** `internal/features/scripting/infrastructure/persistence/repo.go`:
```go
package persistence

import (
    "context"

    "gorm.io/gorm"
    "network-debugger/internal/features/scripting/domain"
)

type GormScriptRepository struct {
    db *gorm.DB
}

func NewGormScriptRepository(db *gorm.DB) *GormScriptRepository {
    return &GormScriptRepository{db: db}
}

func (r *GormScriptRepository) Save(ctx context.Context, script *domain.Script) error {
    model := FromDomain(script)
    return r.db.WithContext(ctx).Save(model).Error
}

func (r *GormScriptRepository) Get(ctx context.Context, id string) (*domain.Script, error) {
    var model ScriptModel
    if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return model.ToDomain(), nil
}

func (r *GormScriptRepository) List(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
    var models []ScriptModel
    query := r.db.WithContext(ctx).Order("priority DESC, created_at ASC")

    if filter.Enabled != nil {
        query = query.Where("enabled = ?", *filter.Enabled)
    }
    if filter.Runtime != "" {
        query = query.Where("runtime = ?", filter.Runtime)
    }
    if filter.TriggerType != "" {
        query = query.Where("trigger_type = ?", filter.TriggerType)
    }

    if err := query.Find(&models).Error; err != nil {
        return nil, err
    }

    scripts := make([]*domain.Script, len(models))
    for i, m := range models {
        scripts[i] = m.ToDomain()
    }
    return scripts, nil
}

func (r *GormScriptRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&ScriptModel{}, "id = ?", id).Error
}

func (r *GormScriptRepository) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
    return r.db.WithContext(ctx).
        Model(&ScriptModel{}).
        Where("id = ?", id).
        Update("enabled", enabled).Error
}
```

---

### Phase 3: Extism Executor (День 2-3)

**Зависимости**:
```bash
go get github.com/extism/go-sdk
```

**Executor** `internal/features/scripting/infrastructure/extism/executor.go`:
```go
package extism

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    extism "github.com/extism/go-sdk"
    "github.com/tetratelabs/wazero/api"

    "network-debugger/internal/features/scripting/domain"
)

type ExtismExecutor struct {
    plugins  map[string]*extism.Plugin
    hostFns  []extism.HostFunction
}

func NewExtismExecutor() *ExtismExecutor {
    return &ExtismExecutor{
        plugins: make(map[string]*extism.Plugin),
        hostFns: createHostFunctions(),
    }
}

func (e *ExtismExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
    start := time.Now()

    // Get or create plugin
    plugin, err := e.getOrCreatePlugin(req.Script)
    if err != nil {
        return domain.ExecutionResult{}, fmt.Errorf("failed to load plugin: %w", err)
    }

    // Apply timeout
    ctx, cancel := context.WithTimeout(ctx, time.Duration(req.Script.Config.TimeoutMs)*time.Millisecond)
    defer cancel()

    // Execute plugin
    output, err := plugin.Call("process", req.Input)
    duration := time.Since(start)

    if err != nil {
        return domain.ExecutionResult{
            Duration: duration,
            Error:    err.Error(),
        }, nil // Return result with error, not error itself
    }

    return domain.ExecutionResult{
        Output:   output,
        Duration: duration,
    }, nil
}

func (e *ExtismExecutor) Runtime() domain.ScriptRuntime {
    return domain.RuntimeExtism
}

func (e *ExtismExecutor) Validate(ctx context.Context, script domain.Script) error {
    // Try to load plugin
    _, err := e.getOrCreatePlugin(script)
    return err
}

func (e *ExtismExecutor) Close() error {
    for _, plugin := range e.plugins {
        plugin.Close()
    }
    return nil
}

func (e *ExtismExecutor) getOrCreatePlugin(script domain.Script) (*extism.Plugin, error) {
    if plugin, ok := e.plugins[script.ID]; ok {
        return plugin, nil
    }

    ctx := context.Background()

    manifest := extism.Manifest{
        Wasm: []extism.Wasm{
            extism.WasmData{Data: script.Code},
        },
        Memory: &extism.Memory{
            MaxPages: uint32(script.Config.MemoryLimitMB * 16), // 64KB per page
        },
        AllowedHosts: script.Config.AllowedHosts,
    }

    config := extism.PluginConfig{
        EnableWasi: false, // Security: disable WASI by default
    }

    plugin, err := extism.NewPlugin(ctx, manifest, config, e.hostFns)
    if err != nil {
        return nil, err
    }

    e.plugins[script.ID] = plugin
    return plugin, nil
}
```

**Host Functions** `internal/features/scripting/infrastructure/extism/host_functions.go`:
```go
package extism

import (
    "context"
    "fmt"
    "log"

    extism "github.com/extism/go-sdk"
    "github.com/tetratelabs/wazero/api"
)

func createHostFunctions() []extism.HostFunction {
    return []extism.HostFunction{
        // Log function - plugins can log messages
        extism.NewHostFunctionWithStack(
            "log",
            func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
                offset := stack[0]
                msg, err := plugin.ReadString(offset)
                if err == nil {
                    log.Printf("[Script Log] %s", msg)
                }
            },
            []api.ValueType{api.ValueTypeI64},
            []api.ValueType{},
        ),

        // HTTP fetch function - plugins can make HTTP requests
        extism.NewHostFunctionWithStack(
            "http_fetch",
            func(ctx context.Context, plugin *extism.CurrentPlugin, stack []uint64) {
                urlOffset := stack[0]
                url, _ := plugin.ReadString(urlOffset)

                // TODO: Implement HTTP client with proper security controls
                response := fmt.Sprintf(`{"status": 200, "body": "Mock response for %s"}`, url)

                plugin.ReturnString(response)
            },
            []api.ValueType{api.ValueTypeI64},
            []api.ValueType{api.ValueTypeI64},
        ),
    }
}
```

---

### Phase 4: Dart Executor (День 4-5)

**Executor** `internal/features/scripting/infrastructure/dart/executor.go`:
```go
package dart

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os/exec"
    "sync"
    "time"

    "network-debugger/internal/features/scripting/domain"
)

type DartExecutor struct {
    processPool *ProcessPool
}

func NewDartExecutor(maxProcesses int, scriptRunnerPath string) (*DartExecutor, error) {
    pool, err := NewProcessPool(maxProcesses, scriptRunnerPath)
    if err != nil {
        return nil, err
    }

    return &DartExecutor{
        processPool: pool,
    }, nil
}

func (e *DartExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
    start := time.Now()

    // Get process from pool
    proc, err := e.processPool.Get(ctx)
    if err != nil {
        return domain.ExecutionResult{}, err
    }
    defer e.processPool.Release(proc)

    // Create JSON-RPC request
    rpcReq := map[string]any{
        "jsonrpc": "2.0",
        "method":  "execute",
        "params": map[string]any{
            "code":  string(req.Script.Code),
            "input": string(req.Input),
        },
        "id": time.Now().UnixNano(),
    }

    // Send request
    if err := json.NewEncoder(proc.stdin).Encode(rpcReq); err != nil {
        return domain.ExecutionResult{}, fmt.Errorf("failed to send request: %w", err)
    }

    // Read response with timeout
    type response struct {
        result domain.ExecutionResult
        err    error
    }

    respChan := make(chan response, 1)
    go func() {
        result, err := proc.readResponse()
        respChan <- response{result, err}
    }()

    select {
    case <-ctx.Done():
        return domain.ExecutionResult{
            Duration: time.Since(start),
            Error:    "timeout",
        }, nil
    case resp := <-respChan:
        resp.result.Duration = time.Since(start)
        return resp.result, resp.err
    }
}

func (e *DartExecutor) Runtime() domain.ScriptRuntime {
    return domain.RuntimeDart
}

func (e *DartExecutor) Validate(ctx context.Context, script domain.Script) error {
    // TODO: Syntax validation
    return nil
}

func (e *DartExecutor) Close() error {
    return e.processPool.Close()
}
```

**Process Pool** `internal/features/scripting/infrastructure/dart/process_pool.go`:
```go
package dart

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os/exec"
    "sync"

    "network-debugger/internal/features/scripting/domain"
)

type ProcessPool struct {
    processes       chan *DartProcess
    maxSize         int
    scriptRunnerPath string
    mu              sync.Mutex
}

type DartProcess struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout *bufio.Scanner
}

func NewProcessPool(maxSize int, scriptRunnerPath string) (*ProcessPool, error) {
    pool := &ProcessPool{
        processes:       make(chan *DartProcess, maxSize),
        maxSize:         maxSize,
        scriptRunnerPath: scriptRunnerPath,
    }

    // Pre-spawn processes
    for i := 0; i < maxSize; i++ {
        proc, err := pool.startProcess()
        if err != nil {
            return nil, fmt.Errorf("failed to start Dart process: %w", err)
        }
        pool.processes <- proc
    }

    return pool, nil
}

func (p *ProcessPool) Get(ctx context.Context) (*DartProcess, error) {
    select {
    case proc := <-p.processes:
        return proc, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (p *ProcessPool) Release(proc *DartProcess) {
    select {
    case p.processes <- proc:
    default:
        // Pool full, kill process
        proc.cmd.Process.Kill()
    }
}

func (p *ProcessPool) Close() error {
    close(p.processes)
    for proc := range p.processes {
        proc.cmd.Process.Kill()
    }
    return nil
}

func (p *ProcessPool) startProcess() (*DartProcess, error) {
    cmd := exec.Command("dart", "run", p.scriptRunnerPath)

    stdin, err := cmd.StdinPipe()
    if err != nil {
        return nil, err
    }

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }

    if err := cmd.Start(); err != nil {
        return nil, err
    }

    return &DartProcess{
        cmd:    cmd,
        stdin:  stdin,
        stdout: bufio.NewScanner(stdout),
    }, nil
}

func (p *DartProcess) readResponse() (domain.ExecutionResult, error) {
    if !p.stdout.Scan() {
        return domain.ExecutionResult{}, errors.New("no response from Dart process")
    }

    var rpcResp struct {
        Result struct {
            Output string   `json:"output"`
            Logs   []string `json:"logs"`
        } `json:"result"`
        Error *struct {
            Message string `json:"message"`
        } `json:"error"`
    }

    if err := json.Unmarshal(p.stdout.Bytes(), &rpcResp); err != nil {
        return domain.ExecutionResult{}, err
    }

    if rpcResp.Error != nil {
        return domain.ExecutionResult{
            Error: rpcResp.Error.Message,
        }, nil
    }

    return domain.ExecutionResult{
        Output: []byte(rpcResp.Result.Output),
        Logs:   rpcResp.Result.Logs,
    }, nil
}
```

**Dart Script Runner** `scripts/dart/script_runner.dart`:
```dart
import 'dart:io';
import 'dart:convert';
import 'dart:isolate';

void main() async {
  // JSON-RPC server over stdin/stdout
  final stdinStream = stdin
      .transform(utf8.decoder)
      .transform(const LineSplitter())
      .map((line) => json.decode(line));

  await for (final request in stdinStream) {
    try {
      final method = request['method'] as String;
      final params = request['params'] as Map<String, dynamic>;
      final id = request['id'];

      if (method == 'execute') {
        final code = params['code'] as String;
        final input = params['input'] as String;

        try {
          // Execute Dart code
          final result = await executeScript(code, input);

          // Send success response
          final response = {
            'jsonrpc': '2.0',
            'result': {
              'output': result,
              'logs': [],
            },
            'id': id,
          };

          stdout.writeln(json.encode(response));
        } catch (e, stack) {
          // Send error response
          final response = {
            'jsonrpc': '2.0',
            'error': {
              'code': -32000,
              'message': e.toString(),
              'data': stack.toString(),
            },
            'id': id,
          };

          stdout.writeln(json.encode(response));
        }
      }
    } catch (e) {
      stderr.writeln('Error processing request: $e');
    }
  }
}

Future<String> executeScript(String code, String input) async {
  // For MVP: Simple function call execution
  // For production: Use dart_eval or Isolate-based execution

  // This is a simplified implementation
  // In real scenario, you'd compile and execute the Dart code safely

  return 'Dart execution not yet implemented';
}
```

---

### Phase 5: Use Case Layer (День 6)

**Service** `internal/features/scripting/usecase/service.go`:
```go
package usecase

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "network-debugger/internal/features/scripting/domain"
)

type ScriptService struct {
    executors map[domain.ScriptRuntime]domain.ScriptExecutor
    repo      domain.ScriptRepository
}

func NewScriptService(repo domain.ScriptRepository) *ScriptService {
    return &ScriptService{
        executors: make(map[domain.ScriptRuntime]domain.ScriptExecutor),
        repo:      repo,
    }
}

// RegisterExecutor adds a runtime executor (Dependency Inversion)
func (s *ScriptService) RegisterExecutor(executor domain.ScriptExecutor) {
    s.executors[executor.Runtime()] = executor
    log.Printf("Registered script executor: %s", executor.Runtime())
}

// ExecuteForRequest executes scripts for request hook
func (s *ScriptService) ExecuteForRequest(ctx context.Context, req *domain.HTTPRequest, session *domain.SessionInfo) (*domain.HTTPRequest, error) {
    enabled := true
    scripts, err := s.repo.List(ctx, domain.ScriptFilter{
        Enabled:     &enabled,
        TriggerType: domain.TriggerRequest,
    })
    if err != nil {
        return req, err
    }

    // Filter scripts matching this request
    matchedScripts := s.filterMatchingScripts(scripts, req)

    currentReq := req
    for _, script := range matchedScripts {
        executor, ok := s.executors[script.Runtime]
        if !ok {
            log.Printf("No executor for runtime %s, skipping script %s", script.Runtime, script.Name)
            continue
        }

        // Prepare input
        scriptCtx := domain.ScriptContext{
            Request: currentReq,
            Session: session,
        }
        input, _ := json.Marshal(scriptCtx)

        // Execute script
        result, err := executor.Execute(ctx, domain.ExecutionRequest{
            Script: *script,
            Input:  input,
        })

        if err != nil {
            log.Printf("Script %s execution failed: %v", script.Name, err)
            continue // Don't break chain on error
        }

        if result.Error != "" {
            log.Printf("Script %s returned error: %s", script.Name, result.Error)
            continue
        }

        // Parse result
        var scriptResult domain.ScriptResult
        if err := json.Unmarshal(result.Output, &scriptResult); err != nil {
            log.Printf("Failed to parse script result: %v", err)
            continue
        }

        // Apply modifications
        if scriptResult.Modified && scriptResult.ModifiedRequest != nil {
            currentReq = scriptResult.ModifiedRequest
        }

        // Log script logs
        for _, logMsg := range result.Logs {
            log.Printf("[Script %s] %s", script.Name, logMsg)
        }
    }

    return currentReq, nil
}

// ExecuteForResponse executes scripts for response hook
func (s *ScriptService) ExecuteForResponse(ctx context.Context, req *domain.HTTPRequest, resp *domain.HTTPResponse, tx *domain.TransactionInfo) (*domain.HTTPResponse, error) {
    enabled := true
    scripts, err := s.repo.List(ctx, domain.ScriptFilter{
        Enabled:     &enabled,
        TriggerType: domain.TriggerResponse,
    })
    if err != nil {
        return resp, err
    }

    matchedScripts := s.filterMatchingScripts(scripts, req)

    currentResp := resp
    for _, script := range matchedScripts {
        executor, ok := s.executors[script.Runtime]
        if !ok {
            continue
        }

        scriptCtx := domain.ScriptContext{
            Request:     req,
            Response:    currentResp,
            Transaction: tx,
        }
        input, _ := json.Marshal(scriptCtx)

        result, err := executor.Execute(ctx, domain.ExecutionRequest{
            Script: *script,
            Input:  input,
        })

        if err != nil || result.Error != "" {
            log.Printf("Script %s failed: %v %s", script.Name, err, result.Error)
            continue
        }

        var scriptResult domain.ScriptResult
        if err := json.Unmarshal(result.Output, &scriptResult); err != nil {
            continue
        }

        if scriptResult.Modified && scriptResult.ModifiedResponse != nil {
            currentResp = scriptResult.ModifiedResponse
        }
    }

    return currentResp, nil
}

// filterMatchingScripts filters scripts by match rules
func (s *ScriptService) filterMatchingScripts(scripts []*domain.Script, req *domain.HTTPRequest) []*domain.Script {
    // TODO: Implement pattern matching logic (similar to mapping feature)
    // For now, return all scripts
    return scripts
}

// CRUD methods
func (s *ScriptService) CreateScript(ctx context.Context, script *domain.Script) error {
    if err := script.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // Validate syntax
    executor, ok := s.executors[script.Runtime]
    if !ok {
        return fmt.Errorf("no executor for runtime: %s", script.Runtime)
    }

    if err := executor.Validate(ctx, *script); err != nil {
        return fmt.Errorf("script validation failed: %w", err)
    }

    return s.repo.Save(ctx, script)
}

func (s *ScriptService) GetScript(ctx context.Context, id string) (*domain.Script, error) {
    return s.repo.Get(ctx, id)
}

func (s *ScriptService) ListScripts(ctx context.Context) ([]*domain.Script, error) {
    return s.repo.List(ctx, domain.ScriptFilter{})
}

func (s *ScriptService) UpdateScript(ctx context.Context, script *domain.Script) error {
    if err := script.Validate(); err != nil {
        return err
    }
    return s.repo.Save(ctx, script)
}

func (s *ScriptService) DeleteScript(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *ScriptService) ToggleScript(ctx context.Context, id string, enabled bool) error {
    return s.repo.UpdateEnabled(ctx, id, enabled)
}
```

---

### Phase 6: HTTP API & Integration (День 7-8)

**Handlers** `internal/infrastructure/httpapi/script_handlers.go`:
```go
package httpapi

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/google/uuid"
    "network-debugger/internal/features/scripting/domain"
    "network-debugger/internal/features/scripting/usecase"
)

type ScriptHandlers struct {
    service *usecase.ScriptService
}

func NewScriptHandlers(service *usecase.ScriptService) *ScriptHandlers {
    return &ScriptHandlers{service: service}
}

// POST /_api/v1/scripts
func (h *ScriptHandlers) CreateScript(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name        string   `json:"name"`
        Description string   `json:"description"`
        Runtime     string   `json:"runtime"`
        Code        string   `json:"code"`
        Language    string   `json:"language"`
        TriggerType string   `json:"triggerType"`
        Priority    int      `json:"priority"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    script := &domain.Script{
        ID:          uuid.New().String(),
        Name:        req.Name,
        Description: req.Description,
        Runtime:     domain.ScriptRuntime(req.Runtime),
        Code:        []byte(req.Code),
        Language:    req.Language,
        TriggerType: domain.TriggerType(req.TriggerType),
        Priority:    req.Priority,
        Config: domain.ScriptConfig{
            TimeoutMs:     5000,
            MemoryLimitMB: 10,
        },
        Enabled:   true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := h.service.CreateScript(r.Context(), script); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(script)
}

// GET /_api/v1/scripts
func (h *ScriptHandlers) ListScripts(w http.ResponseWriter, r *http.Request) {
    scripts, err := h.service.ListScripts(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(scripts)
}

// GET /_api/v1/scripts/{id}
// PUT /_api/v1/scripts/{id}
// DELETE /_api/v1/scripts/{id}
// PATCH /_api/v1/scripts/{id}/toggle
// ...
```

**Router Integration** - модифицировать `internal/infrastructure/httpapi/router.go`:
```go
// Add to Deps
type Deps struct {
    // ... existing fields
    ScriptSvc *usecase.ScriptService
}

// Add routes
mux.HandleFunc("POST /_api/v1/scripts", scriptHandlers.CreateScript)
mux.HandleFunc("GET /_api/v1/scripts", scriptHandlers.ListScripts)
// ...
```

**Proxy Integration** - модифицировать `internal/infrastructure/httpapi/httpproxy.go`:
```go
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. Execute request scripts (FIRST)
    if p.scriptService != nil {
        modifiedReq, err := p.scriptService.ExecuteForRequest(r.Context(), toScriptReq(r), sessionInfo)
        if err == nil && modifiedReq != nil {
            r = fromScriptReq(modifiedReq)
        }
    }

    // 2. UI Intercept (SECOND - existing code)
    if p.interceptor.ShouldIntercept(r) {
        // ... existing intercept logic
    }

    // 3. Proxy request
    resp := p.proxy(r)

    // 4. Execute response scripts
    if p.scriptService != nil {
        modifiedResp, err := p.scriptService.ExecuteForResponse(r.Context(), toScriptReq(r), toScriptResp(resp), txInfo)
        if err == nil && modifiedResp != nil {
            resp = fromScriptResp(modifiedResp)
        }
    }

    writeResponse(w, resp)
}
```

---

## 4. Testing Strategy

### Unit Tests

**Domain Layer**:
```go
func TestScript_Validate(t *testing.T) {
    script := &domain.Script{
        Name: "",
        Code: []byte("test"),
    }

    err := script.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "name is required")
}
```

**Executors**:
```go
func TestExtismExecutor_Execute(t *testing.T) {
    executor := extism.NewExtismExecutor()

    script := domain.Script{
        Code: loadTestWASM("hello.wasm"),
        Config: domain.ScriptConfig{TimeoutMs: 1000},
    }

    result, err := executor.Execute(context.Background(), domain.ExecutionRequest{
        Script: script,
        Input:  []byte(`{"test": "input"}`),
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, result.Output)
}
```

### Integration Tests

**End-to-End**:
```go
func TestScriptExecution_ModifyRequest(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    repo := persistence.NewGormScriptRepository(db)
    service := usecase.NewScriptService(repo)

    extismExec := extism.NewExtismExecutor()
    service.RegisterExecutor(extismExec)

    // Create script
    script := &domain.Script{
        Runtime: domain.RuntimeExtism,
        Code:    loadTestWASM("add_header.wasm"),
        TriggerType: domain.TriggerRequest,
    }
    service.CreateScript(context.Background(), script)

    // Execute
    req := &domain.HTTPRequest{
        Method: "GET",
        URL:    "http://example.com",
        Headers: map[string][]string{},
    }

    modifiedReq, err := service.ExecuteForRequest(context.Background(), req, nil)

    // Assert
    assert.NoError(t, err)
    assert.Contains(t, modifiedReq.Headers, "X-Modified")
}
```

---

## 5. Example Usage

### JavaScript Plugin (AssemblyScript)

**plugin.ts**:
```typescript
// @extism/as-pdk
import { Host } from "@extism/as-pdk";

class ScriptContext {
  request: HTTPRequest;
  response: HTTPResponse;
}

class HTTPRequest {
  method: string;
  url: string;
  headers: Map<string, string>;
  body: string;
}

export function process(input: string): string {
  const ctx: ScriptContext = JSON.parse(input);

  // Modify request
  ctx.request.headers["X-Modified"] = "true";
  ctx.request.headers["X-Timestamp"] = Date.now().toString();

  // Log
  Host.log("Modified request headers");

  return JSON.stringify({
    modified: true,
    modifiedRequest: ctx.request,
  });
}
```

**Compile**:
```bash
npm install @extism/as-pdk
asc plugin.ts -o plugin.wasm --optimize
```

**Upload via API**:
```bash
curl -X POST http://localhost:9092/_api/v1/scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Add Headers",
    "runtime": "extism",
    "language": "javascript",
    "code": "<base64-encoded WASM>",
    "triggerType": "request"
  }'
```

### Rust Plugin

**plugin.rs**:
```rust
use extism_pdk::*;
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
struct ScriptContext {
    request: HTTPRequest,
}

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: std::collections::HashMap<String, String>,
    body: Vec<u8>,
}

#[plugin_fn]
pub fn process(input: String) -> FnResult<String> {
    let mut ctx: ScriptContext = serde_json::from_str(&input)?;

    // Add CORS headers
    ctx.request.headers.insert("Access-Control-Allow-Origin".to_string(), "*".to_string());

    let result = serde_json::json!({
        "modified": true,
        "modifiedRequest": ctx.request,
    });

    Ok(result.to_string())
}
```

**Compile**:
```bash
cargo install cargo-wasi
cargo build --target wasm32-wasi --release
```

---

## 6. Deployment

### Single Binary

```bash
# Build with Extism support (pure Go, no CGO)
go build -o network-debugger ./cmd/network-debugger

# For Dart support, ensure Dart SDK is in PATH
export PATH="$PATH:/path/to/dart-sdk/bin"

./network-debugger
```

### Docker

```dockerfile
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o network-debugger ./cmd/network-debugger

# Include Dart runtime for Dart scripts
FROM dart:stable
COPY --from=builder /app/network-debugger /usr/local/bin/
COPY scripts/dart/script_runner.dart /app/scripts/

CMD ["network-debugger"]
```

---

## 7. Performance Considerations

### Metrics

**Expected Performance**:
- Extism (WASM): 20-50% overhead vs native
- Dart Subprocess: 100-200ms startup, then ~10-50ms per call
- Recommended: Use Extism for hot path, Dart for specific tasks

**Optimizations**:
1. **Plugin Caching**: Keep loaded WASM modules in memory
2. **Process Pooling**: Reuse Dart VM processes
3. **Lazy Loading**: Only load scripts when needed
4. **Parallel Execution**: Execute independent scripts concurrently

### Monitoring

```go
// Add metrics
type ExecutionMetrics struct {
    ScriptID     string
    Runtime      string
    Duration     time.Duration
    Success      bool
    Error        string
}

// Log metrics
log.Printf("Script %s executed in %v (success: %v)",
    metrics.ScriptID, metrics.Duration, metrics.Success)
```

---

## 8. Security

### Sandboxing

**Extism (WASM)**:
- ✅ Memory limits via `MaxPages`
- ✅ CPU timeout via context
- ✅ No filesystem access (WASI disabled)
- ✅ Controlled network access via host functions

**Dart Subprocess**:
- ✅ Process isolation
- ✅ Limited permissions
- ✅ Timeout via context
- ⚠️ Review Dart script code before execution

### Best Practices

1. **Always validate scripts** before saving
2. **Set reasonable timeouts** (default: 5s)
3. **Limit memory** (default: 10MB)
4. **Whitelist allowed hosts** for http_fetch
5. **Review user-submitted scripts** before enabling

---

## 9. Future Enhancements

### Phase 7: Python Support (Optional)

Add Python via Extism or subprocess similar to Dart

### Phase 8: Script Marketplace

- Community-contributed scripts
- Rating and reviews
- One-click install

### Phase 9: Advanced Features

- Script debugging console
- Hot reload from filesystem
- Script versioning
- Performance profiling

---

## 10. Success Metrics

**Technical**:
- ✅ Script execution < 100ms (p95)
- ✅ Support 3+ languages (JS, Rust, Dart minimum)
- ✅ Memory isolation working
- ✅ 100% test coverage for core

**Business**:
- ✅ Competitive score increase: 8.4 → 8.8+
- ✅ Unique differentiator vs Proxyman
- ✅ Community engagement (plugin contributions)

---

## 11. Timeline Summary

| Phase | Task | Duration |
|-------|------|----------|
| 1 | Domain Layer | 1 день |
| 2 | Database Schema | 1 день |
| 3 | Extism Executor | 2 дня |
| 4 | Dart Executor | 2 дня |
| 5 | Use Case Layer | 1 день |
| 6 | API & Integration | 1 день |
| **TOTAL** | **MVP Implementation** | **8 дней** |

---

## 12. SOLID Principles Compliance

✅ **Single Responsibility**:
- Domain: бизнес-логика
- Infrastructure: технические детали
- Use Case: оркестрация

✅ **Open/Closed**:
- Новый runtime? Реализуй ScriptExecutor interface

✅ **Liskov Substitution**:
- ExtismExecutor ↔ DartExecutor полностью взаимозаменяемы

✅ **Interface Segregation**:
- Минимальные интерфейсы (Execute, Validate, Close)

✅ **Dependency Inversion**:
- Use Case зависит от портов (интерфейсов), не от адаптеров

---

## 13. Clean Architecture Compliance

```
internal/features/scripting/
├── domain/              # Entities, Value Objects, Ports (NO dependencies)
├── usecase/             # Application Logic (depends on domain)
└── infrastructure/      # Adapters (depends on domain)
```

**Dependency Rule**: Зависимости направлены внутрь (к domain)

---

## Conclusion

Эта архитектура обеспечивает:
- ✅ **Battle-tested компоненты**
- ✅ **Простую реализацию** (8 дней)
- ✅ **Масштабируемость** (легко добавить языки)
- ✅ **Гибкость** (WASM + subprocess)
- ✅ **Безопасность** (sandboxing)
- ✅ **Clean Architecture** (чистое разделение слоёв)
- ✅ **SOLID принципы** (dependency inversion, open/closed)
- ✅ **DDD подход** (entities, value objects, repositories)

**Готово к реализации!** 🚀
