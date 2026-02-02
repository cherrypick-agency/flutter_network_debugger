import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';
import 'package:web_socket/web_socket.dart' as ws;
import 'package:web_socket_debugger/web_socket_debugger.dart';

void main() {
  group('e2e (web_socket_debugger) → edge cases', () {
    HttpServer? upstream;
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path.startsWith('/echo') &&
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

    test('preserves baseUrl query params and still records ws session',
        () async {
      final upstreamWs =
          'ws://127.0.0.1:${upstream!.port}/echo?space=a%20b&unicode=%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82';

      final cfg = WebSocketDebugger.attach(
        baseUrl: upstreamWs,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: 'http://localhost:9091/wsproxy?junk=1',
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

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('binary frames are proxied (echo) and recorded', () async {
      final upstreamWs = 'ws://127.0.0.1:${upstream!.port}/echo-binary';
      final cfg = WebSocketDebugger.attach(
        baseUrl: upstreamWs,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: '/wsproxy/',
        mode: 'reverse',
        enabled: true,
      );

      final socket = await WebSocketDebugger.connect(config: cfg);
      final echoed = Completer<Uint8List>();
      socket.events.listen((e) {
        switch (e) {
          case ws.BinaryDataReceived(data: final data):
            if (!echoed.isCompleted) echoed.complete(Uint8List.fromList(data));
          default:
        }
      });

      final payload = Uint8List.fromList([0, 1, 2, 3, 255]);
      socket.sendBytes(payload);
      expect(await echoed.future.timeout(const Duration(seconds: 10)), payload);
      await socket.close();

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('/echo-binary')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('two connections create at least two ws sessions', () async {
      final a = 'ws://127.0.0.1:${upstream!.port}/echo/a';
      final b = 'ws://127.0.0.1:${upstream!.port}/echo/b';

      final cfgA = WebSocketDebugger.attach(
        baseUrl: a,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      final cfgB = WebSocketDebugger.attach(
        baseUrl: b,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );

      final s1 = await WebSocketDebugger.connect(config: cfgA);
      final s2 = await WebSocketDebugger.connect(config: cfgB);
      await s1.close();
      await s2.close();

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      final count = sessions.where((s) => s['kind']?.toString() == 'ws').length;
      expect(count, greaterThanOrEqualTo(2));
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
