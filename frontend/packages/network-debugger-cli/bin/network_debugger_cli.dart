import 'dart:async';
import 'dart:io';

import 'package:args/args.dart';

import '../lib/src/platform.dart';
import '../lib/src/runner.dart';
import '../lib/src/downloader.dart';
import '../lib/src/server.dart';

Future<int> main(List<String> args) async {
  // Команды:
  // network-debugger [flags]
  // network-debugger serve-artifacts --dir <path> [--port 8099]
  final parser =
      ArgParser()
        ..addOption('version', abbr: 'v')
        ..addFlag(
          'force',
          abbr: 'f',
          help: 'Force re-download even if cached',
          defaultsTo: false,
        )
        ..addOption(
          'base-url',
          abbr: 'b',
          help: 'Base URL or file:// for artifacts',
        )
        ..addOption(
          'local-dir',
          abbr: 'L',
          help: 'Local directory with binary or archives',
        )
        ..addOption(
          'artifact-name',
          abbr: 'n',
          help: 'Override artifact filename',
        )
        ..addFlag(
          'no-remote',
          help: 'Do not fallback to remote sources',
          defaultsTo: false,
        )
        ..addFlag(
          'allow-remote',
          help: 'Allow remote fallback even when local provided',
          defaultsTo: false,
        )
        ..addFlag(
          'no-cache',
          help: 'Ignore cache and always re-fetch/extract',
          defaultsTo: false,
        )
        ..addFlag('help', abbr: 'h', negatable: false);

  final serve =
      ArgParser()
        ..addOption(
          'dir',
          abbr: 'd',
          help: 'Directory to serve',
          defaultsTo: '.',
        )
        ..addOption('port', abbr: 'p', help: 'Port', defaultsTo: '8099');
  parser.addCommand('serve-artifacts', serve);

  final res = parser.parse(args);
  if (res.command?.name == 'serve-artifacts') {
    final cmd = res.command!;
    final dir = (cmd['dir'] as String).trim();
    final port = int.tryParse((cmd['port'] as String).trim()) ?? 8099;
    await serveArtifacts(dir: dir, port: port);
    return 0;
  }
  if (res['help'] == true) {
    stdout.writeln('network-debugger - launcher');
    stdout.writeln(parser.usage);
    return 0;
  }

  final requestedVersion = (res['version'] as String?)?.trim();
  final force = res['force'] as bool;
  final baseUrl =
      (res['base-url'] as String?)?.trim() ??
      Platform.environment['NETWORK_DEBUGGER_BASE_URL'];
  final localDir =
      (res['local-dir'] as String?)?.trim() ??
      Platform.environment['NETWORK_DEBUGGER_LOCAL_DIR'];
  final artifactName =
      (res['artifact-name'] as String?)?.trim() ??
      Platform.environment['NETWORK_DEBUGGER_ARTIFACT_NAME'];
  final noRemote =
      (res['no-remote'] as bool) ||
      (Platform.environment['NETWORK_DEBUGGER_NO_REMOTE'] == 'true');
  final allowRemote =
      (res['allow-remote'] as bool) ||
      (Platform.environment['NETWORK_DEBUGGER_ALLOW_REMOTE'] == 'true');
  final noCache =
      (res['no-cache'] as bool) ||
      (Platform.environment['NETWORK_DEBUGGER_NO_CACHE'] == 'true');

  try {
    final plat = detectPlatform();
    final cacheDir = await getCacheDir();
    final repo =
        Platform.environment['NETWORK_DEBUGGER_GITHUB_REPO'] ??
        'belief/network-debugger';
    final exec = await ensureBinary(
      platform: plat,
      cacheDir: cacheDir,
      repo: repo,
      version: requestedVersion,
      force: force,
      baseUrl: baseUrl,
      localDir: localDir,
      artifactName: artifactName,
      noRemote: noRemote,
      allowRemote: allowRemote,
      noCache: noCache,
    );
    final code = await runBinary(execPath: exec);
    return code;
  } catch (e, st) {
    stderr.writeln('[network-debugger] Ошибка: $e');
    stderr.writeln(st);
    return 1;
  }
}
