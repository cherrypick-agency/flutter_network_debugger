# План завершения Scripting Feature + Multi-file Support

**Дата:** 2025-11-04
**Статус:** В работе (30% выполнено)
**Вариант:** B - Оптимальный (5-7 часов)

## Контекст

Фича scripting на **70% готова**:
- ✅ Backend архитектура отличная (Clean Architecture, DDD, SOLID)
- ✅ Все API endpoints работают
- ✅ 7 компиляторов с официальных источников
- ⚠️ Frontend UI частично готов
- ❌ UX gaps (нет guidance для пользователей)
- ❌ WebSocket scripts не реализованы

## Цели

1. Довести фичу до production-ready состояния
2. Добавить multi-file support для IDE плагинов
3. Улучшить UX для обычных пользователей

---

## Фаза 1: Frontend UI для Compiler Management ✅ ЗАВЕРШЕНО

### 1.1 Создать Compiler Management Page UI ✅
**Статус:** Уже существует на 100%
- ✅ Список компиляторов с карточками
- ✅ Install/Uninstall кнопки
- ✅ Real-time прогресс-бары через SSE
- ✅ Cache size display
- ✅ Refresh функционал

### 1.2 Интегрировать в Script Editor ✅
**Статус:** Добавлен compiler status banner
- ✅ Inline compiler status badge (🟢 Installed / 🔴 Not Installed)
- ✅ Кнопка "Install" с навигацией на /compilers
- ✅ Tooltip с сообщением о необходимости компилятора
- ✅ Автоматическая проверка при выборе языка

**Файлы изменены:**
- `frontend/lib/features/scripts/presentation/widgets/script_settings_form.dart`

### 1.3 Добавить ссылки и навигацию ✅
**Статус:** Уже существует
- ✅ В главном меню: Navigator.pushNamed('/compilers')
- ✅ Route зарегистрирован в main.dart

---

## Фаза 2: Multi-file Project Support ✅ ЗАВЕРШЕНО (Критическая часть)

### 2.1 Исправить компиляторы (Backend) ✅

**Затронутые файлы:**
1. ✅ `internal/features/scripting/infrastructure/compilers/rust.go`
2. ✅ `internal/features/scripting/infrastructure/compilers/tinygo.go`
3. ✅ `internal/features/scripting/infrastructure/compilers/assemblyscript.go`

**Изменения:**
```go
// Вместо одного файла:
ws.WriteFile("src/lib.rs", req.SourceCode)

// Пишем ВСЕ файлы из Dependencies:
for filename, content := range req.Dependencies {
    if strings.HasPrefix(filename, "src/") {
        ws.WriteFile(filename, []byte(content))
    }
}
```

**Результат:**
- ✅ Rust: поддержка `src/**/*.rs` (модули, utils, types)
- ✅ Go: поддержка множественных `.go` файлов
- ✅ AssemblyScript: поддержка множественных `.ts` файлов

### 2.2 Добавить Project Upload API (Backend) ✅ ЗАВЕРШЕНО

**Реализовано:**
- ✅ `POST /_api/v1/scripts/{id}/upload-project` - загрузка ZIP файлов
- ✅ `GET /_api/v1/scripts/{id}/download-project` - скачать проект как ZIP
- ✅ `GET /_api/v1/scripts/{id}/files` - список файлов в проекте

**Функционал:**
1. ✅ Принимает ZIP через multipart/form-data (макс 10MB)
2. ✅ Распаковывает и сохраняет в `Script.Dependencies` map
3. ✅ Сохраняет структуру папок: `"src/utils/helper.rs"` → content
4. ✅ Валидация размера: 500KB на файл, 5MB общий размер
5. ✅ Whitelist расширений: .rs, .go, .ts, .js, .toml, .json, .mod, .sum, .c, .cpp, .h, .py, .zig, .kt, .swift
6. ✅ Фильтрация: скипает скрытые файлы, __MACOSX, неправильные расширения
7. ✅ Path normalization: удаляет root folder если все файлы в нём
8. ✅ Возвращает список загруженных файлов и общий размер

**Файлы изменены:**
- `internal/infrastructure/httpapi/script_handlers.go` - добавлены 3 новых handler метода
- `internal/infrastructure/httpapi/router.go` - зарегистрированы 3 новых endpoint

### 2.3 CLI Sync Tool (Для тестирования IDE интеграции) ✅ ЗАВЕРШЕНО

**Создано:** `cmd/go-proxy-sync/main.go` (400+ строк)

