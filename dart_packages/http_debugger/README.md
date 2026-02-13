# http_debugger

<p align="left">
  <a href="https://pub.dev/packages/http_debugger"><img src="https://img.shields.io/pub/v/http_debugger.svg" alt="pub version" /></a>
  <a href="https://pub.dev/packages/http_debugger/score"><img src="https://img.shields.io/pub/likes/http_debugger" alt="likes" /></a>
  <a href="https://pub.dev/packages/http_debugger/score"><img src="https://img.shields.io/pub/points/http_debugger" alt="pub points" /></a>
  <a href="https://pub.dev/packages/http_debugger/score"><img src="https://img.shields.io/pub/popularity/http_debugger" alt="popularity" /></a>
</p>

## Getting started with Network Debugger

1) CLI (WEB UI in browser) — fastest way to start

```bash
dart pub global activate network_debugger
network_debugger
```

This starts the proxy and opens the UI in your browser:
- UI: `http://localhost:9092/`
- Proxy base (HTTP/WebSocket forward): `http://localhost:9091`

2) Desktop App (Native GUI)

Download the desktop application from
[GitHub Releases](https://github.com/cherrypick-agency/flutter_network_debugger/releases).
It bundles the proxy server and UI.

> This package is for integrating Network Debugger with `dart:io` HTTP and/or
> `package:http`.

> Part of the [network_debugger](https://pub.dev/packages/network_debugger) ecosystem

Small utility that globally enables forward or reverse HTTP proxying via
`HttpOverrides`, so all `dart:io` HTTP traffic in your app can be routed
through a local proxy (e.g. `network-debugger` on `http://localhost:9091`).

## Installation

```yaml
dependencies:
  http_debugger: ^0.2.0
```

## Starting the Proxy

Before using `http_debugger`, you need to start the network debugger proxy server. Install and run it with:

```bash
# Install the CLI globally
dart pub global activate network_debugger

# Start the proxy (proxy port 9091, UI opens on 9092)
network_debugger
```

Proxy base will be `http://localhost:9091`. The web UI opens on `http://localhost:9092`.

For more options and programmatic usage, see the [network_debugger package documentation](https://pub.dev/packages/network_debugger).

## Quick start 

Start with sensible defaults in one line (reverse by default; falls back to forward if `upstreamBaseUrl` is not provided):

```dart
void main() {
  HttpDebugger.enable();
}
```

Advanced:

```dart
HttpDebugger.enable(
  // reverse by default
  upstreamBaseUrl: 'https://api.example.test',
  proxyBaseUrl: 'http://localhost:9091',      // Android emulator: 10.0.2.2:9091
  proxyHttpPath: '/httpproxy',
  allowBadCertificates: true,
  skipPaths: ['/metrics'],
  skipHosts: ['auth.local'],
  skipMethods: ['OPTIONS'],
  // allow* work as whitelist if provided
  // allowPaths: [...],
  // allowHosts: [...],
  // allowMethods: ['GET','POST'],
  // forward‑proxy extras
  // mode: 'forward',
  // bypassHosts: ['localhost', '127.0.0.1'],
);
```

## Auto mode via ENV/--dart-define

```dart
void main() {
  HttpDebugger.enableAuto();
}
```

Supported variables

- `HTTP_PROXY_MODE|PROXY_MODE`: `reverse|forward|none` (defaults to `reverse`)
- `DIO_DEBUGGER_ENABLED|HTTP_PROXY_ENABLED`: on/off (defaults to on)
- Reverse mode:
  - `UPSTREAM_BASE_URL|API_HOST`
  - `PROXY_BASE_URL|HTTP_PROXY`
  - `PROXY_HTTP_PATH|HTTP_PROXY_PATH` (defaults to `/httpproxy`)
- Forward mode:
  - `HTTP_PROXY|PROXY_BASE_URL` (`http://host:port` or `host:port`)
  - `HTTP_PROXY_ALLOW_BAD_CERTS`

Defaults when ENV is not set:
- If mode resolves to `forward` and no proxy is provided, uses
  `http://localhost:9091` (or `http://10.0.2.2:9091` on Android emulator).
- If mode resolves to `reverse` but `UPSTREAM_BASE_URL` is missing, the library
  safely falls back to forward proxy with the default proxy above.

## Forward proxy

```dart
import 'package:http_debugger/http_debugger.dart';

void main() {
  HttpDebugger.enableForwardProxy(
    const HttpDebuggerConfig(
      // Android emulator: use 10.0.2.2:9091
      proxyHostPort: 'localhost:9091',
      allowBadCertificates: true, // convenient for local dev
      bypassHosts: ['localhost', '127.0.0.1'],
    ),
  );
}
```

## Reverse proxy

```dart
import 'package:http_debugger/http_debugger.dart';

void main() {
  HttpDebugger.enableReverseProxy(
    const HttpReverseProxyConfig(
      upstreamBaseUrl: 'https://api.example.test',
      proxyBaseUrl: 'http://localhost:9091',
      proxyHttpPath: '/httpproxy',
      allowBadCertificates: true,
      // optional filters
      skipPaths: ['/metrics'],
      skipMethods: ['OPTIONS'],
    ),
  );
}
```

After enabling, a request `GET https://api.example.test/path` becomes:
```
http://localhost:9091/httpproxy?_target=https://api.example.test/path
```

## Isolated client (package:http) — Web/IO

If you need to test an isolated `package:http` client (and also support Web), wrap your client:

```dart
import 'package:http/http.dart' as http;
import 'package:http_debugger/http_debugger.dart';

final base = http.Client();
final client = HttpDebuggerClient.wrap(
  base,
  upstreamBaseUrl: 'https://api.example.test',
  proxyBaseUrl: 'http://localhost:9091', // Android emulator: 10.0.2.2:9091
  proxyHttpPath: '/httpproxy',            // or '/proxy' to unify HTTP/WS endpoint
  // Optional filters:
  // skipPaths: ['/metrics'],
  // skipHosts: ['auth.local'],
  // skipMethods: ['OPTIONS'],
);

// Use `client` everywhere instead of top-level http.get(...)
final resp = await client.get(Uri.parse('/users?limit=10'));
```

This approach works on Web and IO, since URL rewriting happens before the request is sent.

## Notes

- Global `HttpOverrides` works on `dart:io` platforms (iOS/Android/macOS/Windows/Linux). On Web it is a no‑op.
- For Web, use the isolated `package:http` wrapper above; it requires you to pass the wrapped client through your code (not global).
- Affects all clients using the standard `HttpClient` under the hood (including
  `package:http`, Dio with `IOHttpClientAdapter`, etc.).
- For Web you need a separate intercepting approach (e.g. patching `fetch` via
  `dart:js_interop`).


