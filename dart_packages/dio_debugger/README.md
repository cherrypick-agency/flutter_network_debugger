# dio_debugger

Lightweight utility that patches the provided `Dio` and attaches a reverse/forward proxy interceptor. Useful for local debugging, traffic interception, and bypassing CORS/certificates via your local proxy.

## Features
- One-liner attach: `DioDebugger.attach(dio)`
- Config sources (priority):
  1) `attach` arguments
  2) `--dart-define` (`UPSTREAM_BASE_URL`, `PROXY_BASE_URL`, `PROXY_HTTP_PATH`, `DIO_DEBUGGER_ENABLED`/`HTTP_PROXY_ENABLED`)
  3) OS ENV (via conditional import; web-safe)
- Handles absolute URLs in `RequestOptions.path` — if `path` is already `http(s)://…`, it is proxied as is.
- Interceptor ordering: `insertFirst` (default `true`) — places the interceptor first.
- Skip/allow filters: `skip*`/`allow*` by paths/hosts/methods.

## Installation
Add to your `pubspec.yaml`:

```yaml
dependencies:
  dio: ^5.4.0
  dio_debugger: ^0.1.0
```

## Quick start

```dart
import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

final dio = Dio(BaseOptions(baseUrl: 'https://41098f05e20d.ngrok-free.app'));
// Explicit attach with proxy params
DioDebugger.attach(
  dio,
  upstreamBaseUrl: 'https://41098f05e20d.ngrok-free.app',
  proxyBaseUrl: 'http://localhost:9091',
  proxyHttpPath: '/httpproxy',
);
```

### Advanced options
```dart
DioDebugger.attach(
  dio,
  insertFirst: true,         // place interceptor first
  enabled: null,             // if null — read from env: DIO_DEBUGGER_ENABLED/HTTP_PROXY_ENABLED (true|1|yes|on)
  skipPaths: ["/metrics"],  // bypass proxy for these paths
  skipHosts: ["auth.local"],
  skipMethods: ["OPTIONS"],
  allowPaths: null,          // when allow* is set, only matching requests go through proxy
  allowHosts: null,
  allowMethods: null,
  upstreamBaseUrl: 'https://41098f05e20d.ngrok-free.app',
  proxyBaseUrl: 'http://localhost:9091',
  proxyHttpPath: '/httpproxy',
);
```

### Configuration examples
- Via `--dart-define`:
```bash
--dart-define=UPSTREAM_BASE_URL=https://dev.api.padelme.app \
--dart-define=PROXY_BASE_URL=http://localhost:9091 \
--dart-define=PROXY_HTTP_PATH=/httpproxy \
--dart-define=DIO_DEBUGGER_ENABLED=true
```

- Via OS ENV (on platforms with `dart:io`):
```
UPSTREAM_BASE_URL=https://dev.api.padelme.app
PROXY_BASE_URL=http://localhost:9091
PROXY_HTTP_PATH=/httpproxy
DIO_DEBUGGER_ENABLED=true
```

After attach a request `GET /path` will go to:
```
http://localhost:9091/httpproxy?_target=https://dev.api.padelme.app/path
```
If `options.path` is already an absolute `http(s)://…`, it is proxied without concatenating with `upstreamBaseUrl`.

## Notes
- The proxy must expose an endpoint `/httpproxy` that accepts `_target` query and forwards the request.
- If `upstreamBaseUrl` or `proxyBaseUrl` is empty, the package is a no‑op (safe for prod).
- If the proxy is provided without scheme and with port `:443`, `https` will be used automatically.

## License
MIT
