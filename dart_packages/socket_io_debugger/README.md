# socket_io_debugger

<p align="left">
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/socket_io_debugger.yml"><img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/socket_io_debugger.yml/badge.svg?branch=main" alt="CI" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License" /></a>
</p>

A helper package to attach network-debugger proxy to `socket_io_client` for local debugging and Socket.IO traffic interception.

## Version Compatibility

**Important**: Choose the version based on your Socket.IO server:

| socket_io_debugger | socket_io_client | Socket.IO Server | Engine.IO |
| ------------------ | ---------------- | ---------------- | --------- |
| **1.0.0+**         | ^3.1.4           | v4.7+            | v4        |
| **^0.1.0**         | ^2.0.3           | v2.*/v3.*/v4.6   | v3/v4     |

## Installation

```yaml
dependencies:
  socket_io_client: ^3.1.4
  socket_io_debugger: ^1.1.0
```

## Starting the Proxy

Before using, start the network-debugger:

```bash
# Install CLI globally
dart pub global activate network_debugger

# Start proxy (port 9091, UI on 9092)
network_debugger
```

## Quick Start

```dart
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

void main() {
  // Configure proxy
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
    path: '/socket.io/',  // Socket.IO path on server
    proxyBaseUrl: 'http://localhost:9091',
    proxyHttpPath: '/wsproxy',
);

  // Create socket with config from attach()
final socket = io.io(
  cfg.effectiveBaseUrl,
  io.OptionBuilder()
    .setTransports(['websocket'])
    .setPath(cfg.effectivePath)
    .setQuery(cfg.query)
    .build(),
);

  socket.onConnect((_) => print('Connected!'));
  socket.on('message', (data) => print('Message: $data'));
  
  // Forward mode requires HttpOverrides
if (cfg.useForwardOverrides) {
    HttpOverrides.runZoned(
    () => socket.connect(),
    createHttpClient: (_) => cfg.httpClientFactory!(),
  );
} else {
  socket.connect();
  }
}
```

## API

### `SocketIoDebugger.attach()`

Creates a configuration for connecting through the proxy.

| Parameter       | Type      | Default                 | Description                          |
| --------------- | --------- | ----------------------- | ------------------------------------ |
| `baseUrl`       | `String`  | required                | Server URL (with namespace in path)  |
| `path`          | `String`  | `/socket.io/`           | Engine.IO endpoint path              |
| `proxyBaseUrl`  | `String?` | `http://localhost:9091` | Proxy address                        |
| `proxyHttpPath` | `String?` | `/wsproxy`              | Proxy WS endpoint path               |
| `enabled`       | `bool?`   | `true`                  | Enable/disable                       |

### Returned `SocketIoConfig`

| Field                 | Type                       | Description                            |
| --------------------- | -------------------------- | -------------------------------------- |
| `effectiveBaseUrl`    | `String`                   | URL to connect (proxy or original)     |
| `effectivePath`       | `String`                   | Path for `setPath()`                   |
| `query`               | `Map<String, dynamic>`     | Query params (including `_target`)     |
| `useForwardOverrides` | `bool`                     | Whether `HttpOverrides` is needed      |
| `httpClientFactory`   | `HttpClient Function()?`   | Factory for forward mode               |

## Path vs Namespace

These are different concepts in Socket.IO:

### Path (the `path` parameter)

**Path** is the HTTP endpoint where Engine.IO handshake occurs.

- Default: `/socket.io/`
- Must match between client and server
- Example: `path: '/my-custom-path/'`

### Namespace (part of `baseUrl`)

**Namespace** is a logical channel for organizing events.

- Default: `/`
- Specified as **part of the baseUrl path**
- Example: `baseUrl: 'https://example.com/admin'` -> namespace `/admin`

### Examples

```dart
// Default namespace (/), default path
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  // path defaults to '/socket.io/'
);

// Custom namespace /chat, default path
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com/chat',
);

// Custom namespace /admin, custom path
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com/admin',
  path: '/_api/v1/ws/',
);
```

## Modes

### Reverse (default)

Client connects to proxy, proxy forwards to upstream:

```
Client -> ws://proxy:9091/wsproxy?_target=http://example.com/socket.io/?EIO=4&transport=websocket -> Proxy -> example.com
```

### Forward

Client connects directly, traffic goes through system proxy (dart:io only).

## Platform Behavior

| Feature                 | dart:io (mobile/desktop) | Web (dart:js_interop)     |
| ----------------------- | ------------------------ | ------------------------- |
| Reverse mode            | yes                      | yes                       |
| Forward mode            | yes                      | no (no HttpOverrides)     |
| Custom headers          | yes                      | no (browser limitation)   |
| Self-signed certs       | yes (with flag)          | no                        |
| Read ENV                | yes                      | no (`--dart-define` only) |

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
| `SOCKET_UPSTREAM_PATH`         | Socket.IO path on upstream            |
| `SOCKET_UPSTREAM_TARGET`       | Full `_target` value                  |

## Android Emulator

On Android emulator `localhost` doesn't work — use `10.0.2.2`:

```dart
import 'dart:io' show Platform;

final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  proxyBaseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9091' 
    : 'http://localhost:9091',
);
```

## Known Issues

### Port 0 in socket_io_client

There's a bug in `socket_io_client`: if no port is specified in URL, `Uri.port` returns `0`, and the client doesn't set default 80/443. This package **automatically adds explicit port** in `_target` to avoid the issue.

## Links

- [socket_io_client on pub.dev](https://pub.dev/packages/socket_io_client)
- [network_debugger](https://pub.dev/packages/network_debugger)

## License

Apache-2.0
