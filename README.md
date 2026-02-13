# Network Debugger

<p align="center">
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License" />
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
<img width="180" alt="image" src="https://github.com/user-attachments/assets/3eb43e0d-e0ce-4c6d-a0c2-0fb5985ad8f9" />
<img width="180" alt="image" src="https://github.com/user-attachments/assets/029cecea-cde6-466b-9e5c-06f5a78bba50" />
<img width="180" height="815" alt="image" src="https://github.com/user-attachments/assets/c9b753a8-874a-414d-a628-4b794e1e3319" />
<img width="180"  alt="image" src="https://github.com/user-attachments/assets/43044ece-e6b4-4702-80bc-0584e844c042"  />


<!-- Screen recording (2025-10-02) -->

Free tool for debugging HTTP, **WebSocket** (killer feature), SOCKS which is MUCH BETTER than the built-in Flutter Netwrok Devtools.

Suitable for local development and test environments. Has crossplatform interface: WEB, desktop (MacOS/Windows/Linux) and CLI (with code highlight, many options).

### Features
- Intercept and view HTTP(S) and WebSockets/Socket.io traffic (WS supports very nice)
- Waterfall timeline of requests
- grouping by domain/route
- Filters: method, status, MIME, minimum duration, by headers...
- Convenient search with highlighting
- HTTP details: headers (with sensitive data masking), body (pretty/JSON tree/HEX), TTFB/Total...
- CORS/Cache hints, cookies and TLS summary
- WebSocket details: events/frames, pings/pongs, payload preview, json highligh, nice global/local search
- import/export HAR
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
- Compose / Request Builder + Tree Library
- Edit requests using Compose interface
- Mapping (Map Local / Map Remote)
- Throttling with profiles
- Tags & Annotations: Tag and annotate sessions for better organization
- Performance Insights dashboard: real-time latency, throughput, error rates, and endpoint hotspots (beta)
- Privacy first. Works offline.
- Scriping (language agnostic! Go, Rust, JS/TS, C/C++, even Swift, Kotlin, Dart and more) Currently in beta. Plugins system is planned.

...

### Install the right Dart package (integration)

If you want to capture traffic from your Flutter/Dart app, install the package
that matches your networking stack:

- **Dio**: use `dio_debugger`

```bash
dart pub add dio_debugger
```

- **package:http**: use `http_debugger`

```bash
dart pub add http_debugger
```

- **WebSocket (killer feature)**: pick the one you use in the app:
  - `web_socket_debugger` (dart:io WebSocket)
  - `web_socket_channel_debugger` (package:web_socket_channel)
  - `socket_io_debugger` (Socket.IO client)

```bash
dart pub add web_socket_debugger
# or
dart pub add web_socket_channel_debugger
# or
dart pub add socket_io_debugger
```

See the full list in the [Dart Packages](#dart-packages) section below.

### Quick start

1) CLI (WEB UI in browser) — fastest way to start

Via CLI tool:

```bash
dart pub global activate network_debugger
network_debugger
```

This starts the proxy and opens the UI in your browser:
- UI: `http://localhost:9092/`
- Proxy base (HTTP/WebSocket forward): `http://localhost:9091`

2) Desktop App (Native GUI)

**Download standalone desktop application** for macOS, Windows, or Linux from [GitHub Releases](https://github.com/cherrypick-agency/flutter_network_debugger/releases):

- **macOS**: Download `.dmg` for your architecture (Intel or Apple Silicon)
  - Open DMG and drag app to Applications
  - Note: the app is not signed/notarized yet, so macOS may show a security
    prompt on first launch. If it gets blocked, go to System Settings →
    Privacy & Security → **Open Anyway**
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
- OS Forward Proxy mode (system-wide proxy)
- Startup dialog for port configuration
- Auto-update from GitHub Releases
- Cross-platform support

See [docs/DESKTOP_SETUP.md](docs/DESKTOP_SETUP.md) for detailed setup and development guide.

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

See [docs/CLI_SESSIONS_MODE.md](docs/CLI_SESSIONS_MODE.md).


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

See [docs/SETTINGS.md](docs/SETTINGS.md).

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
