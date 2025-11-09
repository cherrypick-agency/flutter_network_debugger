# Nice to Have Features для Scripting Feature

**Дата:** 2025-11-04
**Статус:** Документация всех опциональных улучшений
**Текущая версия:** Scripting Feature 100% готов, эти фичи - дополнительные улучшения

---

## Контекст

Scripting Feature полностью завершена и production-ready на 100%. Этот документ содержит список **опциональных улучшений (Nice to Have)**, которые можно добавлять постепенно по мере необходимости.

**Всего фич:** 24
**Общее время:** ~300+ часов
**Категории:** UI/UX, DevEx, Advanced, Performance

---

## Категория А: UI/UX Improvements (5 фич)

### 1. Examples Library ⭐ HIGH PRIORITY

**Описание:**
Встроенная библиотека готовых примеров скриптов с one-click import. Пользователь открывает dialog, выбирает пример (Add Header, Rate Limiting, Transform Response), смотрит preview кода, нажимает "Use This Template" → создается новый скрипт из примера.

**Зачем:**
- Сильно улучшает onboarding для новых пользователей
- Quick start без чтения документации
- Обучение по примерам (Beginner → Intermediate → Advanced)

**Где упоминается:**
- `docs/plans/scripts-ui-implementation.md:276-282`
- `docs/plans/scripts-ui-implementation.md:433`

**Реализация:**
1. Frontend: создать `examples_library_dialog.dart`
2. Categories: Beginner, Intermediate, Advanced
3. Embedded examples из `examples/scripts/` (уже существуют!)
4. Preview: код + README + кнопка "Use This Template"
5. Integration с ScriptsStore: `createFromTemplate(example)`

**Сложность:** Easy
**Время:** 4-6 часов
**Приоритет:** HIGH
**Зависимости:** Нет (примеры уже существуют)

**Файлы:**
- `frontend/lib/features/scripts/presentation/widgets/examples_library_dialog.dart` (новый)
- `frontend/lib/features/scripts/presentation/pages/scripts_page_full.dart` (кнопка "Examples")
- `frontend/lib/features/scripts/application/stores/scripts_store.dart` (`createFromTemplate` action)

---

### 2. Import/Export Scripts 📦 MEDIUM PRIORITY

**Описание:**
Возможность экспортировать скрипты в JSON/ZIP файлы и импортировать их обратно. Для sharing, backup, version control, migration между инстансами.

**Зачем:**
- **Sharing** - поделиться скриптами с коллегами (отправить JSON файл)
- **Backup** - сохранить скрипты в git repo (version control)
- **Migration** - перенести скрипты между инстансами debugger'а
- **Templates** - создать библиотеку готовых скриптов для команды

**Где упоминается:**
- `docs/plans/scripts-ui-implementation.md:272-273`
- `docs/plans/scripts-ui-implementation.md:434`

**Что экспортируется:**
```json
{
  "name": "Add Custom Header",
  "description": "Adds X-Custom-Id header",
  "language": "rust",
  "runtime": "extism",
  "sourceCode": "...",
  "dependencies": {
    "Cargo.toml": "...",
    "src/lib.rs": "..."
  },
  "matchRules": { ... },
  "config": { ... },
  "priority": 100,
  "enabled": true
}
```

**Реализация:**

**Backend:**
1. `GET /_api/v1/scripts/{id}/export` - экспорт как JSON
2. `POST /_api/v1/scripts/import` - импорт из JSON
3. Использовать существующий Project Download API для ZIP экспорта

**Frontend:**
1. Кнопка "Export" на каждой карточке скрипта → скачать JSON/ZIP
2. Кнопка "Import Script" в AppBar + drag-n-drop zone
3. `script_import_dialog.dart` - preview, validation, confirmation
4. Integration с ScriptsStore: `exportScript()`, `importScript()`

**Сложность:** Medium
**Время:** 6-8 часов
**Приоритет:** Medium
**Зависимости:** Project Upload/Download API (✅ уже реализовано!)

