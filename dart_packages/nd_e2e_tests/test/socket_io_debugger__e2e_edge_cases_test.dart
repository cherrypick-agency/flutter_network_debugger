import 'dart:async';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (socket_io_debugger) → edge cases', () {
    GoNetworkDebuggerProcess? proxy;
    GoSocketIoUpstreamProcess? upstream;

    setUp(() async {
      proxy = await GoNetworkDebuggerProcess.start(logLevel: 'info');
      upstream =
          await GoSocketIoUpstreamProcess.start(port: await pickFreePort());
      await clearSessions(proxy!.apiBase);
    });

    tearDown(() async {
      await upstream?.stop();
      await proxy?.stop();
    });

    Future<void> _connectAndEcho(SocketIoConfig cfg) async {
      final gotInit = Completer<void>();
      final gotEcho = Completer<void>();
      final connectError = Completer<Object?>();

      final socket = io.io(
        cfg.effectiveBaseUrl,
        io.OptionBuilder()
            .setTransports(['websocket'])
            .setPath(cfg.effectivePath)
            .setQuery(cfg.query)
            .disableAutoConnect()
            .build(),
      );

      socket.onConnect((_) {
        if (!gotInit.isCompleted) gotInit.complete();
        socket.emit('hello', 'hi');
      });
      socket.onConnectError((data) {
        if (!connectError.isCompleted) connectError.complete(data);
      });
      socket.onError((data) {
        if (!connectError.isCompleted) connectError.complete(data);
      });
      socket.onAny((event, data) {
        final ok = data == 'hi' ||
            (data is List && data.isNotEmpty && data.first.toString() == 'hi');
        if (event.toString() == 'hello' && ok && !gotEcho.isCompleted) {
          gotEcho.complete();
        }
      });

      socket.connect();
      final first = await Future.any<Object?>([
        gotInit.future,
        connectError.future,
      ]).timeout(const Duration(seconds: 15));
      if (first != null && !gotInit.isCompleted) {
        throw StateError('connect error: $first');
      }
      await gotEcho.future.timeout(const Duration(seconds: 15));
      socket.dispose();
    }

    test('proxy path with trailing slash still works', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy/',
        enabled: true,
      );

      await _connectAndEcho(cfg);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('127.0.0.1:${upstream!.port}') &&
            (s['target'] as String).contains('/socket.io/')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('socket path without slashes is normalized and still works', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: 'socket.io',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      await _connectAndEcho(cfg);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('127.0.0.1:${upstream!.port}') &&
            (s['target'] as String).contains('/socket.io/')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('proxyHttpPath passed as full URL with query is normalized', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '${proxy!.proxyHttpBase}/wsproxy?junk=1',
        enabled: true,
      );

      await _connectAndEcho(cfg);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('127.0.0.1:${upstream!.port}') &&
            (s['target'] as String).contains('/socket.io/')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
