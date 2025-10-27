/// Launcher for network-debugger binary with automatic download and caching.
library network_debugger;

export 'src/debugger_process.dart' show DebuggerInstance, ProcessException;
export 'src/binary_downloader.dart' show ProgressCallback;
export 'src/platform_detector.dart' show PlatformDetector;
export 'src/binary_cache.dart' show BinaryCache, CacheException;

import 'dart:io';
import 'src/platform_detector.dart';
import 'src/github_release.dart';
import 'src/binary_cache.dart';
import 'src/binary_downloader.dart';
import 'src/debugger_process.dart';

/// Main entry point for launching the network debugger.
class NetworkDebugger {
  static const String _defaultOwner = 'cherrypick-agency';
  static const String _defaultRepo = 'flutter_network_debugger';
  static const String _binaryBaseName = 'network-debugger-web';

  /// Launches the network debugger.
  ///
  /// - [version]: Specific version to use (e.g., 'v1.0.0'). If null, uses latest release.
  /// - [port]: Port to run the debugger on (default: 9091).
  /// - [autoOpenBrowser]: Whether to automatically open browser (default: true).
  /// - [onProgress]: Optional callback for download progress.
  /// - [owner]: GitHub repository owner (default: 'cherrypick-agency').
  /// - [repo]: GitHub repository name (default: 'flutter_network_debugger').
  /// - [environment]: Additional environment variables to pass to the process.
  ///
  /// Returns a [DebuggerInstance] that can be used to manage the process.
  ///
  /// Throws:
  /// - [UnsupportedError] if the platform is not supported.
  /// - [GitHubReleaseException] if fetching release information fails.
  /// - [DownloadException] if downloading or extracting fails.
  /// - [ProcessException] if starting the process fails.
  static Future<DebuggerInstance> launch({
    String? version,
    int port = 9091,
    bool autoOpenBrowser = true,
    ProgressCallback? onProgress,
    String owner = _defaultOwner,
    String repo = _defaultRepo,
    Map<String, String>? environment,
  }) async {
    // Check platform support
    if (!PlatformDetector.isSupported()) {
      throw UnsupportedError(
        'Platform not supported: ${Platform.operatingSystem}',
      );
    }

    final binaryName = PlatformDetector.getBinaryName(
      binaryBaseName: _binaryBaseName,
    );

    // Determine version to use
    final githubClient = GitHubRelease(owner: owner, repo: repo);
    final release = version == null
        ? await githubClient.getLatestRelease()
        : await githubClient.getRelease(version);

    final versionTag = release.tagName;

    // Check cache first
    var binaryPath = await BinaryCache.getBinaryPath(versionTag, binaryName);

    if (binaryPath != null) {
      // Validate cached binary
      if (await BinaryCache.validateBinary(binaryPath)) {
        // Binary found in cache and valid
        if (onProgress != null) {
          onProgress(1, 1); // Report 100% since we're using cache
        }
      } else {
        // Invalid binary, download fresh
        binaryPath = null;
      }
    }

    // Download if not in cache
    if (binaryPath == null) {
      final archiveName = PlatformDetector.getArchiveName(
        binaryBaseName: _binaryBaseName,
      );

      final downloadUrl = githubClient.findAssetUrl(release, archiveName);
      if (downloadUrl == null) {
        throw DownloadException(
          'Asset not found for platform: $archiveName',
        );
      }

      final cacheDir = await BinaryCache.ensureVersionCacheDir(versionTag);
      final downloader = BinaryDownloader();

      binaryPath = await downloader.downloadAndExtract(
        downloadUrl,
        cacheDir,
        expectedBinaryName: binaryName,
        onProgress: onProgress,
      );
    }

    // Launch the process
    final process = DebuggerProcess(
      binaryPath: binaryPath,
      port: port,
      autoOpenBrowser: autoOpenBrowser,
      environment: environment,
    );

    await process.start();

    final instance = DebuggerInstance(process);

    // Wait for the debugger to be ready
    final isReady = await instance.waitUntilReady();
    if (!isReady) {
      await instance.stop();
      throw ProcessException(
        'Debugger failed to become ready within timeout',
      );
    }

    return instance;
  }

  /// Clears the cache.
  ///
  /// If [version] is specified, only that version is cleared.
  /// Otherwise, all cached versions are cleared.
  static Future<void> clearCache({String? version}) async {
    if (version != null) {
      await BinaryCache.clearVersion(version);
    } else {
      await BinaryCache.clearAll();
    }
  }

  /// Lists all cached versions.
  static Future<List<String>> listCachedVersions() async {
    return BinaryCache.listVersions();
  }

  /// Gets the total size of the cache in bytes.
  static Future<int> getCacheSize() async {
    return BinaryCache.getCacheSize();
  }

  /// Gets a human-readable string of the cache size.
  static Future<String> getCacheSizeFormatted() async {
    final size = await getCacheSize();
    return BinaryCache.formatBytes(size);
  }

  /// Gets the cache directory path.
  static Future<String> getCacheDirectory() async {
    return BinaryCache.getCacheDir();
  }

  /// Gets platform information.
  static String getPlatformInfo() {
    return PlatformDetector.getPlatformDescription();
  }
}
