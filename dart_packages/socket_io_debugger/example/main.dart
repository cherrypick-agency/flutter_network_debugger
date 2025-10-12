import 'dart:io';

import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';

Future<void> main() async {
  // Attach proxy config (reverse by default). Adjust values for your local proxy.
  final cfg = SocketIoDebugger.attach(
    baseUrl: 'https://chat.example.com',
    socketPath: '/socket.io',
    proxyBaseUrl: 'http://localhost:9091',
    proxyHttpPath: '/wsproxy',
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
    await HttpOverrides.runZoned(() async {
      socket.connect();
      await Future<void>.delayed(const Duration(seconds: 1));
      socket.dispose();
    }, createHttpClient: (_) => cfg.httpClientFactory!());
  } else {
    socket.connect();
    await Future<void>.delayed(const Duration(seconds: 1));
    socket.dispose();
  }
}
