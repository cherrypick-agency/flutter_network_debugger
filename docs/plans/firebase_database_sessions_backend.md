# Firebase Realtime Database (Flutter `firebase_database`): backend‑план для нового типа сессий

**Дата**: 2026-02-15  
**Статус**: Имплементировано (с доп. усилениями)  
**Оценка сложности (backend)**: 4/10  

## Executive summary

Мы добавляем в Go‑бэкенд новый “источник данных”, который **не** перехватывает сеть,
а принимает **структурированные события** от Flutter‑обёртки `firebase_database_debugger`.

События складываются в существующую модель:
- `domain.Session` (как “подключение”, аналог WS session)
- `domain.Frame` (как “сообщение/ивент”, аналог WS frame)

Благодаря этому UI может отображать RTDB‑дебаг так же, как WS: timeline, поиск, фильтры,
и live‑обновления через monitor.

Ключевые тонкости, которые план закрывает:
- **visibility / capture**: сессии не должны пропадать из списка, если запись выключена
- **idempotency**: повторные POST’ы не создают дублей
- **память/размеры**: preview режем и редактируем, большие тела уносим в `BodyFile`
- **безопасность**: ingest не должен стать “дырой”, но должен работать с телефона/эмулятора

---

## 1) Ограничения и предположения

- Flutter `firebase_database` общается через нативные SDK/JS SDK, поэтому сырой WebSocket
  трафик на стороне Dart **недоступен**.
- Мы дебажим **операции и события** публичного API (`set/get/update/onValue/...`) и
  отправляем их на бэкенд.
- На бэке пока основной storage — `memory.Store` (как сейчас для HTTP/WS).

---

## 2) Definition of Done (backend)

- В `GET /_api/v1/sessions` появляются элементы с `kind="firebase_database"`.
- В `GET /_api/v1/sessions/{id}/frames` видны фреймы с событиями/операциями.
- Live‑обновления в UI работают через monitor:
  - `session_started`, `frame_added`, `session_ended`.
- Для больших payload:
  - preview в `Frame.Preview` обрезан/редактирован,
  - полный body доступен через `GET /_api/v1/sessions/{id}/frames/{frameId}/body`.
- Фильтр `types=firebase` (и варианты) работает.
- Firebase‑сессии **не теряются** из‑за capture state (даже если recording выключен).

---

## 3) Как это ложится на текущую архитектуру

Сейчас бэк уже умеет:
- хранить `Session` и `Frame` (`internal/domain`)
- добавлять фреймы (`SessionService.AddFrame`)
- пушить live‑события в UI (`MonitorHub.Broadcast`)
- отдавать list/details/frames через `/_api/v1/sessions...`

Значит нам нужен только:
- новый ingest endpoint + handler,
- небольшая правка тегов для фильтра `types=...`,
- (важно) **идемпотентность CreateSession** для `memory.Store`.

---

## 4) Семантика Firebase RTDB “сессии”

### 4.1 Kind

- `Session.Kind = "firebase_database"`

### 4.2 Target (важно для UI и aggregate)

`computeBaseTags` и `aggregate` парсят `Session.Target` как URL. Поэтому:
- `target` должен быть **валидным URL**.

Рекомендация (для клиента):
- `target = <databaseURL>/<path>`  
  пример: `https://my-app-default-rtdb.europe-west1.firebasedatabase.app/users/123`

Для query/listen можно добавлять query‑параметры (уже канонизированные):
- `...?orderBy=key&limitToFirst=50`

### 4.3 Frames: direction/opcode

Чтобы UI мог “читать” направление:
- исходящая операция (`set/get/update/remove/transaction`) → `client->upstream`
- входящее событие подписки (`onValue/onChildAdded/...`) → `upstream->client`
- `opcode` почти всегда `"text"`

---

## 5) Ingest API (V1)

### 5.1 Endpoint

Делаем отдельный endpoint, чтобы не конфликтовать с `/_api/v1/sessions/{id}`:

