# Configuration Guide

This guide covers all configuration options for the Network Debugger Dart packages.

## Configuration Methods

There are three ways to configure the debugger packages:

1. **Runtime parameters** - Pass values directly to `attach()` or constructor
2. **`--dart-define`** - Compile-time constants (works on all platforms)
3. **Environment variables** - Runtime config (dart:io platforms only)

### Priority Order

When multiple sources provide the same setting, priority is:

1. Runtime parameters (highest)
2. `--dart-define` values
3. Environment variables (lowest, dart:io only)

## Common Configuration Options

### Proxy Server Settings

| Setting | dart-define | ENV Variable | Default |
| ------- | ----------- | ------------ | ------- |
| Proxy URL | `SOCKET_PROXY` | `SOCKET_PROXY` | `http://localhost:9091` |
| Proxy HTTP path | `SOCKET_PROXY_PATH` | `SOCKET_PROXY_PATH` | `/httpproxy` or `/wsproxy` |
| Enabled | `SOCKET_PROXY_ENABLED` | `SOCKET_PROXY_ENABLED` | `true` |
| Mode | `SOCKET_PROXY_MODE` | `SOCKET_PROXY_MODE` | `reverse` |

### Upstream Settings

| Setting | dart-define | ENV Variable | Default |
| ------- | ----------- | ------------ | ------- |
| Upstream URL | `SOCKET_UPSTREAM_URL` | `SOCKET_UPSTREAM_URL` | - |
| Upstream path | `SOCKET_UPSTREAM_PATH` | `SOCKET_UPSTREAM_PATH` | - |
| Explicit target | `SOCKET_UPSTREAM_TARGET` | `SOCKET_UPSTREAM_TARGET` | - |

### Security Settings

| Setting | dart-define | ENV Variable | Default |
| ------- | ----------- | ------------ | ------- |
| Allow bad certs | `SOCKET_PROXY_ALLOW_BAD_CERTS` | `SOCKET_PROXY_ALLOW_BAD_CERTS` | `false` |

## Using dart-define

The `--dart-define` flag passes compile-time constants that work on all platforms including Web.

### Command Line

```bash
flutter run \
  --dart-define=SOCKET_PROXY=http://localhost:9091 \
  --dart-define=SOCKET_PROXY_MODE=reverse \
  --dart-define=SOCKET_PROXY_ENABLED=true
```

### In VS Code (launch.json)

```json
{
  "configurations": [
    {
      "name": "Debug with Proxy",
      "request": "launch",
      "type": "dart",
      "args": [
        "--dart-define=SOCKET_PROXY=http://localhost:9091",
        "--dart-define=SOCKET_PROXY_ENABLED=true"
      ]
    }
  ]
}
```

### In Android Studio

Add to **Run/Debug Configurations > Additional run args**:

```
--dart-define=SOCKET_PROXY=http://localhost:9091
```

## Using Environment Variables

Environment variables work only on dart:io platforms (mobile, desktop). They are **not available on Web**.

### Setting ENV Variables

```bash
# Unix/macOS
export SOCKET_PROXY=http://localhost:9091
export SOCKET_PROXY_MODE=reverse
flutter run

# Windows PowerShell
$env:SOCKET_PROXY = "http://localhost:9091"
flutter run

# Windows CMD
set SOCKET_PROXY=http://localhost:9091
flutter run
```

### In a .env File

Create a `.env` file and source it before running:

```bash
# .env
SOCKET_PROXY=http://localhost:9091
SOCKET_PROXY_MODE=reverse
SOCKET_PROXY_ENABLED=true
```

```bash
source .env && flutter run
```

## Package-Specific Configuration

### dio_debugger

```dart
DioDebugger.interceptor(
  proxyBaseUrl: 'http://localhost:9091',  // Proxy server URL
  proxyHttpPath: '/httpproxy',            // HTTP proxy endpoint
  enabled: true,                          // Enable/disable
  mode: 'reverse',                        // 'reverse', 'forward', or 'none'
  skipPaths: ['/health', '/metrics'],     // Paths to bypass
  skipHosts: ['localhost'],               // Hosts to bypass
  allowPaths: ['/api/'],                  // Only proxy these paths
);
```

### http_debugger

```dart
HttpDebugger.enable(
  mode: 'reverse',
  upstreamBaseUrl: 'https://api.example.com',
  proxyBaseUrl: 'http://localhost:9091',
  proxyHttpPath: '/httpproxy',
  allowBadCertificates: false,
  skipPaths: [RegExp(r'^/health')],
  skipHosts: ['localhost', '127.0.0.1'],
);
```

### web_socket_debugger / web_socket_channel_debugger

```dart
WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',        // Target WebSocket URL
  proxyBaseUrl: 'http://localhost:9091',  // Proxy server URL
  proxyPath: '/wsproxy',                  // WebSocket proxy endpoint
  enabled: true,                          // Enable/disable
  mode: 'reverse',                        // 'reverse', 'forward', or 'none'
);
```

### socket_io_debugger

```dart
SocketIoDebugger.attach(
  baseUrl: 'https://example.com',         // Socket.IO server URL
  path: '/socket.io/',                    // Engine.IO endpoint path
  proxyBaseUrl: 'http://localhost:9091',  // Proxy server URL
  proxyHttpPath: '/wsproxy',              // WebSocket proxy endpoint
  enabled: true,                          // Enable/disable
);
```

## Android Emulator Configuration

The Android emulator cannot access `localhost` on the host machine. Use `10.0.2.2` instead:

```dart
import 'dart:io' show Platform;

final proxyUrl = Platform.isAndroid 
  ? 'http://10.0.2.2:9091' 
  : 'http://localhost:9091';

DioDebugger.interceptor(
  proxyBaseUrl: proxyUrl,
);
```

Or use `--dart-define`:

```bash
# For Android emulator
flutter run --dart-define=SOCKET_PROXY=http://10.0.2.2:9091
```

## iOS Simulator Configuration

iOS Simulator can use `localhost` directly:

```dart
final proxyUrl = 'http://localhost:9091';
```

## Conditional Configuration

### Based on Build Mode

```dart
import 'package:flutter/foundation.dart';

DioDebugger.interceptor(
  proxyBaseUrl: 'http://localhost:9091',
  enabled: kDebugMode,  // Only in debug builds
);
```

### Based on Flavor/Environment

```dart
const flavor = String.fromEnvironment('FLAVOR', defaultValue: 'dev');

DioDebugger.interceptor(
  proxyBaseUrl: 'http://localhost:9091',
  enabled: flavor == 'dev',
);
```

```bash
flutter run --dart-define=FLAVOR=dev
```

## Next Steps

- [Proxy Modes](./proxy-modes.md) - Understanding reverse vs forward proxy
- [Platform Support](./platform-support.md) - Platform-specific differences
- [Troubleshooting](./troubleshooting.md) - Common configuration issues
