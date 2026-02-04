import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (dio_debugger) → real Go proxy', () {
    HttpServer? upstream;
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path == '/hello') {
          req.response.headers.contentType =
              ContentType('application', 'json', charset: 'utf-8');
          req.response.write(jsonEncode({
            'ok': true,
            'path': req.uri.path,
            'query': req.uri.queryParameters,
            'method': req.method,
          }));
          await req.response.close();
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

    test('reverse proxy: request reaches upstream and is recorded in Go',
        () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';

      final dio = Dio(BaseOptions(baseUrl: upstreamBase));
      DioDebugger.attach(
        dio,
        upstreamBaseUrl: upstreamBase,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/httpproxy',
        enabled: true,
      );

      final resp = await dio.get(
        '/hello',
        queryParameters: const {'q': '1'},
      );
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();
      expect(decoded['ok'], isTrue);
      expect(decoded['path'], '/hello');
      expect((decoded['query'] as Map)['q'], '1');

      final sessions = await listSessions(proxy!.apiBase, types: 'http');
      expect(
        sessions.any((s) =>
            (s['kind'] == 'http') &&
            (s['target'] as String).contains('$upstreamBase/hello')),
        isTrue,
        reason:
            'did not find http-session associated with request through /httpproxy',
      );
    }, timeout: const Timeout(Duration(seconds: 60)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
