# Firebase Realtime Database (Flutter `firebase_database`): план фронта + клиентской интеграции

Связанный документ (backend):
- [`firebase_database_sessions_backend.md`](./firebase_database_sessions_backend.md)

## 0) Текущее состояние и “дыры”, которые надо закрыть

Backend ingest для `kind="firebase_database"` уже готов, но UI сейчас:
- **ошибочно считает** такие сессии HTTP и открывает `HttpDetailsPanel`,
  потому что в `DetailsContainer` WS определяется только как `kind == 'ws'` или “старый ws без kind”.
- quick‑фильтры завязаны на **две разные системы**:
  - серверный параметр `types=...` в `GET /_api/v1/sessions`
  - локальные теги в `_tagsForSession()` (если их не добавить — UI отфильтрует всё в ноль)

Значит “фича” end-to-end = UI поправки + отдельный клиентский пакет, который начнёт слать ingest.

## 1) Definition of Done (frontend + dart)

- В списке сессий видны `kind="firebase_database"` и есть быстрый фильтр “Firebase”.
- Деталка Firebase‑сессии открывается и показывает фреймы (MVP: через `WsDetailsPanel`).
- `types=firebase` реально фильтрует список и на сервере, и локально в UI.
- В приложении интеграция включается флагом и не влияет на прод‑сборки.
- Поддержан `X-Admin-Token` (если задан `ADMIN_TOKEN` на бэке).

## 2) Что сделать в UI (frontend/)

### 2.1 Быстрые фильтры: добавить Firebase/RTDB (и сделать так, чтобы они реально работали)

Файлы:
- `frontend/lib/features/inspector/presentation/widgets/quick_filters_bar.dart`
- `frontend/lib/features/inspector/presentation/utils/sessions_filtering.dart`
- `frontend/lib/features/inspector/data/inspector_repository_impl.dart` (проверка поведения `types=...`)

Как сейчас устроено:
- `QuickFiltersBar` управляет `HomeUiStore.quickTypes`.
- `InspectorRepositoryImpl.listSessions()` подмешивает `types` из quickTypes в query параметр `types=...`.
- `filterVisibleSessions()` дополнительно режет список локально по `_tagsForSession()`.

Что сделать:
1) `quick_filters_bar.dart`
   - добавить в `typeMap` новые ключи:
     - `firebase` → `Firebase`
     - `rtdb` → `RTDB` (опционально)
     - `firebase_database` → `Firebase DB` (опционально)

2) `sessions_filtering.dart`
   - в `_tagsForSession()` добавить теги для Firebase‑сессий:
     - если `s.kind == 'firebase_database'`:
       - `firebase`
       - `rtdb`
       - `firebase_database`
   - (желательно) эвристика по хосту:
     - если `uri.host` оканчивается на `firebaseio.com` или `firebasedatabase.app` → добавить `firebase`

Почему так:
- ключ `firebase` должен совпасть с серверным тегом и с `types=firebase`, который уже поддержан бэком.
- локальный `_tagsForSession` нужен, иначе quick‑фильтр будет “пустым” из-за двойной фильтрации.

### 2.2 Деталка: Firebase должен открываться как “realtime frames” (MVP = WS панель)

Файлы:
- `frontend/lib/features/inspector/presentation/pages/home/widgets/details_container.dart`
- `frontend/lib/features/inspector/presentation/widgets/details/details_tabs.dart`

Где сейчас ошибка:
- в `DetailsContainer` вычисляется:
  - `isWs = (kind == 'ws') || (method.isEmpty && kind == null)`
  - `selIsHttp = !isWs`
  - для `kind="firebase_database"` это даёт `selIsHttp=true` → UI рендерит `HttpDetailsPanel`

Что сделать:
- в `DetailsContainer` расширить правило “показывать как ws”:
  - `kind == 'firebase_database'` → `showWs=true`, `showHttp=false`

Замечание:
- `DetailsTabs` уже умеет работать без табов: если `showWs && !showHttp` — сразу отдаёт `WsDetailsPanel`.

### 2.3 Список сессий/таймлайн: не считать Firebase “HTTP”

Файлы:
- `frontend/lib/features/inspector/presentation/widgets/sessions_column.dart`
- `frontend/lib/features/inspector/presentation/widgets/waterfall_timeline.dart`

Зачем:
- в этих местах есть эвристика `isWs`, которая влияет на визуальные элементы и поведение.

Что сделать:
- синхронизировать логику с `DetailsContainer`:
  - `kind == 'firebase_database'` считать “realtime” (как минимум для выбора панели/лейбла).

## 3) Клиентская интеграция: отдельный пакет в `dart_packages/`

### 3.1 Почему отдельный пакет

Такой же подход уже используется в проекте:
- `dio_debugger`
- `web_socket_channel_debugger`

Firebase‑интеграция должна быть:
- опциональной зависимостью для приложений;
- без форка `firebase_database`;
- без нативных плагинов (чистый Dart).

### 3.2 Новый пакет: `dart_packages/firebase_database_debugger`

Новый пакет:
- `dart_packages/firebase_database_debugger/`

Зависимости (MVP):
- `firebase_database`
- `http`

Платформенность:
- без `dart:html` (если нужен web‑специфичный код — через conditional import и `package:web`),
  но в MVP можно обойтись чистым `http`.

### 3.3 Контракт ingest (важные лимиты, чтобы не словить 400/413)

Backend ограничения (см. backend‑план):
- до **200 frames** в одном запросе (`firebaseIngestMaxFrames`)
- лимит request body **16MB**

