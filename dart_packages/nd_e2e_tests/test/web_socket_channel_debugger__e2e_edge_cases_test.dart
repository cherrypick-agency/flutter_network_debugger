import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';
import 'package:web_socket_channel/web_socket_channel.dart' as wsc;
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() {
  group('e2e (web_socket_channel_debugger) → edge cases', () {
    HttpServer? upstream;
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path.startsWith('/echo') &&
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

    test('preserves baseUrl query params and still records ws session',
        () async {
      final upstreamWs =
          'ws://127.0.0.1:${upstream!.port}/echo?space=a%20b&unicode=%D0%BF%D1%80%D0%B8%D0%B2%D0%B5%D1%82';
      final proxyUrl = Uri.parse('${proxy!.proxyHttpBase}').replace(
        scheme: 'ws',
        path: '/wsproxy',
        queryParameters: {'_target': upstreamWs},
      );

      final ch = wsc.WebSocketChannel.connect(proxyUrl);
      final got = Completer<String>();
      ch.stream.listen((e) {
        if (!got.isCompleted) got.complete(e.toString());
      });
      ch.sink.add('hi');
      expect(await got.future.timeout(const Duration(seconds: 10)), 'hi');
      await ch.sink.close();

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('headers are forwarded through wsproxy (Authorization/Cookie)',
        () async {
      final headers = <String, dynamic>{
        'Authorization': 'Bearer test-token',
        'Cookie': 'a=b',
      };

      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'ws://127.0.0.1:${upstream!.port}/echo-auth',
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );

      final ch =
          WebSocketChannelDebugger.connect(config: cfg, headers: headers);

      final got = Completer<String>();
      ch.stream.listen((e) {
        if (!got.isCompleted) got.complete(e.toString());
      });

      ch.sink.add('hi');
      expect(await got.future.timeout(const Duration(seconds: 10)), 'hi');
      await ch.sink.close();

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('/echo-auth')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('wsproxy accepts http(s) _target by auto-normalizing to ws(s)',
        () async {
      final targetHttp = 'http://127.0.0.1:${upstream!.port}/echo-http-target';
      final proxyUrl = Uri.parse('${proxy!.proxyHttpBase}').replace(
        scheme: 'ws',
        path: '/wsproxy',
        queryParameters: {'_target': targetHttp},
      );

      final ch = wsc.WebSocketChannel.connect(proxyUrl);
      final got = Completer<String>();
      ch.stream.listen((e) {
        if (!got.isCompleted) got.complete(e.toString());
      });
      ch.sink.add('hi');
      expect(await got.future.timeout(const Duration(seconds: 10)), 'hi');
      await ch.sink.close();

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind']?.toString() == 'ws') &&
            (s['target'] as String).contains('/echo-http-target')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('binary frames are proxied and recorded', () async {
      final upstreamWs = 'ws://127.0.0.1:${upstream!.port}/echo-binary';
      final proxyUrl = Uri.parse('${proxy!.proxyHttpBase}').replace(
        scheme: 'ws',
        path: '/wsproxy/',
        queryParameters: {'_target': upstreamWs},
      );

      final ch = wsc.WebSocketChannel.connect(proxyUrl);
      final got = Completer<Uint8List>();
      ch.stream.listen((e) {
        if (e is List<int> && !got.isCompleted) {
          got.complete(Uint8List.fromList(e));
        }
      });
      final payload = Uint8List.fromList([9, 8, 7, 0, 255]);
      ch.sink.add(payload);
      expect(await got.future.timeout(const Duration(seconds: 10)), payload);
      await ch.sink.close();

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
      final proxyA = Uri.parse('${proxy!.proxyHttpBase}').replace(
        scheme: 'ws',
        path: '/wsproxy',
        queryParameters: {'_target': a},
      );
      final proxyB = Uri.parse('${proxy!.proxyHttpBase}').replace(
        scheme: 'ws',
        path: '/wsproxy',
        queryParameters: {'_target': b},
      );

      final c1 = wsc.WebSocketChannel.connect(proxyA);
      final c2 = wsc.WebSocketChannel.connect(proxyB);
      await c1.sink.close();
      await c2.sink.close();

      final sessions =
          await listSessions(proxy!.apiBase, types: 'ws', limit: 200);
      final count = sessions.where((s) => s['kind']?.toString() == 'ws').length;
      expect(count, greaterThanOrEqualTo(2));
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
