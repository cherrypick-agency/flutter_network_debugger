# План покрытия тестами для 4 независимых агентов

**Текущее покрытие:** 46.7%  
**Целевое покрытие:** 52%  
**Необходимо добавить:** ~5.3%

## Принципы разделения

1. **Независимость**: Каждый агент работает с разными пакетами/файлами
2. **Баланс**: Примерно равная сложность работы для каждого агента
3. **Изоляция**: Нет пересечений файлов между агентами
4. **Маркировка**: Все новые тесты помечаются комментарием `// Composer 1.`

---

## АГЕНТ 1: Downloaders Infrastructure

### Целевой пакет
`internal/features/scripting/infrastructure/downloaders`

### Текущее покрытие
8.4%

### Задачи

#### 1. BaseDownloader - DownloadFile и attemptDownload
**Файл:** `base.go`  
**Методы для покрытия:**
- `DownloadFile()` - основной метод загрузки с retry логикой
- `attemptDownload()` - внутренний метод с resume support
- Тестировать:
  - Успешная загрузка
  - Retry при сетевых ошибках
  - Resume при прерывании
  - Cancellation через context
  - Disk space checking
  - Progress callback вызовы
  - Различные HTTP статус коды

**Тестовый файл:** `base_test.go` (дополнить)

#### 2. Специфичные downloaders
**Файлы для покрытия:**
- `tinygo_downloader.go` - методы Download, Install
- `rust_downloader.go` - методы Download, Install  
- `assemblyscript_downloader.go` - методы Download, Install
- `swift_downloader.go` - методы Download, Install
- `kotlin_downloader.go` - методы Download, Install
- `zig_downloader.go` - методы Download, Install
- `wasisdk_downloader.go` - методы Download, Install

**Тестовые файлы:** Создать `*_downloader_test.go` для каждого

**Что тестировать:**
- Конструкторы (New*Downloader)
- Методы Download с моками HTTP клиента
- Методы Install с моками Extract методов
- Обработка ошибок
- Проверка checksums

### Ожидаемый результат
Покрытие downloaders: ~40-50%

---

## АГЕНТ 2: Compilers Infrastructure

### Целевой пакет
`internal/features/scripting/infrastructure/compilers`

### Текущее покрытие
16.9%

### Задачи

#### 1. TinyGoCompiler - методы Compile
**Файл:** `tinygo.go`  
**Методы для покрытия:**
- `Compile()` - основной метод компиляции
- Тестировать:
  - Успешная компиляция
  - Обработка зависимостей (go.mod)
  - Ошибки компиляции
  - Timeout через context
  - Работа с workspace

**Тестовый файл:** `tinygo_test.go` (создать)

#### 2. RustCompiler - методы Compile
**Файл:** `rust.go`  
**Методы для покрытия:**
- `Compile()` - основной метод компиляции
- Тестировать:
  - Успешная компиляция
  - Cargo.toml генерация
  - Ошибки компиляции
  - Environment variables (RUSTUP_HOME, CARGO_HOME)

**Тестовый файл:** `rust_test.go` (создать)

#### 3. Другие компиляторы - базовые методы
**Файлы:**
- `assemblyscript.go` - Compile, IsAvailable
- `c_cpp.go` - Compile, IsAvailable
- `python.go` - Compile, IsAvailable
- `kotlin.go` - Compile, IsAvailable, ValidateDependencies
- `swift.go` - Compile, IsAvailable
- `zig.go` - Compile, IsAvailable

**Тестовые файлы:** Создать `*_test.go` для каждого

**Что тестировать:**
- Методы Compile с моками workspace и exec.Command
- Методы IsAvailable
- ValidateDependencies где применимо
- Обработка ошибок

### Ожидаемый результат
Покрытие compilers: ~40-50%

---

## АГЕНТ 3: Dart Executor + Process Detector

### Целевые пакеты
1. `internal/features/scripting/infrastructure/dart`
2. `internal/features/process/infrastructure/detector`

### Текущее покрытие
- dart: 27.4%
- detector: 5.6%

### Задачи

#### 1. Dart Executor - Execute и ProcessPool
**Файл:** `executor.go`  
**Методы для покрытия:**
- `Execute()` - выполнение скриптов через JSON-RPC
- `ProcessPool.Get()` - получение процесса из пула
- `ProcessPool.Release()` - возврат процесса в пул
- `ProcessPool.Close()` - закрытие пула
- Тестировать:
  - Успешное выполнение скрипта
  - Обработка JSON-RPC запросов/ответов
  - Timeout при выполнении
  - Ошибки выполнения
  - Pool exhaustion
  - Process reuse
  - Graceful shutdown

**Тестовый файл:** `executor_test.go` (дополнить)

