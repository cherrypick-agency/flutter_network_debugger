# API Reference

Complete API reference for all Network Debugger Dart packages.

## Package Index

- [dio_debugger](#dio_debugger)
- [http_debugger](#http_debugger)
- [web_socket_debugger](#web_socket_debugger)
- [web_socket_channel_debugger](#web_socket_channel_debugger)
- [socket_io_debugger](#socket_io_debugger)

---

## dio_debugger

### DioDebugger

```dart
class DioDebugger {
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
}
```

**Returns:** `Interceptor` - Add to `Dio.interceptors`

---

## http_debugger

### HttpDebugger

```dart
class HttpDebugger {
  // Enable global forward proxy
  static void enableForwardProxy(HttpDebuggerConfig config);
  
  // Enable global reverse proxy
  static void enableReverseProxy(HttpReverseProxyConfig config);
  
  // Universal enable method
  static void enable({
    String mode = 'reverse',
    String? upstreamBaseUrl,
    String? proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    bool allowBadCertificates = false,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
    List<Pattern> bypassHosts = const [],
  });
  
  // Disable global proxy
  static void disable();
  
  // Run code in zone with reverse proxy
  static T runZonedWithReverseProxy<T>(
    HttpReverseProxyConfig config,
    T Function() body,
  );
}
```

### HttpDebuggerClient

```dart
class HttpDebuggerClient {
  static http.Client wrap(
    http.Client inner, {
    required String proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    String? upstreamBaseUrl,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
  });
}
```

### HttpDebuggerConfig

```dart
class HttpDebuggerConfig {
  final String proxyHostPort;
  final bool allowBadCertificates;
  final List<Pattern> bypassHosts;
  
  const HttpDebuggerConfig({
    required this.proxyHostPort,
    this.allowBadCertificates = false,
    this.bypassHosts = const [],
  });
}
```

### HttpReverseProxyConfig

```dart
class HttpReverseProxyConfig {
  final String upstreamBaseUrl;
  final String proxyBaseUrl;
  final String proxyHttpPath;
  final bool allowBadCertificates;
  final List<Pattern>? skipPaths;
  final List<Pattern>? skipHosts;
  final List<String>? skipMethods;
  final List<Pattern>? allowPaths;
  final List<Pattern>? allowHosts;
  final List<String>? allowMethods;
  
  const HttpReverseProxyConfig({
    required this.upstreamBaseUrl,
    required this.proxyBaseUrl,
    this.proxyHttpPath = '/httpproxy',
    this.allowBadCertificates = false,
    this.skipPaths,
    this.skipHosts,
    this.skipMethods,
    this.allowPaths,
    this.allowHosts,
    this.allowMethods,
  });
}
```

---

## web_socket_debugger

### WebSocketDebugger

```dart
class WebSocketDebugger {
  static WebSocketProxyConfig attach({
    required String baseUrl,
    String proxyBaseUrl = 'http://localhost:9091',
    String proxyPath = '/wsproxy',
    bool? enabled,
    String? mode,
  });
  
  static Future<WebSocket> connect({
    required WebSocketProxyConfig config,
    Map<String, dynamic>? headers,
  });
}
```

### WebSocketProxyConfig

```dart
class WebSocketProxyConfig {
  final Uri connectUrl;
  final Map<String, dynamic> query;
  final bool useForwardOverrides;
  final Object Function()? httpClientFactory;
  
  const WebSocketProxyConfig({
    required this.connectUrl,
    required this.query,
    required this.useForwardOverrides,
    this.httpClientFactory,
  });
}
```

---

## web_socket_channel_debugger

### WebSocketChannelDebugger

```dart
class WebSocketChannelDebugger {
  static WscProxyConfig attach({
    required String baseUrl,
    String proxyBaseUrl = 'http://localhost:9091',
    String proxyPath = '/wsproxy',
    bool? enabled,
    String? mode,
  });
  
  static WebSocketChannel connect({
    required WscProxyConfig config,
    Map<String, dynamic>? headers,
  });
}
```

### WscProxyConfig

```dart
class WscProxyConfig {
  final Uri connectUrl;
  final Map<String, dynamic> query;
  final bool useForwardOverrides;
  final Object Function()? httpClientFactory;
  
  const WscProxyConfig({
    required this.connectUrl,
    required this.query,
    required this.useForwardOverrides,
    this.httpClientFactory,
  });
}
```

---

## socket_io_debugger

### SocketIoDebugger

```dart
class SocketIoDebugger {
  static SocketIoConfig attach({
    required String baseUrl,
    String path = '/socket.io/',
    String? proxyBaseUrl,
    String? proxyHttpPath,
    bool? enabled,
  });
}
```

### SocketIoConfig

```dart
class SocketIoConfig {
  final String effectiveBaseUrl;
  final String effectivePath;
  final Map<String, dynamic> query;
  final bool useForwardOverrides;
  final HttpClient Function()? httpClientFactory;
  
  const SocketIoConfig({
    required this.effectiveBaseUrl,
    required this.effectivePath,
    required this.query,
    required this.useForwardOverrides,
    this.httpClientFactory,
  });
}
```

---

## Common Types

### Pattern Matching

Parameters accepting `List<Pattern>` support both:
- **String** - Exact match or contains
- **RegExp** - Regular expression match

```dart
// String patterns
skipPaths: ['/health', '/metrics']

// RegExp patterns  
skipPaths: [RegExp(r'^/api/v\d+/health')]

// Mixed
skipPaths: ['/health', RegExp(r'^/internal/')]
```

### Mode Values

| Value | Description |
| ----- | ----------- |
| `'reverse'` | Default. Client connects to proxy, proxy forwards to upstream |
| `'forward'` | Client connects to upstream via system proxy (dart:io only) |
| `'none'` | Bypass proxy completely |

### Boolean String Values

Environment variables and dart-define accept these truthy values:
- `'true'`, `'1'`, `'yes'`, `'on'` - Enabled
- `'false'`, `'0'`, `'no'`, `'off'` - Disabled

---

## Environment Variables Reference

| Variable | Packages | Description |
| -------- | -------- | ----------- |
| `SOCKET_PROXY` | All | Proxy server URL |
| `SOCKET_PROXY_PATH` | All | Proxy endpoint path |
| `SOCKET_PROXY_MODE` | All | `reverse`, `forward`, `none` |
| `SOCKET_PROXY_ENABLED` | All | Enable/disable |
| `SOCKET_PROXY_ALLOW_BAD_CERTS` | All | Allow self-signed certs |
| `SOCKET_UPSTREAM_URL` | WS packages | Override upstream URL |
| `SOCKET_UPSTREAM_PATH` | socket_io | Override Socket.IO path |
| `SOCKET_UPSTREAM_TARGET` | WS packages | Full `_target` value |
| `DIO_DEBUGGER_ENABLED` | dio, socket_io | Alternative enable flag |

---

## Proxy Endpoints

| Endpoint | Protocol | Usage |
| -------- | -------- | ----- |
| `/httpproxy` | HTTP | dio_debugger, http_debugger |
| `/wsproxy` | WebSocket | web_socket_debugger, web_socket_channel_debugger, socket_io_debugger |
| `/healthz` | HTTP | Health check |
| `/_api/v1/sessions` | HTTP | Session management API |

---

## See Also

- [Package Documentation](./packages/)
- [Configuration Guide](./configuration.md)
- [Troubleshooting](./troubleshooting.md)
