import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:http_debugger/http_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (http_debugger) → edge cases', () {
    HttpServer? upstream;
    GoNetworkDebuggerProcess? proxy;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path == '/echo') {
          final body = await utf8.decodeStream(req);
          req.response.headers.contentType =
              ContentType('application', 'json', charset: 'utf-8');
          req.response.write(jsonEncode({
            'ok': true,
            'method': req.method,
            'path': req.uri.path,
            'rawQuery': req.uri.query,
            'queryAll': req.uri.queryParametersAll,
            'xTest': req.headers.value('x-test'),
            'bodyLen': body.length,
          }));
          await req.response.close();
          return;
        }
        if (req.uri.path == '/redirect') {
          req.response.statusCode = 302;
          req.response.headers.set('location', '/echo?from=redirect');
          await req.response.close();
          return;
        }
        if (req.uri.path == '/direct') {
          req.response.headers.contentType =
              ContentType('application', 'json', charset: 'utf-8');
          req.response.write(jsonEncode({'ok': true, 'path': '/direct'}));
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

    test('query encoding: repeated keys, unicode, spaces', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final url = Uri.parse(
        '$upstreamBase/echo?a=1&a=2&space=${Uri.encodeQueryComponent('a b')}&unicode=${Uri.encodeQueryComponent('привет')}',
      );

      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy/',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(url);
          req.headers.set('x-test', '1');
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );

      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['ok'], isTrue);
      expect(decoded['xTest'], '1');
      final queryAll = (decoded['queryAll'] as Map).cast<String, dynamic>();
      expect((queryAll['a'] as List).map((e) => e.toString()).toList(),
          ['1', '2']);
      expect((queryAll['space'] as List).single.toString(), 'a b');
      expect((queryAll['unicode'] as List).single.toString(), 'привет');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) => (s['target'] as String).contains('/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('proxy merges query from _target and incoming query params', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final proxyBase = proxy!.proxyHttpBase.toString();

      final proxyUrl = Uri.parse(proxyBase).replace(
        path: '/httpproxy',
        queryParameters: {
          'extra': '2',
          '_target': '$upstreamBase/echo?base=1&override=a',
        },
      );

      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxyBase,
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(proxyUrl);
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );

      final decoded = jsonDecode(body) as Map<String, dynamic>;
      final queryAll = (decoded['queryAll'] as Map).cast<String, dynamic>();
      expect((queryAll['base'] as List).single.toString(), '1');
      expect((queryAll['extra'] as List).single.toString(), '2');
      expect((queryAll['override'] as List).single.toString(), 'a');
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('redirect follow works through proxy and records sessions', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/redirect'));
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );
      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['path'], '/echo');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/redirect')),
          isTrue);
      expect(sessions.any((s) => (s['target'] as String).contains('/echo')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('POST body passes through and is recorded', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final big = List<int>.filled(128 * 1024, 'b'.codeUnitAt(0));
      final payload = jsonEncode({'blob': base64Encode(big)});

      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.postUrl(Uri.parse('$upstreamBase/echo'));
          req.headers.set('content-type', 'application/json');
          req.add(utf8.encode(payload));
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );

      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['method'], 'POST');
      expect((decoded['bodyLen'] as num).toInt(), greaterThan(1024));

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/echo')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('skipPaths bypasses proxy (request succeeds, but no session)',
        () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
          skipPaths: [RegExp(r'^/direct$')],
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/direct'));
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );
      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['path'], '/direct');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/direct')),
          isFalse);
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('concurrency: multiple requests are all recorded', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final futures = List.generate(
        10,
        (i) => HttpDebugger.runZonedWithReverseProxy<Future<void>>(
          HttpReverseProxyConfig(
            upstreamBaseUrl: upstreamBase,
            proxyBaseUrl: proxy!.proxyHttpBase.toString(),
            proxyHttpPath: '/httpproxy',
          ),
          () async {
            final c = HttpClient();
            final req = await c.getUrl(Uri.parse('$upstreamBase/echo?n=$i'));
            final resp = await req.close();
            await resp.drain<void>();
            c.close(force: true);
          },
        ),
      );
      await Future.wait(futures);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 500);
      final echo = sessions
          .where((s) => (s['target'] as String).contains('/echo'))
          .toList();
      expect(echo.length, greaterThanOrEqualTo(10));
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('already-proxy URL is not re-written (no double proxying)', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final proxyBase = proxy!.proxyHttpBase.toString();

      final proxyUrl = Uri.parse(proxyBase).replace(
        path: '/httpproxy',
        queryParameters: {'_target': '$upstreamBase/echo?via=direct'},
      );

      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxyBase,
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(proxyUrl);
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );

      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['path'], '/echo');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) => (s['target'] as String).contains('/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
