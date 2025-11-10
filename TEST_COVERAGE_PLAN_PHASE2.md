# План покрытия тестами для следующих 4 агентов (Фаза 2)

**Текущее покрытие (после фазы 1):** ~52% (ожидаемое)  
**Целевое покрытие:** ~58-60%  
**Необходимо добавить:** ~6-8%

## Принципы разделения

1. **Независимость**: Каждый агент работает с разными пакетами/файлами
2. **Баланс**: Примерно равная сложность работы для каждого агента
3. **Изоляция**: Нет пересечений файлов между агентами
4. **Маркировка**: Все новые тесты помечаются комментарием `// Composer 1.`

---

## АГЕНТ 5: HTTP API Handlers - Scripting & Compilation

### Целевой пакет
`internal/infrastructure/httpapi`

### Текущее покрытие
55.0%

### Задачи

#### 1. Script Handlers - CRUD операции
**Файл:** `script_handlers.go`  
**Методы для покрытия:**
- `CreateScript()` - создание скрипта через API
- `GetScript()` - получение скрипта по ID
- `ListScripts()` - список скриптов с фильтрами
- `UpdateScript()` - обновление скрипта
- `DeleteScript()` - удаление скрипта
- `ToggleScript()` - включение/выключение скрипта
- `TestScript()` - тестирование скрипта
- `ListExamples()` - список примеров скриптов
- `UploadProject()` - загрузка проекта (multi-file)
- `DownloadProject()` - скачивание проекта
- `ListProjectFiles()` - список файлов проекта
- `ExportScriptAsZip()` - экспорт скрипта как ZIP
- `ImportScriptFromZip()` - импорт скрипта из ZIP

**Тестовый файл:** `script_handlers_test.go` (создать)

**Что тестировать:**
- Успешные операции (200 OK)
- Валидация входных данных (400 Bad Request)
- Не найденные ресурсы (404 Not Found)
- Ошибки сервиса (500 Internal Server Error)
- JSON сериализация/десериализация
- Multipart form data для upload
- ZIP файлы для export/import

#### 2. Compilation Handlers
**Файл:** `compilation_handlers.go`  
**Методы для покрытия:**
- `CompileScript()` - компиляция исходного кода
- `GetCompilationStatus()` - статус компиляции
- `GetCompilationLogs()` - логи компиляции

**Тестовый файл:** `compilation_handlers_test.go` (создать)

**Что тестировать:**
- Успешная компиляция
- Ошибки компиляции
- Timeout при компиляции
- Проверка статуса компиляции
- Получение логов

#### 3. Compiler Installation Handlers
**Файл:** `compiler_installation_handlers.go`  
**Методы для покрытия:**
- `InstallCompiler()` - установка компилятора
- `GetCompilerStatus()` - статус компилятора
- `ListAvailableCompilers()` - список доступных компиляторов
- `UninstallCompiler()` - удаление компилятора

**Тестовый файл:** `compiler_installation_handlers_test.go` (создать)

**Что тестировать:**
- Установка компилятора
- Проверка статуса
- Список компиляторов
- Ошибки установки
- Progress callbacks для установки

### Ожидаемый результат
Покрытие httpapi: ~65-70%

---

## АГЕНТ 6: HTTP API Handlers - Mapping, Tags, Performance, Process

### Целевой пакет
`internal/infrastructure/httpapi`

### Текущее покрытие
55.0% (будет улучшено после агента 5)

### Задачи

#### 1. Mapping Handlers
**Файл:** `mapping_handlers.go`  
**Методы для покрытия:**
- `CreateMapRule()` - создание правила маппинга
- `GetMapRule()` - получение правила
- `ListMapRules()` - список правил
- `UpdateMapRule()` - обновление правила
- `DeleteMapRule()` - удаление правила
- `TestMapRule()` - тестирование правила

**Тестовый файл:** `mapping_handlers_test.go` (создать/дополнить)

**Что тестировать:**
- CRUD операции
- Валидация правил маппинга
- Тестирование правил на примерах
- Обработка ошибок

#### 2. Tags Handlers
**Файл:** `tags_handlers.go`  
**Методы для покрытия:**
- `ListPredefinedTags()` - список предопределенных тегов
- `CreatePredefinedTag()` - создание тега
- `DeletePredefinedTag()` - удаление тега
- `GetSessionTags()` - теги сессии
- `AddTagToSession()` - добавление тега
- `RemoveTagFromSession()` - удаление тега
- `BulkAddTags()` - массовое добавление
- `BulkRemoveTags()` - массовое удаление
- `GetSessionAnnotations()` - аннотации сессии
- `UpsertAnnotation()` - создание/обновление аннотации
- `DeleteAnnotation()` - удаление аннотации

