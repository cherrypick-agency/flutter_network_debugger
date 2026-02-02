import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (dio_debugger) → edge cases', () {
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
      await upstream?.close(force: true);
      await proxy?.stop();
    });

    Dio _dio({
      List<Pattern>? skipPaths,
      String proxyHttpPath = '/httpproxy',
    }) {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = Dio(BaseOptions(baseUrl: upstreamBase));
      DioDebugger.attach(
        dio,
        upstreamBaseUrl: upstreamBase,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: proxyHttpPath,
        enabled: true,
        skipPaths: skipPaths,
      );
      return dio;
    }

    test('query encoding: list values, spaces, unicode, reserved chars',
        () async {
      final dio = _dio(proxyHttpPath: '/httpproxy/');

      final resp = await dio.get(
        '/echo',
        queryParameters: const {
          'a': ['1', '2'],
          'space': 'a b',
          'unicode': 'привет',
          'reserved': 'a&b=c',
        },
        options: Options(headers: const {'x-test': '1'}),
      );
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();

      expect(decoded['ok'], isTrue);
      expect(decoded['path'], '/echo');
      expect(decoded['xTest'], '1');

      final queryAll = (decoded['queryAll'] as Map).cast<String, dynamic>();
      expect((queryAll['a'] as List).map((e) => e.toString()).toList(),
          ['1', '2']);
      expect((queryAll['space'] as List).single.toString(), 'a b');
      expect((queryAll['unicode'] as List).single.toString(), 'привет');
      expect((queryAll['reserved'] as List).single.toString(), 'a&b=c');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind'] == 'http') &&
            (s['target'] as String)
                .contains('127.0.0.1:${upstream!.port}/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));

    test(
        'upstreamBaseUrl query is preserved and merged (request overrides keys)',
        () async {
      final upstreamBase =
          'http://127.0.0.1:${upstream!.port}?base=1&override=a';
      final dio = Dio(BaseOptions(baseUrl: upstreamBase));
      DioDebugger.attach(
        dio,
        upstreamBaseUrl: upstreamBase,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/httpproxy',
        enabled: true,
      );

      final resp = await dio.get('/echo', queryParameters: const {
        'override': 'b',
        'extra': '2',
      });
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();
      final queryAll = (decoded['queryAll'] as Map).cast<String, dynamic>();
      expect((queryAll['base'] as List).single.toString(), '1');
      expect((queryAll['extra'] as List).single.toString(), '2');
      expect((queryAll['override'] as List).single.toString(), 'b');
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('absolute URL in path is proxied too', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio();

      final resp = await dio.get(
        '$upstreamBase/echo',
        queryParameters: const {'q': '1'},
      );
      expect(resp.statusCode, 200);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind'] == 'http') &&
            (s['target'] as String).contains('$upstreamBase/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('query key with leading "?" is normalized', () async {
      final dio = _dio();

      final resp = await dio.get(
        '/echo',
        queryParameters: const {'?bad': '1'},
      );
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();
      final queryAll = (decoded['queryAll'] as Map).cast<String, dynamic>();
      expect((queryAll['bad'] as List).single.toString(), '1');
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('redirect follow works through proxy and records sessions', () async {
      final dio = _dio();
      final resp = await dio.get('/redirect');
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();
      expect(decoded['path'], '/echo');
      expect(((decoded['queryAll'] as Map)['from'] as List).single, 'redirect');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) =>
            (s['kind'] == 'http') &&
            (s['target'] as String).contains('/redirect')),
        isTrue,
      );
      expect(
        sessions.any((s) =>
            (s['kind'] == 'http') && (s['target'] as String).contains('/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('already-proxy request is not rewritten (no double proxying)',
        () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';

      final dio = Dio(BaseOptions(baseUrl: proxy!.proxyHttpBase.toString()));
      DioDebugger.attach(
        dio,
        upstreamBaseUrl: upstreamBase,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/httpproxy',
        enabled: true,
      );

      final resp = await dio.get(
        '/httpproxy',
        queryParameters: {'_target': '$upstreamBase/echo?via=direct'},
      );
      expect(resp.statusCode, 200);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) => (s['target'] as String).contains('/echo')),
        isTrue,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('POST body passes through and is recorded', () async {
      final dio = _dio();

      final big = List<int>.filled(256 * 1024, 'a'.codeUnitAt(0));
      final payload = {'blob': base64Encode(big)};

      final resp = await dio.post(
        '/echo',
        data: jsonEncode(payload),
        options: Options(
          headers: const {'content-type': 'application/json'},
        ),
      );
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();
      expect(decoded['method'], 'POST');
      expect((decoded['bodyLen'] as num).toInt(), greaterThan(1024));

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.where((s) => (s['kind'] == 'http')).length,
        greaterThanOrEqualTo(1),
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('skipPaths bypasses proxy (request succeeds, but no session)',
        () async {
      final dio = _dio(skipPaths: [RegExp(r'^/direct$')]);
      final resp = await dio.get('/direct');
      expect(resp.statusCode, 200);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(
        sessions.any((s) => (s['target'] as String).contains('/direct')),
        isFalse,
      );
    }, timeout: const Timeout(Duration(seconds: 60)));

    test('concurrency: multiple requests are all recorded', () async {
      final dio = _dio();

      final futures = List.generate(
        12,
        (i) => dio.get('/echo', queryParameters: {'n': '$i'}),
      );
      await Future.wait(futures);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 500);
      final echo = sessions
          .where((s) =>
              (s['kind'] == 'http') &&
              (s['target'] as String).contains('/echo'))
          .toList();
      expect(echo.length, greaterThanOrEqualTo(12));
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
