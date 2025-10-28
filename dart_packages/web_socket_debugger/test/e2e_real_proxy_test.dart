import 'dart:async';
import 'dart:io';

import 'package:test/test.dart';
import 'package:web_socket_debugger/web_socket_debugger.dart';

String _findProxyBinaryPath() {
  final override = Platform.environment['NETWORK_DEBUGGER_BIN'];
  if (override != null && override.isNotEmpty && File(override).existsSync()) {
    return override;
  }
  final exe = Platform.isWindows
      ? 'network-debugger_windows_amd64.exe'
      : 'network-debugger';
  var dir = Directory.current;
  for (int i = 0; i < 6; i++) {
    final candidate = File(
        '${dir.path}${Platform.pathSeparator}bin${Platform.pathSeparator}$exe');
    if (candidate.existsSync()) return candidate.path;
    dir = dir.parent;
  }
  return '';
}

Future<Process> _startProxy({required int port}) async {
  final binPath = _findProxyBinaryPath();
  expect(binPath.isNotEmpty, isTrue, reason: 'proxy binary not found');
  final bin = File(binPath);
  final env = Map<String, String>.from(Platform.environment);
  env['ADDR'] = '127.0.0.1:$port';
  env['DEV_MODE'] = '1';
  final p = await Process.start(bin.path, [],
      environment: env, workingDirectory: bin.parent.path);
  final ready = Completer<void>();
  p.stdout.transform(SystemEncoding().decoder).listen((s) {
    if (s.contains('starting network-debugger')) ready.complete();
  });
  p.stderr.listen((_) {});
  await ready.future.timeout(const Duration(seconds: 3));
  return p;
}

void main() {
  group('e2e with real proxy (web_socket)', () {
    HttpServer? upstream;
    Process? proxy;
    late int proxyPort;

    setUp(() async {
      upstream = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
      upstream!.listen((req) async {
        if (req.uri.path == '/echo' &&
            WebSocketTransformer.isUpgradeRequest(req)) {
          final ws = await WebSocketTransformer.upgrade(req);
          ws.listen((d) => ws.add(d));
        } else {
          req.response.statusCode = 404;
          await req.response.close();
        }
      });

      proxyPort = 39000 + (DateTime.now().microsecondsSinceEpoch % 1000);
      proxy = await _startProxy(port: proxyPort);
    });

    tearDown(() async {
      await upstream?.close(force: true);
      proxy?.kill(ProcessSignal.sigterm);
    });

    test('echo through real /wsproxy', () async {
      final upstreamWs = 'ws://127.0.0.1:${upstream!.port}/echo';
      final cfg = WebSocketDebugger.attach(
        baseUrl: upstreamWs,
        proxyBaseUrl: 'http://127.0.0.1:$proxyPort',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );

      final socket = await WebSocketDebugger.connect(config: cfg);
      final c = Completer<String>();
      socket.events.listen((e) {
        // Т.к. типизированные события импортировать не обязательно, берём toString
        final s = e.toString();
        if (s.contains('TextDataReceived')) {
          c.complete('ok');
        }
      });
      socket.sendText('hi');
      await c.future.timeout(const Duration(seconds: 4));
      await socket.close();
    });
  });
}
