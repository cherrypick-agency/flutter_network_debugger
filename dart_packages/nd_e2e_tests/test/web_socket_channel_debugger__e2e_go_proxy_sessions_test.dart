import 'dart:async';
import 'dart:io';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';
import 'package:web_socket_channel/web_socket_channel.dart' as wsc;

void main() {
  group('e2e (web_socket_channel_debugger) → go run proxy + sessions API', () {
    HttpServer? upstream;
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path == '/echo' &&
            WebSocketTransformer.isUpgradeRequest(req)) {
          final ws = await WebSocketTransformer.upgrade(req);
          ws.listen((d) => ws.add(d));
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
      final proxyUrl = Uri.parse('${proxy!.proxyHttpBase}')
          .replace(scheme: 'ws', path: '/wsproxy', queryParameters: {
        '_target': upstreamWs,
      });

      final ch = wsc.WebSocketChannel.connect(proxyUrl);
      final c = Completer<String>();
      ch.stream.listen((e) {
        if (!c.isCompleted) c.complete(e.toString());
      });
      ch.sink.add('hi');
      expect(await c.future.timeout(const Duration(seconds: 10)), 'hi');
      await ch.sink.close();

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