#### 2. Process Detector - методы детекции
**Файл:** `detector.go`  
**Методы для покрытия:**
- `DetectByPort()` - детекция по порту
- `DetectByPID()` - детекция по PID
- `RequiresPrivileges()` - проверка привилегий
- Тестировать:
  - Успешная детекция
  - Обработка ошибок
  - Различные платформы (unix/windows)
  - Mock gopsutil адаптера

**Тестовый файл:** `detector_test.go` (дополнить)

#### 3. Platform-specific detectors
**Файлы:**
- `detector_unix.go` - Unix-специфичная логика
- `detector_windows.go` - Windows-специфичная логика

**Тестовые файлы:** Создать `*_test.go` с build tags

### Ожидаемый результат
Покрытие dart: ~50-60%  
Покрытие detector: ~40-50%

---

## АГЕНТ 4: Icon Extractor + Extism + Scripting Usecase Edge Cases

### Целевые пакеты
1. `internal/features/process/infrastructure/icon`
2. `internal/features/scripting/infrastructure/extism`
3. `internal/features/scripting/usecase` (edge cases)

### Текущее покрытие
- icon: 35.7%
- extism: 51.3%
- usecase: 40.2%

### Задачи

#### 1. Icon Extractor - методы извлечения
**Файлы:**
- `extractor_darwin.go` - macOS извлечение
- `extractor_linux.go` - Linux извлечение
- `extractor_windows.go` - Windows извлечение
- `extractor_stub.go` - Stub для неподдерживаемых платформ
- `extractor.go` - Factory функция

**Методы для покрытия:**
- `ExtractByPID()` - извлечение по PID
- `ExtractByPath()` - извлечение по пути
- `Extract()` - общий метод
- `FindAppBundle()` (macOS) - поиск app bundle

**Тестовые файлы:** Дополнить существующие `*_test.go`

**Что тестировать:**
- Успешное извлечение иконок
- Обработка ошибок (процесс не найден, файл не найден)
- Различные форматы (PNG, ICO)
- Platform-specific логика

#### 2. Extism Executor - методы выполнения
**Файл:** `executor.go`  
**Методы для покрытия:**
- `Execute()` - выполнение WASM плагинов
- `Validate()` - валидация WASM
- `RemovePlugin()` - удаление плагина из кеша
- Тестировать:
  - Успешное выполнение плагина
  - Обработка ошибок выполнения
  - Plugin pool management
  - WASM validation
  - Timeout handling

**Тестовый файл:** `executor_test.go` (создать/дополнить)

#### 3. Scripting Usecase - Edge Cases
**Файл:** `service.go`  
**Методы для покрытия:**
- `ExecuteForRequest()` - edge cases:
  - Множественные скрипты с разными runtime
  - Скрипты с ошибками выполнения (не должны ломать цепочку)
  - Скрипты модифицирующие request последовательно
- `ExecuteForResponse()` - edge cases:
  - Аналогично для response
- `filterMatchingScripts()` - edge cases:
  - Комбинации match rules (method + path + host)
  - Regex patterns с ошибками
  - Wildcard edge cases

**Тестовый файл:** `service_test.go` (дополнить)

### Ожидаемый результат
Покрытие icon: ~50-60%  
Покрытие extism: ~70-80%  
Покрытие usecase: ~50-60%

---

## Общие инструкции для всех агентов

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
go test ./internal/features/scripting/infrastructure/downloaders/... -v

# С покрытием
go test -coverprofile=coverage.out ./internal/features/scripting/infrastructure/downloaders/...
go tool cover -func=coverage.out | grep total
```

### Проверка перед коммитом
1. Все тесты проходят: `go test ./...`
2. Форматирование: `go fmt ./...`
3. Покрытие выросло на целевом пакете

### Избегайте конфликтов
- Каждый агент работает только со своими файлами
- Не редактируйте файлы других агентов
- Если нужен общий mock - создайте в своем пакете

---

## Метрики успеха

После завершения работы всех агентов ожидается:
- **Общее покрытие:** ≥52%
- **downloaders:** ~40-50% (было 8.4%)
- **compilers:** ~40-50% (было 16.9%)
- **dart:** ~50-60% (было 27.4%)
- **detector:** ~40-50% (было 5.6%)
- **icon:** ~50-60% (было 35.7%)
- **extism:** ~70-80% (было 51.3%)
- **usecase:** ~50-60% (было 40.2%)

---

## Порядок работы агентов

Агенты могут работать **параллельно**, так как работают с разными файлами. Рекомендуемый порядок:

1. **Агент 1** → **Агент 2** → **Агент 3** → **Агент 4** (последовательно)
2. Или все **параллельно** (если уверены в изоляции)

---

## Контрольные точки

После завершения каждого агента:
1. Запустить тесты: `go test ./...`
2. Проверить покрытие целевого пакета
3. Убедиться что общее покрытие выросло
4. Зафиксировать результат