**Файлы:**
- Backend: `internal/infrastructure/httpapi/script_handlers.go`, `router.go`
- Frontend: `scripts_page_full.dart`, `script_import_dialog.dart`, `scripts_store.dart`, `scripts_api_service.dart`

---

### 3. Keyboard Shortcuts ⌨️ LOW PRIORITY

**Описание:**
Hotkeys для частых действий в Script Editor. Power users любят shortcuts.

**Shortcuts:**
- `Ctrl+S` / `Cmd+S` → Save script
- `Ctrl+T` / `Cmd+T` → Run test
- `Esc` → Close dialog (с подтверждением если есть изменения)
- `Ctrl+Shift+C` / `Cmd+Shift+C` → Compile (только для Extism)

**Зачем:**
Продуктивность для power users, меньше кликов мышкой.

**Где упоминается:**
- `docs/plans/scripts-ui-implementation.md:322-325`
- `docs/plans/scripts-ui-implementation.md:436`

**Реализация:**
1. Обернуть `script_editor_dialog.dart` в `Shortcuts` widget
2. Добавить hint tooltips: "Save (Ctrl+S)"
3. Опционально: "Keyboard Shortcuts" help dialog (список всех shortcuts)

**Сложность:** Easy
**Время:** 2-3 часа
**Приоритет:** Low
**Зависимости:** Нет

**Файлы:**
- `script_editor_dialog.dart` - добавить Shortcuts widget
- `script_settings_form.dart`, `script_test_tab.dart` - tooltips

---

### 4. Diff View в Test Tab 🔄 MEDIUM PRIORITY

**Описание:**
Side-by-side diff view для сравнения "Original" vs "Modified" request/response. Визуализация изменений, которые сделал скрипт.

**Зачем:**
- Легко увидеть что изменилось (зеленый = added, красный = removed)
- Debugging - понять почему скрипт работает не так
- Better UX - не нужно вручную сравнивать JSON

**Где упоминается:**
- `docs/plans/scripting-feature-completion.md:309`

**Реализация:**
1. Добавить dependency: `diff_match_patch` или `flutter_diff_view`
2. Создать `diff_viewer_widget.dart`:
   - Split view: Left = Original, Right = Modified
   - Syntax highlighting для JSON/XML
   - Line-by-line diff с colors
   - Toggle: Side-by-side / Inline / Raw
3. Интегрировать в `script_test_tab.dart`

**Сложность:** Medium
**Время:** 4-5 часов
**Приоритет:** Medium
**Зависимости:** Diff library для Flutter

**Файлы:**
- `pubspec.yaml` - добавить dependency
- `diff_viewer_widget.dart` - новый виджет
- `script_test_tab.dart` - интеграция

---

### 5. Copy as cURL 📋 LOW PRIORITY

**Описание:**
Экспорт test request как cURL команду для CLI testing. Кнопка "Copy as cURL" → clipboard.

**Зачем:**
- Протестировать в терминале
- Сохранить в документацию
- Запустить на другом сервере

**Где упоминается:**
- `docs/plans/scripts-ui-implementation.md:274`
- `docs/plans/scripting-feature-completion.md:312`

**Реализация:**
1. Helper метод `_generateCurlCommand(request)`:
   ```
   curl -X POST https://api.example.com/path \
     -H "Content-Type: application/json" \
     -d '{"key": "value"}'
   ```
2. Кнопка "Copy as cURL" в Test Tab
3. `Clipboard.setData()` + SnackBar notification

**Сложность:** Easy
**Время:** 2 часа
**Приоритет:** Low
**Зависимости:** Нет

**Файлы:**
- `script_test_tab.dart` - добавить метод и кнопку

---

## Категория Б: Developer Experience (5 фич)

### 6. WebSocket Compilation Logs 🔌 LOW PRIORITY

**Описание:**
Real-time streaming compilation logs через WebSocket вместо polling. Более быстрая доставка сообщений.

