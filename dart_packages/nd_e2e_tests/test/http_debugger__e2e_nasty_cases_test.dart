import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:http_debugger/http_debugger.dart';
import 'package:nd_test_support/nd_test_support.dart';
import 'package:test/test.dart';

void main() {
  group('e2e (http_debugger) → nasty cases', () {
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
      HttpDebugger.disable();
      await upstream?.close(force: true);
      await proxy?.stop();
    });

    test('chunked/streaming response passes through', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final proxyBase = proxy!.proxyHttpBase.toString();

      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxyBase,
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/stream'));
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );
      expect(body, 'hello world');
      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/stream')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('gzip response passes through and is readable', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/gzip'));
          final resp = await req.close();
          final txt = await utf8.decodeStream(resp);
          c.close(force: true);
          return txt;
        },
      );
      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['enc'], 'gzip');
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('deflate response passes through and is readable', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final body = await HttpDebugger.runZonedWithReverseProxy<Future<String>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient()..autoUncompress = false;
          final req = await c.getUrl(Uri.parse('$upstreamBase/deflate'));
          final resp = await req.close();
          final raw =
              await resp.fold<List<int>>(<int>[], (p, e) => p..addAll(e));
          final decoded = ZLibCodec().decode(raw);
          final txt = utf8.decode(decoded);
          c.close(force: true);
          return txt;
        },
      );
      final decoded = jsonDecode(body) as Map<String, dynamic>;
      expect(decoded['enc'], 'deflate');
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('PUT and PATCH are proxied and recorded', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';
      final proxyBase = proxy!.proxyHttpBase.toString();

      await HttpDebugger.runZonedWithReverseProxy<Future<void>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxyBase,
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final put = await c.openUrl('PUT', Uri.parse('$upstreamBase/echo'));
          put.add(utf8.encode('x'));
          final putResp = await put.close();
          await putResp.drain<void>();
          final patch =
              await c.openUrl('PATCH', Uri.parse('$upstreamBase/echo'));
          patch.add(utf8.encode('y'));
          final patchResp = await patch.close();
          await patchResp.drain<void>();
          c.close(force: true);
        },
      );

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 500);
      expect(
        sessions.where((s) => (s['target'] as String).contains('/echo')).length,
        greaterThanOrEqualTo(2),
      );
    }, timeout: const Timeout(Duration(seconds: 90)));

    test('double redirects are followed and all hops are recorded', () async {
      final upstreamBase = 'http://127.0.0.1:${upstream!.port}';

      await HttpDebugger.runZonedWithReverseProxy<Future<void>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/r1'));
          final resp = await req.close();
          await resp.drain<void>();
          c.close(force: true);
        },
      );

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
      final long = List.filled(6000, 'x').join();

      await HttpDebugger.runZonedWithReverseProxy<Future<void>>(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstreamBase,
          proxyBaseUrl: proxy!.proxyHttpBase.toString(),
          proxyHttpPath: '/httpproxy',
        ),
        () async {
          final c = HttpClient();
          final req = await c.getUrl(Uri.parse('$upstreamBase/echo?big=$long'));
          final resp = await req.close();
          await resp.drain<void>();
          c.close(force: true);
        },
      );

      final sessions =
          await listSessions(proxy!.apiBase, types: 'http', limit: 200);
      expect(sessions.any((s) => (s['target'] as String).contains('/echo')),
          isTrue);
    }, timeout: const Timeout(Duration(seconds: 90)));
  }, skip: GoNetworkDebuggerProcess.hasGo() ? false : 'go not found');
}
