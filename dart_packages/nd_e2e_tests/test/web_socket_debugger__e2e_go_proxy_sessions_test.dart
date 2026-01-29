import 'dart:async';
import 'dart:io';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';
import 'package:web_socket_debugger/web_socket_debugger.dart';
import 'package:web_socket/web_socket.dart' as ws;

void main() {
  group('e2e (web_socket_debugger) → go run proxy + sessions API', () {
    HttpServer? upstream;
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path == '/echo' &&
            WebSocketTransformer.isUpgradeRequest(req)) {
          final socket = await WebSocketTransformer.upgrade(req);
          socket.listen((data) => socket.add(data));
          return;
        }
        req.response.statusCode = 404;
        await req.response.close();
      });

      proxy = await GoNetworkDebuggerProcess.start();
      await clearSessions(proxy!.apiBase);
    });

    tearDown(() async {
      await upstream?.close(force: true);
      await proxy?.stop();
    });

    test('echo through /wsproxy записывает ws-сессию в Go', () async {
      final upstreamWs = 'ws://127.0.0.1:${upstream!.port}/echo';
      final cfg = WebSocketDebugger.attach(
        baseUrl: upstreamWs,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );

      final socket = await WebSocketDebugger.connect(config: cfg);
      final echoed = Completer<String>();
      socket.events.listen((e) {
        switch (e) {
          case ws.TextDataReceived(text: final text):
            if (!echoed.isCompleted) echoed.complete(text);
          default:
        }
      });
      socket.sendText('hi');
      expect(await echoed.future.timeout(const Duration(seconds: 10)), 'hi');
      await socket.close();

      final sessions = await listSessions(proxy!.apiBase, types: 'ws');
      expect(
        sessions.any((s) {
          final kind = s['kind']?.toString();
          final target = s['target']?.toString() ?? '';
          return kind == 'ws' &&
              target.contains('127.0.0.1:${upstream!.port}') &&
              target.contains('/echo');
        }),
        isTrue,
        reason: 'не нашли ws-сессию, связанную с подключением через /wsproxy',
      );
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
