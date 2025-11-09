# План реализации UI для управления скриптами (Scripts Feature)

**Дата:** 2025-11-03
**Статус:** В разработке
**Архитектура:** Clean Architecture + DDD + SOLID + DRY

---

## Общее описание

Полноценный UI для создания, редактирования и управления скриптами (Extism WASM и Dart) с:
- Monaco Editor для редактирования кода (Rust/JavaScript/Dart)
- Визуальным конструктором match rules
- Панелью тестирования скриптов
- Drag-n-drop загрузкой WASM файлов

---

## Архитектура

```
frontend/lib/features/scripts/
├── domain/          # Entities, Value Objects, Repository Interfaces
├── data/            # Repository Implementations, DTOs, API
├── application/     # Use Cases, Services, Stores (MobX)
└── presentation/    # Pages, Widgets, UI Logic
```

---

## Phase 1: Domain Layer (Чистая бизнес-логика)

### 1.1 Entities
- `script.dart` - Script entity (id, name, runtime, code, enabled, priority, etc.)
- `match_rules.dart` - MatchRules value object (methods, pathPattern, patternType)
- `script_config.dart` - ScriptConfig value object (timeoutMs, memoryLimitMB)
- `script_test_result.dart` - Результат тестирования скрипта

### 1.2 Repository Interface
- `scripts_repository.dart` - Контракт для работы с данными
  - `Future<List<Script>> getAll()`
  - `Future<Script> getById(String id)`
  - `Future<Script> create(Script script)`
  - `Future<Script> update(String id, Script script)`
  - `Future<void> delete(String id)`
  - `Future<void> toggle(String id, bool enabled)`
  - `Future<ScriptTestResult> test(Script script, TestRequest request)`

### 1.3 Use Cases (Single Responsibility)
- `create_script_usecase.dart` - Создание скрипта
- `update_script_usecase.dart` - Обновление
- `delete_script_usecase.dart` - Удаление
- `toggle_script_usecase.dart` - Включение/выключение
- `test_script_usecase.dart` - Тестирование перед применением

---

## Phase 2: Data Layer (Реализация инфраструктуры)

### 2.1 DTOs (Data Transfer Objects)
- `script_dto.dart` - Маппинг JSON ↔ Entity
- `match_rules_dto.dart` - Маппинг match rules
- `script_config_dto.dart` - Маппинг config

### 2.2 API Service
- `scripts_api_service.dart` - HTTP клиент для REST API
  - `GET /_api/v1/scripts`
  - `POST /_api/v1/scripts`
  - `GET /_api/v1/scripts/{id}`
  - `PUT /_api/v1/scripts/{id}`
  - `DELETE /_api/v1/scripts/{id}`
  - `PATCH /_api/v1/scripts/{id}/toggle`
  - `POST /_api/v1/scripts/test` (новый endpoint для тестирования)

### 2.3 Repository Implementation
- `scripts_repository_impl.dart` - Реализация с Dependency Injection

---

## Phase 3: Application Layer (Оркестрация)

### 3.1 MobX Store
- `scripts_store.dart` - Главный store
  - Observable: `scripts`, `selectedScript`, `isLoading`, `error`
  - Actions: `loadScripts()`, `createScript()`, `updateScript()`, etc.
  - Computed: `enabledScripts`, `filteredScripts`

### 3.2 Editor Store (локальный для dialog)
- `script_editor_store.dart`
  - State для формы создания/редактирования
  - Валидация полей
  - Управление Monaco editor state

---

## Phase 4: Presentation - Core UI

### 4.1 Main Page
- `scripts_page.dart`
  - AppBar с кнопкой "+ Create Script"
  - Поиск и фильтры
  - ScriptsList widget
  - FAB для быстрого создания

### 4.2 Scripts List
- `scripts_list.dart` - Список скриптов
- `script_card.dart` - Карточка скрипта с:
  - Название, описание, язык
  - Runtime badge (Extism/Dart)
  - Enabled toggle
  - Quick actions: Edit, Delete, Duplicate, Test
  - Priority indicator
  - Match rules preview

### 4.3 Empty State
- `scripts_empty_state.dart` - Когда скриптов нет
  - Приветственный текст
  - Кнопки: "Create First Script", "Import Example"

---

## Phase 5: Monaco Editor Integration

### 5.1 Package Setup
- Добавить `flutter_monaco` в `pubspec.yaml`
- Настроить WebView permissions (если нужно)

