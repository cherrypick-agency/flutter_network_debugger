/// Launcher for network-debugger binary with automatic download and caching.
library network_debugger;

export 'src/debugger_process.dart' show DebuggerInstance, ProcessException;
export 'src/binary_downloader.dart'
    show ProgressCallback, RetryCallback, ChecksumCallback;
export 'src/platform_detector.dart' show PlatformDetector;
export 'src/binary_cache.dart' show BinaryCache, CacheException;
export 'src/logger.dart' show Logger, LogLevel;

import 'dart:io';
import 'src/platform_detector.dart';
import 'src/github_release.dart';
import 'src/binary_cache.dart';
import 'src/binary_downloader.dart';
import 'src/debugger_process.dart';
import 'src/error_formatter.dart';
import 'src/logger.dart';

/// Main entry point for launching the network debugger.
class NetworkDebugger {
  static const String _defaultOwner = 'cherrypick-agency';
  static const String _defaultRepo = 'flutter_network_debugger';
  static const String _binaryBaseName = 'network-debugger-web';

  /// Launches the network debugger.
  ///
  /// - [version]: Specific version to use (e.g., 'v1.0.0'). If null, uses latest release.
  /// - [port]: Port to run the UI/REST on (default: 9092). Forward proxy runs on 9091.
  /// - [autoOpenBrowser]: Whether to automatically open browser (default: true).
  /// - [onProgress]: Optional callback for download progress.
  /// - [onRetry]: Optional callback for retry attempts.
  /// - [onChecksum]: Optional callback for checksum validation.
  /// - [skipChecksumValidation]: Skip SHA256 checksum validation (NOT RECOMMENDED, default: false).
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
  /// - [ChecksumValidationException] if checksum validation fails.
  /// - [ProcessException] if starting the process fails.
  static Future<DebuggerInstance> launch({
    String? version,
    int port = 9092,
    bool autoOpenBrowser = true,
    ProgressCallback? onProgress,
    RetryCallback? onRetry,
    ChecksumCallback? onChecksum,
    bool skipChecksumValidation = false,
    String owner = _defaultOwner,
    String repo = _defaultRepo,
    Map<String, String>? environment,
  }) async {
    final logger = Logger('NetworkDebugger');

    try {
      logger.info('Launching network debugger on port $port');
      logger.debug('Platform: ${PlatformDetector.getPlatformDescription()}');

      // Check platform support
      if (!PlatformDetector.isSupported()) {
        logger.error('Platform not supported: ${Platform.operatingSystem}');
        throw UnsupportedError(
          'Platform not supported: ${Platform.operatingSystem}',
        );
      }

      final binaryName = PlatformDetector.getBinaryName(
        binaryBaseName: _binaryBaseName,
      );

      // Determine version to use
      final githubClient = GitHubRelease(owner: owner, repo: repo);
      logger.debug('Fetching release information...');
      final release = version == null
          ? await githubClient.getLatestRelease()
          : await githubClient.getRelease(version);

      final versionTag = release.tagName;
      logger.info('Using version: $versionTag');

      // Check cache first
      logger.debug('Checking cache for version $versionTag...');
      var binaryPath = await BinaryCache.getBinaryPath(versionTag, binaryName);

      if (binaryPath != null) {
        logger.debug('Found cached binary: $binaryPath');
        // Validate cached binary
        if (await BinaryCache.validateBinary(binaryPath)) {
          logger.info('Using cached binary (validated)');
          // Binary found in cache and valid
          if (onProgress != null) {
            onProgress(1, 1); // Report 100% since we're using cache
          }
        } else {
          logger.warning('Cached binary invalid, will re-download');
          // Invalid binary, download fresh
          binaryPath = null;
        }
      }

      // Download if not in cache
      if (binaryPath == null) {
        logger.info('Binary not in cache, downloading...');
        final archiveName = PlatformDetector.getArchiveName(
          binaryBaseName: _binaryBaseName,
        );

        final downloadUrl = githubClient.findAssetUrl(release, archiveName);
        if (downloadUrl == null) {
          throw DownloadException(
            'Asset not found for platform: $archiveName',
          );
        }

        // Get all asset URLs for checksum validation
        final allAssetUrls = githubClient.getAllAssetUrls(release);

        final cacheDir = await BinaryCache.ensureVersionCacheDir(versionTag);
        final downloader = BinaryDownloader();

        binaryPath = await downloader.downloadAndExtract(
          downloadUrl,
          cacheDir,
          expectedBinaryName: binaryName,
          onProgress: onProgress,
          onRetry: onRetry,
          onChecksum: onChecksum,
          availableAssetUrls: allAssetUrls,
          skipChecksumValidation: skipChecksumValidation,
        );
      }

      // Launch the process
      logger.info('Starting debugger process...');
      logger.debug('Binary path: $binaryPath');

      final process = DebuggerProcess(
        binaryPath: binaryPath,
        port: port,
        autoOpenBrowser: autoOpenBrowser,
        environment: environment,
      );

      await process.start();

      final instance = DebuggerInstance(process);

      // Wait for the debugger to be ready
      logger.debug('Waiting for debugger to be ready...');
      final isReady = await instance.waitUntilReady();
      if (!isReady) {
        logger.error('Debugger failed to become ready');
        await instance.stop();
        throw ProcessException(
          'Debugger failed to become ready within timeout',
        );
      }

      logger.info(
        'Network debugger started successfully on http://localhost:$port',
      );
      return instance;
    } catch (e) {
      logger.error('Failed to launch debugger', e);
      // Format error with user-friendly message
      final formattedError = ErrorFormatter.format(e);
      throw Exception(formattedError);
    }
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
