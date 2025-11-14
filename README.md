# Network Debugger

<p align="center">
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/cherrypick-agency/flutter_network_debugger" alt="License" />
  </a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/coverage.yml">
    <img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/cherrypick-agency/flutter_network_debugger/gh-pages/coverage.json" alt="coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/cherrypick-agency/flutter_network_debugger">
    <img src="https://goreportcard.com/badge/github.com/cherrypick-agency/flutter_network_debugger" alt="Go Report Card" />
  </a>

  <br>

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
</p>

<img width="1312" height="815" alt="image" src="https://github.com/user-attachments/assets/d9c95c20-f79b-45da-94a7-c341cd33a388" />
<img width="500"  alt="image" src="https://github.com/user-attachments/assets/43044ece-e6b4-4702-80bc-0584e844c042"  />

<!--![Запись экрана 2025-10-02 в 13 06 06](https://github.com/user-attachments/assets/43044ece-e6b4-4702-80bc-0584e844c042)-->

Free tool for debugging HTTP, WebSocket, SOCKS which is MUCH BETTER than the built-in Flutter Netwrok Devtools.

Suitable for local development and test environments. Has crossplatform interface: WEB, desktop (MacOS/Windows/Linux) and CLI (with code highlight, many options).

### Features
- Intercept and view HTTP(S) and WebSockets/Socket.io traffic (WS supports very nice)
- Waterfall timeline of requests
- grouping by domain/route
- Filters: method, status, MIME, minimum duration, by headers...
- Convenient search with highlighting
- HTTP details: headers (with sensitive data masking), body (pretty/JSON tree), TTFB/Total
- CORS/Cache hints, cookies and TLS summary
- WebSocket details: events/frames, pings/pongs, payload preview, json highligh
- HAR/Curl export
- Artificial response delay (useful for simulating "slow networks")
- Record/stop and records management
- HTML responses preview
- Form Data (show files) For example Flutter devtools don't show at all
- You can proxy only app requests or all OS requests (forward proxy)
- Crossplatform (WEB, Desktop, CLI)
- The fastest GO backend for processing requests compared to competitors
- 10,000+ req/sec throughput (5x faster than Charles, 2x faster than Proxyman)
- An independent GO backend that can be run anywhere
- Well covered with tests
- Compose / Request Builder + Library
- Mapping (Map Local / Map Remote)
- Throttling with profiles
- Tags & Annotations: Tag and annotate sessions for better organization
- Performance Insights dashboard: real-time latency, throughput, error rates, and endpoint hotspots (beta)
- Privacy first. Works offline.
- Scriping (language agnostic! Go, Rust, JS/TS, C/C++, even Swift, Kotlin, Dart and more) Currently in beta. Plugins system is planned.

...

### Quick start

#### Desktop App (Native GUI)

**Download standalone desktop application** for macOS, Windows, or Linux from [GitHub Releases](https://github.com/cherrypick-agency/flutter_network_debugger/releases):

- **macOS**: Download `.dmg` for your architecture (Intel or Apple Silicon)
  - Open DMG and drag app to Applications
  - Auto-update support via GitHub Releases

- **Windows**: Download `.zip` archive
  - Extract and run `install.bat`
  - Creates shortcuts on Desktop and Start Menu

- **Linux**: Download `.deb` or `.tar.gz`
  ```bash
  # Debian/Ubuntu
  sudo dpkg -i network-debugger_*_amd64.deb

  # Other distros
  tar -xzf NetworkDebugger-*-linux-amd64.tar.gz
  cd NetworkDebugger-*
  ./install.sh
  ```

Desktop app features:
- Native UI with Flutter
- Integrated Go proxy server (single process)
- Startup dialog for port configuration
- Auto-update from GitHub Releases
- Cross-platform support

See [docs/DESKTOP_SETUP.md](docs/DESKTOP_SETUP.md) for detailed setup and development guide.

#### CLI (automatically downloads binary)

- Via CLI tool:
  ```bash
  dart pub global activate network_debugger
  network_debugger
  ```

#### Docker

- Docker Compose:
  ```bash
  docker compose -f deploy/docker-compose.yml up -d
  ```

#### From Source (Go)

- Build from source:
  ```bash
  # server/desktop binary
  go build -o ./network-debugger ./cmd/network-debugger
  ./network-debugger

  # web version that opens browser automatically
  go build -o ./network-debugger-web ./cmd/network-debugger-web
  ./network-debugger-web
  ```

### CLI sessions mode (colored console output)

Run server with CLI output enabled (no browser auto‑open):

```bash
./network-debugger --cli --cli-preset basic
# or fine‑tune fields:
./network-debugger --cli \
  --cli-fields line,sizes,timings,req-headers,resp-headers \
  --cli-color auto \
  --cli-filter "/api/" \
  --cli-body-bytes 50000
```

Presets: `minimal | basic | advanced | full` (fields listed below). `--cli-fields` overrides preset entirely.
Color modes: `auto | always | never`. Body preview default uses `PREVIEW_MAX_BYTES`.

Flags
- `--cli`: enable CLI mode (disables auto‑open browser)
- `--cli-preset`: one of `minimal|basic|advanced|full`
- `--cli-fields`: comma‑separated list of sections to show (overrides preset)
- `--cli-body-bytes`: body preview limit (bytes); 0 = use `PREVIEW_MAX_BYTES`
- `--cli-color`: `auto|always|never`
- `--cli-filter`: substring filter (matches URL/method/status)

Fields (sections)
- `line`: single‑line summary (time, METHOD, URL, STATUS, totalMs, sizes)
- `sizes`: request/response byte sizes (also included in `line` summary)
- `timings`: HTTP timings (DNS/Connect/TLS/TTFB/Total) when available
- `req-headers`, `resp-headers`: selected headers with masking of sensitive values
- `req-body`, `resp-body`: pretty JSON or raw preview, trimmed by `--cli-body-bytes`
- `tls`: TLS peer info (version/cipher/ALPN/certs summary) when available
- `cookies`: Set‑Cookie flags summary (counts of Secure/HttpOnly/SameSite)
- `ids`: internal IDs (session/tx) for cross‑referencing

Presets mapping
- minimal: `line`
- basic: `line,sizes`
- advanced: `line,sizes,timings,req-headers,resp-headers`
- full: `advanced + req-body,resp-body,tls,cookies,ids`

Notes
- HTTP request/response bodies shown are previews; they may be truncated and/or decompressed for readability.
- Sensitive headers are masked by default; enable raw exposure via server config if needed.
- Colors are enabled automatically for TTY; force with `--cli-color always`.

Examples
```bash
# Show only single-line summaries, always with colors
./network-debugger --cli --cli-preset minimal --cli-color always

# Full details but limit body previews to 16KB
./network-debugger --cli --cli-preset full --cli-body-bytes 16384

# Custom selection with filtering for API routes
./network-debugger --cli \
  --cli-fields line,timings,req-headers,resp-headers \
  --cli-filter "/v1/"
```

Where UI opens
- By default server listens on :9092 (UI), proxy (forward) is on :9091:
  - UI: http://localhost:9092/
  - Proxy base (HTTP/WebSocket forward): http://localhost:9091


### Dart Packages

| Package                                                                             | Version                                                                                                                      | Pub Points                                                                                                                                 | Description                            |
| ----------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------- |
| [network_debugger](https://pub.dev/packages/network_debugger)                       | [![pub](https://img.shields.io/pub/v/network_debugger.svg)](https://pub.dev/packages/network_debugger)                       | [![pub points](https://img.shields.io/pub/points/network_debugger)](https://pub.dev/packages/network_debugger/score)                       | Core CLI tool for starting the proxy   |
| [dio_debugger](https://pub.dev/packages/dio_debugger)                               | [![pub](https://img.shields.io/pub/v/dio_debugger.svg)](https://pub.dev/packages/dio_debugger)                               | [![pub points](https://img.shields.io/pub/points/dio_debugger)](https://pub.dev/packages/dio_debugger/score)                               | Interceptor for Dio HTTP client        |
| [http_debugger](https://pub.dev/packages/http_debugger)                             | [![pub](https://img.shields.io/pub/v/http_debugger.svg)](https://pub.dev/packages/http_debugger)                             | [![pub points](https://img.shields.io/pub/points/http_debugger)](https://pub.dev/packages/http_debugger/score)                             | Wrapper for package:http client        |
| [web_socket_debugger](https://pub.dev/packages/web_socket_debugger)                 | [![pub](https://img.shields.io/pub/v/web_socket_debugger.svg)](https://pub.dev/packages/web_socket_debugger)                 | [![pub points](https://img.shields.io/pub/points/web_socket_debugger)](https://pub.dev/packages/web_socket_debugger/score)                 | Wrapper for dart:io WebSocket          |
| [web_socket_channel_debugger](https://pub.dev/packages/web_socket_channel_debugger) | [![pub](https://img.shields.io/pub/v/web_socket_channel_debugger.svg)](https://pub.dev/packages/web_socket_channel_debugger) | [![pub points](https://img.shields.io/pub/points/web_socket_channel_debugger)](https://pub.dev/packages/web_socket_channel_debugger/score) | Wrapper for package:web_socket_channel |
| [socket_io_debugger](https://pub.dev/packages/socket_io_debugger)                   | [![pub](https://img.shields.io/pub/v/socket_io_debugger.svg)](https://pub.dev/packages/socket_io_debugger)                   | [![pub points](https://img.shields.io/pub/points/socket_io_debugger)](https://pub.dev/packages/socket_io_debugger/score)                   | Wrapper for socket.io client           |


### Settings

Main settings (ENV)
- `ADDR` — server address (default :9092)
- `DEV_MODE` — development mode (1/true)
- `NO_BROWSER` — disable automatic browser opening (1/true)
- `DEFAULT_TARGET` — default target upstream
- `CAPTURE_BODIES` — save request/response bodies (1/true)
- `RESPONSE_DELAY_MS` — fixed or range, e.g. `1000` or `1000-3000`
- `INSECURE_TLS` — trust self-signed certificates (1/true)

Network throttling (bandwidth/reliability/latency)
- `THROTTLE_ENABLE` — enable throttling globally (1/true)
- `THROTTLE_DOWN_KBPS` — downstream speed in kbit/s (server→client)
- `THROTTLE_UP_KBPS` — upstream speed in kbit/s (client→server)
- `THROTTLE_PACKET_LOSS` — packet loss percent (0..100), best‑effort
- `THROTTLE_LATENCY_MS` — base latency in ms (RTT/ping simulation)
- `THROTTLE_LATENCY_JITTER` — random jitter ± ms (e.g., 20 = varies ±20ms)
- `THROTTLE_OFFLINE` — simulate offline (reject new requests)

Runtime API
- `GET /_api/v1/throttle` — current values and preset hints
- `POST /_api/v1/throttle` — set values
  ```json
  {"enabled":true,"downKbps":3000,"upKbps":3000,"packetLossPct":0,"latencyMs":100,"latencyJitter":20,"offline":false}
  ```

SOCKS/HTTP Proxy Runtime (ports)
- Управляется из UI: Settings → Proxy
- API:
  - `GET /_api/v1/proxy/config`
  - `POST /_api/v1/proxy/config`
    ```json
    {
      "forward": {"enabled": true, "port": 8888},
      "socks": {"enabled": true, "port": 8889, "authMode": "none"}
    }
    ```
  - Изменения применяются на лету без рестарта процесса (graceful shutdown/start).

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
Database migrations
- Local/dev: AutoMigrate is enabled only when `DEV_MODE=1`.
- Prod/Test: apply SQL migrations from `./migrations` using goose or golang-migrate in CI/CD.
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
- `PREVIEW_MAX_BYTES` — preview limit for text payloads (default 50000)
- `WS_PREVIEW_MAX_BYTES` — WS preview limit (fallback to PREVIEW_MAX_BYTES)
- `WS_DEFLATE_PREVIEW` — try to decompress permessage-deflate for preview (default 1)
- `WS_CAPTURE_BODIES` — save WS message bodies to spool (default 0)
- `WS_BODY_MAX_BYTES` — spool size limit for WS message body (default 1 MiB)

### Local development (without GitHub)
- Ready binary/archive in `./dist`:
  ```bash
  network-debugger --local-dir ./dist --no-remote
  ```
- Local artifacts server:
  ```bash
  network-debugger serve-artifacts --dir ./dist --port 8099
  network-debugger --base-url http://127.0.0.1:8099 --no-remote
  ```

### TODO
- Plugins marketplace (write/install plugins on almost any language!)
- Analytics (alredy started)
- 

### Useful to know
- Binary cache: macOS/Linux `~/.cache/network_debugger/`, Windows `%LOCALAPPDATA%\network_debugger\Cache\`
- Binary name: `network-debugger-web` (Windows — `network-debugger-web.exe`)

<!-- https://github.com/flutter/devtools/issues/8223 -->
