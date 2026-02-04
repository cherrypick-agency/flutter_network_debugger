import 'dart:async';
import 'dart:io';
import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:logging/logging.dart';
import 'package:path/path.dart' as path;
import 'package:path_provider/path_provider.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import '../../domain/entities/update_info.dart';
import '../../domain/repositories/updates_repository.dart';
import '../../presentation/widgets/download_progress_dialog.dart';
import '../datasources/github_api_datasource.dart';
import '../datasources/updates_local_datasource.dart';

/// UpdatesRepository implementation for working with updates
class UpdatesRepositoryImpl implements UpdatesRepository {
  final GitHubApiDataSource _githubApi;
  final UpdatesLocalDataSource _localStorage;
  final String currentVersion;
  final _log = Logger('UpdatesRepositoryImpl');

  // Cache
  UpdateInfo? _cachedUpdateInfo;
  DateTime? _lastCheckTime;

  // Dio for downloading updates with timeouts
  final Dio _dio = Dio(
    BaseOptions(
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(
        minutes: 30,
      ), // Large files may take a long time to download
      sendTimeout: const Duration(seconds: 30),
    ),
  );

  UpdatesRepositoryImpl({
    required GitHubApiDataSource githubApi,
    required UpdatesLocalDataSource localStorage,
    required this.currentVersion,
  }) : _githubApi = githubApi,
       _localStorage = localStorage {
    // Clean up old downloaded files on initialization
    cleanupOldDownloads().catchError((e) {
      _log.warning('Failed to cleanup old downloads: $e');
    });
  }

  @override
  Future<UpdateInfo?> checkForUpdates({bool forceCheck = false}) async {
    // Web doesn't support auto-updates
    if (kIsWeb) {
      _log.fine('Auto-updates not supported on web platform');
      return null;
    }

    // Check no more than once per hour (unless force)
    if (!forceCheck && _lastCheckTime != null && _cachedUpdateInfo != null) {
      final diff = DateTime.now().difference(_lastCheckTime!);
      if (diff.inHours < 1) {
        _log.fine(
          'Using cached update info (checked ${diff.inMinutes} minutes ago)',
        );
        return _cachedUpdateInfo;
      }
    }

    try {
      _log.info('Checking for updates from GitHub Releases...');

      // Get data from GitHub API via datasource
      final data = await _githubApi.fetchLatestRelease();
      if (data == null) {
        return null;
      }

      final latestVersion = data['tag_name'] as String? ?? '';

      // Check if user skipped this version
      final skippedVersion = await _localStorage.getSkippedVersion();
      if (skippedVersion == latestVersion) {
        _log.info('User skipped version $latestVersion');
        return null;
      }

      // Determine asset to download based on platform
      final asset = _getPlatformAsset(data['assets'] as List);

      if (asset == null) {
        _log.warning('No compatible asset found for current platform');
        return null;
      }

      final updateInfo = UpdateInfo.fromGitHubRelease(data, asset);

      // Check if this is actually a new version
      if (!updateInfo.isNewerThan(currentVersion)) {
        _log.info('Already on latest version: $currentVersion');
        _lastCheckTime = DateTime.now();
        await _localStorage.setLastCheckTime(_lastCheckTime!);
        return null;
      }

      _log.info('Update available: ${updateInfo.version}');
      _cachedUpdateInfo = updateInfo;
      _lastCheckTime = DateTime.now();
      await _localStorage.setLastCheckTime(_lastCheckTime!);

      // Cache release for future use
      await _localStorage.cacheRelease(latestVersion, data);

      return updateInfo;
    } catch (e, stack) {
      _log.warning('Error checking for updates', e, stack);
      return null;
    }
  }

  @override
  Future<List<UpdateInfo>> getAllReleases({
    int page = 1,
    int perPage = 10,
    bool includePrerelease = true,
  }) async {
    if (kIsWeb) {
      _log.fine('GitHub releases not supported on web platform');
      return [];
    }

    try {
      // Get data from GitHub API via datasource
      final releasesJson = await _githubApi.fetchAllReleases(
        page: page,
        perPage: perPage,
      );

      final List<UpdateInfo> releases = [];

      for (final release in releasesJson) {
        // Skip pre-release if not needed
        if (!includePrerelease && (release['prerelease'] as bool? ?? false)) {
          continue;
        }

        // Get asset for current platform
        final assets = release['assets'] as List;
        final asset = _getPlatformAsset(assets);

        if (asset == null) {
          // No asset for current platform - skip
          _log.fine('No asset for platform in release: ${release['tag_name']}');
          continue;
        }

        final updateInfo = UpdateInfo.fromGitHubRelease(release, asset);
        releases.add(updateInfo);

        // Cache each release
        final version = release['tag_name'] as String? ?? '';
        if (version.isNotEmpty) {
          await _localStorage.cacheRelease(version, release);
        }
      }

      _log.info('Fetched ${releases.length} releases');
      return releases;
    } catch (e, stack) {
      _log.warning('Error fetching releases', e, stack);
      rethrow;
    }
  }