**Зачем:**
- Меньше latency (нет polling interval)
- Двусторонняя связь (можно отменять компиляцию)
- Меньше нагрузки на сервер (persistent connection)

**Где упоминается:**
- `docs/plans/WASM_COMPILATION_COMPLETE.md:260`

**Реализация:**

**Backend:**
1. WebSocket endpoint: `/_api/v1/compilers/{language}/logs/ws`
2. Upgrade HTTP → WebSocket
3. Stream compilation logs: `{"type": "log", "level": "info", "message": "..."}`
4. Heartbeat/ping для keep-alive

**Frontend:**
1. WebSocket client в `compiler_api.dart`
2. Fallback на SSE если WebSocket недоступен
3. Auto-reconnect logic

**Сложность:** Hard
**Время:** 8-10 часов
**Приоритет:** Low (текущий SSE подход работает хорошо)
**Зависимости:** WebSocket infrastructure

**Файлы:**
- Backend: `script_handlers.go`, `router.go`
- Frontend: `compiler_api.dart`, `installation_progress_store.dart`, `compilers_page.dart`

---

### 7. Hot Reload from Filesystem 🔥 LOW PRIORITY

**Описание:**
Watch file system и автоматически reload скрипт при изменениях. Для локальной разработки.

**Зачем:**
- Instant feedback при редактировании в IDE
- Не нужно вручную нажимать "Reload"

**Где упоминается:**
- `docs/plans/scripting-api-implementation.md:1764`

**Реализация:**
1. Backend: file watcher для scripts directory (fsnotify)
2. API endpoint: `POST /_api/v1/scripts/{id}/reload`
3. Frontend: toggle "Auto-reload" в settings

**Сложность:** Medium
**Время:** 6-8 часов
**Приоритет:** Low (уже есть CLI sync tool с watch mode)
**Зависимости:** fsnotify или аналог

---

### 8. Script Debugging Console 🐛 LOW PRIORITY

**Описание:**
Interactive console для пошаговой отладки скриптов. Breakpoints, variable inspection, call stack.

**Зачем:**
- Пошаговая отладка сложных скриптов
- Inspect переменных в runtime
- Понять почему скрипт не работает

**Где упоминается:**
- `docs/plans/scripting-api-implementation.md:1763`

**Реализация:**
1. WASM debugger (Wasmtime Debugger или custom)
2. Source maps для WASM → language mapping
3. Frontend: DevTools-like UI с breakpoints, call stack, variables

**Сложность:** Very Hard
**Время:** 40+ часов
**Приоритет:** Low
**Зависимости:** WASM debugger integration, source maps

---

### 9. Script Versioning 📜 MEDIUM PRIORITY

**Описание:**
Track история изменений скриптов. Version control, rollback, diff между версиями.

**Зачем:**
- Откатиться к предыдущей версии если что-то сломалось
- Audit trail - кто и когда изменил
- Сравнить изменения между версиями

**Где упоминается:**
- `docs/plans/scripting-api-implementation.md:1765`

**Реализация:**
1. Backend: `script_versions` table (version number, timestamp, snapshot)
2. API: `GET /_api/v1/scripts/{id}/versions`, `POST /_api/v1/scripts/{id}/rollback`
3. Frontend: Version history viewer, diff между версиями

**Сложность:** Medium
**Время:** 10-12 часов
**Приоритет:** Medium
**Зависимости:** Database schema changes

---

### 10. Performance Profiling 📊 LOW PRIORITY

**Описание:**
Детальные metrics выполнения скрипта. CPU time breakdown, memory usage, bottlenecks.

**Зачем:**
- Оптимизировать медленные скрипты
- Найти memory leaks
- Понять почему скрипт тормозит

**Где упоминается:**
- `docs/plans/scripting-api-implementation.md:1766`

**Реализация:**
1. Backend: Extism profiling API или custom instrumentation
2. Metrics: CPU time, memory allocations, syscall counts
3. Frontend: Flame graph или timeline visualization

