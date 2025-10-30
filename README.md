# Network Debugger

<p align="center">
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/network_debugger.yml">
    <img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/network_debugger.yml/badge.svg?branch=main" alt="network_debugger CI" />
  </a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/dio_debugger.yml">
    <img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/dio_debugger.yml/badge.svg?branch=main" alt="dio_debugger CI" />
  </a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_debugger.yml">
    <img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_debugger.yml/badge.svg?branch=main" alt="web_socket_debugger CI" />
  </a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_channel_debugger.yml">
    <img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_channel_debugger.yml/badge.svg?branch=main" alt="web_socket_channel_debugger CI" />
  </a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/socket_io_debugger.yml">
    <img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/socket_io_debugger.yml/badge.svg?branch=main" alt="socket_io_debugger CI" />
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/cherrypick-agency/flutter_network_debugger" alt="License" />
  </a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/coverage.yml">
    <img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/cherrypick-agency/flutter_network_debugger/gh-pages/coverage.json" alt="coverage" />
  </a>
</p>

### Dart Packages

| Package | Version | Description |
|---------|---------|-------------|
| [network_debugger](https://pub.dev/packages/network_debugger) | [![pub](https://img.shields.io/pub/v/network_debugger.svg)](https://pub.dev/packages/network_debugger) | Core CLI tool for starting the proxy |
| [dio_debugger](https://pub.dev/packages/dio_debugger) | [![pub](https://img.shields.io/pub/v/dio_debugger.svg)](https://pub.dev/packages/dio_debugger) | Interceptor for Dio HTTP client |
| [http_debugger](https://pub.dev/packages/http_debugger) | [![pub](https://img.shields.io/pub/v/http_debugger.svg)](https://pub.dev/packages/http_debugger) | Wrapper for package:http client |
| [web_socket_debugger](https://pub.dev/packages/web_socket_debugger) | [![pub](https://img.shields.io/pub/v/web_socket_debugger.svg)](https://pub.dev/packages/web_socket_debugger) | Wrapper for dart:io WebSocket |
| [web_socket_channel_debugger](https://pub.dev/packages/web_socket_channel_debugger) | [![pub](https://img.shields.io/pub/v/web_socket_channel_debugger.svg)](https://pub.dev/packages/web_socket_channel_debugger) | Wrapper for package:web_socket_channel |
| [socket_io_debugger](https://pub.dev/packages/socket_io_debugger) | [![pub](https://img.shields.io/pub/v/socket_io_debugger.svg)](https://pub.dev/packages/socket_io_debugger) | Wrapper for socket.io client |

![Запись экрана 2025-10-02 в 13 06 06](https://github.com/user-attachments/assets/43044ece-e6b4-4702-80bc-0584e844c042)

Free tool for debugging HTTP and WebSocket which is MUCH BETTER than the built-in Flutter Netwrok Devtools. 

Suitable for local development and test environments. Has web interface (opens in browser), desktop and CLI.

What it can do
- Intercept and view HTTP and WebSocket traffic
- Waterfall timeline of requests
- grouping by domain/route
- Filters: method, status, MIME, minimum duration, by headers
- Convenient search with highlighting
- HTTP details: headers (with sensitive data masking), body (pretty/JSON tree), TTFB/Total
- CORS/Cache hints, cookies and TLS summary
- WebSocket details: events/frames, pings/pongs, payload preview
- HAR export
- Artificial response delay (useful for simulating "slow networks")
- Record/stop and records management
- HTML preview
- Form Data (show files) For example Flutter devtools don't show at all
- You can proxy only app requests or all OS requests (forward proxy)
- Crossplatform

...

Quick start
- Via CLI (automatically downloads binary and opens UI):
  ```bash
  dart pub global activate network_debugger
  network_debugger
  ```
- Docker:
  ```bash
  docker compose -f deploy/docker-compose.yml up -d
  ```
- From source (Go):
  ```bash
  # server/desktop binary
  go build -o ./network-debugger ./cmd/network-debugger
  ./network-debugger

  # web version that opens browser automatically
  go build -o ./network-debugger-web ./cmd/network-debugger-web
  ./network-debugger-web
  ```

Where UI opens
- By default server listens on :9091, UI is available at:
  - http://localhost:9091/_ui (or root if auto-redirect is enabled)

Main settings (ENV)
- `ADDR` — server address (default :9091)
- `DEV_MODE` — development mode (1/true)
- `NO_BROWSER` — disable automatic browser opening (1/true)
- `DEFAULT_TARGET` — default target upstream
- `CAPTURE_BODIES` — save request/response bodies (1/true)
- `RESPONSE_DELAY_MS` — fixed or range, e.g. `1000` or `1000-3000`
- `INSECURE_TLS` — trust self-signed certificates (1/true)

Cookies and stealth (reverse proxy /httpproxy)
- `STEALTH_HEADERS` — hide proxy headers (Via, X-Forwarded-*) on /httpproxy (default 1)
- `COOKIES_MODE` — `isolate` | `auto` | `off` (default `isolate`)
- `COOKIES_DOMAIN_STRATEGY` — `hostOnly` | `proxyHost` (default `hostOnly`)
- `COOKIES_PATH_STRATEGY` — `prefix` | `root` (default `prefix`)

Per-request overrides (query params)
- `_cookie_mode=isolate|auto|off`
- `_stealth=1|0`

Notes
- For `SameSite=None` and `__Secure-`/`__Host-` cookies to be accepted by the browser, client→proxy must be HTTPS.
- In `isolate` mode cookie names are namespaced in the browser storage and unwrapped towards upstream, so different `_target` do not collide.

WebSocket preview settings
- `PREVIEW_MAX_BYTES` — preview limit for text payloads (default 50000)
- `WS_PREVIEW_MAX_BYTES` — WS preview limit (fallback to PREVIEW_MAX_BYTES)
- `WS_DEFLATE_PREVIEW` — try to decompress permessage-deflate for preview (default 1)
- `WS_CAPTURE_BODIES` — save WS message bodies to spool (default 0)
- `WS_BODY_MAX_BYTES` — spool size limit for WS message body (default 1 MiB)

Local development (without GitHub)
- Ready binary/archive in `./dist`:
  ```bash
  network-debugger --local-dir ./dist --no-remote
  ```
- Local artifacts server:
  ```bash
  network-debugger serve-artifacts --dir ./dist --port 8099
  network-debugger --base-url http://127.0.0.1:8099 --no-remote
  ```

Useful to know
- Binary cache: macOS/Linux `~/.cache/network_debugger/`, Windows `%LOCALAPPDATA%\network_debugger\Cache\`
- Binary name: `network-debugger-web` (Windows — `network-debugger-web.exe`)
