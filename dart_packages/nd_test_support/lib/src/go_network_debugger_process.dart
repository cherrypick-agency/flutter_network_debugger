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

  static bool hasGo() {
    try {
      final r = Process.runSync('go', const ['version']);
      return r.exitCode == 0;
    } catch (_) {
      return false;
    }
  }

  static Future<GoNetworkDebuggerProcess> start({
    String logLevel = 'error',
  }) async {
    final repoRoot = findRepoRoot();
    final apiPort = await pickFreePort();
    final proxyPort = await pickFreePort();
    final dataDir = await Directory.systemTemp.createTemp('nd-e2e-');

    final env = Map<String, String>.from(Platform.environment);
    env['NO_BROWSER'] = '1';
    env['DEV_MODE'] = '1';
    env['LOG_LEVEL'] = logLevel;

    final stdoutLines = <String>[];

    final p = await Process.start(
      'go',
      [
        'run',
        './cmd/network-debugger',
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

    await _waitHealthy(apiPort);

    return GoNetworkDebuggerProcess._(
      process: p,
      apiPort: apiPort,
      proxyPort: proxyPort,
      stdoutLines: stdoutLines,
    );
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
    final deadline = DateTime.now().add(const Duration(seconds: 30));
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
}
