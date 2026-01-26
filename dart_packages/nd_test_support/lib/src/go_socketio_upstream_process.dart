import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'repo_root.dart';

class GoSocketIoUpstreamProcess {
  GoSocketIoUpstreamProcess._({
    required this.process,
    required this.port,
    required this.stdoutLines,
  });

  final Process process;
  final int port;
  final List<String> stdoutLines;

  Uri get httpBase => Uri.parse('http://127.0.0.1:$port');

  static Future<GoSocketIoUpstreamProcess> start({required int port}) async {
    final repoRoot = findRepoRoot();
    final env = Map<String, String>.from(Platform.environment);

    final stdoutLines = <String>[];
    final p = await Process.start(
      'go',
      [
        'run',
        './cmd/test-socketio-upstream',
        '--addr',
        '127.0.0.1:$port',
      ],
      workingDirectory: repoRoot.path,
      environment: env,
    );

    final ready = Completer<void>();
    p.stdout
        .transform(utf8.decoder)
        .transform(const LineSplitter())
        .listen((line) {
      stdoutLines.add(line);
      if (line.contains('listening')) {
        if (!ready.isCompleted) ready.complete();
      }
    });
    p.stderr.listen((_) {});

    await ready.future.timeout(const Duration(seconds: 15));

    return GoSocketIoUpstreamProcess._(
      process: p,
      port: port,
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
}
