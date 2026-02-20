# http_debugger

HTTP debugging for [package:http](https://pub.dev/packages/http) and `dart:io HttpClient`.

## Installation

```yaml
dependencies:
  http: ^1.0.0
  http_debugger: ^0.2.0
```

## Usage Options

http_debugger provides multiple integration methods:

1. **Client Wrapper** - Wrap `package:http` Client
2. **Global HttpOverrides** - Intercept all dart:io HttpClient traffic
3. **Zoned Execution** - Scoped interception

## Client Wrapper (Recommended)

Best for `package:http` users. Works on all platforms.

```dart
import 'package:http/http.dart' as http;
import 'package:http_debugger/http_debugger.dart';

// Create wrapped client
final client = HttpDebuggerClient.wrap(
  http.Client(),
  proxyBaseUrl: 'http://localhost:9091',
);

// Use as normal http.Client
final response = await client.get(Uri.parse('https://api.example.com/users'));
print(response.body);
```

### HttpDebuggerClient.wrap()

```dart
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
```

## Global HttpOverrides (dart:io only)

Intercepts all HTTP traffic from `HttpClient` and libraries using it.

```dart
import 'package:http_debugger/http_debugger.dart';

void main() {
  // Enable globally
  HttpDebugger.enable(
    mode: 'reverse',
    upstreamBaseUrl: 'https://api.example.com',
    proxyBaseUrl: 'http://localhost:9091',
  );
  
  // All HttpClient requests now go through proxy
  final client = HttpClient();
  final request = await client.getUrl(Uri.parse('https://api.example.com/users'));
  final response = await request.close();
  
  // Disable when done
  HttpDebugger.disable();
}
```

### HttpDebugger.enable()

```dart
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
```

| Parameter | Type | Default | Description |
| --------- | ---- | ------- | ----------- |
| `mode` | `String` | `reverse` | `reverse`, `forward`, or `none` |
| `upstreamBaseUrl` | `String?` | - | Base URL for upstream server |
| `proxyBaseUrl` | `String?` | `http://localhost:9091` | Proxy server URL |
| `proxyHttpPath` | `String` | `/httpproxy` | Proxy endpoint path |
| `allowBadCertificates` | `bool` | `false` | Accept self-signed certs |
| `skipPaths` | `List<Pattern>?` | - | Paths to bypass |
| `skipHosts` | `List<Pattern>?` | - | Hosts to bypass |
| `bypassHosts` | `List<Pattern>` | `[]` | Forward mode: hosts to connect directly |

## Zoned Execution

Scope proxy to specific code blocks:

```dart
import 'package:http_debugger/http_debugger.dart';

Future<void> fetchData() async {
  final result = await HttpDebugger.runZonedWithReverseProxy(
    HttpReverseProxyConfig(
      upstreamBaseUrl: 'https://api.example.com',
      proxyBaseUrl: 'http://localhost:9091',
    ),
    () async {
      // Only requests in this block go through proxy
      final client = HttpClient();
      final request = await client.getUrl(Uri.parse('https://api.example.com/users'));
      final response = await request.close();
      return utf8.decodeStream(response);
    },
  );
  print(result);
}
```

### HttpDebugger.runZonedWithReverseProxy()

```dart
static T runZonedWithReverseProxy<T>(
  HttpReverseProxyConfig config,
  T Function() body,
);
```

## Configuration Classes

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
}
```

### HttpDebuggerConfig (Forward Mode)

```dart
class HttpDebuggerConfig {
  final String proxyHostPort;  // e.g., "localhost:9091"
  final bool allowBadCertificates;
  final List<Pattern> bypassHosts;
}
```

## Examples

### Skip Health Checks

```dart
HttpDebugger.enable(
  proxyBaseUrl: 'http://localhost:9091',
  skipPaths: ['/health', '/healthz', '/metrics'],
);
```

### Whitelist API Endpoints

```dart
HttpDebugger.enable(
  proxyBaseUrl: 'http://localhost:9091',
  allowPaths: [RegExp(r'^/api/')],
  allowHosts: ['api.example.com', 'api.staging.example.com'],
);
```

### Forward Proxy Mode

```dart
HttpDebugger.enable(
  mode: 'forward',
  proxyBaseUrl: 'http://localhost:9091',
  bypassHosts: ['localhost', '127.0.0.1'],
  allowBadCertificates: true,
);
```

### Android Emulator

```dart
import 'dart:io' show Platform;

HttpDebugger.enable(
  proxyBaseUrl: Platform.isAndroid 
    ? 'http://10.0.2.2:9091' 
    : 'http://localhost:9091',
);
```

## Platform Support

| Feature | dart:io | Web |
| ------- | ------- | --- |
| Client Wrapper | Yes | Yes |
| Global HttpOverrides | Yes | No |
| Zoned Execution | Yes | No |
| Forward Mode | Yes | No |

## Environment Variables

| Variable | Description |
| -------- | ----------- |
| `SOCKET_PROXY` | Proxy server URL |
| `SOCKET_PROXY_PATH` | Proxy endpoint path |
| `SOCKET_PROXY_MODE` | `reverse`, `forward`, `none` |
| `SOCKET_PROXY_ENABLED` | `true`/`false` |

## Troubleshooting

### Stack Overflow with HttpOverrides

If you see stack overflow errors, ensure you're not creating recursive HttpClient calls:

```dart
// Wrong - causes recursion
HttpOverrides.global = MyOverrides();

// Correct - use provided methods
HttpDebugger.enable(...);
```

### Requests Not Intercepted

1. Verify `HttpDebugger.enable()` called before making requests
2. Check skip/allow rules
3. For zoned execution, ensure request is inside the zone

## See Also

- [Quick Start Guide](../quick-start.md)
- [Configuration Guide](../configuration.md)
- [Troubleshooting](../troubleshooting.md)