**Сложность:** Hard
**Время:** 12-15 часов
**Приоритет:** Low
**Зависимости:** Profiling instrumentation в WASM runtime

---

## Категория В: Advanced Features (5 фич)

### 11. Bulk Operations ✅ MEDIUM PRIORITY

**Описание:**
Выбрать несколько скриптов и применить действие (enable/disable/delete) ко всем сразу.

**Зачем:**
- Управление большим количеством скриптов (10+)
- Быстро включить/выключить группу
- Массовое удаление тестовых скриптов

**Где упоминается:**
- `docs/plans/scripts-ui-implementation.md:292-295`
- `docs/plans/scripts-ui-implementation.md:435`

**Реализация:**
1. Frontend: Checkbox mode в списке скриптов
2. Multi-select UI (Ctrl+Click, Shift+Click)
3. Bulk action bar: "Enable All (5 selected)", "Delete All"
4. Backend: loop или `POST /_api/v1/scripts/bulk` (опционально)

**Сложность:** Medium
**Время:** 4-6 часов
**Приоритет:** Medium
**Зависимости:** Нет

---

### 12. Template System 🎨 MEDIUM PRIORITY

**Описание:**
Интерактивный wizard для создания скриптов из templates. Не просто examples, а параметризованные шаблоны.

**Зачем:**
- Быстрое создание типичных скриптов без написания кода
- Заполнить параметры → автогенерация кода
- Пример: "Add Header" template → ввести header name/value → готов код на Rust

**Где упоминается:**
- `docs/plans/WASM_COMPILATION_COMPLETE.md:263`

**Реализация:**
1. Template wizard: Step 1 - выбрать тип (Add Header, Transform Body, etc.)
2. Step 2 - заполнить параметры (header name, value, etc.)
3. Step 3 - автогенерация кода на основе параметров
4. Backend: Template engine (Go templates или Jinja2-like)

**Сложность:** Medium
**Время:** 8-10 часов
**Приоритет:** Medium
**Зависимости:** Examples Library

---

### 13. Script Marketplace 🛒 LOW PRIORITY

**Описание:**
Community-contributed scripts с rating, reviews, one-click install. Как VS Code Extensions Marketplace.

**Зачем:**
- Sharing скриптов с community
- Discovery - найти готовые решения
- Monetization - платные premium скрипты (опционально)

**Где упоминается:**
- `docs/plans/WASM_COMPILATION_COMPLETE.md:264`
- `docs/plans/scripting-api-implementation.md:1755-1757`

**Реализация:**
1. Backend: Marketplace API (publish, search, download, rate, review)
2. Database: marketplace_scripts table
3. Frontend: Marketplace page с категориями, рейтингами, поиском
4. Security: Code review, sandboxing, malware detection
5. User auth, moderation tools

**Сложность:** Very Hard
**Время:** 60+ часов (полноценная отдельная фича)
**Приоритет:** Low
**Зависимости:** Auth system, Script publishing API, Moderation tools

---

### 14. Python Support 🐍 LOW PRIORITY

**Описание:**
Поддержка Python скриптов через Pyodide WASM или subprocess execution.

**Зачем:**
- Python - популярный язык для scripting
- Много библиотек (requests, pandas, etc.)
- Знаком многим пользователям

**Где упоминается:**
- `docs/plans/WASM_COMPILATION_COMPLETE.md:262`
- `docs/plans/scripting-api-implementation.md:1751-1753`

**Реализация:**
1. Backend: PythonCompiler implementation (Pyodide или subprocess)
2. Extism plugin для Python (если через WASM)
3. Dependency management (pip packages)
4. Frontend: Python language support

**Сложность:** Very Hard
**Время:** 20-30 часов
**Приоритет:** Low (уже есть Rust, Go, JavaScript, Dart)
**Зависимости:** Pyodide или Python runtime

---

### 15. WebSocket Scripts 🔌 LOW PRIORITY