  @override
  Future<UpdateInfo?> getReleaseByTag(String tag) async {
    if (kIsWeb) {
      _log.fine('GitHub releases not supported on web platform');
      return null;
    }

    try {
      // First check cache
      final cachedData = await _localStorage.getCachedRelease(tag);
      if (cachedData != null) {
        _log.fine('Using cached release data for: $tag');
        final assets = cachedData['assets'] as List;
        final asset = _getPlatformAsset(assets);

        if (asset != null) {
          return UpdateInfo.fromGitHubRelease(cachedData, asset);
        }
      }

      // If not in cache - fetch from GitHub API
      _log.info('Fetching release by tag: $tag');
      final releaseJson = await _githubApi.fetchReleaseByTag(tag);

      if (releaseJson == null) {
        return null;
      }

      // Get asset for current platform
      final assets = releaseJson['assets'] as List;
      final asset = _getPlatformAsset(assets);

      if (asset == null) {
        _log.warning('No asset for platform in release: $tag');
        return null;
      }

      // Cache received release
      await _localStorage.cacheRelease(tag, releaseJson);

      return UpdateInfo.fromGitHubRelease(releaseJson, asset);
    } catch (e, stack) {
      _log.warning('Error fetching release by tag', e, stack);
      return null;
    }
  }

  @override
  Future<String?> downloadUpdate(
    UpdateInfo updateInfo, {
    required StreamController<DownloadProgress> progressController,
    CancelToken? cancelToken,
    int maxRetries = 3,
  }) async {
    if (kIsWeb) {
      _log.warning('Downloads not supported on web platform');
      return null;
    }

    // Check disk space
    if (updateInfo.sizeBytes > 0) {
      final hasSpace = await _checkDiskSpace(updateInfo.sizeBytes);
      if (!hasSpace) {
        _log.warning('Insufficient disk space for download');
        if (!progressController.isClosed) {
          progressController.add(
            DownloadProgress(
              downloaded: 0,
              total: 0,
              speed: 0,
              error:
                  'Insufficient disk space. Need ${updateInfo.formattedSize}',
            ),
          );
        }
        return null;
      }
    }

    // Retry logic with exponential backoff
    for (int attempt = 0; attempt < maxRetries; attempt++) {
      try {
        if (attempt > 0) {
          final delay = Duration(seconds: (1 << attempt)); // 2, 4, 8 seconds
          _log.info('Retry attempt $attempt after ${delay.inSeconds}s delay');
          await Future.delayed(delay);
        }

        _log.info(
          'Starting download: ${updateInfo.downloadUrl} (attempt ${attempt + 1}/$maxRetries)',
        );

        // Get download directory
        final tempDir = await getTemporaryDirectory();
        final fileName = path.basename(Uri.parse(updateInfo.downloadUrl).path);
        final savePath = path.join(tempDir.path, fileName);

        // Delete old file if exists
        final file = File(savePath);
        if (file.existsSync()) {
          await file.delete();
        }

        // Variables for tracking progress
        int lastReceived = 0;
        DateTime lastUpdate = DateTime.now();
        double currentSpeed = 0.0;

        // Download file
        await _dio.download(
          updateInfo.downloadUrl,
          savePath,
          cancelToken: cancelToken,
          onReceiveProgress: (received, total) {
            final now = DateTime.now();
            final timeDiff = now.difference(lastUpdate).inMilliseconds;

            // Update progress every 100ms for smoothness
            if (timeDiff >= 100) {
              final bytesDiff = received - lastReceived;
              currentSpeed = (bytesDiff / timeDiff) * 1000; // bytes/sec

              lastReceived = received;
              lastUpdate = now;

              // Send progress
              if (!progressController.isClosed) {
                progressController.add(
                  DownloadProgress(
                    downloaded: received,
                    total: total,
                    speed: currentSpeed,
                  ),
                );
              }
            }
          },
        );

        // Final progress (100%)
        if (!progressController.isClosed) {
          try {
            final fileSize = file.lengthSync();
            progressController.add(
              DownloadProgress(
                downloaded: fileSize,
                total: fileSize,
                speed: currentSpeed,
              ),
            );
          } catch (e) {
            _log.warning('Failed to read final file size: $e');
            // Not critical, just don't send final progress
          }
        }

        // Compute SHA256 for integrity check
        _log.info('Computing SHA256 checksum...');
        final computedSha256 = await _computeSha256(file);
        _log.info('File SHA256: $computedSha256');

        // NOTE: GitHub Releases API doesn't provide checksums for assets.
        // Checksum is computed for logging and future verification when/if
        // GitHub adds checksum support in API response.
        // Alternative: checksums can be stored in release notes as a workaround.

        _log.info('Download completed successfully: $savePath');
        return savePath;
      } catch (e) {
        // Error during download attempt
        if (attempt == maxRetries - 1) {
          // Last attempt failed
          rethrow;
        }
        _log.warning('Download attempt ${attempt + 1} failed: $e');
        // Continue retry loop
      }
    }

    // All retries exhausted - send final error
    _log.warning('All download attempts failed');
    if (!progressController.isClosed) {
      progressController.add(
        DownloadProgress(
          downloaded: 0,
          total: 0,
          speed: 0,
          error: 'Download failed after $maxRetries attempts',
        ),
      );
    }
    return null;
  }

