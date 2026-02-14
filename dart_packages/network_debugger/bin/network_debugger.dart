#!/usr/bin/env dart
// ignore_for_file: avoid_print

import 'dart:io';

import 'package:args/args.dart';
import 'package:network_debugger/network_debugger.dart';

const String version = '0.2.3';

void main(List<String> arguments) async {
  final parser = ArgParser()
    ..addOption(
      'port',
      abbr: 'p',
      defaultsTo: '9092',
      help: 'Port to run the debugger on',
    )
    ..addOption(
      'binary-version',
      help: 'Specific binary version to use (e.g., v1.0.0)',
    )
    ..addOption(
      'log-level',
      abbr: 'l',
      defaultsTo: 'info',
      allowed: ['debug', 'info', 'warning', 'error', 'none'],
      help: 'Log level: debug, info, warning, error, none',
    )
    ..addOption(
      'github-token',
      help:
          'GitHub token for API requests (optional). Also supports GITHUB_TOKEN env var.',
    )
    ..addFlag(
      'no-browser',
      negatable: false,
      help: 'Do not automatically open browser',
    )
    ..addFlag(
      'clear-cache',
      negatable: false,
      help: 'Clear cached binaries and exit',
    )
    ..addOption(
      'clear-cache-version',
      help: 'Clear cache for a specific version and exit (e.g., v1.0.0)',
    )
    ..addFlag(
      'verbose',
      negatable: false,
      help: 'Enable verbose logging (same as --log-level=debug)',
    )
    ..addFlag(
      'quiet',
      abbr: 'q',
      negatable: false,
      help: 'Quiet mode - only show errors (same as --log-level=error)',
    )
    ..addFlag(
      'debugger-logs',
      defaultsTo: true,
      help:
          'Show logs from the debugger process (use --no-debugger-logs to disable)',
    )
    ..addFlag(
      'help',
      abbr: 'h',
      negatable: false,
      help: 'Show this help message',
    )
    ..addFlag(
      'version',
      abbr: 'v',
      negatable: false,
      help: 'Show version information',
    );

  late final ArgResults args;
  try {
    args = parser.parse(arguments);
  } catch (e) {
    print('Error parsing arguments: $e\n');
    _printUsage(parser);
    exit(1);
  }

  if (args['help'] as bool) {
    _printUsage(parser);
    exit(0);
  }

  if (args['version'] as bool) {
    print('network_debugger version $version');
    exit(0);
  }

  final clearCache = args['clear-cache'] as bool;
  final clearCacheVersion = args['clear-cache-version'] as String?;
  if (clearCacheVersion != null && clearCacheVersion.trim().isNotEmpty) {
    await NetworkDebugger.clearCache(version: clearCacheVersion.trim());
    print('Cache cleared for version: ${clearCacheVersion.trim()}');
    exit(0);
  }
  if (clearCache) {
    await NetworkDebugger.clearCache();
    print('Cache cleared');
    exit(0);
  }

  // Configure logging based on flags
  final verbose = args['verbose'] as bool;
  final quiet = args['quiet'] as bool;
  final logLevelArg = args['log-level'] as String;
  final debuggerLogsFlag = args['debugger-logs'] as bool;

  if (verbose) {
    Logger.enableVerboseMode();
  } else if (quiet) {
    Logger.enableQuietMode();
  } else {
    Logger.setLevelFromString(logLevelArg);
  }

  final port = int.tryParse(args['port'] as String);
  if (port == null || port < 1 || port > 65535) {
    print('Error: Invalid port number "${args['port']}"\n');
    _printUsage(parser);
    exit(1);
  }

  final binaryVersion = args['binary-version'] as String?;
  final noBrowser = args['no-browser'] as bool;
  final githubToken = (args['github-token'] as String?)?.trim();

  final showDebuggerLogs = args.wasParsed('debugger-logs')
      ? debuggerLogsFlag
      : !(quiet || logLevelArg == 'none');

  print('Network Debugger Launcher v$version\n');
  print('Platform: ${NetworkDebugger.getPlatformInfo()}');
  print('Port: $port');
  print('Browser: ${noBrowser ? 'disabled' : 'auto-open'}');
  if (binaryVersion != null) {
    print('Binary version: $binaryVersion');
  }
  print('');

  DebuggerInstance? debugger;

  // Setup signal handlers for graceful shutdown
  ProcessSignal.sigint.watch().listen((signal) async {
    print('\n\nShutting down...');
    if (debugger != null) {
      await debugger.stop();
    }
    exit(0);
  });

  ProcessSignal.sigterm.watch().listen((signal) async {
    print('\n\nShutting down...');
    if (debugger != null) {
      await debugger.stop();
    }
    exit(0);
  });

  try {
    // Launch the debugger
    print('Checking cache and downloading if needed...');
    debugger = await NetworkDebugger.launch(
      port: port,
      version: binaryVersion,
      autoOpenBrowser: !noBrowser,
      showDebuggerProcessLogs: showDebuggerLogs,
      githubToken:
          (githubToken != null && githubToken.isNotEmpty) ? githubToken : null,
      onProgress: (received, total) {
        final percent = ((received / total) * 100).toStringAsFixed(1);
        stdout.write('\rDownload progress: $percent%');
        if (received == total) {
          stdout.write('\n');
        }
      },
    );

    print('Network debugger is running!');
    print('');
    print('Web UI: ${debugger.url}');
    print('Process ID: ${debugger.pid}');
    print('');
    print('Press Ctrl+C to stop');
    print('');

    // Keep running until signal
    while (debugger.isRunning) {
      await Future.delayed(const Duration(seconds: 1));
    }
  } catch (e) {
    print('\n❌ Error: $e');
    exit(1);
  }
}

void _printUsage(ArgParser parser) {
  print('Network Debugger - Launch network debugging proxy with web UI\n');
  print('Usage: network_debugger [options]\n');
  print('Options:');
  print(parser.usage);
  print('\nExamples:');
  print(
    '  network_debugger                          # Default: port 9092 (UI), proxy on 9091, auto-open browser',
  );
  print('  network_debugger --port 8080              # Custom port');
  print(
    '  network_debugger --no-browser             # Without opening browser',
  );
  print('  network_debugger --verbose                # Enable verbose logging');
  print(
    '  network_debugger --quiet                  # Quiet mode (errors only)',
  );
  print('  network_debugger --log-level=debug        # Set custom log level');
  print(
    '  network_debugger --binary-version v1.0.0  # Specific binary version',
  );
  print(
    '  network_debugger --github-token <token>  # GitHub API token (rate limit mitigation)',
  );
  print(
    '  network_debugger --version                # Show version information',
  );
}
