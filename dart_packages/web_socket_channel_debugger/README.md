# web_socket_channel_debugger

<p align="left">
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_channel_debugger.yml"><img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/web_socket_channel_debugger.yml/badge.svg?branch=main" alt="CI" /></a>
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

> This package is for integrating Network Debugger with
> `package:web_socket_channel`.

A helper package to attach network-debugger proxy to `package:web_socket_channel` for local debugging and WebSocket traffic interception.

## Installation

```yaml
dependencies:
  web_socket_channel: ^3.0.3
  web_socket_channel_debugger: ^0.2.0
```

## Quick Start

```dart
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() async {
  const upstream = 'wss://echo.websocket.events';
  
  final cfg = WebSocketChannelDebugger.attach(
    baseUrl: upstream,
    proxyBaseUrl: 'http://localhost:9091',
    proxyPath: '/wsproxy',
  );

  final channel = WebSocketChannelDebugger.connect(config: cfg);
  await channel.ready;
  
  channel.stream.listen((message) => print('Received: \$message'));
  channel.sink.add('hello');
}
```

## API

### `WebSocketChannelDebugger.attach()`

Creates a configuration for connecting through the proxy.

| Parameter      | Type      | Default                 | Description                                |
| -------------- | --------- | ----------------------- | ------------------------------------------ |
| `baseUrl`      | `String`  | required                | Target WebSocket URL (`ws://` or `wss://`) |
| `proxyBaseUrl` | `String`  | `http://localhost:9091` | Proxy server address                       |
| `proxyPath`    | `String`  | `/wsproxy`              | Proxy WebSocket endpoint path              |
| `enabled`      | `bool?`   | `true`                  | Enable/disable proxy                       |
| `mode`         | `String?` | `reverse`               | Mode: `reverse`, `forward`, `none`         |

### `WebSocketChannelDebugger.connect()`

| Parameter | Type                    | Description                 |
| --------- | ----------------------- | --------------------------- |
| `config`  | `WscProxyConfig`        | Config from `attach()`      |
| `headers` | `Map<String, dynamic>?` | HTTP headers (dart:io only) |

## Platform Behavior

| Feature           | dart:io (mobile/desktop) | Web (dart:js_interop)     |
| ----------------- | ------------------------ | ------------------------- |
| Reverse mode      | yes                      | yes                       |
| Forward mode      | yes                      | no (no HttpOverrides)     |
| Custom headers    | yes                      | no (browser limitation)   |
| Self-signed certs | yes (with flag)          | no                        |
| Read ENV          | yes                      | no (`--dart-define` only) |

## Environment Variables

| Variable                       | Description                           |
| ------------------------------ | ------------------------------------- |
| `SOCKET_PROXY`                 | Proxy server URL                      |
| `SOCKET_PROXY_PATH`            | WS endpoint path (usually `/wsproxy`) |
| `SOCKET_PROXY_MODE`            | `reverse` / `forward` / `none`        |
| `SOCKET_PROXY_ENABLED`         | `true` / `false`                      |
| `SOCKET_PROXY_ALLOW_BAD_CERTS` | Allow self-signed (forward mode)      |
| `SOCKET_UPSTREAM_URL`          | Explicit upstream URL                 |
| `SOCKET_UPSTREAM_TARGET`       | Full `_target` for reverse mode       |

## Links

- [web_socket_channel on pub.dev](https://pub.dev/packages/web_socket_channel)
- [network_debugger](https://pub.dev/packages/network_debugger)

## License

MIT