**Описание:**
Поддержка скриптов для модификации WebSocket messages (не только HTTP).

**Зачем:**
- Модифицировать WebSocket frames
- Intercept WebSocket handshake
- Testing WebSocket-based APIs

**Где упоминается:**
- `docs/plans/scripting-feature-completion.md:15`
- `docs/plans/scripting-feature-completion.md:316`

**Реализация:**
1. Backend: WebSocket message interception
2. Script API: onWebSocketMessage(), modifyWebSocketFrame()
3. Frontend: WebSocket script editor с отдельными настройками

**Сложность:** Very Hard
**Время:** 40+ часов
**Приоритет:** Low (намеренно отложено как "слишком большая задача")
**Зависимости:** WebSocket proxy infrastructure

---

## Категория Г: Performance/Reliability (9 фич)

### 16. Compilation Retry Logic 🔄 LOW PRIORITY

**Описание:**
3 попытки с exponential backoff (1s, 2s, 4s) при ошибках компиляции.

**Зачем:**
- Обработка transient errors (network issues)
- Retry при temporary compiler failures

**Где упоминается:**
- `docs/plans/scripting-feature-completion.md:229-231`

**Реализация:**
1. Backend: retry loop в CompilationService
2. Backoff strategy: 1s, 2s, 4s
3. Frontend: показать "Retrying (2/3)..."

**Сложность:** Easy
**Время:** 2-3 часа
**Приоритет:** Low (компиляция обычно детерминированная)
**Зависимости:** Нет

---

### 17. Disk Space Check 💾 MEDIUM PRIORITY

**Описание:**
Проверка свободного места перед скачиванием компилятора.

**Зачем:**
- Предотвратить "No space left on device" errors
- Показать понятное сообщение + кнопка "Clear Cache"

**Где упоминается:**
- `docs/plans/scripting-feature-completion.md:234-237`
- `docs/plans/compiler-download-on-demand-system.md:1022-1043`

**Реализация:**
1. Backend: `unix.Statfs()` для проверки
2. Error message: "Not enough disk space. Need 600MB, available 200MB"
3. Frontend: показать ошибку + кнопку "Clear Cache"

**Сложность:** Easy
**Время:** 2-3 часа
**Приоритет:** Medium
**Зависимости:** Нет

---

### 18. Resume Support для Downloads ⏯️ LOW PRIORITY

**Описание:**
HTTP Range header для продолжения прерванной загрузки компилятора.

**Зачем:**
- Не начинать заново при обрыве соединения
- Полезно для медленных/нестабильных сетей

**Где упоминается:**
- `docs/plans/scripting-feature-completion.md:239-242`

**Реализация:**
1. Backend: сохранять partial downloads в temp файлы
2. HTTP Range header: `Range: bytes=1024000-`
3. Resume logic: проверить checksum и продолжить

**Сложность:** Medium
**Время:** 6-8 часов
**Приоритет:** Low (компиляторы скачиваются 1 раз)
**Зависимости:** Нет

---

### 19. Compilation Metrics 📈 LOW PRIORITY

**Описание:**
Статистика компиляций (duration, success rate, cache hit rate).

**Зачем:**
- Monitoring и observability
- Оптимизация компиляции
- Alerts при проблемах

**Где упоминается:**
- `docs/plans/WASM_COMPILATION_COMPLETE.md:261`

**Реализация:**
1. Backend: metrics collection (Prometheus client)
2. Metrics: compilation_duration_seconds, compilation_success_rate, cache_hit_rate
3. Frontend: Stats page с графиками

**Сложность:** Medium
**Время:** 6-8 часов
**Приоритет:** Low
**Зависимости:** Metrics infrastructure (Prometheus или custom)

---

### 20. Compiler Version Pinning 📌 MEDIUM PRIORITY

**Описание:**
Закрепить конкретную версию компилятора вместо всегда latest.

**Зачем:**
- Reproducible builds
- Избежать breaking changes в новых версиях
- Pin версию для production

