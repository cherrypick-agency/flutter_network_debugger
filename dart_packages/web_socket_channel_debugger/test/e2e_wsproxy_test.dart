import 'dart:async';
import 'dart:io';

import 'package:test/test.dart';
import 'package:web_socket_channel/web_socket_channel.dart' as wsc;

void main() {
  group('e2e via /wsproxy with _target', () {
    HttpServer? http;
    HttpServer? proxy;

    setUp(() async {
      // echo ws upstream
      http = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      http!.listen((req) async {
        if (req.uri.path == '/echo') {
          if (WebSocketTransformer.isUpgradeRequest(req)) {
            final socket = await WebSocketTransformer.upgrade(req);
            socket.listen((data) {
              socket.add(data);
            });
          } else {
            req.response.statusCode = 400;
            await req.response.close();
          }
        } else {
          req.response.statusCode = 404;
          await req.response.close();
        }
      });

      // lightweight proxy handler for test: /wsproxy just redirects to _target
      proxy = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      proxy!.listen((req) async {
        if (req.uri.path == '/wsproxy') {
          final target = req.uri.queryParameters['_target'];
          if (target == null) {
            req.response.statusCode = 400;
            await req.response.close();
            return;
          }
          // proxy: upgrade and pipe to target
          if (!WebSocketTransformer.isUpgradeRequest(req)) {
            req.response.statusCode = 400;
            await req.response.close();
            return;
          }
          final client = await WebSocket.connect(target);
          final incoming = await WebSocketTransformer.upgrade(req);
          // bidirectional piping
          incoming.listen((d) => client.add(d), onDone: () => client.close());
          client.listen((d) => incoming.add(d), onDone: () => incoming.close());
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

    test('echo through proxy', () async {
      final upstreamWs = 'ws://localhost:${http!.port}/echo';
      final proxyUrl = Uri.parse('ws://localhost:${proxy!.port}/wsproxy')
          .replace(queryParameters: {'_target': upstreamWs});

      final channel = wsc.WebSocketChannel.connect(proxyUrl);
      final completer = Completer<String>();
      channel.stream.listen((event) {
        completer.complete(event.toString());
      });
      channel.sink.add('ping');
      final echoed = await completer.future.timeout(const Duration(seconds: 2));
      expect(echoed, 'ping');
      await channel.sink.close();
    });
  });
}
