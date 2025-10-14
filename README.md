# Network Debugger

![Запись экрана 2025-10-02 в 13 06 06](https://github.com/user-attachments/assets/43044ece-e6b4-4702-80bc-0584e844c042)


Simple universal proxy for debugging HTTP and WebSocket. Suitable for local development and test environments. Has web interface (opens in browser), desktop and CLI.

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
- Record/stop and record management
...

Quick start
- Via CLI (automatically downloads binary and opens UI):
  ```bash
  dart pub global activate network_debugger_cli
  network-debugger
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
- Binary cache: macOS `~/Library/Caches/network-debugger/bin-cache`, Linux `~/.cache/network-debugger/bin-cache`, Windows `%LOCALAPPDATA%/network-debugger/bin-cache`
- Binary name: `network-debugger-web` (Windows — `network-debugger-web.exe`)