**Где упоминается:**
- `docs/plans/compiler-download-on-demand-system.md:1109`

**Реализация:**
1. Backend: поддержка `{cache}/compilers/{lang}/{version}/`
2. API: `GET /_api/v1/compilers/{lang}/versions` - список доступных версий
3. Frontend: dropdown для выбора версии при установке

**Сложность:** Medium
**Время:** 8-10 часов
**Приоритет:** Medium
**Зависимости:** Metadata API для доступных версий

---

### 21. Auto-Update Compilers 🔄 LOW PRIORITY

**Описание:**
Проверка новых версий компиляторов и уведомление пользователя.

**Зачем:**
- Всегда использовать latest features
- Security patches автоматически
- Удобство - не нужно вручную проверять

**Где упоминается:**
- `docs/plans/compiler-download-on-demand-system.md:1110`

**Реализация:**
1. Backend: cron job для проверки новых версий (GitHub releases API)
2. Notification system: уведомить пользователя о доступном обновлении
3. Frontend: badge "Update Available" + кнопка "Update"

**Сложность:** Medium
**Время:** 6-8 часов
**Приоритет:** Low
**Зависимости:** Version Pinning

---

### 22. Bandwidth Limiting 🚦 LOW PRIORITY

**Описание:**
Ограничение скорости загрузки компиляторов (для медленных соединений).

**Зачем:**
- Не забивать весь канал при download
- Полезно для shared connections

**Где упоминается:**
- `docs/plans/compiler-download-on-demand-system.md:1111`

**Реализация:**
1. Backend: rate limiter для HTTP downloads
2. Settings: "Max Download Speed" (MB/s)
3. Library: `golang.org/x/time/rate` или custom

**Сложность:** Easy
**Время:** 3-4 часа
**Приоритет:** Low
**Зависимости:** Нет

---

### 23. Proxy Support для Downloads 🌐 MEDIUM PRIORITY

**Описание:**
Скачивание компиляторов через HTTP proxy (для корпоративных сетей).

**Зачем:**
- Работа в корпоративных сетях с proxy
- Compliance с network policies

**Где упоминается:**
- `docs/plans/compiler-download-on-demand-system.md:1112`

**Реализация:**
1. Backend: HTTP client с proxy support (`http.Client.Transport.Proxy`)
2. Settings: "HTTP Proxy URL" (optional)
3. Environment variable: `HTTP_PROXY`, `HTTPS_PROXY`

**Сложность:** Easy
**Время:** 2-3 часа
**Приоритет:** Medium (важно для корпоративных users)
**Зависимости:** Нет

---

### 24. Offline Mode 📴 LOW PRIORITY

**Описание:**
Pre-download компиляторов для offline использования.

**Зачем:**
- Работа без интернета
- Air-gapped environments
- Predictable setup time

**Где упоминается:**
- `docs/plans/compiler-download-on-demand-system.md:1113`

**Реализация:**
1. Backend: кнопка "Download All Compilers" (pre-download)
2. Fallback: если download failed, использовать cached версию
3. Frontend: indicator "Offline Mode" + cached compilers list

**Сложность:** Medium
**Время:** 4-6 часов
**Приоритет:** Low
**Зависимости:** Нет

---

## Итоговая таблица (сортировано по приоритету)