**Тестовый файл:** `tags_handlers_test.go` (создать)

**Что тестировать:**
- Все CRUD операции
- Bulk операции
- Валидация данных
- Обработка ошибок

#### 3. Performance Handlers
**Файл:** `performance_handlers.go`  
**Методы для покрытия:**
- `GetPerformanceOverview()` - обзор производительности
- `GetPerformanceMetrics()` - метрики производительности
- `GetSlowRequests()` - медленные запросы
- `GetErrorRate()` - частота ошибок

**Тестовый файл:** `performance_handlers_test.go` (создать)

**Что тестировать:**
- Получение метрик
- Фильтрация по времени
- Агрегация данных
- Обработка пустых данных

#### 4. Process Handlers
**Файл:** `process_handlers.go`  
**Методы для покрытия:**
- `GetProcessConfig()` - получение конфигурации
- `UpdateProcessConfig()` - обновление конфигурации
- `GetHelperStatus()` - статус helper tool
- `InstallHelper()` - установка helper
- `DetectProcess()` - детекция процесса

**Тестовый файл:** `process_handlers_test.go` (создать)

**Что тестировать:**
- CRUD конфигурации
- Статус helper
- Установка helper
- Детекция процессов

### Ожидаемый результат
Покрытие httpapi: ~75-80%

---

## АГЕНТ 7: Process Helper Infrastructure

### Целевой пакет
`internal/features/process/infrastructure/helper`

### Текущее покрытие
0.0%

### Задачи

#### 1. Helper Client - IPC коммуникация
**Файл:** `client.go`  
**Методы для покрытия:**
- `NewClient()` - создание клиента
- `IsRunning()` - проверка доступности daemon
- `DetectProcess()` - детекция процесса
- `ExtractIcon()` - извлечение иконки
- `Ping()` - health check
- `Close()` - закрытие соединения
- `sendRequest()` - отправка запроса (internal)
- `connect()` - подключение к socket (internal)
- `reconnect()` - переподключение (internal)
- `generateID()` - генерация ID (internal)

**Тестовый файл:** `client_test.go` (создать)

**Что тестировать:**
- Успешные IPC запросы с моками Unix socket
- Обработка ошибок подключения
- Timeout при запросах
- Reconnect при ошибках
- JSON сериализация/десериализация
- Thread safety (concurrent requests)
- Закрытие соединения

#### 2. Helper Client - Platform-specific
**Файл:** `client_darwin.go`  
**Методы для покрытия:**
- Platform-specific логика (если есть)

**Тестовый файл:** `client_darwin_test.go` (создать с build tag)

#### 3. Helper Installer - Interface implementation
**Файл:** `installer_darwin.go`  
**Методы для покрытия:**
- `IsInstalled()` - проверка установки
- `Install()` - установка helper
- `Uninstall()` - удаление helper
- `GetVersion()` - получение версии

**Тестовый файл:** `installer_darwin_test.go` (создать с build tag)

**Что тестировать:**
- Проверка установки (с моками файловой системы)
- Установка helper (с моками sudo/правами)
- Удаление helper
- Получение версии
- Обработка ошибок прав доступа

#### 4. Helper Installer - Other platforms
**Файл:** `installer_other.go`  
**Методы для покрытия:**
- Stub implementation для не-Darwin платформ

**Тестовый файл:** `installer_other_test.go` (создать)

**Что тестировать:**
- Stub методы возвращают правильные значения
- Ошибки при попытке установки на неподдерживаемой платформе

### Ожидаемый результат
Покрытие helper: ~70-80%

---

## АГЕНТ 8: Mapping Runtime + Performance Usecase + Edge Cases

### Целевые пакеты
1. `internal/features/mapping/runtime`
2. `internal/features/performance/usecase`
3. Другие утилиты и edge cases

### Текущее покрытие
- mapping/runtime: 81.3%
- performance/usecase: 78.2%

### Задачи

#### 1. Mapping Runtime - Edge Cases
**Файл:** `manager.go`  
**Методы для покрытия:**
- `ApplyRules()` - edge cases:
  - Множественные правила с разными приоритетами
  - Правила с одинаковым приоритетом
  - Правила с regex patterns
  - Правила с wildcard patterns
  - Невалидные правила
  - Правила без совпадений
- `MatchRule()` - edge cases:
  - Все типы patterns (exact, prefix, wildcard, regex)
  - Invalid regex patterns
  - Empty patterns
  - Case sensitivity

**Тестовый файл:** `manager_test.go` (дополнить)

**Что тестировать:**
- Приоритизация правил
- Обработка невалидных правил
- Regex edge cases
- Wildcard edge cases
- Performance с большим количеством правил

