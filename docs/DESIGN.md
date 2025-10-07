# Design Doc 

Goals: локальный HTTP/WS прокси‑инспектор (в т.ч. Socket.IO) с минимальными оверхедами, живыми апдейтами и простым API/UI.

Flow: client ⇄ network-debugger ⇄ upstream (http/https/ws/wss). Прокси зеркалит трафик, пишет превью фреймов/событий и метаданные в in‑memory store, даёт REST для сессий/фреймов/событий/HTTP‑транзакций и мониторинг по WS/SSE.

Entities: Session, Frame, Event, HTTPTransaction.
- Session.kind: ws|http
- Frame: направление, opcode, size, preview (обрезка; редактирование чувствительных полей)
- Event: Socket.IO best‑effort парсер (v4, частично v3)
- HTTPTransaction: метод, статус, mime, размеры, тайминги (DNS/Connect/TLS/TTFB/Total)

Architecture: Clean Architecture (domain/usecase/adapters/infrastructure). Interfaces у потребителя (usecase). Хранилище: память (ring buffer + TTL).

Key services:
- Reverse proxy: `GET /httpproxy[/path]?_target=<url>`
- WS proxy: `GET /wsproxy?_target=<ws(s)://...>`
- Unified: `GET /proxy` — определяет по Upgrade (ws → WS proxy; иначе HTTP reverse)
- Sessions REST:
  - `GET /_api/v1/sessions?limit&offset&q&_target` — список сессий (с httpMeta/sizes)
  - `GET /_api/v1/sessions/{id}` — детали, `DELETE` — удаление
  - `GET /_api/v1/sessions/{id}/frames|events|http` — курсорные выборки
  - `GET /_api/v1/sessions/aggregate?groupBy=domain` — простая агрегация
  - SSE: `GET /api/sessions_stream/{id}` (live апдейты для конкретной сессии)
- Monitor WS: `/_api/v1/monitor/ws` (глобальные события)
- Capture control: `POST /_api/v1/capture {action:start|stop}`; `GET /_api/v1/captures` (история/статус)
- Settings: `GET /_api/v1/settings` (runtime настройки: задержки ответа и т.п.)

Notable decisions:
- Query параметр прокси: только `_target` (во избежание коллизий с пользовательскими query)
- Маскирование чувствительных заголовков в превью (Authorization/Cookie/*token*/...)
- Порог/обрезка превью тел: `PREVIEW_MAX_BYTES` (по умолчанию 4096)
- Искусственная задержка ответа для визуализации таймлайна: `RESPONSE_DELAY_MS` (или min-max диапазон)
- CORS упрощённый: `Access-Control-Allow-Origin` на весь API (dev‑режим)

Constraints: локальная отладка (без auth/TLS по умолчанию), >=10k msg/min, средний оверхед <5ms, in‑memory only.

Risks/Workarounds:
- Socket.IO v4/v3 различия: используем best‑effort парсер + эвристики (ack/id, 42/43 префиксы)
- Память: ограниченные буферы, TTL‑эвикшн, усечённые превью
- Backpressure: буфер канала монитора, дроп на медленных потребителях

Build/Run:
- Веб‑фронт встраивается в бинарь (go:embed). Один бинарь запускает API и SPA.
- Порт по умолчанию: `ADDR=:9091`. Статика и API разруливаются роутером (API до статических).
