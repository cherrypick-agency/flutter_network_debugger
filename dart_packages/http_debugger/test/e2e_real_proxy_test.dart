import 'dart:convert';
import 'dart:io';

import 'package:http_debugger/http_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (http_debugger) → real Go proxy', () {
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
      HttpDebugger.disable();
      await upstream?.close(force: true);
      await proxy?.stop();
    });

    test('reverse proxy: request reaches upstream and is recorded in Go',
        () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';

      // Enable reverse-proxy only for the duration of the request to avoid intercepting API calls.
      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/hello?q=1'));
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );

      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['ok'], isTrue);
      expect(decoded['path'], '/hello');
      expect((decoded['query'] as Map)['q'], '1');

      // Verify via API that Go actually recorded the session.
      final sessions = await listSessions(proxy!.apiBase, types: 'http');
      expect(
        sessions.any((s) =>
            (s['kind'] == 'http') &&
            (s['target'] as String).contains('$upstreamBase/hello')),
        isTrue,
        reason: 'http session linked to request via /httpproxy not found',
      );
    }, timeout: const Timeout(Duration(seconds: 60)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
