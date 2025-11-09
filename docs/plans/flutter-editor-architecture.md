Отлично. Ниже — **архитектура плагинов для обычного Flutter-пакета**: без изоляторов и отдельных процессов **сейчас**, но с закладками, чтобы завтра безболезненно прикрутить Extism/WASM.

---

# Ключевые ориентиры из практики

* **Separation of concerns** и чистые границы модулей — официальные рекомендации Flutter. ([docs.flutter.dev][1])
* **Ports & Adapters (Hexagonal)**: ядро не зависит от инфраструктуры; расширения сидят за интерфейсами. Это будущая «стыковка» с любым рантаймом (в т.ч. Extism). ([Java Code Geeks][2])
* **Манифест/activation events** как у VS Code — активируйте плагины ровно тогда, когда они нужны (у нас это будет синхронный реестр без отдельного хоста, но тот же принцип). ([Visual Studio Code][3])
* **Federated структура пакетов** — привычный во Flutter способ делить API/реализации/адаптеры, чтобы проект втаскивал только нужное. ([Habr][4])
* **SemVer и стабильный публичный API** — критично для экосистемы плагинов. ([Dart][5])

---

# Цели и ограничения

* **Нет изоляторов/процессов**: плагины — обычные Dart-пакеты в том же изоляте.
* **Масштабируемость/гибкость**: контракты ядра стабильны, расширяемы через «extension points».
* **Будущая интеграция с Extism**: API и типы готовим «к сериализации» (JSON-совместимые DTO), вводим «capabilities» (права), чтобы потом отдать их WASI/Extism. ([GitHub][6])

---

# Структура репо (минимум)

```
packages/
  editor_core/          # домен, состояние, команды, события, extension points (SPI)
  editor_api/           # public API плагинов (интерфейсы, DTO, токены capabilities)
  editor_widgets/       # UI-компоненты редактора (не знают о конкретных плагинах)
  editor_plugin_registry/ # реализация реестра (in-process)
  plugins/
    plugin_markdown/    # пример плагина (Dart)
    plugin_s3sync/      # пример плагина: backend sync
```

* Приложение конечного пользователя ставит `editor_core + editor_widgets` и **любые** плагины; собирает реестр.

---

# Ядро расширяемости: команды, события, точки расширения

### 1) Командная шина (sync, без RPC)

Идея как у VS Code: «команды» по строковому `id` + типобезопасные обёртки.

```dart
// editor_api
typedef Json = Map<String, Object?>;

abstract class Command<TIn extends Object?, TOut extends Object?> {
  String get id; // 'editor.file.create'
  Future<TOut> handle(TIn input, EditorHost host);
}

abstract class EditorHost {
  // Мини-хост API: файловые операции, сеть, таймеры и т.п.
  Future<String> readFile(Uri uri);
  Future<void> writeFile(Uri uri, String contents);
  // ...расширяемо
}

// editor_plugin_registry
class CommandBus {
  final _map = <String, Command<Object?, Object?>>{};
  void register(Command cmd) => _map[cmd.id] = cmd as dynamic;
  Future<O?> exec<I, O>(String id, I input, EditorHost host) async {
    final c = _map[id]; if (c == null) throw StateError('No command $id');
    return await c.handle(input as Object?, host) as O?;
  }
}
```

### 2) События (event bus)

Синхронные/асинхронные события для UI/плагинов:

```dart
class EventBus {
  final _streams = <String, StreamController<Json>>{};
  Stream<Json> on(String topic) =>
      (_streams[topic] ??= StreamController.broadcast()).stream;
  void emit(String topic, Json payload) =>
      (_streams[topic] ??= StreamController.broadcast()).add(payload);
}
```

### 3) Extension Points (SPI)

Определяем **точки расширения** как интерфейсы + реестр:

```dart
abstract class FileTypeContribution {
  bool supports(String ext);                // .md, .json...
  Widget buildEditor(EditorContext ctx);    // вернуть виджет редактора
}

abstract class ToolbarContribution {
  List<ToolbarAction> actions(EditorContext ctx);
}

// Реестр
class ExtensionPoints {
  final fileEditors = <FileTypeContribution>[];
  final toolbars    = <ToolbarContribution>[];
  // add more: formatters, linters, sync providers, indexers...
}
```

### 4) Плагин-контракт

Плагин — обычный Dart-класс, который **регистрирует** себя в реестре:

```dart
abstract class EditorPlugin {
  String get id;
  VersionConstraint get api;     // совместимость с editor_api
  Set<String> get capabilities;  // 'fs:read', 'net:fetch', ...
  Future<void> activate(PluginContext ctx);
}

abstract class PluginContext {
  CommandBus get commands;
  EventBus get events;
  ExtensionPoints get points;
  EditorHost get host; // огранённый хост с проверкой прав
}
```

---

# Регистрация плагинов (без манифест-читалок и без изолятов)

Просто создайте **реестр** и покличьте `activate()` при старте:

```dart
final registry = PluginRegistry(
  host: CapabilityGuardHost(baseHost: DefaultHost(), grants: {/*...*/}),
);

await registry.use(PluginMarkdown());
await registry.use(PluginS3Sync());
```

* Позже можно добавить **манифест `plugin.yaml`** и lazy-activation «по событиям» (`onFileSystem`, `onLanguage:md`) — как у VS Code, только «внутри одного процесса». ([Visual Studio Code][3])

---

# Capabilities (права) — готовим почву для WASI/Extism

