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
</p>

<p align="center">
  <a href="https://pub.dev/packages/network_debugger"><img src="https://img.shields.io/pub/v/network_debugger.svg" alt="network_debugger on pub.dev" /></a>
  <a href="https://pub.dev/packages/dio_debugger"><img src="https://img.shields.io/pub/v/dio_debugger.svg" alt="dio_debugger on pub.dev" /></a>
  <a href="#"><img src="https://img.shields.io/badge/web_socket_debugger-pending-lightgrey" alt="web_socket_debugger" /></a>
  <a href="#"><img src="https://img.shields.io/badge/web_socket_channel_debugger-pending-lightgrey" alt="web_socket_channel_debugger" /></a>
</p>

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
- `DEFAULT_TARGET` — default target upstream
- `CAPTURE_BODIES` — save request/response bodies (1/true)
- `RESPONSE_DELAY_MS` — fixed or range, e.g. `1000` or `1000-3000`
- `INSECURE_TLS` — trust self-signed certificates (1/true)

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
