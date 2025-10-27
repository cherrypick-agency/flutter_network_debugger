#!/usr/bin/env dart
// ignore_for_file: avoid_print

import 'dart:io';

import 'package:args/args.dart';
import 'package:network_debugger/network_debugger.dart';

const String version = '0.1.0';

void main(List<String> arguments) async {
  final parser = ArgParser()
    ..addOption(
      'port',
      abbr: 'p',
      defaultsTo: '9091',
      help: 'Port to run the debugger on',
    )
    ..addOption(
      'version',
      abbr: 'v',
      help: 'Specific binary version to use (e.g., v1.0.0)',
    )
    ..addFlag(
      'no-browser',
      negatable: false,
      help: 'Do not automatically open browser',
    )
    ..addFlag(
      'help',
      abbr: 'h',
      negatable: false,
      help: 'Show this help message',
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

  final port = int.tryParse(args['port'] as String);
  if (port == null || port < 1 || port > 65535) {
    print('Error: Invalid port number "${args['port']}"\n');
    _printUsage(parser);
    exit(1);
  }

  final binaryVersion = args['version'] as String?;
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
      if (line.contains('"level":"error"') ||
          line.contains('"level":"warn"')) {
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
  print('  network_debugger                    # Default: port 9091, auto-open browser');
  print('  network_debugger --port 8080        # Custom port');
  print('  network_debugger --no-browser       # Without opening browser');
  print('  network_debugger --version v1.0.0   # Specific binary version');
}