### 5.2 Monaco Wrapper
- `monaco_code_editor.dart` - Обертка с настройками:
  - Языки: `rust`, `javascript`, `typescript`, `dart`
  - Темы: `vs-dark`, `vs-light` (из app settings)
  - Options: line numbers, minimap, word wrap
  - Callbacks: `onChanged`, `onSave`

### 5.3 Language Support
- Rust syntax highlighting
- JavaScript/TypeScript highlighting
- Dart highlighting
- Автоопределение языка по полю `language`

---

## Phase 6: Script Editor Dialog

### 6.1 Main Dialog
- `script_editor_dialog.dart` - Full-screen dialog
  - Tabs: "Code", "Settings", "Match Rules", "Test"
  - Save/Cancel кнопки
  - Validation feedback

### 6.2 Code Tab
- Monaco editor (во весь экран)
- Language selector dropdown
- Runtime selector (Extism/Dart)
- Upload WASM button (если Extism)

### 6.3 Settings Tab
- `script_settings_form.dart`
  - Name (required)
  - Description (optional)
  - Priority (slider 0-100)
  - Trigger Type (radio: request/response/both)
  - Enabled toggle
  - Timeout (milliseconds)
  - Memory limit (MB)

---

## Phase 7: Visual Match Rules Builder

### 7.1 Match Rules Form
- `match_rules_form.dart`
  - Section: "When to execute this script"
  - Toggle: "Apply to all requests" vs "Custom rules"

### 7.2 HTTP Methods Selector
- `http_methods_multi_select.dart`
  - Chips: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
  - Multi-select
  - "Select All" / "Clear All"

### 7.3 Pattern Inputs
- `pattern_input.dart`
  - Pattern Type selector (exact, prefix, wildcard)
  - Path pattern input с примерами
  - Host pattern input
  - Validation (wildcard syntax)

### 7.4 Preview Panel
- `match_rules_preview.dart`
  - Примеры URLs которые совпадут
  - Примеры URLs которые НЕ совпадут
  - Live validation

---

## Phase 8: WASM Upload (Drag-n-Drop)

### 8.1 Upload Zone
- `wasm_upload_zone.dart`
  - Drag-n-drop область
  - File picker кнопка (fallback)
  - Поддержка `.wasm` и `.wat` файлов
  - Preview: имя файла, размер

### 8.2 Processing
- Чтение файла в bytes
- Конвертация в base64
- Заполнение поля `code` в форме
- Автоопределение языка (по имени файла или metadata)

### 8.3 Visual Feedback
- Progress indicator при загрузке
- Success/Error toast
- Размер файла лимит (например, 5MB)

---

## Phase 9: Script Testing Panel

### 9.1 Test Panel UI
- `script_test_panel.dart` - В отдельной tab "Test"
  - Sample request builder
  - "Run Test" кнопка
  - Results viewer

### 9.2 Sample Request Builder
- `test_request_builder.dart`
  - HTTP Method selector
  - URL path input
  - Headers editor (key-value pairs)
  - Body editor (JSON)
  - "Use real request from history" кнопка

### 9.3 Test Execution
- API call к backend для тестирования
- Показать loading state
- Обработка ошибок (compilation errors, runtime errors)

### 9.4 Results Viewer
- `test_result_viewer.dart`
  - Split view: "Original" | "Modified"
  - Diff highlighting для изменений
  - Logs viewer (stdout/stderr)
  - Execution time badge
  - Error panel (если есть)

---

## Phase 10: Navigation & Integration

### 10.1 Main Navigation
- Добавить "Scripts" в главное меню/drawer
- Иконка: `Icons.code` или `Icons.extension`
- Route: `/scripts`

### 10.2 Deep Linking
- `/scripts` - список
- `/scripts/create` - создание
- `/scripts/:id/edit` - редактирование

### 10.3 Permissions & Guards
- Проверка доступа (если есть auth)
- Route guards

---

## Phase 11: Additional Features

### 11.1 Quick Actions
- Duplicate script (копирование с новым именем)
- Export script (JSON файл)
- Import script (JSON upload)
- Export as curl command

### 11.2 Examples Library
- `examples_dialog.dart`
- Встроенные примеры скриптов:
  - "Add Header" (Rust)
  - "Log Request" (JavaScript)
  - "Modify Response" (Dart)
- One-click import

### 11.3 Filters & Search
- `scripts_filters.dart`
  - Search by name
  - Filter by runtime (Extism/Dart)
  - Filter by trigger type
  - Filter by enabled/disabled
  - Sort by: name, priority, created date

