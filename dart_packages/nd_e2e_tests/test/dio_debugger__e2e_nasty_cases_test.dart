import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (dio_debugger) → nasty cases', () {
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
            'queryAll': req.uri.queryParametersAll,
            'bodyLen': body.length,
          }));
          await req.response.close();
          return;
        }
        if (req.uri.path == '/stream') {
          req.response.headers.contentType =
              ContentType('text', 'plain', charset: 'utf-8');
          req.response.write('hello ');
          await req.response.flush();
          await Future<void>.delayed(const Duration(milliseconds: 80));
          req.response.write('world');
          await req.response.close();
          return;
        }
        if (req.uri.path == '/gzip') {
          final raw = utf8.encode(jsonEncode({'ok': true, 'enc': 'gzip'}));
          final gz = gzip.encode(raw);
          req.response.headers.set('content-encoding', 'gzip');
          req.response.headers.contentType =
              ContentType('application', 'json', charset: 'utf-8');
          req.response.add(gz);
          await req.response.close();
          return;
        }
        if (req.uri.path == '/deflate') {
          final raw = utf8.encode(jsonEncode({'ok': true, 'enc': 'deflate'}));
          final def = ZLibCodec().encode(raw);
          req.response.headers.set('content-encoding', 'deflate');
          req.response.headers.contentType =
              ContentType('application', 'json', charset: 'utf-8');
          req.response.add(def);
          await req.response.close();
          return;
        }
        if (req.uri.path == '/r1') {
          req.response.statusCode = 302;
          req.response.headers.set('location', '/r2');
          await req.response.close();
          return;
        }
        if (req.uri.path == '/r2') {
          req.response.statusCode = 302;
          req.response.headers.set('location', '/echo?from=r2');
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

    Dio _dio(String upstreamBase) {
      final dio = Dio(BaseOptions(baseUrl: upstreamBase));
      DioDebugger.attach(
        dio,
        upstreamBaseUrl: upstreamBase,
        proxyBaseUrl: proxy!.proxyHttpBase.toString(),
        proxyHttpPath: '/httpproxy',
        enabled: true,
      );
      return dio;
    }

    test('chunked/streaming response passes through', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio(upstreamBase);

      final resp = await dio.get<String>(
        '/stream',
        options: Options(responseType: ResponseType.plain),
      );
      expect(resp.statusCode, 200);
      expect(resp.data, 'hello world');

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/stream')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('gzip response passes through and is readable', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio(upstreamBase);

      final resp = await dio.get('/gzip');
      expect(resp.statusCode, 200);
      final decoded = (resp.data is String)
          ? jsonDecode(resp.data as String) as Map<String, dynamic>
          : (resp.data as Map).cast<String, dynamic>();
      expect(decoded['enc'], 'gzip');
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('deflate response passes through and is readable', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio(upstreamBase);

      final resp = await dio.get<List<int>>(
        '/deflate',
        options: Options(responseType: ResponseType.bytes),
      );
      expect(resp.statusCode, 200);
      final raw = resp.data ?? const <int>[];
      final decodedBytes = ZLibCodec().decode(raw);
      final decoded =
          jsonDecode(utf8.decode(decodedBytes)) as Map<String, dynamic>;
      expect(decoded['enc'], 'deflate');
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('PUT and PATCH are proxied and recorded', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio(upstreamBase);

      final put = await dio.put('/echo', data: 'x');
      expect(put.statusCode, 200);
      final patch = await dio.patch('/echo', data: 'y');
      expect(patch.statusCode, 200);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 500);
      expect(
        sessions.where((s) => (s['target'] as String).contains('/echo')).length,
        greaterThanOrEqualTo(2),
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('double redirects are followed and all hops are recorded', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio(upstreamBase);

      final resp = await dio.get('/r1');
      expect(resp.statusCode, 200);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 500);
      expect(
          sessions.any((s) => (s['target'] as String).contains('/r1')), isTrue);
      expect(
          sessions.any((s) => (s['target'] as String).contains('/r2')), isTrue);
      expect(sessions.any((s) => (s['target'] as String).contains('/echo')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('very long query works', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final dio = _dio(upstreamBase);

      final long = List.filled(6000, 'x').join();
      final resp = await dio.get('/echo', queryParameters: {'big': long});
      expect(resp.statusCode, 200);

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/echo')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
