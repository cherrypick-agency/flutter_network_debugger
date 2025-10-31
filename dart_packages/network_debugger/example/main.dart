// ignore_for_file: avoid_print

import 'package:network_debugger/network_debugger.dart';

Future<void> main() async {
  print('Network Debugger Launcher Example\n');

  // Display platform information
  print('Platform: ${NetworkDebugger.getPlatformInfo()}');

  // Display cache information
  final cacheDir = await NetworkDebugger.getCacheDirectory();
  print('Cache directory: $cacheDir');

  final cachedVersions = await NetworkDebugger.listCachedVersions();
  print(
    'Cached versions: ${cachedVersions.isEmpty ? "none" : cachedVersions.join(", ")}',
  );

  if (cachedVersions.isNotEmpty) {
    final cacheSize = await NetworkDebugger.getCacheSizeFormatted();
    print('Cache size: $cacheSize');
  }

  print('\nLaunching network debugger...');

  // Launch the debugger
  final debugger = await NetworkDebugger.launch(
    port: 9092,
    autoOpenBrowser: true,
    onProgress: (received, total) {
      final percent = ((received / total) * 100).toStringAsFixed(1);
      print('Download progress: $percent%');
    },
  );

  print('Network debugger is running at: ${debugger.url}');
  print('Process ID: ${debugger.pid}');

  // Subscribe to output streams
  debugger.stdout.listen((line) {
    print('[DEBUGGER] $line');
  });

  debugger.stderr.listen((line) {
    print('[DEBUGGER ERROR] $line');
  });

  // Keep running for 30 seconds
  print('\nDebugger will run for 30 seconds...');
  await Future.delayed(const Duration(seconds: 30));

  // Stop the debugger
  print('\nStopping debugger...');
  await debugger.stop();

  print('Debugger stopped.');
  print('Is running: ${debugger.isRunning}');
}
