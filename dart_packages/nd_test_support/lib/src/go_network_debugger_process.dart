import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'ports.dart';
import 'repo_root.dart';

class GoNetworkDebuggerProcess {
  GoNetworkDebuggerProcess._({
    required this.process,
    required this.apiPort,
    required this.proxyPort,
    required this.stdoutLines,
  });

  final Process process;
  final int apiPort;
  final int proxyPort;
  final List<String> stdoutLines;

  Uri get apiBase => Uri.parse('http://127.0.0.1:$apiPort');
  Uri get proxyHttpBase => Uri.parse('http://127.0.0.1:$proxyPort');

  static Future<String>? _binaryPathFuture;

  static bool hasGo() {
    try {
      final r = Process.runSync('go', const ['version']);
      return r.exitCode == 0;
    } catch (_) {
      return false;
    }
  }

  static Future<String> _ensureBuiltBinary() {
    final existing = _binaryPathFuture;
    if (existing != null) return existing;

    _binaryPathFuture = () async {
      final repoRoot = findRepoRoot();
      final outDir = await Directory.systemTemp.createTemp('nd-go-bin-');
      final exe =
          Platform.isWindows ? 'network-debugger.exe' : 'network-debugger';
      final outPath = '${outDir.path}${Platform.pathSeparator}$exe';

      final r = await Process.run(
        'go',
        [
          'build',
          '-o',
          outPath,
          './cmd/network-debugger',
        ],
        workingDirectory: repoRoot.path,
      ).timeout(const Duration(minutes: 2));

      if (r.exitCode != 0) {
        throw StateError('go build failed: ${r.stderr}');
      }
      return outPath;
    }();

    return _binaryPathFuture!;
  }

  static Future<GoNetworkDebuggerProcess> start({
    String logLevel = 'error',
  }) async {
    final repoRoot = findRepoRoot();
    final binPath = await _ensureBuiltBinary();
    final env = Map<String, String>.from(Platform.environment)
      ..['NO_BROWSER'] = '1'
      ..['DEV_MODE'] = '1'
      ..['LOG_LEVEL'] = logLevel;

    for (var attempt = 1; attempt <= 5; attempt++) {
      final apiPort = await pickFreePort();
      final proxyPort = await pickFreePort();
      final dataDir = await Directory.systemTemp.createTemp('nd-e2e-');
      final stdoutLines = <String>[];

      final p = await Process.start(
        binPath,
        [
          '--api-port',
          '$apiPort',
          '--proxy-port',
          '$proxyPort',
          '--data-dir',
          dataDir.path,
        ],
        workingDirectory: repoRoot.path,
        environment: env,
      );

      p.stdout
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen(stdoutLines.add);
      p.stderr.transform(utf8.decoder).listen((_) {});

      try {
        await _waitHealthy(apiPort);
        await _waitPortOpen(proxyPort);
        return GoNetworkDebuggerProcess._(
          process: p,
          apiPort: apiPort,
          proxyPort: proxyPort,
          stdoutLines: stdoutLines,
        );
      } catch (_) {
        p.kill(ProcessSignal.sigterm);
        try {
          await p.exitCode.timeout(const Duration(seconds: 2));
        } catch (_) {
          p.kill(ProcessSignal.sigkill);
        }
      }
    }

    throw TimeoutException('failed to start network-debugger after retries');
  }

  Future<void> stop() async {
    process.kill(ProcessSignal.sigterm);
    try {
      await process.exitCode.timeout(const Duration(seconds: 5));
    } catch (_) {
      process.kill(ProcessSignal.sigkill);
    }
  }

  static Future<void> _waitHealthy(int apiPort) async {
    final uri = Uri.parse('http://127.0.0.1:$apiPort/healthz');
    final deadline = DateTime.now().add(const Duration(seconds: 60));
    while (DateTime.now().isBefore(deadline)) {
      try {
        final client = HttpClient();
        final req = await client.getUrl(uri);
        final resp = await req.close();
        await resp.drain<void>();
        client.close(force: true);
        if (resp.statusCode == 200) return;
      } catch (_) {}
      await Future<void>.delayed(const Duration(milliseconds: 150));
    }
    throw TimeoutException('network-debugger did not become healthy: $uri');
  }

  static Future<void> _waitPortOpen(int port) async {
    final deadline = DateTime.now().add(const Duration(seconds: 30));
    while (DateTime.now().isBefore(deadline)) {
      try {
        final s = await Socket.connect(
          InternetAddress.loopbackIPv4,
          port,
          timeout: const Duration(milliseconds: 500),
        );
        await s.close();
        return;
      } catch (_) {}
      await Future<void>.delayed(const Duration(milliseconds: 120));
    }
    throw TimeoutException('proxy port did not open: 127.0.0.1:$port');
  }
}
