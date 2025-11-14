# dio_debugger

<p align="left">
  <a href="https://pub.dev/packages/dio_debugger"><img src="https://img.shields.io/pub/v/dio_debugger.svg" alt="pub version" /></a>
  <a href="https://pub.dev/packages/dio_debugger/score"><img src="https://img.shields.io/pub/likes/dio_debugger" alt="likes" /></a>
  <a href="https://pub.dev/packages/dio_debugger/score"><img src="https://img.shields.io/pub/points/dio_debugger" alt="pub points" /></a>
  <a href="https://pub.dev/packages/dio_debugger/score"><img src="https://img.shields.io/pub/popularity/dio_debugger" alt="popularity" /></a>
  <a href="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/dio_debugger.yml"><img src="https://github.com/cherrypick-agency/flutter_network_debugger/actions/workflows/dio_debugger.yml/badge.svg?branch=main" alt="CI" /></a>
</p>

> Part of the [network_debugger](https://pub.dev/packages/network_debugger) ecosystem

Lightweight utility that patches the provided `Dio` and attaches a reverse/forward proxy interceptor. Useful for local debugging, traffic interception, and bypassing CORS/certificates via your local proxy.

## Features
- One-liner attach: `DioDebugger.attach(dio)`
- Config sources (priority):
  1) `attach` arguments
  2) `dio.options.baseUrl` (fallback)
  3) `--dart-define` (`UPSTREAM_BASE_URL`, `PROXY_BASE_URL`, `PROXY_HTTP_PATH`, `DIO_DEBUGGER_ENABLED`/`HTTP_PROXY_ENABLED`)
  4) OS ENV (via conditional import; web-safe)
- Handles absolute URLs in `RequestOptions.path` — if `path` is already `http(s)://…`, it is proxied as is.
- Interceptor ordering: `insertFirst` (default `true`) — places the interceptor first.
- Skip/allow filters: `skip*`/`allow*` by paths/hosts/methods.

## Installation
Add to your `pubspec.yaml`:

```yaml
dependencies:
  dio: ^5.4.0
  dio_debugger: ^0.1.2
```

## Starting the Proxy

Before using `dio_debugger`, you need to start the network debugger proxy server. Install and run it with:

```bash
# Install the CLI globally
dart pub global activate network_debugger

# Start the proxy (proxy port 9091, UI opens on 9092)
network_debugger
```

Proxy base will be `http://localhost:9091`. The web UI opens on `http://localhost:9092`.

For more options and programmatic usage, see the [network_debugger package documentation](https://pub.dev/packages/network_debugger).

## Quick start (only 2 lines)

```dart
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:dio_debugger/dio_debugger.dart'; // 1

final dio = Dio(
  BaseOptions(baseUrl: 'https://api.example.com'),
);

if (kDebugMode) DioDebugger.attach(dio); // 2
```

### Advanced options
```dart
DioDebugger.attach(
  dio,
  insertFirst: true,         // place interceptor first
  enabled: null,             // if null — read from env: DIO_DEBUGGER_ENABLED/HTTP_PROXY_ENABLED (true|1|yes|on)
  skipPaths: ['/metrics'],  // bypass proxy for these paths
  skipHosts: ['auth.local'],
  skipMethods: ['OPTIONS'],
  allowPaths: null,          // when allow* is set, only matching requests go through proxy
  allowHosts: null,
  allowMethods: null,
  upstreamBaseUrl: 'https://api.example.test',
  // Custom URL where debug proxy is running
  // Use 'http://10.0.2.2:9091 if Android emulator
  proxyBaseUrl: 'http://localhost:9091',
  proxyHttpPath: '/httpproxy',
);
```

### Auto-reset capture on hot restart

During development, you may want to separate network traffic captured from different hot restarts. Use the `resetCaptureOnHotRestart` parameter to automatically clear previous sessions and start a new capture:

```dart
if (kDebugMode) {
  DioDebugger.attach(
    dio,
    resetCaptureOnHotRestart: true,  // Clear sessions on each hot restart
  );
}
```

This adds a `_resetCapture=true` query parameter to the first proxied request, which triggers the proxy to:
- Clear all previous sessions
- Increment the capture ID
- Start a new capture session

This approach has **zero latency** — no separate HTTP request is made, the reset happens inline with your first network call. The proxy will safely ignore this parameter if the reset feature is not supported.

### Configuration examples
- Via `--dart-define`:
```bash
--dart-define=UPSTREAM_BASE_URL=https://api.example.test \
--dart-define=PROXY_BASE_URL=http://localhost:9091 \
--dart-define=PROXY_HTTP_PATH=/httpproxy \
--dart-define=DIO_DEBUGGER_ENABLED=true
```

- Via OS ENV (on platforms with `dart:io`):
```
UPSTREAM_BASE_URL=https://api.example.test
PROXY_BASE_URL=http://localhost:9091
PROXY_HTTP_PATH=/httpproxy
DIO_DEBUGGER_ENABLED=true
```

After attach a request `GET /path` will go to:
```
http://localhost:9091/httpproxy?_target=https://api.example.test/path
```
If `options.path` is already an absolute `http(s)://…`, it is proxied without concatenating with `upstreamBaseUrl`.

## Notes
- The proxy must expose an endpoint `/httpproxy` that accepts `_target` query and forwards the request.
- If `upstreamBaseUrl` or `proxyBaseUrl` is empty, the package is a no‑op (safe for prod).
- If the proxy is provided without scheme and with port `:443`, `https` will be used automatically.

## License
MIT