Следствия для клиента:
- `maxBatchFrames <= 200` (по умолчанию 50–100)
- `flushInterval` 100–300мс (чтобы не отправлять на каждый ивент)
- большой payload класть в `body` (base64) и помнить, что бэк режет/спулит по лимитам

### 3.4 Структура API пакета (MVP)

Публичные классы:
- `FirebaseDatabaseDebuggerConfig`
  - `debuggerBaseUrl` (например, `http://localhost:9092`)
  - `enabled` (по умолчанию `true` в debug)
  - `adminToken` (опционально)
  - `flushInterval`
  - `maxBatchFrames`
- `FirebaseDatabaseDebugger`
  - `ref(DatabaseReference)` / `query(Query)` → возвращает обёртку
  - очередь фреймов и `flush()`
- `FirebaseIngestClient`
  - `ingest(session, frames, close, error)`

Минимальный пример:

```dart
final db = FirebaseDatabase.instance;
final dbg = FirebaseDatabaseDebugger(
  config: FirebaseDatabaseDebuggerConfig(
    debuggerBaseUrl: 'http://localhost:9092',
    enabled: kDebugMode,
    adminToken: const String.fromEnvironment('ND_ADMIN_TOKEN'),
    maxBatchFrames: 100,
  ),
);

final ref = dbg.ref(db.ref('users/123'));
await ref.set({'name': 'Alex'});
```

### 3.5 Как маппить операции `firebase_database` → session/frames

Session:
- `kind`: всегда `firebase_database` (на сервере это фиксировано)
- `id`: стабильный для одной “линии” событий (подписка/реф)
  - для операций `set/get/update/remove` можно использовать `refKey + path`
  - для подписок — отдельный id на каждую подписку (чтобы stop/cancel закрывал именно её)
- `target`: валидный URL:
  - `<databaseURL>/<path>?<query>`
  - хост/схема нужны для фильтров и группировки по домену в UI

Frame:
- `id`: уникальный (uuid/монотонный id)
- `ts`: `DateTime.now().toUtc()`
- `direction`:
  - операции → `client->upstream`
  - события подписки → `upstream->client`
- `opcode`: `text`
- `preview`: JSON‑конверт (строка)
- `body`: опционально (когда payload большой)

Рекомендуемый формат `preview` (чтобы UI мог парсить красиво позже):

```json
{
  "type": "firebase_database",
  "op": "set",
  "path": "/users/123",
  "query": {"orderBy":"key"},
  "ok": true,
  "durationMs": 12,
  "error": null
}
```

### 3.6 Потоки (listen) и отмена подписки

Требование:
- подписка должна логировать:
  - событие “listen_start”
  - входящие value events (onValue/onChild*)
  - “listen_cancel” (и по желанию `close=true` по session)

Практика:
- обёртка над `Stream` должна прокидывать события дальше без изменения поведения приложения.
- cancellation должен быть best‑effort: если отправить close не удалось — это не должно ронять приложение.

## 4) Этапы внедрения (чтобы быстро получить пользу)

1) UI‑правки (без пакета) — проверить на ручном ingest через curl/тестовый клиент
   - quick‑фильтры “Firebase”
   - открытие Firebase как WS‑панели
2) Создать `dart_packages/firebase_database_debugger` (MVP: set/get/update/remove + onValue)
3) Добавить пример в README/интеграции (как подключать и куда слать)
4) Опционально: “красивый” рендер preview в WS‑панели

## 5) Тест‑план

### 5.1 UI (ручной)
- Отправить ingest в `POST /_api/v1/ingest/firebase_database`, убедиться:
  - сессия видна в списке
  - чип `Firebase` показывает только эти сессии
  - деталка открывается как WS и показывает frames

### 5.2 Dart пакет (unit)
- сериализация preview/bodyEncoding
- батчинг: не больше 200 frames/запрос
- ретраи/идемпотентность по `frame.id` (повторный flush не плодит дубль на сервере)

## 6) Практика подключения (чтобы не упереться в сеть/безопасность)

### 6.1 Базовый сценарий (эмулятор/симулятор)

- `debuggerBaseUrl`: `http://localhost:9092`
- если на бэке включён `ADMIN_TOKEN` → передавать `X-Admin-Token`

### 6.2 Физический телефон

Обычно `localhost` на телефоне — это не машина разработчика. Нужно:
- поднять backend на машине,
- использовать LAN IP машины, например `http://192.168.1.10:9092`

Безопасность ingest:
- если `ADMIN_TOKEN` не задан, бэк по умолчанию принимает ingest только с loopback/локальной сети
  (это ок для телефона в той же Wi‑Fi сети).
- если нужно осознанно открыть наружу — `INGEST_ALLOW_REMOTE=1`, но лучше всё же включить токен.

## 7) Риски и компромиссы

- **Шум от подписок**: `onValue` может стрелять часто. Поэтому батчинг и `flushInterval` обязателен.
- **Секреты**: бэк редактирует JSON в preview, но лучше не отправлять чувствительные поля в `body`,
  либо предусмотреть клиентскую опцию “не слать body”.
- **Покрытие API**: у `firebase_database` много методов, MVP стоит делать инкрементально
  (сначала основные операции + `onValue`).
- **Это не “перехват сети”**: мы логируем публичный API и события, а не сырой WebSocket трафик нативного SDK.

## 8) Ссылки

- backend ingest, безопасность и лимиты: [`firebase_database_sessions_backend.md`](./firebase_database_sessions_backend.md)

