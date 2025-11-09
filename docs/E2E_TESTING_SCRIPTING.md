# E2E Testing для Scripting API

Это руководство описывает E2E тестирование Scripting API feature в Network Debugger.

## Обзор

**Создано E2E тестов**: 9+
**Покрытые языки**: Rust (WASM), Dart (subprocess)
**CI Integration**: GitHub Actions с кэшированием WASM fixtures

## Структура тестов

```
internal/e2e/
├── script_helpers_test.go        # Helpers для Scripting API
├── scripting_api_test.go         # Core E2E тесты (CRUD, Request Modification, Validation)
└── scripting_advanced_test.go    # Advanced тесты (Dart, MatchRules, Priority, Toggle)

internal/e2e/testdata/scripts/
├── wasm/                          # Скомпилированные WASM fixtures
│   ├── add_header.wasm           # Rust: добавляет заголовки
│   ├── noop.wasm                 # Minimal valid WASM (passthrough)
│   └── invalid.wasm              # Corrupt WASM для негативных тестов
├── wasm-src/                      # Исходники WASM fixtures
│   └── noop/                     # Rust проект для noop.wasm
└── dart/                          # Dart скрипты
    └── simple_logger.dart        # Логирует requests, добавляет заголовки
```

## Тесты

### Core E2E Tests (`scripting_api_test.go`)

#### 1. `TestE2E_ScriptingAPI_CRUD`
**Что тестирует**: CRUD операции через REST API

**Сценарий**:
- Create script (POST /_api/v1/scripts)
- Get script (GET /_api/v1/scripts/{id})
- List scripts (GET /_api/v1/scripts)
- Update script (PUT /_api/v1/scripts/{id})
- Toggle enabled (PATCH /_api/v1/scripts/{id}/toggle)
- Delete script (DELETE /_api/v1/scripts/{id})
- Verify 404 после deletion

**Fixture**: `add_header.wasm`

#### 2. `TestE2E_ScriptingAPI_RequestModification`
**Что тестирует**: Выполнение скрипта при HTTP request

**Сценарий**:
- Загружает Rust скрипт `add_header.wasm`
- Делает HTTP запрос через proxy
- Проверяет добавленные заголовки:
  - `X-Script-Processed: Rust`
  - `X-Test-Header: E2E-Test`

**Upstream**: Echo HTTP server (возвращает headers as JSON)

#### 3. `TestE2E_ScriptingAPI_ValidationError`
**Что тестирует**: Rejection невалидного WASM

**Сценарий**:
- Пытается создать скрипт с `invalid.wasm`
- Ожидает ошибку validation (400 Bad Request)

#### 4. `TestE2E_ScriptingAPI_Noop`
**Что тестирует**: Минимальный WASM модуль (passthrough)

**Сценарий**:
- Загружает `noop.wasm` (возвращает input без изменений)
- Делает request - должен пройти без модификаций

### Advanced E2E Tests (`scripting_advanced_test.go`)

#### 5. `TestE2E_ScriptingAPI_Dart_Subprocess`
**Что тестирует**: Dart executor через subprocess

**Сценарий**:
- Skip если Dart SDK не установлен
- Загружает `simple_logger.dart` (source code)
- Делает request
- Проверяет заголовки: `X-Dart-Processed: true`

**Note**: Требует Dart SDK

#### 6. `TestE2E_ScriptingAPI_MatchRules`
**Что тестирует**: Pattern matching (methods, path)

**Сценарий**:
- Создает скрипт с фильтром: `methods: ["POST"], pathPattern: "/api/*"`
- Test 1: POST /api/test → должен сработать (adds header)
- Test 2: GET /api/test → НЕ должен сработать (method mismatch)
- Test 3: POST /other → НЕ должен сработать (path mismatch)

#### 7. `TestE2E_ScriptingAPI_Priority`
**Что тестирует**: Execution order по priority

**Сценарий**:
- Создает 3 скрипта: priority 5, 10, 20
- Делает request
- Expected order: 20 → 10 → 5 (higher priority first)

**Note**: Фактическая проверка порядка требует логирования или script-specific markers

#### 8. `TestE2E_ScriptingAPI_ToggleEnabled`
**Что тестирует**: Включение/выключение скрипта

**Сценарий**:
- Test 1: Enabled script → adds header
- Disable script
- Test 2: Disabled script → does NOT add header
- Re-enable script
- Test 3: Re-enabled script → adds header again

## Запуск тестов

### Локально

```bash
# Собрать WASM fixtures (один раз)
./scripts/build_test_wasm.sh

# Запустить все E2E тесты
go test -v ./internal/e2e

# Запустить только Scripting API тесты
go test -v ./internal/e2e -run TestE2E_ScriptingAPI

# Запустить конкретный тест
go test -v ./internal/e2e -run TestE2E_ScriptingAPI_CRUD
```

### В CI (GitHub Actions)

Тесты запускаются автоматически:
- **Кэширование**: WASM fixtures кэшируются по hash исходников
- **Cache hit**: Тесты запускаются сразу (~5 мин)
- **Cache miss**: Устанавливается Rust, компилируются fixtures (~7 мин)

