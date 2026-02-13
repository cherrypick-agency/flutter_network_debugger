# Settings

This document describes the runtime configuration options for Network Debugger.

## Main settings (ENV)

- `ADDR` — server address (default `:9092`)
- `DEV_MODE` — development mode (`1`/`true`)
- `NO_BROWSER` — disable automatic browser opening (`1`/`true`)
- `OPEN_BROWSER` — open browser on start for `network-debugger` (`1`/`true`);
  `network-debugger-web` opens by default
- `DEFAULT_TARGET` — default upstream target
- `CAPTURE_BODIES` — save request/response bodies (`1`/`true`)
- `RESPONSE_DELAY_MS` — fixed value or range, e.g. `1000` or `1000-3000`
- `INSECURE_TLS` — trust self-signed certificates (`1`/`true`)

## Network throttling (bandwidth / reliability / latency)

- `THROTTLE_ENABLE` — enable throttling globally (`1`/`true`)
- `THROTTLE_DOWN_KBPS` — downstream speed in kbit/s (server→client)
- `THROTTLE_UP_KBPS` — upstream speed in kbit/s (client→server)
- `THROTTLE_PACKET_LOSS` — packet loss percent (0..100), best-effort
- `THROTTLE_LATENCY_MS` — base latency in ms (RTT/ping simulation)
- `THROTTLE_LATENCY_JITTER` — random jitter ± ms (e.g. `20` = varies ±20ms)
- `THROTTLE_OFFLINE` — simulate offline (reject new requests)

## Runtime API

- `GET /_api/v1/throttle` — current values and preset hints
- `POST /_api/v1/throttle` — set values

Example:

```json
{"enabled":true,"downKbps":3000,"upKbps":3000,"packetLossPct":0,"latencyMs":100,"latencyJitter":20,"offline":false}
```

## SOCKS / HTTP Proxy runtime (ports)

Configured from UI: Settings → Proxy.

API:
- `GET /_api/v1/proxy/config`
- `POST /_api/v1/proxy/config`

Example:

```json
{
  "forward": {"enabled": true, "port": 8888},
  "socks": {"enabled": true, "port": 8889, "authMode": "none"}
}
```

Changes are applied on-the-fly (graceful shutdown/start), no process restart is
required.

## Cookies and stealth (reverse proxy /httpproxy)

- `STEALTH_HEADERS` — hide proxy headers (Via, X-Forwarded-*) on /httpproxy
  (default `1`)
- `COOKIES_MODE` — `isolate` | `auto` | `off` (default `isolate`)
- `COOKIES_DOMAIN_STRATEGY` — `hostOnly` | `proxyHost` (default `hostOnly`)
- `COOKIES_PATH_STRATEGY` — `prefix` | `root` (default `prefix`)

Per-request overrides (query params):
- `_cookie_mode=isolate|auto|off`
- `_stealth=1|0`

Notes:
- For `SameSite=None` and `__Secure-`/`__Host-` cookies to be accepted by the
  browser, client→proxy must be HTTPS.
- In `isolate` mode cookie names are namespaced in the browser storage and
  unwrapped towards upstream, so different `_target` do not collide.

## WebSocket preview settings

- `PREVIEW_MAX_BYTES` — preview limit for text payloads (default `50000`)
- `WS_PREVIEW_MAX_BYTES` — WebSocket preview limit (fallback to
  `PREVIEW_MAX_BYTES`)
- `WS_DEFLATE_PREVIEW` — try to decompress permessage-deflate for preview
  (default `1`)
- `WS_CAPTURE_BODIES` — save WebSocket message bodies to spool (default `0`)
- `WS_BODY_MAX_BYTES` — spool size limit for WebSocket message body (default
  `1 MiB`)

## Database migrations

- Local/dev: AutoMigrate is enabled only when `DEV_MODE=1`.
- Prod/Test: apply SQL migrations from `./migrations` using `goose` or
  `golang-migrate` in CI/CD.

Example (goose):

```bash
# install once: go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir ./migrations sqlite3 ./data/network_debugger.db up
```

Example (golang-migrate):

```bash
# install once: brew install golang-migrate
migrate -path ./migrations -database sqlite3://./data/network_debugger.db up
```

