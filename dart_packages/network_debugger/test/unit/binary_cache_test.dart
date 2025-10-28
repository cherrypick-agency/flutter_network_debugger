import 'dart:io';
import 'package:test/test.dart';
import 'package:network_debugger/src/binary_cache.dart';

void main() {
  group('BinaryCache', () {
    test('getCacheDir returns valid path', () async {
      final cacheDir = await BinaryCache.getCacheDir();
      expect(cacheDir, isNotEmpty);
      expect(cacheDir, contains('network_debugger'));
    });

    test('getVersionCacheDir includes version', () async {
      final versionDir = await BinaryCache.getVersionCacheDir('v1.0.0');
      expect(versionDir, contains('v1.0.0'));
      expect(versionDir, contains('network_debugger'));
    });

    test('formatBytes formats correctly', () {
      expect(BinaryCache.formatBytes(500), equals('500 B'));
      expect(BinaryCache.formatBytes(1024), equals('1.00 KB'));
      expect(BinaryCache.formatBytes(1024 * 1024), equals('1.00 MB'));
      expect(BinaryCache.formatBytes(1024 * 1024 * 1024), equals('1.00 GB'));
    });

    test('ensureVersionCacheDir creates directory', () async {
      final testVersion =
          'test-version-${DateTime.now().millisecondsSinceEpoch}';
      final versionDir = await BinaryCache.ensureVersionCacheDir(testVersion);

      expect(await Directory(versionDir).exists(), isTrue);

      // Cleanup
      await BinaryCache.clearVersion(testVersion);
    });

    test('listVersions returns list of versions', () async {
      // Create a test version
      final testVersion =
          'test-version-${DateTime.now().millisecondsSinceEpoch}';
      await BinaryCache.ensureVersionCacheDir(testVersion);

      final versions = await BinaryCache.listVersions();
      expect(versions, contains(testVersion));

      // Cleanup
      await BinaryCache.clearVersion(testVersion);
    });

    test('clearVersion removes version directory', () async {
      final testVersion =
          'test-version-${DateTime.now().millisecondsSinceEpoch}';
      final versionDir = await BinaryCache.ensureVersionCacheDir(testVersion);

      expect(await Directory(versionDir).exists(), isTrue);

      await BinaryCache.clearVersion(testVersion);

      expect(await Directory(versionDir).exists(), isFalse);
    });

    test('getBinaryPath returns null when binary not in cache', () async {
      final path = await BinaryCache.getBinaryPath(
        'nonexistent-version',
        'nonexistent-binary',
      );
      expect(path, isNull);
    });

    test('validateBinary returns false for nonexistent file', () async {
      final isValid = await BinaryCache.validateBinary('/nonexistent/path');
      expect(isValid, isFalse);
    });
  });
}
