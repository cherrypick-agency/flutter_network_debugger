import 'dart:async';

import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:socket_io_debugger/socket_io_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (socket_io_debugger) → real Go wsproxy', () {
    GoNetworkDebuggerProcess? proxy;
    GoSocketIoUpstreamProcess? upstream;

    setUp(() async {
      // Logging is needed here for diagnostics: we want to see that wsproxy actually connected to upstream.
      proxy = await GoNetworkDebuggerProcess.start(logLevel: 'info');
      upstream =
          await GoSocketIoUpstreamProcess.start(port: await pickFreePort());
      await clearSessions(proxy!.apiBase);
    });

    tearDown(() async {
      await upstream?.stop();
      await proxy?.stop();
    });

    test('connects to Go Socket.IO server through /wsproxy and _target',
        () async {
      // We spin up a separate upstream (mini engine.io/socket.io) so that Go monitors it
      // and records the ws-session.
      final cfg = SocketIoDebugger.attach(
        baseUrl: upstream!.httpBase.toString(),
        path: '/socket.io/',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      expect(cfg.effectiveBaseUrl, proxy!.proxyHttpBase.toString());
      expect(cfg.effectivePath, '/wsproxy');
      expect(cfg.query.containsKey('_target'), isTrue);

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
      try {
        final first = await Future.any<Object?>([
          gotInit.future,
          connectError.future,
        ]).timeout(const Duration(seconds: 15));
        if (first != null && !gotInit.isCompleted) {
          throw StateError('connect error: $first');
        }
      } catch (_) {
        // ignore: avoid_print
        print('Upstream stdout:\\n${upstream!.stdoutLines.join("\\n")}');
        // ignore: avoid_print
        print('Go proxy stdout (tail):\\n'
            '${proxy!.stdoutLines.reversed.take(200).toList().reversed.join("\\n")}');
        rethrow;
      }
      try {
        await gotEcho.future.timeout(const Duration(seconds: 15));
      } catch (_) {
        // ignore: avoid_print
        print('Upstream stdout:\\n${upstream!.stdoutLines.join("\\n")}');
        // ignore: avoid_print
        print('Go proxy stdout (tail):\\n'
            '${proxy!.stdoutLines.reversed.take(200).toList().reversed.join("\\n")}');
        rethrow;
      }
      socket.dispose();

      // Minimal check that wsproxy actually dialed the upstream.
      expect(
        proxy!.stdoutLines.any((l) => l.contains('connected to upstream')),
        isTrue,
        reason: 'did not see wsproxy connection to upstream in logs',
      );

      // Main point: Go actually recorded the ws-session.
      final sessions = await listSessions(proxy!.apiBase, types: 'ws');
      expect(
        sessions.any((s) {
          final kind = s['kind']?.toString();
          final target = s['target']?.toString() ?? '';
          return kind == 'ws' &&
              target.contains('127.0.0.1:${upstream!.port}') &&
              target.contains('/socket.io/');
        }),
        isTrue,
        reason:
            'did not find ws-session associated with connection through /wsproxy',
      );
    }, timeout: const Timeout(Duration(seconds: 60)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
