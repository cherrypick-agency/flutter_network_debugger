import 'dart:async';
import 'package:dio/dio.dart';
import '../entities/update_info.dart';
import '../../presentation/widgets/download_progress_dialog.dart';

/// Repository interface for working with updates
abstract class UpdatesRepository {
  /// Checks for updates via GitHub Releases API
  Future<UpdateInfo?> checkForUpdates({bool forceCheck = false});

  /// Gets list of all releases from GitHub with pagination
  Future<List<UpdateInfo>> getAllReleases({
    int page = 1,
    int perPage = 10,
    bool includePrerelease = true,
  });

  /// Gets specific release by tag (version)
  Future<UpdateInfo?> getReleaseByTag(String tag);

  /// Downloads update with progress
  /// Returns path to downloaded file or null on error
  Future<String?> downloadUpdate(
    UpdateInfo updateInfo, {
    required StreamController<DownloadProgress> progressController,
    CancelToken? cancelToken,
    int maxRetries = 3,
  });

  /// Marks version as skipped
  Future<void> skipVersion(String version);

  /// Clears skipped version
  Future<void> clearSkippedVersion();

  /// Gets last check time
  Future<DateTime?> getLastCheckTime();

  /// Cleans up old downloaded installers
  Future<void> cleanupOldDownloads({Duration maxAge = const Duration(days: 7)});

  /// Releases resources
  void dispose();
}
