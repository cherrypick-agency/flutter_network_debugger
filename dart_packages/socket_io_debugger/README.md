socket_io_debugger
===================

One-call helper to attach a proxy to a Socket.IO client (reverse/forward modes).

Example:

```dart
final cfg = SocketIoDebugger.attach(
  baseUrl: 'https://example.com',
  socketPath: '/socket.io',
);
final socket = io.io(
  cfg.effectiveBaseUrl,
  io.OptionBuilder()
    .setTransports(['websocket'])
    .setPath(cfg.effectivePath)
    .setQuery(cfg.query)
    .build(),
);
if (cfg.useForwardOverrides) {
  await HttpOverrides.runZoned(() => socket.connect(), createHttpClient: (_) => cfg.httpClientFactory!());
} else {
  socket.connect();
}
```

Environment variables:
- HTTP_PROXY_MODE=reverse|forward|none
- HTTP_PROXY, SOCKET_PROXY
- HTTP_PROXY_PATH, SOCKET_PROXY_PATH
- HTTP_PROXY_ALLOW_BAD_CERTS=true|false
- HTTP_PROXY_ENABLED=true|false


