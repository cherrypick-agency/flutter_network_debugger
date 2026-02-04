import 'dart:async';
import 'package:dio/dio.dart';
import '../../domain/entities/update_info.dart';
import '../../domain/repositories/updates_repository.dart';
import '../../presentation/widgets/download_progress_dialog.dart';

/// Service for managing application updates
///
/// This is a high-level service that coordinates update operations.
/// Additional business logic, analytics, etc. can be added here in the future.
class UpdatesService {
  final UpdatesRepository _repository;

  UpdatesService({required UpdatesRepository repository})
    : _repository = repository;

  /// Checks for available updates
  Future<UpdateInfo?> checkForUpdates({bool forceCheck = false}) async {
    return await _repository.checkForUpdates(forceCheck: forceCheck);
  }

  /// Gets list of all releases with pagination
  Future<List<UpdateInfo>> getAllReleases({
    int page = 1,
    int perPage = 10,
    bool includePrerelease = true,
  }) async {
    return await _repository.getAllReleases(
      page: page,
      perPage: perPage,
      includePrerelease: includePrerelease,
    );
  }

  /// Gets specific release by tag
  Future<UpdateInfo?> getReleaseByTag(String tag) async {
    return await _repository.getReleaseByTag(tag);
  }

  /// Downloads update with progress
  Future<String?> downloadUpdate(
    UpdateInfo updateInfo, {
    required StreamController<DownloadProgress> progressController,
    CancelToken? cancelToken,
    int maxRetries = 3,
  }) async {
    return await _repository.downloadUpdate(
      updateInfo,
      progressController: progressController,
      cancelToken: cancelToken,
      maxRetries: maxRetries,
    );
  }

  /// Marks version as skipped
  Future<void> skipVersion(String version) async {
    await _repository.skipVersion(version);
  }

  /// Clears skipped version
  Future<void> clearSkippedVersion() async {
    await _repository.clearSkippedVersion();
  }

  /// Gets last check time
  Future<DateTime?> getLastCheckTime() async {
    return await _repository.getLastCheckTime();
  }

  /// Cleans up old downloaded installers
  Future<void> cleanupOldDownloads({
    Duration maxAge = const Duration(days: 7),
  }) async {
    await _repository.cleanupOldDownloads(maxAge: maxAge);
  }

  /// Releases resources
  void dispose() {
    _repository.dispose();
  }
}