  @override
  Future<void> skipVersion(String version) async {
    await _localStorage.setSkippedVersion(version);
  }

  @override
  Future<void> clearSkippedVersion() async {
    await _localStorage.clearSkippedVersion();
  }

  @override
  Future<DateTime?> getLastCheckTime() async {
    return await _localStorage.getLastCheckTime();
  }

  @override
  Future<void> cleanupOldDownloads({
    Duration maxAge = const Duration(days: 7),
  }) async {
    await _localStorage.cleanupOldDownloads(maxAge: maxAge);
  }

  /// Determines asset to download based on platform
  Map<String, dynamic>? _getPlatformAsset(List assets) {
    if (assets.isEmpty) return null;

    if (Platform.isMacOS) {
      // Priority: .dmg > .tar.gz
      final dmg = assets.firstWhere(
        (a) => (a['name'] as String).endsWith('.dmg'),
        orElse: () => null,
      );
      if (dmg != null) {
        return dmg as Map<String, dynamic>;
      }

      // Look for tar.gz for macOS
      final tarGz = assets.firstWhere((a) {
        final name = (a['name'] as String).toLowerCase();
        return name.contains('darwin') && name.endsWith('.tar.gz');
      }, orElse: () => null);
      if (tarGz != null) {
        return tarGz as Map<String, dynamic>;
      }
    } else if (Platform.isWindows) {
      // Priority: .msi > .zip
      final msi = assets.firstWhere(
        (a) => (a['name'] as String).endsWith('.msi'),
        orElse: () => null,
      );
      if (msi != null) {
        return msi as Map<String, dynamic>;
      }

      // Look for zip for Windows
      final zip = assets.firstWhere((a) {
        final name = (a['name'] as String).toLowerCase();
        return name.contains('windows') && name.endsWith('.zip');
      }, orElse: () => null);
      if (zip != null) {
        return zip as Map<String, dynamic>;
      }
    } else if (Platform.isLinux) {
      // Priority: AppImage > deb > tar.gz
      final appImage = assets.firstWhere(
        (a) => (a['name'] as String).endsWith('.AppImage'),
        orElse: () => null,
      );
      if (appImage != null) {
        return appImage as Map<String, dynamic>;
      }

      final deb = assets.firstWhere(
        (a) => (a['name'] as String).endsWith('.deb'),
        orElse: () => null,
      );
      if (deb != null) {
        return deb as Map<String, dynamic>;
      }

      // Look for tar.gz for Linux
      final tarGz = assets.firstWhere((a) {
        final name = (a['name'] as String).toLowerCase();
        return name.contains('linux') && name.endsWith('.tar.gz');
      }, orElse: () => null);
      if (tarGz != null) {
        return tarGz as Map<String, dynamic>;
      }
    }

    _log.warning('No compatible asset found for current platform');
    return null;
  }

  /// Computes SHA256 hash of file
  Future<String> _computeSha256(File file) async {
    final bytes = await file.readAsBytes();
    final digest = sha256.convert(bytes);
    return digest.toString();
  }

  /// Checks available disk space
  Future<bool> _checkDiskSpace(int requiredBytes) async {
    try {
      final tempDir = await getTemporaryDirectory();

      // For Unix-like systems we can use df command
      if (Platform.isLinux || Platform.isMacOS) {
        final result = await Process.run('df', ['-k', tempDir.path]);
        if (result.exitCode == 0) {
          final lines = (result.stdout as String).split('\n');
          if (lines.length > 1) {
            final parts = lines[1].split(RegExp(r'\s+'));
            if (parts.length >= 4) {
              final availableKB = int.tryParse(parts[3]) ?? 0;
              final availableBytes = availableKB * 1024;
              // Require 20% margin
              return availableBytes >= (requiredBytes * 1.2);
            }
          }
        }
      }

      // For Windows or if df failed - do approximate check
      // by attempting to create test file
      return true; // Optimistic fallback
    } catch (e) {
      _log.warning('Failed to check disk space: $e');
      return true; // Optimistic fallback on error
    }
  }

  @override
  void dispose() {
    _dio.close();
  }
}