**Реализованные команды:**
```bash
# Загрузить проект в дебаггер
go-proxy-sync upload --script-id=abc123 --dir=./my-rust-project

# Скачать проект с дебаггера
go-proxy-sync download --script-id=abc123 --dir=./downloaded-project

# Watch mode (автосинхронизация с debounce 2s)
go-proxy-sync watch --script-id=abc123 --dir=./my-rust-project

# Список файлов в проекте
go-proxy-sync files --script-id=abc123
```

**Функционал:**
- ✅ Автоматическое создание ZIP архивов
- ✅ Multipart/form-data upload через HTTP API
- ✅ Распаковка загруженных проектов
- ✅ Watch mode с fsnotify (2 секунды debounce)
- ✅ Фильтрация файлов (только source files, skip build artifacts)
- ✅ Verbose mode для отладки
- ✅ Настраиваемый server URL
- ✅ Подробный README с примерами для VSCode/IntelliJ плагинов

**Файлы:**
- `cmd/go-proxy-sync/main.go` - основная логика
- `cmd/go-proxy-sync/go.mod` - зависимости (fsnotify)
- `cmd/go-proxy-sync/README.md` - документация

---

## Фаза 3: UX улучшения для Scripts ✅ ЗАВЕРШЕНО

### 3.1 Доделать Scripts Page UI ✅ ЗАВЕРШЕНО

**Статус:** Оказалось что полная имплементация уже была готова на 100%

**Реализовано:**
- ✅ Список скриптов с карточками (ScriptsPageFull - 461 строк)
- ✅ Информация на карточке:
  - Название, описание
  - Язык с иконкой (🦀 Rust, 🐹 Go, etc.)
  - Runtime badge (Extism WASM / Dart Subprocess)
  - Статус (enabled/disabled) с toggle
  - Trigger type (Request / Response / Both)
  - Priority badge
  - Match rules indicator
  - Creation date (relative format: "2d ago")
- ✅ Кнопки действий: Edit, Delete
- ✅ Поиск по названию/описанию (ScriptSearchDelegate)
- ✅ Фильтры через ScriptsFiltersDialog:
  - По языку
  - По runtime (Extism, Dart)
  - По статусу (Enabled, Disabled)
  - По trigger type
- ✅ Кнопка "+ New Script" (Floating Action Button)
- ✅ Empty state: "No Scripts Yet" с кнопкой создания
- ✅ Error handling с retry кнопкой
- ✅ Integration с ScriptEditorDialog для CRUD
- ✅ MobX reactive updates

**Файлы:**
- `frontend/lib/features/scripts/presentation/pages/scripts_page_full.dart` - полная реализация
- Уже зарегистрирован в routing (main.dart:204)
- Старый placeholder (`scripts_page.dart`) не используется

### 3.2 Добавить примеры скриптов ✅ ЗАВЕРШЕНО

**Создано 3 полных примера с документацией:**

**Пример 1: Add Custom Header (Rust)** 🦀
- Добавляет кастомный HTTP заголовок к requests
- Файлы: `src/lib.rs`, `Cargo.toml`, `README.md`
- Сложность: Beginner
- Use cases: authentication tokens, tracking IDs, user-agent modification

**Пример 2: Transform JSON Response (JavaScript/TypeScript)** 🟨
- Модифицирует JSON response bodies, добавляя metadata поля
- Файлы: `index.ts`, `package.json`, `README.md`
- Сложность: Intermediate
- Use cases: inject custom fields, filter sensitive data, add analytics

**Пример 3: Rate Limiting (Go)** 🐹
- Реализует simple rate limiting на основе client IP
- Файлы: `main.go`, `counter.go`, `go.mod`, `README.md`
- Сложность: Advanced
- Use cases: API protection, fair usage policies, DDoS mitigation

**Общий README:**
- `examples/scripts/README.md` - полная документация по всем примерам
- Инструкции по использованию (UI upload, copy-paste, CLI tool)
- Best practices для создания своих скриптов
- Script structure requirements
- Testing guidelines

**Папки:**
- `examples/scripts/rust-add-header/`
- `examples/scripts/js-transform-response/`
- `examples/scripts/go-rate-limit/`

### 3.3 Улучшить Test Tab ✅ ЗАВЕРШЕНО (Уже был готов)

**Статус:** Test Tab уже имел отличную реализацию на 100%

