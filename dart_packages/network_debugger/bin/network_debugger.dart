#!/usr/bin/env dart
// ignore_for_file: avoid_print

import 'dart:io';

import 'package:args/args.dart';
import 'package:network_debugger/network_debugger.dart';

const String version = '0.1.3';

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
    ..addFlag(
      'no-browser',
      negatable: false,
      help: 'Do not automatically open browser',
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

  // Configure logging based on flags
  final verbose = args['verbose'] as bool;
  final quiet = args['quiet'] as bool;
  final logLevelArg = args['log-level'] as String;

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

  print('🚀 Network Debugger Launcher v$version\n');
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
    print('\n\n⏹️  Shutting down...');
    if (debugger != null) {
      await debugger.stop();
    }
    exit(0);
  });

  ProcessSignal.sigterm.watch().listen((signal) async {
    print('\n\n⏹️  Shutting down...');
    if (debugger != null) {
      await debugger.stop();
    }
    exit(0);
  });

  try {
    // Launch the debugger
    print('📥 Checking cache and downloading if needed...');
    debugger = await NetworkDebugger.launch(
      port: port,
      version: binaryVersion,
      autoOpenBrowser: !noBrowser,
      onProgress: (received, total) {
        final percent = ((received / total) * 100).toStringAsFixed(1);
        stdout.write('\r📦 Download progress: $percent%');
        if (received == total) {
          stdout.write('\n');
        }
      },
    );

    print('✅ Network debugger is running!');
    print('');
    print('🌐 Web UI: ${debugger.url}');
    print('🔌 Process ID: ${debugger.pid}');
    print('');
    print('💡 Press Ctrl+C to stop');
    print('');

    // Subscribe to output
    debugger.stdout.listen((line) {
      // Only show important logs
      if (line.contains('"level":"error"') || line.contains('"level":"warn"')) {
        print('[LOG] $line');
      }
    });

    debugger.stderr.listen((line) {
      print('[ERROR] $line');
    });

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
    '  network_debugger --version                # Show version information',
  );
}
