# web_socket_channel_debugger

WebSocket debugging for [package:web_socket_channel](https://pub.dev/packages/web_socket_channel).

## Installation

```yaml
dependencies:
  web_socket_channel: ^3.0.3
  web_socket_channel_debugger: ^0.1.0
```

## Basic Usage

```dart
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() async {
  // Configure proxy
  final config = WebSocketChannelDebugger.attach(
    baseUrl: 'wss://echo.websocket.events',
    proxyBaseUrl: 'http://localhost:9091',
  );

  // Connect through proxy
  final channel = WebSocketChannelDebugger.connect(config: config);
  
  // Wait for connection
  await channel.ready;
  
  // Listen for messages
  channel.stream.listen(
    (message) => print('Received: $message'),
    onError: (error) => print('Error: $error'),
    onDone: () => print('Connection closed'),
  );
  
  // Send message
  channel.sink.add('Hello, WebSocket!');
  
  // Close when done
  await channel.sink.close();
}
```

## API Reference

### WebSocketChannelDebugger.attach()

Creates a proxy configuration for WebSocket connection.

```dart
static WscProxyConfig attach({
  required String baseUrl,
  String proxyBaseUrl = 'http://localhost:9091',
  String proxyPath = '/wsproxy',
  bool? enabled,
  String? mode,
});
```

| Parameter | Type | Default | Description |
| --------- | ---- | ------- | ----------- |
| `baseUrl` | `String` | required | Target WebSocket URL |
| `proxyBaseUrl` | `String` | `http://localhost:9091` | Proxy server URL |
| `proxyPath` | `String` | `/wsproxy` | WebSocket proxy endpoint |
| `enabled` | `bool?` | `true` | Enable/disable proxy |
| `mode` | `String?` | `reverse` | `reverse`, `forward`, `none` |

### WebSocketChannelDebugger.connect()

Creates a WebSocketChannel using the configuration.

```dart
static WebSocketChannel connect({
  required WscProxyConfig config,
  Map<String, dynamic>? headers,
});
```

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `config` | `WscProxyConfig` | Config from `attach()` |
| `headers` | `Map<String, dynamic>?` | HTTP headers (dart:io only) |

### WscProxyConfig

Returned by `attach()`:

```dart
class WscProxyConfig {
  final Uri connectUrl;          // URL to connect to
  final Map<String, dynamic> query;  // Query parameters
  final bool useForwardOverrides;    // Whether forward mode needs HttpOverrides
  final Object Function()? httpClientFactory;  // For forward mode
}
```

## Configuration Examples

### Basic Reverse Proxy

```dart
final config = WebSocketChannelDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
);

final channel = WebSocketChannelDebugger.connect(config: config);
await channel.ready;
```

### With Authentication

```dart
// Option 1: Headers (dart:io only)
final channel = WebSocketChannelDebugger.connect(
  config: config,
  headers: {'Authorization': 'Bearer $token'},
);

// Option 2: Query parameters (all platforms)
final config = WebSocketChannelDebugger.attach(
  baseUrl: 'wss://example.com/ws?token=$token',
  proxyBaseUrl: 'http://localhost:9091',
);
```

### URL Scheme Handling

The package automatically handles scheme conversion:

```dart
// http:// baseUrl -> ws:// target
final config1 = WebSocketChannelDebugger.attach(
  baseUrl: 'http://example.com/ws',  // -> ws://example.com/ws
);

// https:// baseUrl -> wss:// target  
final config2 = WebSocketChannelDebugger.attach(
  baseUrl: 'https://example.com/ws',  // -> wss://example.com/ws
);
```

### Forward Proxy Mode

```dart
final config = WebSocketChannelDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'forward',
);

if (config.useForwardOverrides) {
  HttpOverrides.runZoned(
    () {
      final channel = WebSocketChannelDebugger.connect(config: config);
      // Use channel...
    },
    createHttpClient: (_) => config.httpClientFactory!() as HttpClient,
  );
}
```

### Conditional Debugging

```dart
import 'package:flutter/foundation.dart';

final config = WebSocketChannelDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  enabled: kDebugMode,  // Only in debug builds
);
```

### Android Emulator

```dart
import 'dart:io' show Platform;

final config = WebSocketChannelDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9091' 
    : 'http://localhost:9091',
);
```

## Working with Streams

```dart
final channel = WebSocketChannelDebugger.connect(config: config);
await channel.ready;

// Stream operations
channel.stream
  .where((msg) => msg is String)
  .map((msg) => jsonDecode(msg as String))
  .listen((data) {
    print('Parsed: $data');
  });

// Sink operations
final data = {'action': 'subscribe', 'channel': 'updates'};
channel.sink.add(jsonEncode(data));
```

## Error Handling

```dart
final channel = WebSocketChannelDebugger.connect(config: config);

try {
  await channel.ready;
} on WebSocketChannelException catch (e) {
  print('Connection failed: $e');
  return;
}

channel.stream.listen(
  (message) => print('Message: $message'),
  onError: (error) {
    print('Stream error: $error');
  },
  onDone: () {
    print('Connection closed: ${channel.closeCode} ${channel.closeReason}');
  },
);
```

## Platform Support

| Feature | dart:io | Web |
| ------- | ------- | --- |
| Reverse mode | Yes | Yes |
| Forward mode | Yes | No |
| Custom headers | Yes | No |
| Binary messages | Yes | Yes |

## Environment Variables

| Variable | Description |
| -------- | ----------- |
| `SOCKET_PROXY` | Proxy server URL |
| `SOCKET_PROXY_PATH` | WebSocket endpoint |
| `SOCKET_PROXY_MODE` | `reverse`, `forward`, `none` |
| `SOCKET_PROXY_ENABLED` | `true`/`false` |

## Differences from web_socket_debugger

| Feature | web_socket_debugger | web_socket_channel_debugger |
| ------- | ------------------- | --------------------------- |
| Package | `package:web_socket` | `package:web_socket_channel` |
| API Style | Low-level events | Stream-based |
| Message Types | TextDataReceived/BinaryDataReceived | Dynamic |
| Connection | `Future<WebSocket>` | `WebSocketChannel` |

## See Also

- [Quick Start Guide](../quick-start.md)
- [Proxy Modes](../proxy-modes.md)
- [Platform Support](../platform-support.md)
