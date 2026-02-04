import 'dart:io';
import 'dart:convert';
import 'package:path/path.dart' as path;
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:logging/logging.dart';
import 'package:flutter/foundation.dart' show kIsWeb;

/// DataSource for working with local storage (SharedPreferences + files)
class UpdatesLocalDataSource {
  static const String _skipVersionKey = 'skip_update_version';
  static const String _lastCheckKey = 'last_update_check';
  static const String _cachePrefix = 'release_cache_';

  final _log = Logger('UpdatesLocalDataSource');

  /// Gets skipped version
  Future<String?> getSkippedVersion() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_skipVersionKey);
  }

  /// Saves skipped version
  Future<void> setSkippedVersion(String version) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_skipVersionKey, version);
    _log.info('Skipped version: $version');
  }

  /// Removes skipped version
  Future<void> clearSkippedVersion() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_skipVersionKey);
  }

  /// Gets last check time
  Future<DateTime?> getLastCheckTime() async {
    final prefs = await SharedPreferences.getInstance();
    final timeStr = prefs.getString(_lastCheckKey);
    if (timeStr == null) return null;

    try {
      return DateTime.parse(timeStr);
    } catch (e) {
      return null;
    }
  }

  /// Saves last check time
  Future<void> setLastCheckTime(DateTime time) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_lastCheckKey, time.toIso8601String());
  }

  /// Caches release in SharedPreferences
  Future<void> cacheRelease(
    String version,
    Map<String, dynamic> releaseData,
  ) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cacheData = {
        'data': releaseData,
        'cached_at': DateTime.now().toIso8601String(),
      };
      await prefs.setString(_cachePrefix + version, jsonEncode(cacheData));
      _log.fine('Cached release: $version');
    } catch (e) {
      _log.warning('Failed to cache release: $e');
    }
  }

  /// Gets cached release
  Future<Map<String, dynamic>?> getCachedRelease(
    String version, {
    Duration maxAge = const Duration(days: 1),
  }) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedStr = prefs.getString(_cachePrefix + version);
      if (cachedStr == null) return null;

      final cacheData = jsonDecode(cachedStr) as Map<String, dynamic>;
      final cachedAt = DateTime.parse(cacheData['cached_at'] as String);

      // Check if cache is expired
      if (DateTime.now().difference(cachedAt) > maxAge) {
        _log.fine('Cache expired for version: $version');
        return null;
      }

      _log.fine('Using cached release: $version');
      return cacheData['data'] as Map<String, dynamic>;
    } catch (e) {
      _log.warning('Failed to read cached release: $e');
      return null;
    }
  }

  /// Clears expired release cache
  Future<void> clearExpiredCache({
    Duration maxAge = const Duration(days: 7),
  }) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final keys = prefs.getKeys().where((k) => k.startsWith(_cachePrefix));

      for (final key in keys) {
        final cachedStr = prefs.getString(key);
        if (cachedStr == null) continue;

        try {
          final cacheData = jsonDecode(cachedStr) as Map<String, dynamic>;
          final cachedAt = DateTime.parse(cacheData['cached_at'] as String);

          if (DateTime.now().difference(cachedAt) > maxAge) {
            await prefs.remove(key);
            _log.fine('Removed expired cache: $key');
          }
        } catch (e) {
          // Invalid format - remove
          await prefs.remove(key);
        }
      }
    } catch (e) {
      _log.warning('Failed to clear expired cache: $e');
    }
  }

  /// Cleans up old downloaded installers from temp directory
  Future<void> cleanupOldDownloads({
    Duration maxAge = const Duration(days: 7),
  }) async {
    if (kIsWeb) {
      return; // Web doesn't have file system
    }

    try {
      final tempDir = await getTemporaryDirectory();
      final dir = Directory(tempDir.path);

      if (!dir.existsSync()) {
        return;
      }

      final now = DateTime.now();
      int deletedCount = 0;
      int totalSize = 0;

      // List of installer extensions to clean up
      const installerExtensions = [
        '.dmg',
        '.msi',
        '.deb',
        '.appimage',
        '.tar.gz',
        '.zip',
      ];

      // Scan all files in temp directory
      await for (final entity in dir.list()) {
        if (entity is File) {
          try {
            final fileName = path.basename(entity.path).toLowerCase();

            // Check if file is an installer
            final isInstaller = installerExtensions.any(
              (ext) => fileName.endsWith(ext),
            );

            if (!isInstaller) {
              continue;
            }

            // Check file age
            final stat = await entity.stat();
            final age = now.difference(stat.modified);

            if (age > maxAge) {
              final size = stat.size;
              await entity.delete();
              deletedCount++;
              totalSize += size;
              _log.fine(
                'Deleted old installer: ${path.basename(entity.path)} (age: ${age.inDays} days)',
              );
            }
          } catch (e) {
            _log.warning('Failed to process file ${entity.path}: $e');
          }
        }
      }

      if (deletedCount > 0) {
        _log.info(
          'Cleanup completed: deleted $deletedCount old installer(s), freed ${(totalSize / (1024 * 1024)).toStringAsFixed(1)} MB',
        );
      }
    } catch (e, stack) {
      _log.warning('Error during cleanup', e, stack);
    }
  }
}
