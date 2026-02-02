import 'dart:async';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (socket_io_debugger) → nasty cases', () {
    GoNetworkDebuggerProcess? proxy;
    GoSocketIoUpstreamProcess? upstream;

    setUp(() async {
      proxy = await GoNetworkDebuggerProcess.start(logLevel: 'info');
      upstream = await GoSocketIoUpstreamProcess.start(
        port: await pickFreePort(),
        path: '/custom/socket.io/',
      );
      await clearSessions(proxy!.apiBase);
    });

    tearDown(() async {
      await upstream?.stop();
      await proxy?.stop();
    });

    Future<void> _connectAndEcho(
      SocketIoConfig cfg, {
      String namespace = '',
      Map<String, dynamic>? extraQuery,
    }) async {
      final gotInit = Completer<void>();
      final gotEcho = Completer<void>();
      final connectError = Completer<Object?>();

      final base = namespace.isEmpty
          ? cfg.effectiveBaseUrl
          : '${cfg.effectiveBaseUrl}${namespace.startsWith('/') ? '' : '/'}$namespace';

      final query = <String, dynamic>{...cfg.query, ...?extraQuery};

      final socket = io.io(
        base,
        (io.OptionBuilder()
            .setTransports(['websocket'])
            .setPath(cfg.effectivePath)
            .setQuery(query)
            .disableAutoConnect()
            .build()
          ..['forceNew'] = true),
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

    test('custom socket path works', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/custom/socket.io/',
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
            (s['target'] as String).contains('/custom/socket.io/')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('namespace is preserved (baseUrl path) and still works', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: '${upstream!.httpBase}/chat',
        path: '/custom/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      // attach() already preserves namespace in effectiveBaseUrl
      await _connectAndEcho(cfg);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('/custom/socket.io/')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('extra query params do not break handshake', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/custom/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      await _connectAndEcho(cfg, extraQuery: {'foo': 'bar'});
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('reconnect creates multiple sessions', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/custom/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      await _connectAndEcho(cfg);
      await _connectAndEcho(cfg);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 500);
      final count = sessions
          .where((s) => (s['target'] as String).contains('/custom/socket.io/'))
          .length;
      expect(count, greaterThanOrEqualTo(2));
    }, timeout: const Timeout(Duration(seconds: 120)));

    test('concurrent connects (10) are recorded', () async {
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/custom/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      final futures = List.generate(10, (_) => _connectAndEcho(cfg));
      await Future.wait(futures);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 500);
      final count = sessions
          .where((s) => (s['target'] as String).contains('/custom/socket.io/'))
          .length;
      expect(count, greaterThanOrEqualTo(10));
    }, timeout: const Timeout(Duration(seconds: 120)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