* Дайте плагинам **не весь** `EditorHost`, а обёртку, фильтрующую вызовы по разрешениям (`fs:read:/docs`, `net:fetch:https://api.example.com`).
* Это прямой аналог capability-подхода WASI → в будущем «переедет» в Extism без смены API. ([GitHub][6])

---

# Правила для API и DTO (важно для будущего WASM)

* **Только JSON-совместимые типы** (String, num, bool, List, Map) в сигнатурах команд/событий — легко сериализовать.
* Оборачивайте сложные типы в **DTO** с `toJson/fromJson`.
* **Нуль-безопасность** и **immutable** модели.
* Чётко версионируйте API (`editor_api`) по **SemVer**, фиксируйте совместимость плагинов через `VersionConstraint`. ([Dart][5])

---

# Extension points, которые стоит заложить сразу

1. **Команды** (`commands.register`) — универсальная точка входа. (Аналог команд в VS Code.) ([DEV Community][7])
2. **Редакторы по типу файла** (`fileEditors`) — для UI.
3. **Форматтеры/линтеры** (`formatters`, `linters`) — цепочка обработчиков.
4. **Синх-провайдеры** (`syncProviders`) — тонкие адаптеры к бэку (REST/gRPC/…); UI ничего не знает о транспорте. Это и есть Ports & Adapters. ([Java Code Geeks][2])
5. **Индексаторы/поиск** (`indexers`) — офлайн/кеш.
6. **Toolbar/меню** — вкладка действий, контекстные меню.

---

# Мини-пример плагина (без изоляторов)

```dart
class PluginMarkdown implements EditorPlugin {
  @override String get id => 'plugin.markdown';
  @override VersionConstraint get api => VersionConstraint.parse('^1.0.0');
  @override Set<String> get capabilities => {'fs:read', 'fs:write'};

  @override
  Future<void> activate(PluginContext ctx) async {
    // 1) Команда форматирования
    ctx.commands.register(_FormatCmd());

    // 2) Редактор для .md
    ctx.points.fileEditors.add(_MdEditorContribution());
  }
}

class _FormatCmd extends Command<Map<String, Object?>, void> {
  @override String get id => 'markdown.format';
  @override Future<void> handle(input, host) async {
    final uri = Uri.parse(input['uri'] as String);
    final text = await host.readFile(uri);
    final formatted = _format(text); // sync алгоритм
    await host.writeFile(uri, formatted);
  }
}
```

---

# Как это «подружится» с Extism завтра

* **Слой EditorHost** уже capability-ориентирован и опирается на JSON-DTO → его можно «перекинуть» на Extism host-functions с теми же именами/контрактами. ([GitHub][6])
* **Командная шина/события** останутся прежними; для внешних wasm-плагинов вы просто добавите «адаптер плагина», который проксирует `exec()` ↔ Extism call.
* **Activation events/манифест** можно расширить до полноценных деклараций как в VS Code (onLanguage, onCommand, и т.д.). ([Visual Studio Code][3])
* Если когда-нибудь появится отдельный «extension host», текущие интерфейсы уже совместимы с RPC (идемпотентные JSON-контракты).

---

# Гайд по DX/качество кода

* **Чёткие package-границы** (как federated): `editor_api` (контракты) меняется медленно; `editor_widgets` — отдельно; плагины зависят **только** от `editor_api`. ([DeepWiki][8])
* **Документация extension points** + шаблон плагина (скелет проекта и тесты).
* **Тесты**: unit на команды/хост, golden на виджеты, контрактные тесты для точек расширения.
* **Версионирование**: фиксируйте Breaking/Feature/Patch согласно Dart SemVer. ([Dart][5])

---

# Что получить на выходе

* Ставится «как обычный пакет» в любой Flutter-проект, **без изоляторов и внешних процессов**.
* Чёткие extension points и capability-guarded host API.
* Готовность к росту: от «плагинов-пакетов на Dart» → к wasm/Extism без перелома API.

Хочешь — накину минимальный стартовый монорепо (скелеты `editor_api/core/registry/widgets` + демонстрационный `plugin_markdown` и `plugin_s3sync`) и базовые тесты; можно начать с него и уже оттуда наращивать activation events и капабилити-гварды.

[1]: https://docs.flutter.dev/app-architecture/recommendations?utm_source=chatgpt.com "Architecture recommendations and resources - Flutter"
[2]: https://www.javacodegeeks.com/2025/06/hexagonal-architecture-in-practice-ports-adapters-and-real-use-cases.html?utm_source=chatgpt.com "Hexagonal Architecture in Practice: Ports, Adapters, and Real Use Cases"
[3]: https://code.visualstudio.com/api/references/activation-events?utm_source=chatgpt.com "Activation Events | Visual Studio Code Extension API"
[4]: https://habr.com/ru/companies/friflex/articles/780956/?utm_source=chatgpt.com "Создаем federated plugin для Flutter-проекта / Хабр"
[5]: https://dart.dev/tools/pub/versioning?utm_source=chatgpt.com "Package versioning - Dart"
[6]: https://github.com/WebAssembly/WASI?utm_source=chatgpt.com "GitHub - WebAssembly/WASI: WebAssembly System Interface"
[7]: https://dev.to/karrade7/vs-code-extensions-basic-concepts-architecture-b17?utm_source=chatgpt.com "VS Code Extensions: Basic Concepts & Architecture"
[8]: https://deepwiki.com/fluttercommunity/plus_plugins/1.1-plugin-architecture?utm_source=chatgpt.com "Plugin Architecture | fluttercommunity/plus_plugins | DeepWiki"
