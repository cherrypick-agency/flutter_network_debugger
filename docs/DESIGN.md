# Design Doc 

Goals: local HTTP/WS proxy inspector (incl. Socket.IO) with minimal overhead, live updates and simple API/UI.

Flow: client ⇄ network-debugger ⇄ upstream (http/https/ws/wss). Proxy mirrors traffic, writes frame/event previews and metadata to in-memory store, provides REST for sessions/frames/events/HTTP transactions and monitoring via WS/SSE.

Entities: Session, Frame, Event, HTTPTransaction.
- Session.kind: ws|http
- Frame: direction, opcode, size, preview (truncation; editing sensitive fields)
- Event: Socket.IO best-effort parser (v4, partially v3)
- HTTPTransaction: method, status, mime, sizes, timings (DNS/Connect/TLS/TTFB/Total)

Architecture: Clean Architecture (domain/usecase/adapters/infrastructure). Interfaces at consumer (usecase). Storage: memory (ring buffer + TTL).

Key services:
- Reverse proxy: `GET /httpproxy[/path]?_target=<url>`
- WS proxy: `GET /wsproxy?_target=<ws(s)://...>`
- Unified: `GET /proxy` — determines by Upgrade (ws → WS proxy; otherwise HTTP reverse)
- Sessions REST:
  - `GET /_api/v1/sessions?limit&offset&q&_target` — sessions list (with httpMeta/sizes)
  - `GET /_api/v1/sessions/{id}` — details, `DELETE` — deletion
  - `GET /_api/v1/sessions/{id}/frames|events|http` — cursor selections
  - `GET /_api/v1/sessions/aggregate?groupBy=domain` — simple aggregation
  - SSE: `GET /api/sessions_stream/{id}` (live updates for specific session)
- Monitor WS: `/_api/v1/monitor/ws` (global events)
- Capture control: `POST /_api/v1/capture {action:start|stop}`; `GET /_api/v1/captures` (history/status)
- Settings: `GET /_api/v1/settings` (runtime settings: response delays etc.)

Notable decisions:
- Proxy query parameter: only `_target` (to avoid collisions with user queries)
- Masking sensitive headers in preview (Authorization/Cookie/*token*/...)
- Preview body threshold/truncation: `PREVIEW_MAX_BYTES` (default 4096)
- Artificial response delay for timeline visualization: `RESPONSE_DELAY_MS` (or min-max range)
- CORS simplified: `Access-Control-Allow-Origin` on entire API (dev mode)

Constraints: local debugging (no auth/TLS by default), >=10k msg/min, average overhead <5ms, in-memory only.

Risks/Workarounds:
- Socket.IO v4/v3 differences: use best-effort parser + heuristics (ack/id, 42/43 prefixes)
- Memory: limited buffers, TTL-eviction, truncated previews
- Backpressure: monitor channel buffer, drop on slow consumers

Build/Run:
- Web frontend embedded in binary (go:embed). Single binary runs API and SPA.
- Default port: `ADDR=:9091`. Static and API routed (API before static).
