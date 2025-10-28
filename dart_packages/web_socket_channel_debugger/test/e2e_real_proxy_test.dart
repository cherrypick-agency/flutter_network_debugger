import 'dart:async';
import 'dart:io';

import 'package:test/test.dart';
import 'package:web_socket_channel/web_socket_channel.dart' as wsc;

String _findProxyBinaryPath() {
  final override = Platform.environment['NETWORK_DEBUGGER_BIN'];
  if (override != null && override.isNotEmpty && File(override).existsSync()) {
    return override;
  }
  final exe = Platform.isWindows
      ? 'network-debugger_windows_amd64.exe'
      : 'network-debugger';
  // Try walking up from current directory up to 6 levels to find bin/<exe>
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
  // ждём "starting network-debugger" и биндинг порта
  final ready = Completer<void>();
  p.stdout.transform(SystemEncoding().decoder).listen((s) {
    if (s.contains('starting network-debugger')) {
      ready.complete();
    }
  });
  p.stderr.listen((_) {});
  await ready.future.timeout(const Duration(seconds: 3));
  return p;
}

void main() {
  group('e2e with real proxy (web_socket_channel)', () {
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

      proxyPort = 38000 + (DateTime.now().microsecondsSinceEpoch % 1000);
      proxy = await _startProxy(port: proxyPort);
    });

    tearDown(() async {
      await upstream?.close(force: true);
      proxy?.kill(ProcessSignal.sigterm);
    });

    test('echo through real /wsproxy', () async {
      final upstreamWs = 'ws://127.0.0.1:${upstream!.port}/echo';
      final proxyUrl = Uri.parse('ws://127.0.0.1:$proxyPort/wsproxy')
          .replace(queryParameters: {'_target': upstreamWs});
      final ch = wsc.WebSocketChannel.connect(proxyUrl);
      final c = Completer<String>();
      ch.stream.listen((e) => c.complete(e.toString()));
      ch.sink.add('hi');
      final echoed = await c.future.timeout(const Duration(seconds: 4));
      expect(echoed, 'hi');
      await ch.sink.close();
    });
  });
}