#### 2. Performance Usecase - Edge Cases
**Файл:** `service.go`  
**Методы для покрытия:**
- `GetOverview()` - edge cases:
  - Пустые данные
  - Данные за разные периоды
  - Агрегация больших объемов данных
- `GetSlowRequests()` - edge cases:
  - Различные пороги медленности
  - Сортировка
  - Лимиты результатов
- `GetErrorRate()` - edge cases:
  - Различные временные окна
  - Различные типы ошибок
  - Агрегация по статус кодам

**Тестовый файл:** `service_test.go` (дополнить)

**Что тестировать:**
- Edge cases для всех методов
- Обработка пустых данных
- Граничные значения
- Производительность с большими данными

#### 3. HTTP API - Intercept Handlers
**Файл:** `intercept_handlers.go`  
**Методы для покрытия:**
- `CreateIntercept()` - создание перехвата
- `GetIntercept()` - получение перехвата
- `ListIntercepts()` - список перехватов
- `UpdateIntercept()` - обновление перехвата
- `DeleteIntercept()` - удаление перехвата
- `EnableIntercept()` - включение перехвата
- `DisableIntercept()` - выключение перехвата

**Тестовый файл:** `intercept_handlers_test.go` (дополнить)

**Что тестировать:**
- CRUD операции
- Включение/выключение
- Валидация правил перехвата
- Обработка ошибок

#### 4. HTTP API - MITM Handlers Edge Cases
**Файл:** `mitm_handlers.go`  
**Методы для покрытия:**
- Edge cases для существующих методов
- Обработка ошибок сертификатов
- Обработка различных типов запросов

**Тестовый файл:** `mitm_handlers_test.go` (дополнить)

### Ожидаемый результат
Покрытие mapping/runtime: ~90%  
Покрытие performance/usecase: ~85-90%  
Общее покрытие: ~58-60%

---

## Общие инструкции для всех агентов (Фаза 2)

### Формат тестов
```go
// Composer 1.
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

### Запуск тестов
```bash
# Для конкретного пакета
go test ./internal/infrastructure/httpapi/... -v

# С покрытием
go test -coverprofile=coverage.out ./internal/infrastructure/httpapi/...
go tool cover -func=coverage.out | grep total
```

### Моки для HTTP handlers
Используйте `httptest.NewRecorder()` и `httptest.NewRequest()`:
```go
rr := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/_api/v1/scripts", nil)
handler.ServeHTTP(rr, req)
```

### Моки для IPC Client
Создайте mock Unix socket или используйте `net.Pipe()`:
```go
clientConn, serverConn := net.Pipe()
// Используйте serverConn для эмуляции сервера
```

### Проверка перед коммитом
1. Все тесты проходят: `go test ./...`
2. Форматирование: `go fmt ./...`
3. Покрытие выросло на целевом пакете
4. Нет конфликтов с другими агентами

---

## Метрики успеха (Фаза 2)

После завершения работы всех агентов ожидается:
- **Общее покрытие:** ≥58-60%
- **httpapi:** ~75-80% (было 55.0%)
- **helper:** ~70-80% (было 0.0%)
- **mapping/runtime:** ~90% (было 81.3%)
- **performance/usecase:** ~85-90% (было 78.2%)

---

## Порядок работы агентов (Фаза 2)

Агенты могут работать **параллельно**, так как работают с разными файлами. Рекомендуемый порядок:

1. **Агент 5** → **Агент 6** (последовательно, так как оба работают с httpapi)
2. **Агент 7** → **Агент 8** (параллельно с агентами 5-6)

Или все **параллельно** (если уверены в изоляции файлов).

---

## Контрольные точки (Фаза 2)

После завершения каждого агента:
1. Запустить тесты: `go test ./...`
2. Проверить покрытие целевого пакета
3. Убедиться что общее покрытие выросло
4. Зафиксировать результат
5. Проверить что нет регрессий в существующих тестах

---

## Дополнительные заметки

### HTTP API Handlers
- Используйте моки для сервисов (ScriptService, CompilationService, etc.)
- Тестируйте все HTTP методы (GET, POST, PUT, DELETE, PATCH)
- Проверяйте статус коды ответов
- Тестируйте валидацию входных данных
- Тестируйте обработку ошибок

### Process Helper
- Используйте `net.Pipe()` для эмуляции Unix socket
- Тестируйте thread safety для concurrent requests
- Тестируйте reconnect логику
- Тестируйте timeout handling

### Edge Cases
- Граничные значения
- Пустые данные
- Невалидные данные
- Большие объемы данных
- Concurrent access

