# dio_debugger

HTTP debugging interceptor for [package:dio](https://pub.dev/packages/dio).

## Installation

```yaml
dependencies:
  dio: ^5.0.0
  dio_debugger: ^0.2.0
```

## Basic Usage

```dart
import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

void main() {
  final dio = Dio();
  
  // Add debugger interceptor
  dio.interceptors.add(
    DioDebugger.interceptor(
      proxyBaseUrl: 'http://localhost:9091',
    ),
  );
  
  // Make requests as usual
  dio.get('https://api.example.com/users');
}
```

## API Reference

### DioDebugger.interceptor()

Creates a Dio interceptor that routes requests through the proxy.

```dart
static Interceptor interceptor({
  String proxyBaseUrl = 'http://localhost:9091',
  String proxyHttpPath = '/httpproxy',
  bool? enabled,
  String? mode,
  List<Pattern>? skipPaths,
  List<Pattern>? skipHosts,
  List<String>? skipMethods,
  List<Pattern>? allowPaths,
  List<Pattern>? allowHosts,
  List<String>? allowMethods,
});
```

| Parameter | Type | Default | Description |
| --------- | ---- | ------- | ----------- |
| `proxyBaseUrl` | `String` | `http://localhost:9091` | Proxy server URL |
| `proxyHttpPath` | `String` | `/httpproxy` | HTTP proxy endpoint path |
| `enabled` | `bool?` | `true` | Enable/disable the interceptor |
| `mode` | `String?` | `reverse` | Proxy mode: `reverse`, `forward`, `none` |
| `skipPaths` | `List<Pattern>?` | `null` | URL paths to bypass (exact string or RegExp) |
| `skipHosts` | `List<Pattern>?` | `null` | Hosts to bypass |
| `skipMethods` | `List<String>?` | `null` | HTTP methods to bypass (e.g., `['OPTIONS']`) |
| `allowPaths` | `List<Pattern>?` | `null` | Only proxy these paths (whitelist mode) |
| `allowHosts` | `List<Pattern>?` | `null` | Only proxy these hosts |
| `allowMethods` | `List<String>?` | `null` | Only proxy these methods |

## Configuration Examples

### Basic Setup

```dart
dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
  ),
);
```

### With Skip Rules

```dart
dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
    skipPaths: ['/health', '/metrics', RegExp(r'^/internal/')],
    skipHosts: ['localhost', '127.0.0.1'],
    skipMethods: ['OPTIONS'],
  ),
);
```

### With Allow Rules (Whitelist)

```dart
dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
    allowPaths: ['/api/v1/', '/api/v2/'],
    allowHosts: ['api.example.com'],
  ),
);
```

### Debug Mode Only

```dart
import 'package:flutter/foundation.dart';

dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
    enabled: kDebugMode,
  ),
);
```

### Android Emulator

```dart
import 'dart:io' show Platform;

dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: Platform.isAndroid 
      ? 'http://10.0.2.2:9091' 
      : 'http://localhost:9091',
  ),
);
```

## Environment Variables

| Variable | Description |
| -------- | ----------- |
| `DIO_DEBUGGER_ENABLED` | Enable/disable (`true`/`false`) |
| `SOCKET_PROXY` | Proxy server URL |
| `SOCKET_PROXY_PATH` | Proxy HTTP endpoint path |
| `SOCKET_PROXY_MODE` | Proxy mode |

```bash
flutter run --dart-define=DIO_DEBUGGER_ENABLED=true
```

## How It Works

1. Interceptor catches outgoing requests in `onRequest`
2. Rewrites URL to proxy format: `{proxyBaseUrl}{proxyHttpPath}?_target={originalUrl}`
3. Proxy server extracts original URL and forwards request
4. Response flows back through proxy to client

### Request Transformation

Original request:
```
GET https://api.example.com/users?page=1
```

Transformed request:
```
GET http://localhost:9091/httpproxy?_target=https://api.example.com/users?page=1
```

## Integration with BaseOptions

```dart
final dio = Dio(BaseOptions(
  baseUrl: 'https://api.example.com',
  connectTimeout: Duration(seconds: 30),
));

// Interceptor works with baseUrl
dio.interceptors.add(
  DioDebugger.interceptor(
    proxyBaseUrl: 'http://localhost:9091',
  ),
);

// This becomes: /httpproxy?_target=https://api.example.com/users
dio.get('/users');
```

## Error Handling

The interceptor passes through errors unchanged:

```dart
try {
  await dio.get('https://api.example.com/users');
} on DioException catch (e) {
  // Handle error as usual
  print('Error: ${e.message}');
}
```

## Troubleshooting

### Requests Not Proxied

1. Check interceptor is added before making requests
2. Verify `enabled: true`
3. Check skip/allow rules aren't excluding the request

### Connection Refused

1. Verify proxy server is running: `curl http://localhost:9091/healthz`
2. For Android emulator, use `10.0.2.2` instead of `localhost`

### HTTPS Issues

For development with self-signed certificates, configure Dio:

```dart
(dio.httpClientAdapter as IOHttpClientAdapter).createHttpClient = () {
  final client = HttpClient();
  client.badCertificateCallback = (cert, host, port) => true;
  return client;
};
```

## See Also

- [Quick Start Guide](../quick-start.md)
- [Configuration Guide](../configuration.md)
- [Troubleshooting](../troubleshooting.md)
