# web_socket_debugger

<p align="left">
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_debugger.yml"><img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_debugger.yml/badge.svg?branch=main" alt="CI" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License" /></a>
</p>

## Getting started with Network Debugger

1) CLI (WEB UI in browser) — fastest way to start

```bash
dart pub global activate network_debugger
network_debugger
```

This starts the proxy and opens the UI in your browser:
- UI: `http://localhost:9092/`
- Proxy base (HTTP/WebSocket forward): `http://localhost:9091`

2) Desktop App (Native GUI)

Download the desktop application from
[GitHub Releases](https://github.com/cherrypick-agency/flutter_network_debugger/releases).
It bundles the proxy server and UI.

> This package is for integrating Network Debugger with **WebSockets**.

A helper package to attach network-debugger proxy to `package:web_socket` for local debugging and WebSocket traffic interception.

## Installation

```yaml
dependencies:
  web_socket: ^1.0.1
  web_socket_debugger: ^0.2.0
```

## Quick Start

```dart
import 'package:web_socket_debugger/web_socket_debugger.dart';

Future<void> main() async {
  const upstream = 'wss://ws.postman-echo.com/raw';
  
  // Configure proxy (reverse mode by default)
  final cfg = WebSocketDebugger.attach(
    baseUrl: upstream,
    proxyBaseUrl: 'http://localhost:9091',
    proxyPath: '/wsproxy',
  );

  // Connect through proxy
  final socket = await WebSocketDebugger.connect(config: cfg);
  
  socket.events.listen((e) {
    switch (e) {
      case TextDataReceived(text: final text):
        print('Received: $text');
      case BinaryDataReceived(data: final data):
        print('Binary data: ${data.length} bytes');
      case CloseReceived(code: final code, reason: final reason):
        print('Closed: $code $reason');
    }
  });
  
  socket.sendText('hello');
}
```

## API

### `WebSocketDebugger.attach()`

Creates a configuration for connecting through the proxy.

| Parameter      | Type      | Default                 | Description                                |
| -------------- | --------- | ----------------------- | ------------------------------------------ |
| `baseUrl`      | `String`  | required                | Target WebSocket URL (`ws://` or `wss://`) |
| `proxyBaseUrl` | `String`  | `http://localhost:9091` | Proxy server address                       |
| `proxyPath`    | `String`  | `/wsproxy`              | Proxy WebSocket endpoint path              |
| `enabled`      | `bool?`   | `true`                  | Enable/disable proxy                       |
| `mode`         | `String?` | `reverse`               | Mode: `reverse`, `forward`, `none`         |

### `WebSocketDebugger.connect()`

Creates a WebSocket connection using the configuration.

| Parameter | Type                    | Description                 |
| --------- | ----------------------- | --------------------------- |
| `config`  | `WebSocketProxyConfig`  | Config from `attach()`      |
| `headers` | `Map<String, dynamic>?` | HTTP headers (dart:io only) |

## Modes

### Reverse (default)

Client connects to proxy, proxy forwards to upstream:

```
Client -> ws://proxy:9091/wsproxy?_target=wss://example.com/ws -> Proxy -> wss://example.com/ws
```

```dart
final cfg = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  proxyPath: '/wsproxy',
  mode: 'reverse',
);
```

### Forward

Client connects directly to upstream, traffic goes through system proxy (dart:io only):

```dart
final cfg = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'forward',
);

// Forward mode requires HttpOverrides
if (cfg.useForwardOverrides) {
  HttpOverrides.runZoned(
    () async {
      final socket = await WebSocketDebugger.connect(config: cfg);
      // ...
    },
    createHttpClient: (_) => cfg.httpClientFactory!() as HttpClient,
  );
}
```

## Platform Behavior

| Feature           | dart:io (mobile/desktop) | Web (dart:js_interop)     |
| ----------------- | ------------------------ | ------------------------- |
| Reverse mode      | yes                      | yes                       |
| Forward mode      | yes                      | no (no HttpOverrides)     |
| Custom headers    | yes                      | no (browser limitation)   |
| Self-signed certs | yes (with flag)          | no                        |
| Read ENV          | yes                      | no (`--dart-define` only) |

## Configuration

### Via `--dart-define`

```bash
flutter run \
  --dart-define=SOCKET_PROXY=http://localhost:9091 \
  --dart-define=SOCKET_PROXY_PATH=/wsproxy \
  --dart-define=SOCKET_PROXY_MODE=reverse \
  --dart-define=SOCKET_PROXY_ENABLED=true
```

### Via ENV (dart:io only)

```bash
export SOCKET_PROXY=http://localhost:9091
export SOCKET_PROXY_MODE=reverse
```

### Environment Variables

| Variable                       | Description                           |
| ------------------------------ | ------------------------------------- |
| `SOCKET_PROXY`                 | Proxy server URL                      |
| `SOCKET_PROXY_PATH`            | WS endpoint path (usually `/wsproxy`) |
| `SOCKET_PROXY_MODE`            | `reverse` / `forward` / `none`        |
| `SOCKET_PROXY_ENABLED`         | `true` / `false`                      |
| `SOCKET_PROXY_ALLOW_BAD_CERTS` | Allow self-signed (forward mode)      |
| `SOCKET_UPSTREAM_URL`          | Explicit upstream URL                 |
| `SOCKET_UPSTREAM_TARGET`       | Full `_target` for reverse mode       |

## Android Emulator

On Android emulator `localhost` doesn't work — use `10.0.2.2`:

```dart
import 'dart:io' show Platform;

final cfg = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9091' 
    : 'http://localhost:9091',
);
```

## Links

- [web_socket on pub.dev](https://pub.dev/packages/web_socket)
- [network_debugger](https://pub.dev/packages/network_debugger)

## License

MIT
