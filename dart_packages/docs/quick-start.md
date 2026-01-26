# Quick Start Guide

Get Network Debugger running with your Dart/Flutter app in 5 minutes.

## Step 1: Install and Start the Proxy Server

```bash
# Option A: Install CLI globally
dart pub global activate network_debugger
network_debugger

# Option B: Run from source
cd network_debugger
go run ./cmd/network-debugger
```

The proxy server starts on:
- **Proxy endpoint**: `http://localhost:9091`
- **Web UI**: `http://localhost:9092`

## Step 2: Add the Package to Your Project

Choose the package that matches your networking library:

### For Dio

```yaml
# pubspec.yaml
dependencies:
  dio: ^5.0.0
  dio_debugger: ^0.1.0
```

```dart
import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

final dio = Dio();
dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
  ),
);

// Make requests as usual
final response = await dio.get('https://api.example.com/users');
```

### For package:http

```yaml
dependencies:
  http: ^1.0.0
  http_debugger: ^0.1.0
```

```dart
import 'package:http/http.dart' as http;
import 'package:http_debugger/http_debugger.dart';

// Create wrapped client
final client = HttpDebuggerClient.wrap(
  http.Client(),
  proxyBaseUrl: 'http://localhost:9091',
);

// Make requests as usual
final response = await client.get(Uri.parse('https://api.example.com/users'));
```

### For WebSocket (package:web_socket_channel)

```yaml
dependencies:
  web_socket_channel: ^3.0.0
  web_socket_channel_debugger: ^0.1.0
```

```dart
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

final config = WebSocketChannelDebugger.attach(
  baseUrl: 'wss://echo.websocket.events',
  proxyBaseUrl: 'http://localhost:9091',
);

final channel = WebSocketChannelDebugger.connect(config: config);
await channel.ready;

channel.sink.add('Hello!');
channel.stream.listen((message) => print('Received: $message'));
```

### For Socket.IO

```yaml
dependencies:
  socket_io_client: ^3.1.4
  socket_io_debugger: ^1.0.0
```

```dart
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

final config = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  path: '/socket.io/',
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

socket.onConnect((_) => print('Connected!'));
socket.connect();
```

## Step 3: View Traffic in the Web UI

1. Open `http://localhost:9092` in your browser
2. Make requests from your app
3. See all HTTP/WebSocket traffic in real-time

## Step 4: Disable for Production

Use environment variables or `--dart-define` to disable the proxy in production:

```bash
# Development (proxy enabled)
flutter run

# Production (proxy disabled)
flutter run --dart-define=DIO_DEBUGGER_ENABLED=false
```

Or check build mode in code:

```dart
dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
    enabled: kDebugMode, // Only enable in debug builds
  ),
);
```

## Next Steps

- [Configuration Guide](./configuration.md) - Advanced configuration options
- [Proxy Modes](./proxy-modes.md) - Learn about reverse vs forward proxy
- [Platform Support](./platform-support.md) - Platform-specific considerations
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions
