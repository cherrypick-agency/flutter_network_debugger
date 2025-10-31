# socket_io_debugger

> Part of the [network_debugger](https://pub.dev/packages/network_debugger) ecosystem

One-call helper to attach a proxy to a Socket.IO client (reverse/forward modes). Useful for local debugging, traffic interception, and bypassing CORS/certificates via your local proxy.

## Features

- One-liner attach: `SocketIoDebugger.attach()`
- Supports reverse and forward proxy modes
- Config sources (priority):
  1) `attach` arguments
  2) `--dart-define` (`HTTP_PROXY_MODE`, `HTTP_PROXY`, `SOCKET_PROXY`, `HTTP_PROXY_PATH`, `SOCKET_PROXY_PATH`, `HTTP_PROXY_ALLOW_BAD_CERTS`, `HTTP_PROXY_ENABLED`)
  3) OS ENV (via conditional import; web-safe)
- Automatic HTTP client factory for forward proxy mode
- Support for self-signed certificates
- Zero-config for quick setup

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  socket_io_client: ^2.0.3
  socket_io_debugger: ^0.1.1
```

## Starting the Proxy

Before using `socket_io_debugger`, you need to start the network debugger proxy server. Install and run it with:

```bash
# Install the CLI globally
dart pub global activate network_debugger

# Start the proxy (proxy port 9091, UI opens on 9092)
network_debugger
```

Proxy base will be `http://localhost:9091`. The web UI opens on `http://localhost:9092`.

For more options and programmatic usage, see the [network_debugger package documentation](https://pub.dev/packages/network_debugger).

## Quick start

```dart
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

// Configure the proxy
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  socketPath: '/socket.io',
);

// Create socket with proxy configuration
final socket = io.io(
  cfg.effectiveBaseUrl,
  io.OptionBuilder()
    .setTransports(['websocket'])
    .setPath(cfg.effectivePath)
    .setQuery(cfg.query)
    .build(),
);

// Connect (with HttpOverrides if forward proxy mode)
if (cfg.useForwardOverrides) {
  await HttpOverrides.runZoned(
    () => socket.connect(),
    createHttpClient: (_) => cfg.httpClientFactory!(),
  );
} else {
  socket.connect();
}
```

## Advanced options

### Reverse proxy mode (default)

Reverse proxy mode routes traffic through the proxy's HTTP endpoint:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  socketPath: '/socket.io',
  proxyBaseUrl: 'http://localhost:9092',
  proxyPath: '/wsproxy',  // WebSocket proxy endpoint
  mode: ProxyMode.reverse,
);
```

### Forward proxy mode

Forward proxy mode uses `HttpOverrides` to route traffic:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  socketPath: '/socket.io',
  proxyBaseUrl: 'http://localhost:9091',
  mode: ProxyMode.forward,
  allowBadCerts: true,  // Allow self-signed certificates
);
```

### Android emulator

When testing on Android emulator, use `10.0.2.2` instead of `localhost`:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  socketPath: '/socket.io',
  proxyBaseUrl: Platform.isAndroid
    ? 'http://10.0.2.2:9092'
    : 'http://localhost:9092',
);
```

## Configuration examples

### Via `--dart-define`

```bash
--dart-define=HTTP_PROXY_MODE=reverse \
--dart-define=SOCKET_PROXY=http://localhost:9092 \
--dart-define=SOCKET_PROXY_PATH=/wsproxy \
--dart-define=HTTP_PROXY_ENABLED=true
```

### Via OS ENV (on platforms with `dart:io`)

```bash
HTTP_PROXY_MODE=reverse
SOCKET_PROXY=http://localhost:9092
SOCKET_PROXY_PATH=/wsproxy
HTTP_PROXY_ENABLED=true
HTTP_PROXY_ALLOW_BAD_CERTS=true
```

## Environment variables

- `HTTP_PROXY_MODE` — Proxy mode: `reverse`, `forward`, or `none`
- `HTTP_PROXY` — HTTP proxy URL (fallback for `SOCKET_PROXY`)
- `SOCKET_PROXY` — Socket.IO proxy URL
- `HTTP_PROXY_PATH` — HTTP proxy path (fallback for `SOCKET_PROXY_PATH`)
- `SOCKET_PROXY_PATH` — Socket.IO proxy path
- `HTTP_PROXY_ALLOW_BAD_CERTS` — Allow self-signed certificates: `true`, `1`, `yes`, `on`
- `HTTP_PROXY_ENABLED` — Enable/disable proxy: `true`, `1`, `yes`, `on`

## Notes

- The proxy must expose a WebSocket proxy endpoint (e.g., `/wsproxy`) that accepts `_target` query and forwards the WebSocket connection
- If `proxyBaseUrl` is empty or `HTTP_PROXY_ENABLED=false`, the package is a no-op (safe for prod)
- Forward proxy mode requires `dart:io` and won't work on web platform
- If the proxy is provided without scheme and with port `:443`, `https` will be used automatically

## License

MIT
