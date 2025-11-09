# socket_io_debugger

> Part of the [network_debugger](https://pub.dev/packages/network_debugger) ecosystem

One-call helper to attach a proxy to a Socket.IO client (reverse/forward modes). Useful for local debugging, traffic interception, and bypassing CORS/certificates via your local proxy.

## Version Compatibility

**IMPORTANT**: Choose the correct version based on your Socket.IO server version:

| socket_io_debugger | socket_io_client | Socket.IO Server | Engine.IO |
|-------------------|------------------|------------------|-----------|
| **1.0.0+** | ^3.0.0 | v4.7+ | v4 |
| **^0.1.0** | ^2.0.3 | v2.*/v3.*/v4.6 | v3/v4 |

**How to choose:**
- If your server uses Socket.IO **v4.7 or higher** → use `socket_io_debugger: ^1.0.0`
- If your server uses Socket.IO **v2, v3, or v4.6 and below** → use `socket_io_debugger: ^0.1.0`

**Note**: The API remains the same between versions, so migration only requires updating dependencies.

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

### For Socket.IO server v4.7+

```yaml
dependencies:
  socket_io_client: ^3.0.0
  socket_io_debugger: ^1.0.0
```

### For Socket.IO server v2.*/v3.*/v4.6 and below

```yaml
dependencies:
  socket_io_client: ^2.0.3
  socket_io_debugger: ^0.1.0
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

// Configure the proxy (path parameter is optional, defaults to '/socket.io/')
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  // path: '/socket.io/',  // Optional: Socket.IO path (default)
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

## Understanding Socket.IO Path and Namespace

Socket.IO has two separate concepts that are often confused:

### Path (the `path` parameter)

The **path** is the HTTP endpoint where Socket.IO protocol communication occurs.

- **Default**: `"/socket.io/"`
- **Must match** between client and server
- Set via `path` parameter in `SocketIoDebugger.attach()`
- Example: `path: "/my-custom-path/"`

### Namespace (part of `baseUrl`)

The **namespace** is a logical channel for organizing application logic.

- **Default**: `"/"`
- Specified as **part of the baseUrl path**
- Example: `baseUrl: "https://example.com/admin"` (namespace = `/admin`)

### Complete Example

```dart
// Connect to the '/order' namespace with custom Socket.IO path
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com/order',  // namespace: /order
  path: '/my-custom-path/',              // Socket.IO path option
);
// Results in: GET https://example.com/my-custom-path/?EIO=4&transport=websocket
// Connected to namespace: /order
```

### Examples

**Default namespace, default path:**
```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',     // namespace: / (default)
  // path defaults to '/socket.io/'
);
```

**Custom namespace, default path:**
```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com/admin',  // namespace: /admin
  // path defaults to '/socket.io/'
);
```

**Custom namespace, custom path:**
```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com/chat',  // namespace: /chat
  path: '/_api/v1/monitor/io/',         // custom Socket.IO path
);
```

## Advanced options

### Reverse proxy mode (default)

Reverse proxy mode routes traffic through the proxy's HTTP endpoint:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  path: '/socket.io/',  // Socket.IO path (with trailing slash)
  proxyBaseUrl: 'http://localhost:9092',
  proxyHttpPath: '/wsproxy',  // WebSocket proxy endpoint
);
```

### Forward proxy mode

Forward proxy mode uses `HttpOverrides` to route traffic:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  path: '/socket.io/',  // Socket.IO path (with trailing slash)
  proxyBaseUrl: 'http://localhost:9091',
);
```

### Android emulator

When testing on Android emulator, use `10.0.2.2` instead of `localhost`:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  path: '/socket.io/',  // Socket.IO path (with trailing slash)
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