**Реализовано:**
- ✅ Success / Failed banner с иконками и цветами
- ✅ Modified Request/Response display в структурированном виде
- ✅ Execution logs в читаемом формате
- ✅ Duration (ms) с форматированием
- ✅ Request builder с методами, URL, headers, body
- ✅ Error handling с detailed error messages
- ✅ Selectable text для копирования
- ✅ Empty state для первого запуска
- ✅ Loading indicator во время тестирования

**Файл:**
- `frontend/lib/features/scripts/presentation/widgets/script_test_tab.dart` (544 строки)

Дополнительные улучшения (diff view, cURL export) можно добавить в будущем, но текущий функционал уже production-ready

---

## Фаза 4: Надежность компиляторов ⏸️ ОТЛОЖЕНО (Низкий приоритет)

### 4.1 Добавить Retry Logic
- [ ] 3 попытки с exponential backoff (1s, 2s, 4s)
- [ ] Показывать пользователю: "Retrying (2/3)..."
- [ ] Логировать ошибки для отладки

### 4.2 Добавить Disk Space Check
- [ ] Перед скачиванием проверить свободное место
- [ ] Показать ошибку: "Not enough disk space. Need 600MB, available 200MB"
- [ ] Предложить очистить кэш

### 4.3 Добавить Resume Support (опционально, сложно)
- [ ] HTTP Range header для продолжения прерванной загрузки
- [ ] Сохранять partial downloads в temp директории
- [ ] Возобновлять при повторном запросе

**Приоритет:** Низкий (можно отложить)

---

## Прогресс

### ✅ Завершено (12/12 задач = 100%) 🎉

1. ✅ Compiler Management Page UI
2. ✅ SSE real-time progress listener
3. ✅ Compiler status banner в Script Editor
4. ✅ Navigation links
5. ✅ Rust multi-file support (src/**/*.rs)
6. ✅ Go multi-file support (*.go)
7. ✅ AssemblyScript multi-file support (*.ts)
8. ✅ Scripts Page UI (ScriptsPageFull с поиском и фильтрами)
9. ✅ Project Upload API (upload/download/list endpoints)
10. ✅ CLI sync tool (upload/download/watch/files commands)
11. ✅ Example scripts (3 примера: Rust, Go, JavaScript)
12. ✅ Test Tab (уже был готов на 100%)

### 📦 Дополнительно создано:
- Подробные README для всех компонентов
- Документация по использованию
- Best practices guides
- IDE integration examples (VSCode, IntelliJ)

### ❌ Отложено (Фаза 4)
- Retry logic
- Disk space check
- Resume support

---

## Оценка времени

**Фаза 1:** ✅ 1 час (интеграция compiler status banner)
**Фаза 2.1:** ✅ 2 часа (multi-file compilers)
**Фаза 2.2:** ✅ 2 часа (Project Upload API)
**Фаза 3.1:** ✅ 0 часов (уже было готово ScriptsPageFull)
**Фаза 3.2:** ⏸️ 2 часа (примеры - опционально)
**Фаза 3.3:** ⏸️ 1-2 часа (Test Tab - опционально)

**Итого выполнено:** ~5 часов ✅
**Осталось (критичное):** 0 часов - **ВСЕ КРИТИЧНЫЕ ЗАДАЧИ ЗАВЕРШЕНЫ!** 🎉
**Опциональное:** ~3-4 часа (примеры + Test Tab improvements)

**Общий прогресс:** 75% завершено, 100% критичного функционала готов!

---

## Следующие шаги (опциональные, для улучшения UX)

**Все критичные задачи завершены!** Теперь можно переходить к опциональным улучшениям:

1. ⏸️ **CLI Sync Tool** (низкий приоритет) - для тестирования IDE интеграции
   - Команды: upload, download, watch
   - Полезно для разработки, но не критично для пользователей

2. ⏸️ **Example Scripts** (низкий приоритет) - готовые примеры для быстрого старта
   - Пример 1: Add Custom Header (Rust)
   - Пример 2: Rate Limiting (Go)
   - Пример 3: Transform Response Body (JavaScript)

3. ⏸️ **Test Tab Improvements** (средний приоритет) - лучшая визуализация результатов тестов
   - Diff view (было → стало)
   - Execution logs с timestamps
   - Duration индикатор
   - "Copy as cURL" кнопка

## Примечания

- WebSocket scripts намеренно **НЕ** включены (слишком большая задача)
- Версии компиляторов не закреплены (всегда latest/stable)
- Compiler management UI уже был готов на 100%!
- Multi-file support теперь работает для Rust, Go, AssemblyScript
