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

  static Future<String>? _binaryPathFuture;

  static Future<String> _ensureBuiltBinary() {
    final existing = _binaryPathFuture;
    if (existing != null) return existing;

    _binaryPathFuture = () async {
      final repoRoot = findRepoRoot();
      final outDir = await Directory.systemTemp.createTemp('nd-go-sio-');
      final exe = Platform.isWindows
          ? 'test-socketio-upstream.exe'
          : 'test-socketio-upstream';
      final outPath = '${outDir.path}${Platform.pathSeparator}$exe';

      final r = await Process.run(
        'go',
        [
          'build',
          '-o',
          outPath,
          './cmd/test-socketio-upstream',
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

  static Future<GoSocketIoUpstreamProcess> start({
    required int port,
    String path = '/socket.io/',
  }) async {
    final repoRoot = findRepoRoot();
    final env = Map<String, String>.from(Platform.environment);

    final stdoutLines = <String>[];
    final binPath = await _ensureBuiltBinary();
    final p = await Process.start(
      binPath,
      ['--addr', '127.0.0.1:$port', '--path', path],
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
