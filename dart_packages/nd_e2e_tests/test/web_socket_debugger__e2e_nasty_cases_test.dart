import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';
import 'package:web_socket/web_socket.dart' as ws;
import 'package:web_socket_debugger/web_socket_debugger.dart';

void main() {
  group('e2e (web_socket_debugger) → nasty cases', () {
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

    Future<void> _connectRoundTrip({
      required String path,
      required String text,
      Uint8List? bytes,
    }) async {
      final cfg = WebSocketDebugger.attach(
        baseUrl: 'ws://127.0.0.1:${upstream!.port}$path',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      final socket = await WebSocketDebugger.connect(config: cfg);

      final gotText = Completer<String>();
      final gotBytes = Completer<Uint8List>();
      socket.events.listen((e) {
        switch (e) {
          case ws.TextDataReceived(text: final t):
            if (!gotText.isCompleted) gotText.complete(t);
          case ws.BinaryDataReceived(data: final d):
            if (!gotBytes.isCompleted) gotBytes.complete(Uint8List.fromList(d));
          default:
        }
      });

      socket.sendText(text);
      expect(await gotText.future.timeout(const Duration(seconds: 10)), text);
      if (bytes != null) {
        socket.sendBytes(bytes);
        expect(
          await gotBytes.future.timeout(const Duration(seconds: 10)),
          bytes,
        );
      }
      await socket.close();
    }

    test('mixed text+binary frames pass through', () async {
      await _connectRoundTrip(
        path: '/echo/mixed',
        text: 'hi',
        bytes: Uint8List.fromList([1, 2, 3, 4, 255]),
      );

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
          sessions.any((s) => (s['target'] as String).contains('/echo/mixed')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('reconnect creates multiple ws sessions', () async {
      await _connectRoundTrip(path: '/echo/reconnect', text: 'one');
      await _connectRoundTrip(path: '/echo/reconnect', text: 'two');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 500);
      final count = sessions
          .where((s) => (s['target'] as String).contains('/echo/reconnect'))
          .length;
      expect(count, greaterThanOrEqualTo(2));
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('concurrent connects (20) are recorded', () async {
      final futures = List.generate(
        20,
        (i) => _connectRoundTrip(path: '/echo/c$i', text: 'hi$i'),
      );
      await Future.wait(futures);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 500);
      final count = sessions
          .where((s) => (s['target'] as String).contains('/echo/c'))
          .length;
      expect(count, greaterThanOrEqualTo(20));
    }, timeout: const Timeout(Duration(seconds: 120)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