| # | Фича | Категория | Время | Приоритет | Сложность |
|---|------|-----------|-------|-----------|-----------|
| 1 | Examples Library | UI/UX | 4-6ч | **HIGH** | Easy |
| 17 | Disk Space Check | Performance | 2-3ч | **Medium** | Easy |
| 20 | Compiler Version Pinning | Performance | 8-10ч | **Medium** | Medium |
| 23 | Proxy Support | Performance | 2-3ч | **Medium** | Easy |
| 2 | Import/Export Scripts | UI/UX | 6-8ч | Medium | Medium |
| 4 | Diff View | UI/UX | 4-5ч | Medium | Medium |
| 9 | Script Versioning | DevEx | 10-12ч | Medium | Medium |
| 11 | Bulk Operations | Advanced | 4-6ч | Medium | Medium |
| 12 | Template System | Advanced | 8-10ч | Medium | Medium |
| 3 | Keyboard Shortcuts | UI/UX | 2-3ч | Low | Easy |
| 5 | Copy as cURL | UI/UX | 2ч | Low | Easy |
| 16 | Retry Logic | Performance | 2-3ч | Low | Easy |
| 22 | Bandwidth Limiting | Performance | 3-4ч | Low | Easy |
| 6 | WebSocket Logs | DevEx | 8-10ч | Low | Hard |
| 7 | Hot Reload | DevEx | 6-8ч | Low | Medium |
| 18 | Resume Downloads | Performance | 6-8ч | Low | Medium |
| 19 | Compilation Metrics | Performance | 6-8ч | Low | Medium |
| 21 | Auto-Update Compilers | Performance | 6-8ч | Low | Medium |
| 24 | Offline Mode | Performance | 4-6ч | Low | Medium |
| 10 | Performance Profiling | DevEx | 12-15ч | Low | Hard |
| 14 | Python Support | Advanced | 20-30ч | Low | Very Hard |
| 8 | Debug Console | DevEx | 40+ч | Low | Very Hard |
| 15 | WebSocket Scripts | Advanced | 40+ч | Low | Very Hard |
| 13 | Marketplace | Advanced | 60+ч | Low | Very Hard |

---

## Рекомендации по приоритизации

### 🎯 Quick Wins (10-15 часов, высокая ценность):

1. **Examples Library** (6ч) - сильно улучшит onboarding ⭐
2. **Disk Space Check** (3ч) - предотвратит ошибки установки
3. **Keyboard Shortcuts** (3ч) - для power users
4. **Copy as cURL** (2ч) - полезно для testing
5. **Proxy Support** (3ч) - для корпоративных users

**Итого:** ~17 часов, очень высокая ценность

---

### 🚀 Medium Priority (50 часов, средняя ценность):

6. **Import/Export** (8ч) - для sharing и backup
7. **Compiler Version Pinning** (10ч) - для reproducible builds
8. **Bulk Operations** (6ч) - для управления большим количеством скриптов
9. **Diff View** (5ч) - улучшит Test Tab
10. **Script Versioning** (12ч) - для production environments
11. **Template System** (10ч) - можно использовать Examples Library как замену

**Итого:** ~51 час

---

### ⏳ Long-term / Low Priority (240+ часов):

12. **WebSocket Logs** (10ч) - nice to have, SSE работает хорошо
13. **Performance Profiling** (15ч) - только если есть production performance issues
14. **Python Support** (30ч) - уже есть 4 языка, достаточно
15. **WebSocket Scripts** (40ч) - отдельная большая фича
16. **Marketplace** (60ч) - требует отдельной инфраструктуры
17. **Debug Console** (40ч) - очень сложная задача
18. И остальные...

**Итого:** ~240+ часов

---

## Общая оценка

**Всего Nice to Have фич:** 24
**Всего времени:** ~300+ часов

**По приоритетам:**
- **High Priority:** ~10 часов
- **Medium Priority:** ~50 часов
- **Low Priority:** ~240 часов

---

## Следующие шаги

**Рекомендуется:**
1. ✅ Текущая версия Scripting Feature уже production-ready на 100%
2. Начать с Quick Wins (Examples Library + Disk Space Check) - высокая ценность, мало времени
3. Добавлять Medium Priority фичи по мере необходимости (на основе feedback пользователей)
4. Low Priority фичи - только если есть конкретная потребность

**Не рекомендуется:**
- Начинать с Very Hard фич (Marketplace, Debug Console, WebSocket Scripts)
- Делать все сразу - лучше добавлять постепенно

---

## История изменений

- **2025-11-04:** Создан документ, собраны все Nice to Have фичи из разных планов