### 11.4 Bulk Operations
- Select multiple scripts
- Bulk enable/disable
- Bulk delete

---

## Phase 12: Polish & UX

### 12.1 Loading States
- Skeleton loaders для списка
- Progress indicators
- Optimistic updates

### 12.2 Error Handling
- Network errors
- Validation errors (красная подсветка)
- API errors (toast notifications)
- Retry logic

### 12.3 Tooltips & Help
- Info icons с объяснениями
- Примеры для pattern syntax
- Links на документацию

### 12.4 Responsive Design
- Desktop layout (wide editor)
- Mobile layout (full-screen editor)
- Tablet layout

### 12.5 Keyboard Shortcuts
- Ctrl+S для сохранения
- Ctrl+T для тестирования
- Esc для закрытия dialog

---

## Технические детали

### Dependencies
```yaml
dependencies:
  flutter_monaco: ^0.2.0       # Monaco editor
  mobx: ^2.3.3                 # State management
  flutter_mobx: ^2.2.1
  dio: ^5.4.0                  # HTTP client
  file_picker: ^8.0.0          # File upload
  desktop_drop: ^0.4.4         # Drag-n-drop (desktop)
  freezed_annotation: ^2.4.1   # Immutable models
  json_annotation: ^4.8.1

dev_dependencies:
  mobx_codegen: ^2.6.0
  build_runner: ^2.4.8
  freezed: ^2.4.6
  json_serializable: ^6.7.1
```

### API Extensions (Backend)
Возможно потребуется добавить endpoint для тестирования:
```
POST /_api/v1/scripts/test
Body: { script: {...}, testRequest: {...} }
Response: { modifiedRequest, logs, duration, error }
```

---

## Итоговая структура файлов

```
frontend/lib/features/scripts/
├── domain/
│   ├── entities/
│   │   ├── script.dart
│   │   ├── match_rules.dart
│   │   ├── script_config.dart
│   │   └── script_test_result.dart
│   ├── repositories/
│   │   └── scripts_repository.dart
│   └── usecases/
│       ├── create_script_usecase.dart
│       ├── update_script_usecase.dart
│       ├── delete_script_usecase.dart
│       ├── toggle_script_usecase.dart
│       └── test_script_usecase.dart
├── data/
│   ├── models/
│   │   ├── script_dto.dart
│   │   ├── match_rules_dto.dart
│   │   └── script_config_dto.dart
│   ├── services/
│   │   └── scripts_api_service.dart
│   └── repositories/
│       └── scripts_repository_impl.dart
├── application/
│   └── stores/
│       ├── scripts_store.dart
│       └── script_editor_store.dart
└── presentation/
    ├── pages/
    │   └── scripts_page.dart
    ├── widgets/
    │   ├── script_list.dart
    │   ├── script_card.dart
    │   ├── script_editor_dialog.dart
    │   ├── monaco_code_editor.dart
    │   ├── script_settings_form.dart
    │   ├── match_rules_form.dart
    │   ├── http_methods_multi_select.dart
    │   ├── pattern_input.dart
    │   ├── match_rules_preview.dart
    │   ├── wasm_upload_zone.dart
    │   ├── test_request_builder.dart
    │   ├── test_result_viewer.dart
    │   ├── scripts_filters.dart
    │   ├── examples_dialog.dart
    │   └── scripts_empty_state.dart
    └── utils/
        ├── script_validators.dart
        └── base64_helpers.dart
```

---

## Порядок реализации (приоритеты)

### MVP (Минимальный функционал):
1. ✅ Domain entities + repository interface
2. ✅ Data layer (API + DTOs)
3. ✅ Application store
4. ✅ Main page + list
5. ✅ Basic editor dialog (Monaco + settings)
6. ✅ CRUD operations

### Priority Features:
7. ⬜ Visual match rules builder
8. ⬜ Drag-n-drop WASM upload
9. ⬜ Script testing panel

### Nice to Have:
10. ⬜ Examples library
11. ⬜ Import/Export
12. ⬜ Bulk operations
13. ⬜ Keyboard shortcuts

---

## Оценка времени
- MVP: ~3-4 дня
- Priority Features: ~2-3 дня
- Nice to Have: ~1-2 дня
- **Total: ~6-9 дней** (full-time)

---

## История изменений

- **2025-11-03**: Создание плана, начало реализации