- `POST /_api/v1/ingest/firebase_database`

### 5.2 Контракт запроса (минимум + расширения)

```json
{
  "session": {
    "id": "fb-<uuid>",
    "target": "https://.../users/123",
    "startedAt": "2026-02-15T20:10:00Z",
    "captureId": "current"
  },
  "frames": [
    {
      "id": "fr-<uuid>",
      "ts": "2026-02-15T20:10:01Z",
      "direction": "client->upstream",
      "opcode": "text",
      "contentType": "application/json",
      "preview": "{\"type\":\"firebase_database\",\"op\":\"set\",...}",
      "body": null,
      "bodyEncoding": null
    }
  ],
  "close": false,
  "error": null
}
```

#### Правила валидации (backend)

- **Session**
  - `session.id` обязателен, непустой, разумный лимит (например ≤ 128)
  - `session.target` обязателен, строка, лимит (например ≤ 2048)
  - `session.startedAt` опционален (если нет → `time.Now().UTC()`)
  - `session.captureId` опционален, default `"current"`
    - `"current"` или число в строке
- **Frames**
  - `frames` обязателен, 1..N (лимит N, например 200)
  - `frames[i].id` **обязателен** (мы опираемся на него для идемпотентности/дедупа в будущем)
  - `frames[i].preview` обязателен (не пустой)
  - `frames[i].direction` опционален, default определяется по `preview.op` (если `preview` JSON) иначе `"client->upstream"`
  - `frames[i].opcode` опционален, default `"text"`

### 5.3 Ответы

- `204 No Content` — success
- Ошибки через `writeError()`:
  - `400 BAD_JSON`
  - `400 VALIDATION`
  - `401 UNAUTHORIZED`
  - `409 CONFLICT`
  - `413 PAYLOAD_TOO_LARGE`
  - `500 ...`

---

## 6) Размеры, redaction, и body‑файлы

### 6.1 Preview safety

Backend делает best‑effort защиту независимо от клиента:
- если `preview` похож на JSON → прогоняем через `redact.RedactJSON`
- обрезаем `preview` до `Cfg.PreviewMaxBytes` (жёстко)

### 6.2 BodyFile для больших payload

Если клиент прислал `body`:
- декодируем по `bodyEncoding`:
  - `null` → UTF‑8 bytes
  - `"base64"` → decode base64
- если размер после декодирования > `Cfg.WSBodyMaxBytes` → `413 PAYLOAD_TOO_LARGE`
- пишем bytes в файл через `spoolBodyBytes()` (использует `Cfg.BodySpoolDir`)
- сохраняем путь в `Frame.BodyFile` через `Svc.UpdateFrameBodyFile`
- регистрируем файл через `Svc.AddSpoolFile()` (чтобы GC/удаление почистили)

---

## 7) Idempotency и конкуррентность

### 7.1 Session create

Проблема: `memory.Store.CreateSession()` сейчас не защищён от повторного create с тем же id
(может создать дубли в `order`).

**Обязательная правка для MVP**:
- сделать `CreateSession()` идемпотентным:
  - если `sess.ID` уже существует → вернуть `nil` и не менять `order/items`

Это не ломает текущий код (proxy генерит уникальные ids), но защищает ingest.

### 7.2 Frames

Сделано:
- дедуп на ingest-слое по `frame.id` (повторный POST/ретрай не даёт дубль),
- дедуп в `SessionService.AddFrame` (защита централизованно для всех источников с тем же session/frame id).

Это убирает риск двойного инкремента счётчиков и дублирования фреймов при сетевых ретраях.

---

## 8) Capture / visibility (практическая боль)

По умолчанию sessions list может скрывать “unassigned” (когда recording выключен).
Для ingest это неприемлемо — пользователь подумает, что “ничего не работает”.

Решение:
- ingest handler по умолчанию ставит `CaptureID` в “текущий” capture, даже если recording=false.

