# socket_io_debugger

Socket.IO debugging for [package:socket_io_client](https://pub.dev/packages/socket_io_client).

## Installation

```yaml
dependencies:
  socket_io_client: ^3.1.4
  socket_io_debugger: ^1.1.0
```

## Version Compatibility

| socket_io_debugger | socket_io_client | Socket.IO Server |
| ------------------ | ---------------- | ---------------- |
| **1.0.0+** | ^3.1.4 | v4.7+ |
| **^0.1.0** | ^2.0.3 | v2.*/v3.*/v4.6 |

## Basic Usage

```dart
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

void main() {
  // Configure proxy
  final config = SocketIoDebugger.attach(
    baseUrl: 'https://example.com',
    path: '/socket.io/',
    proxyBaseUrl: 'http://localhost:9091',
  );

  // Create socket with config
  final socket = io.io(
    config.effectiveBaseUrl,
    io.OptionBuilder()
      .setTransports(['websocket'])
      .setPath(config.effectivePath)
      .setQuery(config.query)
      .build(),
  );

  // Event handlers
  socket.onConnect((_) => print('Connected'));
  socket.onDisconnect((_) => print('Disconnected'));
  socket.on('message', (data) => print('Message: $data'));
  
  // Connect
  socket.connect();
}
```

## API Reference

### SocketIoDebugger.attach()

Creates configuration for Socket.IO connection through proxy.

```dart
static SocketIoConfig attach({
  required String baseUrl,
  String path = '/socket.io/',
  String? proxyBaseUrl,
  String? proxyHttpPath,
  bool? enabled,
});
```

| Parameter | Type | Default | Description |
| --------- | ---- | ------- | ----------- |
| `baseUrl` | `String` | required | Socket.IO server URL (with optional namespace) |
| `path` | `String` | `/socket.io/` | Engine.IO endpoint path |
| `proxyBaseUrl` | `String?` | `http://localhost:9091` | Proxy server URL |
| `proxyHttpPath` | `String?` | `/wsproxy` | Proxy WebSocket endpoint |
| `enabled` | `bool?` | `true` | Enable/disable proxy |

### SocketIoConfig

Returned by `attach()`:

```dart
class SocketIoConfig {
  final String effectiveBaseUrl;     // URL for io.io()
  final String effectivePath;        // Path for setPath()
  final Map<String, dynamic> query;  // Query params including _target
  final bool useForwardOverrides;    // If forward mode needs HttpOverrides
  final HttpClient Function()? httpClientFactory;
}
```

## Understanding Path vs Namespace

### Path (Engine.IO Endpoint)

The `path` parameter specifies where Engine.IO handshake occurs:

```dart
// Server configured with: io({ path: '/my-socket/' })
final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  path: '/my-socket/',  // Engine.IO endpoint
);
```

### Namespace (Logical Channel)

Namespace is part of the `baseUrl` path:

```dart
// Connect to /admin namespace
final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com/admin',  // /admin is namespace
  path: '/socket.io/',  // Engine.IO endpoint (usually unchanged)
);
```

### Examples

```dart
// Default namespace (/), default path
SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
);
// -> namespace: /, path: /socket.io/

// Custom namespace /chat
SocketIoDebugger.attach(
  baseUrl: 'https://example.com/chat',
);
// -> namespace: /chat, path: /socket.io/

// Custom namespace /admin, custom path
SocketIoDebugger.attach(
  baseUrl: 'https://example.com/admin',
  path: '/ws/v1/',
);
// -> namespace: /admin, path: /ws/v1/
```

## Configuration Examples

### Basic Setup

```dart
final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  proxyBaseUrl: 'http://localhost:9091',
);

final socket = io.io(
  config.effectiveBaseUrl,
  io.OptionBuilder()
    .setTransports(['websocket'])
    .setPath(config.effectivePath)
    .setQuery(config.query)
    .build(),
);
```

### With Authentication

```dart
final socket = io.io(
  config.effectiveBaseUrl,
  io.OptionBuilder()
    .setTransports(['websocket'])
    .setPath(config.effectivePath)
    .setQuery({
      ...config.query,
      'token': authToken,  // Add your auth params
    })
    .setExtraHeaders({'Authorization': 'Bearer $authToken'})
    .build(),
);
```

### Forward Proxy Mode

```dart
final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  proxyBaseUrl: 'http://localhost:9091',
);

if (config.useForwardOverrides) {
  HttpOverrides.runZoned(
    () {
      final socket = io.io(...);
      socket.connect();
    },
    createHttpClient: (_) => config.httpClientFactory!(),
  );
} else {
  final socket = io.io(...);
  socket.connect();
}
```

### Android Emulator

```dart
import 'dart:io' show Platform;

final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  proxyBaseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9091' 
    : 'http://localhost:9091',
);
```

### Debug Mode Only

```dart
import 'package:flutter/foundation.dart';

final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  proxyBaseUrl: 'http://localhost:9091',
  enabled: kDebugMode,
);
```

## How It Works

1. `attach()` creates config with proxy settings
2. `_target` parameter contains Engine.IO URL with transport params
3. Connection flow:
   ```
   Client -> ws://proxy:9091/wsproxy?_target=http://example.com:443/socket.io/?EIO=4&transport=websocket
   ```
4. Proxy extracts target and forwards WebSocket frames

### Target URL Format

The `_target` includes:
- HTTP scheme (not WS) for Engine.IO
- Explicit port (80 or 443) to work around socket_io_client bug
- Engine.IO params: `EIO=4&transport=websocket`

Example: `http://example.com:443/socket.io/?EIO=4&transport=websocket`

## Platform Support

| Feature | dart:io | Web |
| ------- | ------- | --- |
| Reverse mode | Yes | Yes |
| Forward mode | Yes | No |
| Custom headers | Yes | No |
| Namespaces | Yes | Yes |
| Rooms | Yes | Yes |

## Environment Variables

| Variable | Description |
| -------- | ----------- |
| `SOCKET_PROXY` | Proxy server URL |
| `SOCKET_PROXY_PATH` | Proxy endpoint (`/wsproxy`) |
| `SOCKET_PROXY_MODE` | `reverse`, `forward`, `none` |
| `SOCKET_PROXY_ENABLED` | `true`/`false` |
| `SOCKET_UPSTREAM_URL` | Override upstream URL |
| `SOCKET_UPSTREAM_PATH` | Override Socket.IO path |

## Troubleshooting

### "Incompatible server version" Error

**Cause:** Version mismatch between client and server

**Solution:** Check version compatibility table above

### Connection Timeout

1. Verify proxy server is running
2. Check target URL is accessible directly
3. Verify Engine.IO endpoint path is correct

### Missing Port Error

**Cause:** socket_io_client bug with `Uri.port == 0`

**Solution:** Update to socket_io_debugger 1.0.0+ (includes fix)

### Namespace Not Working

Namespace must be in `baseUrl`, not in `path`:

```dart
// Wrong
SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  path: '/admin/socket.io/',  // /admin is NOT the path
);

// Correct  
SocketIoDebugger.attach(
  baseUrl: 'https://example.com/admin',  // /admin is namespace
  path: '/socket.io/',
);
```

## Complete Example

```dart
import 'dart:io' show Platform;
import 'package:flutter/foundation.dart';
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

class SocketService {
  late io.Socket socket;
  
  void connect() {
    final config = SocketIoDebugger.attach(
      baseUrl: 'https://api.example.com/chat',
      path: '/socket.io/',
      proxyBaseUrl: Platform.isAndroid 
        ? 'http://10.0.2.2:9091' 
        : 'http://localhost:9091',
      enabled: kDebugMode,
    );
    
    socket = io.io(
      config.effectiveBaseUrl,
      io.OptionBuilder()
        .setTransports(['websocket'])
        .setPath(config.effectivePath)
        .setQuery(config.query)
        .enableAutoConnect()
        .enableReconnection()
        .build(),
    );
    
    socket.onConnect((_) => print('Connected to chat'));
    socket.onDisconnect((_) => print('Disconnected'));
    socket.on('new_message', _handleMessage);
    socket.on('error', (e) => print('Socket error: $e'));
  }
  
  void _handleMessage(dynamic data) {
    print('New message: $data');
  }
  
  void sendMessage(String text) {
    socket.emit('send_message', {'text': text});
  }
  
  void dispose() {
    socket.dispose();
  }
}
```

## See Also

- [Quick Start Guide](../quick-start.md)
- [Proxy Modes](../proxy-modes.md)
- [Platform Support](../platform-support.md)
- [Troubleshooting](../troubleshooting.md)
