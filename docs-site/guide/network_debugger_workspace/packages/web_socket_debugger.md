# web_socket_debugger

WebSocket debugging for [package:web_socket](https://pub.dev/packages/web_socket).

## Installation

```yaml
dependencies:
  web_socket: ^1.0.1
  web_socket_debugger: ^0.2.0
```

## Basic Usage

```dart
import 'package:web_socket_debugger/web_socket_debugger.dart';

Future<void> main() async {
  // Configure proxy
  final config = WebSocketDebugger.attach(
    baseUrl: 'wss://echo.websocket.events',
    proxyBaseUrl: 'http://localhost:9091',
  );

  // Connect through proxy
  final socket = await WebSocketDebugger.connect(config: config);
  
  // Listen for messages
  socket.events.listen((event) {
    switch (event) {
      case TextDataReceived(text: final text):
        print('Received: $text');
      case BinaryDataReceived(data: final data):
        print('Binary: ${data.length} bytes');
      case CloseReceived(code: final code, reason: final reason):
        print('Closed: $code $reason');
    }
  });
  
  // Send message
  socket.sendText('Hello, WebSocket!');
}
```

## API Reference

### WebSocketDebugger.attach()

Creates a proxy configuration for WebSocket connection.

```dart
static WebSocketProxyConfig attach({
  required String baseUrl,
  String proxyBaseUrl = 'http://localhost:9091',
  String proxyPath = '/wsproxy',
  bool? enabled,
  String? mode,
});
```

| Parameter | Type | Default | Description |
| --------- | ---- | ------- | ----------- |
| `baseUrl` | `String` | required | Target WebSocket URL (`ws://` or `wss://`) |
| `proxyBaseUrl` | `String` | `http://localhost:9091` | Proxy server URL |
| `proxyPath` | `String` | `/wsproxy` | WebSocket proxy endpoint |
| `enabled` | `bool?` | `true` | Enable/disable proxy |
| `mode` | `String?` | `reverse` | `reverse`, `forward`, or `none` |

### WebSocketDebugger.connect()

Creates a WebSocket connection using the configuration.

```dart
static Future<WebSocket> connect({
  required WebSocketProxyConfig config,
  Map<String, dynamic>? headers,
});
```

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `config` | `WebSocketProxyConfig` | Config from `attach()` |
| `headers` | `Map<String, dynamic>?` | HTTP headers (dart:io only) |

### WebSocketProxyConfig

Returned by `attach()`:

```dart
class WebSocketProxyConfig {
  final Uri connectUrl;          // URL to connect to
  final Map<String, dynamic> query;  // Query parameters
  final bool useForwardOverrides;    // Whether forward mode needs HttpOverrides
  final Object Function()? httpClientFactory;  // For forward mode
}
```

## Configuration Examples

### Basic Reverse Proxy

```dart
final config = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
);

final socket = await WebSocketDebugger.connect(config: config);
```

### With Authentication Headers

```dart
final config = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
);

// Headers only work on dart:io platforms
final socket = await WebSocketDebugger.connect(
  config: config,
  headers: {
    'Authorization': 'Bearer $token',
  },
);
```

### Forward Proxy Mode (dart:io only)

```dart
final config = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'forward',
);

if (config.useForwardOverrides) {
  await HttpOverrides.runZoned(
    () async {
      final socket = await WebSocketDebugger.connect(config: config);
      // Use socket...
    },
    createHttpClient: (_) => config.httpClientFactory!() as HttpClient,
  );
}
```

### Disabled (Bypass)

```dart
final config = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  enabled: false,  // Connect directly, no proxy
);
```

### Android Emulator

```dart
import 'dart:io' show Platform;

final config = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9091' 
    : 'http://localhost:9091',
);
```

## How It Works

### Reverse Proxy Mode

1. `attach()` creates config with proxy URL
2. Original URL embedded in `_target` query parameter
3. Connection goes to: `ws://proxy:9091/wsproxy?_target=wss://example.com/ws`
4. Proxy extracts target and forwards WebSocket frames

### URL Transformation

Original:
```
wss://example.com/ws?token=abc
```

Becomes:
```
ws://localhost:9091/wsproxy?_target=wss://example.com/ws?token=abc
```

## Platform Support

| Feature | dart:io | Web |
| ------- | ------- | --- |
| Reverse mode | Yes | Yes |
| Forward mode | Yes | No |
| Custom headers | Yes | No |
| Self-signed certs | Yes | No |

## Environment Variables

| Variable | Description |
| -------- | ----------- |
| `SOCKET_PROXY` | Proxy server URL |
| `SOCKET_PROXY_PATH` | WebSocket endpoint path |
| `SOCKET_PROXY_MODE` | `reverse`, `forward`, `none` |
| `SOCKET_PROXY_ENABLED` | `true`/`false` |
| `SOCKET_UPSTREAM_URL` | Override upstream URL |
| `SOCKET_UPSTREAM_TARGET` | Full `_target` value |

## Troubleshooting

### Connection Fails Immediately

1. Verify proxy server is running
2. Check WebSocket endpoint is `/wsproxy` not `/httpproxy`
3. Verify target URL is valid WebSocket URL

### Headers Not Working

- Headers are ignored on Web platform (browser limitation)
- Use query parameters or cookies for authentication on Web

### Self-Signed Certificate Errors

For development, use `allowBadCertificates` in forward mode or configure the proxy server with valid certificates.

## See Also

- [Quick Start Guide](../quick-start.md)
- [Proxy Modes](../proxy-modes.md)
- [Platform Support](../platform-support.md)
