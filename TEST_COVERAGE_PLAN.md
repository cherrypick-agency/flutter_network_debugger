# План улучшения покрытия тестами до 71%

**Текущее покрытие:** 65.9%  
**Целевое покрытие:** 71%  
**Необходимо добавить:** ~5.1%

## Агент 1: HTTP API Handlers (Script Handlers & Router)

### Цель: Улучшить покрытие HTTP API handlers с фокусом на script handlers и router

### Файлы для работы:
- `internal/infrastructure/httpapi/script_handlers.go`
  - `ExportScriptAsZip` (65.4%) - добавить тесты для ошибок создания ZIP файлов
  - `DownloadProject` (70.6%) - добавить тесты для edge cases
  - `DeleteScript` (75.0%) - добавить тесты для ошибок удаления

- `internal/infrastructure/httpapi/router.go`
  - `buildBaseMux` (61.8%) - добавить тесты для различных маршрутов и middleware

### Задачи:
1. Добавить тесты для `ExportScriptAsZip`:
   - Тест с пустым скриптом (без sourceCode и dependencies)
   - Тест с большим количеством зависимостей
   - Тест с различными типами языков

2. Добавить тесты для `DownloadProject`:
   - Тест со скриптом без sourceCode
   - Тест со скриптом без dependencies
   - Тест с большим количеством файлов

3. Добавить тесты для `buildBaseMux`:
   - Тест регистрации всех маршрутов
   - Тест middleware цепочки
   - Тест обработки ошибок маршрутизации

### Ожидаемый прирост покрытия: ~1.2-1.5%

---

## Агент 2: Scripting Infrastructure (Compilers & Downloaders)

### Цель: Улучшить покрытие компиляторов и загрузчиков скриптов

### Файлы для работы:
- `internal/features/scripting/infrastructure/compilers/rust.go`
  - `Compile` (53.5%) - добавить тесты для различных сценариев компиляции
  - `OptimizeWASM` (28.6%) - добавить тесты для оптимизации WASM

- `internal/features/scripting/infrastructure/compilers/swift.go`
  - `Compile` (52.3%) - добавить тесты для компиляции Swift
  - `ValidateSyntax` (14.3%) - добавить тесты для валидации синтаксиса

- `internal/features/scripting/infrastructure/compilers/kotlin.go`
  - `Compile` (61.8%) - добавить тесты для компиляции Kotlin
  - `ValidateSyntax` (22.6%) - добавить тесты для валидации синтаксиса

- `internal/features/scripting/infrastructure/compilers/zig.go`
  - `Compile` (58.6%) - добавить тесты для компиляции Zig
  - `ValidateSyntax` (16.7%) - добавить тесты для валидации синтаксиса

- `internal/features/scripting/infrastructure/downloaders/base.go`
  - `attemptDownload` (71.2%) - добавить тесты для retry логики
  - `ExtractTarXz` (73.5%) - добавить тесты для извлечения архивов
  - `ExtractTarGz` (74.3%) - добавить тесты для извлечения архивов
  - `ExtractZip` (71.4%) - добавить тесты для извлечения ZIP

### Задачи:
1. Добавить тесты для компиляторов:
   - Тесты успешной компиляции с различными входными данными
   - Тесты обработки ошибок компиляции
   - Тесты валидации синтаксиса с невалидным кодом
   - Тесты для `OptimizeWASM` с различными входными WASM файлами

2. Добавить тесты для downloaders:
   - Тесты retry логики при ошибках загрузки
   - Тесты извлечения различных типов архивов
   - Тесты обработки поврежденных архивов
   - Тесты обработки сетевых ошибок

### Ожидаемый прирост покрытия: ~1.3-1.6%

---

## Агент 3: HTTP Proxy & Forward Proxy Handlers

### Цель: Улучшить покрытие proxy handlers и forward proxy функциональности

### Файлы для работы:
- `internal/infrastructure/httpapi/httpproxy.go`
  - `handleHTTPProxy` (58.2%) - добавить тесты для различных сценариев проксирования
  - `spoolBody` (75.0%) - добавить тесты для edge cases

- `internal/infrastructure/httpapi/forwardproxy.go`
  - `handleForwardProxy` (62.5%) - добавить тесты для forward proxy
  - `handleHTTPForwardRequest` (45.4%) - добавить тесты для HTTP forward requests
  - `handleHTTPForwardWebSocket` (15.4%) - добавить тесты для WebSocket forwarding
  - `handleConnectTunnel` (43.3%) - добавить тесты для CONNECT туннелирования
  - `handleConnectMITM` (2.9%) - добавить тесты для MITM CONNECT

