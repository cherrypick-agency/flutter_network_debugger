// ignore_for_file: avoid_print

import 'dart:async';
import 'dart:io';
import 'package:test/test.dart';
import 'package:network_debugger/network_debugger.dart';
import 'package:http/http.dart' as http;

void main() {
  group('NetworkDebugger Integration Tests', () {
    test('launch downloads and starts debugger', () async {
      // This test requires internet connection and may take some time
      DebuggerInstance? instance;

      try {
        var downloadProgress = 0;

        instance = await NetworkDebugger.launch(
          port: 19091, // Use non-standard port to avoid conflicts
          autoOpenBrowser: false,
          onProgress: (received, total) {
            final percent = ((received / total) * 100).toInt();
            if (percent > downloadProgress) {
              downloadProgress = percent;
              print('Download progress: $percent%');
            }
          },
        ).timeout(
          const Duration(minutes: 5),
          onTimeout: () {
            throw TimeoutException(
              'Launch timed out after 5 minutes',
            );
          },
        );

        expect(instance.isRunning, isTrue);
        expect(instance.port, equals(19091));
        expect(instance.url, equals('http://localhost:19091'));

        // Wait for the debugger to be fully ready
        final isReady = await instance.waitUntilReady(
          timeout: const Duration(seconds: 30),
        );
        expect(isReady, isTrue);

        // Try to make a request to the debugger
        final response = await http.get(
          Uri.parse(instance.url),
        ).timeout(
          const Duration(seconds: 10),
        );

        expect(response.statusCode, equals(200));
        print('Successfully connected to debugger at ${instance.url}');
      } finally {
        // Always stop the debugger
        if (instance != null) {
          await instance.stop();
          expect(instance.isRunning, isFalse);
        }
      }
    }, timeout: const Timeout(Duration(minutes: 6)),);

    test('cache operations work correctly', () async {
      // Clear cache before test
      await NetworkDebugger.clearCache();

      final versions = await NetworkDebugger.listCachedVersions();
      expect(versions, isEmpty);

      // Launch to populate cache
      DebuggerInstance? instance;
      try {
        instance = await NetworkDebugger.launch(
          port: 19092,
          autoOpenBrowser: false,
        ).timeout(
          const Duration(minutes: 5),
        );

        // Check that cache now has a version
        final versionsAfter = await NetworkDebugger.listCachedVersions();
        expect(versionsAfter, isNotEmpty);

        final cacheSize = await NetworkDebugger.getCacheSize();
        expect(cacheSize, greaterThan(0));

        final cacheSizeFormatted = await NetworkDebugger.getCacheSizeFormatted();
        expect(cacheSizeFormatted, isNotEmpty);
        print('Cache size: $cacheSizeFormatted');

        final cacheDir = await NetworkDebugger.getCacheDirectory();
        expect(cacheDir, isNotEmpty);
        expect(await Directory(cacheDir).exists(), isTrue);
      } finally {
        if (instance != null) {
          await instance.stop();
        }
      }
    }, timeout: const Timeout(Duration(minutes: 6)),);

    test('second launch uses cache', () async {
      DebuggerInstance? instance1;
      DebuggerInstance? instance2;

      try {
        // First launch (may download)
        instance1 = await NetworkDebugger.launch(
          port: 19093,
          autoOpenBrowser: false,
        ).timeout(
          const Duration(minutes: 5),
        );

        await instance1.stop();

        // Second launch should use cache (should be faster)
        final stopwatch = Stopwatch()..start();

        instance2 = await NetworkDebugger.launch(
          port: 19094,
          autoOpenBrowser: false,
        ).timeout(const Duration(minutes: 1));

        stopwatch.stop();

        expect(instance2.isRunning, isTrue);
        print('Second launch took: ${stopwatch.elapsedMilliseconds}ms');

        // Should be much faster (< 10 seconds) when using cache
        expect(stopwatch.elapsedMilliseconds, lessThan(10000));
      } finally {
        if (instance1 != null && instance1.isRunning) {
          await instance1.stop();
        }
        if (instance2 != null && instance2.isRunning) {
          await instance2.stop();
        }
      }
    }, timeout: const Timeout(Duration(minutes: 7)),);

    test('platform info is available', () {
      final platformInfo = NetworkDebugger.getPlatformInfo();
      expect(platformInfo, isNotEmpty);
      print('Platform: $platformInfo');
    });
  });
}
