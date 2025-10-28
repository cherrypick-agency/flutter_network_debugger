import 'dart:io';
import 'package:path/path.dart' as p;

/// Manages cached binary files.
class BinaryCache {
  static const String _cacheSubdir = 'network_debugger';

  /// Gets the cache directory path.
  ///
  /// - macOS/Linux: ~/.cache/network_debugger/
  /// - Windows: %LOCALAPPDATA%\network_debugger\Cache\
  static Future<String> getCacheDir() async {
    if (Platform.isWindows) {
      // On Windows, use LOCALAPPDATA
      final appData = Platform.environment['LOCALAPPDATA'];
      if (appData == null) {
        throw CacheException('LOCALAPPDATA environment variable not found');
      }
      return p.join(appData, _cacheSubdir, 'Cache');
    } else {
      // On macOS/Linux, prefer XDG_CACHE_HOME or fallback to ~/.cache
      final xdgCache = Platform.environment['XDG_CACHE_HOME'];
      if (xdgCache != null) {
        return p.join(xdgCache, _cacheSubdir);
      }

      // Fallback to home directory
      final homeDir = Platform.environment['HOME'];
      if (homeDir == null) {
        throw CacheException('HOME environment variable not found');
      }
      return p.join(homeDir, '.cache', _cacheSubdir);
    }
  }

  /// Gets the version-specific cache directory.
  static Future<String> getVersionCacheDir(String version) async {
    final cacheDir = await getCacheDir();
    return p.join(cacheDir, version);
  }

  /// Checks if a binary exists in cache for the given version and name.
  static Future<String?> getBinaryPath(
      String version, String binaryName) async {
    final versionDir = await getVersionCacheDir(version);
    final binaryPath = p.join(versionDir, binaryName);

    final file = File(binaryPath);
    if (await file.exists()) {
      return binaryPath;
    }

    return null;
  }

  /// Ensures the cache directory exists for a specific version.
  static Future<String> ensureVersionCacheDir(String version) async {
    final versionDir = await getVersionCacheDir(version);
    final dir = Directory(versionDir);

    if (!await dir.exists()) {
      await dir.create(recursive: true);
    }

    return versionDir;
  }

  /// Clears the cache for a specific version.
  static Future<void> clearVersion(String version) async {
    final versionDir = await getVersionCacheDir(version);
    final dir = Directory(versionDir);

    if (await dir.exists()) {
      await dir.delete(recursive: true);
    }
  }

  /// Clears all cached versions.
  static Future<void> clearAll() async {
    final cacheDir = await getCacheDir();
    final dir = Directory(cacheDir);

    if (await dir.exists()) {
      await dir.delete(recursive: true);
    }
  }

  /// Lists all cached versions.
  static Future<List<String>> listVersions() async {
    final cacheDir = await getCacheDir();
    final dir = Directory(cacheDir);

    if (!await dir.exists()) {
      return [];
    }

    final entities = await dir.list().toList();
    final versions = <String>[];

    for (final entity in entities) {
      if (entity is Directory) {
        versions.add(p.basename(entity.path));
      }
    }

    return versions;
  }

  /// Gets the total size of the cache in bytes.
  static Future<int> getCacheSize() async {
    final cacheDir = await getCacheDir();
    final dir = Directory(cacheDir);

    if (!await dir.exists()) {
      return 0;
    }

    var totalSize = 0;

    await for (final entity in dir.list(recursive: true)) {
      if (entity is File) {
        final stat = await entity.stat();
        totalSize += stat.size;
      }
    }

    return totalSize;
  }

  /// Formats bytes to a human-readable string.
  static String formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(2)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(2)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(2)} GB';
  }

  /// Validates that a cached binary is executable.
  static Future<bool> validateBinary(String binaryPath) async {
    final file = File(binaryPath);

    if (!await file.exists()) {
      return false;
    }

    // Check if file is executable (on Unix-like systems)
    if (!Platform.isWindows) {
      final stat = await file.stat();
      // Check if file has execute permission (simplified check)
      final mode = stat.mode;
      final isExecutable =
          (mode & 0x49) != 0; // Check user, group, or other execute bit

      if (!isExecutable) {
        return false;
      }
    }

    return true;
  }
}

/// Exception thrown when cache operations fail.
class CacheException implements Exception {
  final String message;

  CacheException(this.message);

  @override
  String toString() => 'CacheException: $message';
}