- `internal/infrastructure/httpapi/intercept_codec.go`
  - `encodeForIntercept` (66.7%) - добавить тесты для кодирования данных

### Задачи:
1. Добавить тесты для `handleHTTPProxy`:
   - Тесты проксирования различных HTTP методов
   - Тесты обработки заголовков
   - Тесты обработки тела запроса/ответа
   - Тесты обработки ошибок upstream сервера

2. Добавить тесты для forward proxy:
   - Тесты forward proxy для HTTP запросов
   - Тесты WebSocket forwarding
   - Тесты CONNECT туннелирования
   - Тесты MITM CONNECT (если возможно без реального сертификата)

3. Добавить тесты для intercept codec:
   - Тесты кодирования различных типов данных
   - Тесты обработки больших тел запросов/ответов

### Ожидаемый прирост покрытия: ~1.2-1.5%

---

## Агент 4: Process Detection & Scripting Executors

### Цель: Улучшить покрытие детекции процессов и executors для скриптов

### Файлы для работы:
- `internal/features/process/infrastructure/detector/detector_darwin.go`
  - `DetectByPort` (50.0%) - добавить тесты для детекции процессов (если возможно без реальных процессов)

- `internal/features/scripting/infrastructure/dart/executor.go`
  - `NewDartExecutor` (62.5%) - добавить тесты для создания executor
  - `Execute` (13.8%) - добавить тесты для выполнения Dart скриптов
  - `Close` (60.0%) - добавить тесты для закрытия executor

- `internal/features/scripting/infrastructure/extism/executor.go`
  - `Execute` (58.1%) - добавить тесты для выполнения Extism скриптов

- `internal/features/scripting/infrastructure/cache/filesystem_cache.go`
  - `ClearAll` (69.2%) - добавить тесты для очистки кэша
  - `GetCacheSize` (68.8%) - добавить тесты для получения размера кэша

- `internal/features/proxy/infrastructure/persistence/repo.go`
  - `Load` (73.3%) - добавить тесты для загрузки proxy конфигурации

### Задачи:
1. Добавить тесты для process detector (если возможно):
   - Тесты детекции процессов по порту (mock реализация)
   - Тесты обработки ошибок детекции

2. Добавить тесты для executors:
   - Тесты создания Dart executor с различными конфигурациями
   - Тесты выполнения Dart скриптов с различными входными данными
   - Тесты закрытия executor и очистки ресурсов
   - Тесты выполнения Extism скриптов

3. Добавить тесты для cache:
   - Тесты очистки кэша с различными состояниями
   - Тесты получения размера кэша с различными данными

4. Добавить тесты для proxy persistence:
   - Тесты загрузки конфигурации proxy
   - Тесты обработки ошибок загрузки

### Ожидаемый прирост покрытия: ~1.1-1.4%

---

## Общие рекомендации для всех агентов:

1. **Следовать принципам SOLID** - создавать качественные тесты, которые реально тестируют функциональность
2. **Избегать платформо-зависимых тестов** - пропускать тесты, которые могут не работать на всех платформах
3. **Использовать моки** - создавать mock реализации для зависимостей
4. **Тестировать error paths** - покрывать не только happy path, но и обработку ошибок
5. **Запускать тесты** - после добавления тестов запускать `go test` и проверять покрытие
6. **Форматировать код** - всегда запускать `go fmt ./...` после изменений

## Команды для проверки прогресса:

```bash
# Запустить тесты и получить покрытие
go test -short -covermode=atomic -coverprofile=coverage.out $(go list ./... | grep -v /internal/integration | grep -v /internal/e2e)

# Проверить общее покрытие
go tool cover -func=coverage.out | grep "total:" | awk '{print $3}'

# Проверить покрытие конкретного файла
go tool cover -func=coverage.out | grep "script_handlers.go"

# Форматировать код
go fmt ./...
```

## Приоритеты:

1. **Высокий приоритет:** HTTP API handlers (Агент 1) - наибольший потенциал для улучшения
2. **Высокий приоритет:** Scripting Infrastructure (Агент 2) - много файлов с низким покрытием
3. **Средний приоритет:** Proxy Handlers (Агент 3) - сложные интеграционные тесты
4. **Низкий приоритет:** Process Detection (Агент 4) - может быть сложно тестировать без реальных процессов

## Критерии завершения:

- Общее покрытие достигло 71% или выше
- Все добавленные тесты проходят успешно
- Код отформатирован (`go fmt`)
- Нет новых linter ошибок
