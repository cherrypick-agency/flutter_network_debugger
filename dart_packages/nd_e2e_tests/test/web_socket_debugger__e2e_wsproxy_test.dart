import 'dart:async';
import 'dart:io';

import 'package:test/test.dart';
import 'package:web_socket/web_socket.dart' as ws;
import 'package:web_socket_debugger/web_socket_debugger.dart';

void main() {
  group('e2e via /wsproxy with _target (web_socket)', () {
    HttpServer? http;
    HttpServer? proxy;

    setUp(() async {
      // echo ws upstream
      http = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      http!.listen((req) async {
        if (req.uri.path == '/echo') {
          if (WebSocketTransformer.isUpgradeRequest(req)) {
            final socket = await WebSocketTransformer.upgrade(req);
            socket.listen((data) => socket.add(data));
          } else {
            req.response.statusCode = 400;
            await req.response.close();
          }
        } else {
          req.response.statusCode = 404;
          await req.response.close();
        }
      });

      // simple ws proxy
      proxy = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      proxy!.listen((req) async {
        if (req.uri.path == '/wsproxy') {
          final target = req.uri.queryParameters['_target'];
          if (target == null || !WebSocketTransformer.isUpgradeRequest(req)) {
            req.response.statusCode = 400;
            await req.response.close();
            return;
          }
          final upstream = await WebSocket.connect(target);
          final incoming = await WebSocketTransformer.upgrade(req);
          incoming.listen((d) => upstream.add(d),
              onDone: () => upstream.close());
          upstream.listen((d) => incoming.add(d),
              onDone: () => incoming.close());
        } else {
          req.response.statusCode = 404;
          await req.response.close();
        }
      });
    });

    tearDown(() async {
      await http?.close(force: true);
      await proxy?.close(force: true);
    });

    test('echo through proxy (reverse)', () async {
      final upstream = 'ws://localhost:${http!.port}/echo';
      final cfg = WebSocketDebugger.attach(
        baseUrl: upstream,
        proxyBaseUrl: 'http://localhost:${proxy!.port}',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );

      final socket = await WebSocketDebugger.connect(config: cfg);
      final completer = Completer<String>();
      socket.events.listen((e) {
        switch (e) {
          case ws.TextDataReceived(text: final text):
            completer.complete(text);
          default:
        }
      });
      socket.sendText('pong');
      final echoed = await completer.future.timeout(const Duration(seconds: 2));
      expect(echoed, 'pong');
      await socket.close();
    });
  });
}