Технически:
- через type assert к `usecase.CaptureControlRepository` у underlying memory store:
  - `RecordingState() (bool, int)`
- `captureId = "current"` → `Session.CaptureID = &currentCapture`
- `captureId = "<число>"` → `Session.CaptureID = &n`

Default: `"current"`.

---

## 9) Live push в UI (monitor)

В ingest handler:
- при создании → `session_started`
- при добавлении каждого frame → `frame_added` (`ref=frameID`)
- при закрытии → `session_ended`

Если добавляем много frames одним батчем — можно слать `frame_added` на каждый
(как сейчас wsproxy), это ок: monitor уже дропает при медленных клиентах.

---

## 10) Types filter / теги

В `computeBaseTags(v sessionV1)` добавить:
- если `v.Kind == "firebase_database"`:
  - `firebase`, `rtdb`, `firebase_database`

Плюс эвристика по host:
- `*.firebaseio.com`, `*.firebasedatabase.app` → `firebase`

---

## 11) Безопасность

Нужно, чтобы работало с телефона/эмулятора, но не было “дырой”.

Правило:
- если `ADMIN_TOKEN` задан → требуем `X-Admin-Token`
- если `ADMIN_TOKEN` пустой:
  - разрешаем только loopback или private LAN (10/172.16/192.168)
  - для осознанного открытия наружу: `INGEST_ALLOW_REMOTE=1`

Реализация: проверяем `r.RemoteAddr` (без доверия к `X-Forwarded-For`).

---

## 12) Реализация (конкретные точки)

### 12.1 Router

`internal/infrastructure/httpapi/router.go`:
- зарегистрировать `POST /_api/v1/ingest/firebase_database`

### 12.2 Handler

Новый файл:
- `internal/infrastructure/httpapi/firebase_database_ingest.go`

Алгоритм handler’а (псевдокод):

1) auth check (см. раздел 11)  
2) `MaxBytesReader` на request body (например 16MB)  
3) `json.Decode` в `firebaseIngestRequest`  
4) validate session + frames  
5) ensure session exists:
   - `Svc.Get(id)` → если нет → `Svc.Create(session)`
6) принудительно выставить `CaptureID` (раздел 8)  
7) for each frame:
   - sanitize preview (redact + trim)
   - `Svc.AddFrame(...)`
   - если `body != null`: spool → `UpdateFrameBodyFile` → `AddSpoolFile`
   - broadcast `frame_added`
8) close:
   - `Svc.SetClosed(...)`
   - broadcast `session_ended`

### 12.3 Store idempotency

`internal/adapters/storage/memory/store.go`:
- `CreateSession`: если `items[sess.ID]` уже есть → return nil

### 12.4 Tags

`internal/infrastructure/httpapi/sessions_handlers.go`:
- `computeBaseTags` добавить теги (раздел 10)

---

## 13) Тесты

Новый файл:
- `internal/infrastructure/httpapi/firebase_database_ingest_test.go`

Кейсы:
- create + append + close
- validation errors
- auth behavior (token / local network)
- body spool + чтение `/frames/{id}/body`
- `types=firebase` фильтр
- visibility при `recording=false` (captureId=current)
- idempotency `CreateSession` (повторный POST не даёт дублей в list)

---

## 14) Риски и компромиссы

- **длинные подписки**: память ограничена `maxFramesPerSession` (drop-from-head) — ожидаемо
- **frame дедуп**: сделан в ingest-хендлере по `frame.id` (повторные ретраи не плодят дублей)
- **токены/секреты**: редактируем preview, но если клиент кладёт секреты в body — это ответственность клиента

---

## 15) Follow-ups

- WebSocket ingest (меньше overhead и лучше realtime)
- дедуп frames на стороне memory store (сейчас дедуп уже есть на ingest-слое)
- UI‑виджеты для Firebase (path/op/diff)
- План фронта + клиентской интеграции: [`firebase_database_sessions_frontend.md`](./firebase_database_sessions_frontend.md)