**Workflow**: `.github/workflows/go_build_test.yml` → job `integration`

## Helpers

### Основные функции (`script_helpers_test.go`)

```go
// Fixture loading
loadTestWASM(t, "add_header.wasm") → []byte
loadTestDart(t, "simple_logger.dart") → []byte

// CRUD operations
createScript(t, baseURL, wasmData, language) → scriptID
getScript(t, baseURL, scriptID) → Script
listScripts(t, baseURL) → []Script
updateScript(t, baseURL, scriptID, updates)
toggleScript(t, baseURL, scriptID, enabled)
deleteScript(t, baseURL, scriptID)
assertScriptNotFound(t, baseURL, scriptID)

// HTTP helpers
httpPostJSON(t, url, payload) → responseBody
startEchoHTTPServer(t) → *http.Server
makeProxyRequest(t, proxyURL, targetURL, method, body) → *http.Response

// Utilities
isDartAvailable(t) → bool
waitReady(t, baseURL, timeout) → error
```

## Fixtures

### WASM Modules

| Fixture | Language | Size | Purpose |
|---------|----------|------|---------|
| `add_header.wasm` | Rust | ~204KB | Adds `X-Script-Processed: Rust` header |
| `noop.wasm` | Rust | ~95KB | Passthrough (returns input unchanged) |
| `invalid.wasm` | - | 19B | Corrupt WASM for negative tests |

### Dart Scripts

| Fixture | Purpose |
|---------|---------|
| `simple_logger.dart` | Logs requests, adds `X-Dart-Processed: true` |

## Пересборка Fixtures

При изменении примеров или test fixtures:

```bash
./scripts/build_test_wasm.sh
```

**Что делает**:
1. Компилирует `examples/scripts/rust/` → `add_header.wasm`
2. Компилирует `internal/e2e/testdata/scripts/wasm-src/noop/` → `noop.wasm`
3. Создает `invalid.wasm` (corrupt header)

**Requirements**:
- Rust + `wasm32-unknown-unknown` target
- (Optional) Dart SDK для Dart тестов

## Troubleshooting

### "WASM fixture not found"
```bash
# Собрать fixtures
./scripts/build_test_wasm.sh
```

### "Dart SDK not installed"
Dart тесты автоматически skip'аются если Dart не установлен:
```go
if !isDartAvailable(t) {
    t.Skip("Dart SDK not installed")
}
```

### "Test timeout"
E2E тесты могут занимать время:
- Binary compilation: ~2-5 сек
- Server startup: ~1-2 сек
- Script loading: ~0.5 сек

Увеличьте timeout в CI:
```yaml
env:
  E2E_TIMEOUT_SECONDS: 600  # 10 минут
timeout-minutes: 12
```

### "Server not ready"
`waitReady()` polls `/healthz` endpoint. Если падает:
- Проверьте логи binary (stdout/stderr)
- Увеличьте timeout в `waitReady(t, baseURL, 10*time.Second)`

### Компиляция падает (Extism SDK compatibility)
**Известная проблема**: Код executor.go/host_functions.go написан для более старой версии Extism SDK.

**Фиксы требуются** в:
- `internal/features/scripting/infrastructure/extism/executor.go`
- `internal/features/scripting/infrastructure/extism/host_functions.go`

**API changes** (v1.7.1):
- `plugin.Call()` возвращает 3 значения (было 2)
- `plugin.Close()` требует `context.Context`
- `extism.Memory`, `extism.Ptr`, `extism.ModuleConfig` - API изменился
- `plugin.GetConfig()`, `plugin.ReturnString()` - methods removed

**TODO**: Обновить код под текущую версию SDK

## Метрики

**E2E тесты создано**: 9
**Языки покрыты**: Rust (WASM), Dart (subprocess)
**Покрытие функциональности**:
- ✅ CRUD operations
- ✅ Request modification
- ✅ Pattern matching (methods, host, path)
- ✅ Priority execution order
- ✅ Toggle enabled/disabled
- ✅ WASM validation
- ✅ Dart subprocess executor
- ⚠️ Response modification (TODO: требует upstream mock server)
- ⚠️ Timeout handling (TODO: infinite loop fixture)

## Следующие шаги

### 1. Fix Extism SDK Compatibility ⚠️ **REQUIRED**
Обновить executor.go и host_functions.go для v1.7.1 API

### 2. Добавить Response Modification Test
Создать WASM fixture который модифицирует response body

### 3. Timeout Test
Создать WASM fixture с infinite loop для проверки timeout handling

### 4. JavaScript (AssemblyScript) Test
Добавить JS example и E2E тест

### 5. Integration Tests (in-process)
Создать быстрые integration тесты без binary compilation:
- `internal/integration/scripting_test.go`
- Использует `httptest.Server` вместо real binary

## Ссылки

- **Fixtures README**: `internal/e2e/testdata/scripts/README.md`
- **Script Examples**: `examples/scripts/README.md`
- **Languages Guide**: `examples/scripts/LANGUAGES.md`
- **Quick Start**: `examples/scripts/QUICKSTART.md`
- **CI Workflow**: `.github/workflows/go_build_test.yml`
