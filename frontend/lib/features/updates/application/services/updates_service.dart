import 'dart:async';
import 'package:dio/dio.dart';
import '../../domain/entities/update_info.dart';
import '../../domain/repositories/updates_repository.dart';
import '../../presentation/widgets/download_progress_dialog.dart';

/// Сервис для управления обновлениями приложения
///
/// Это высокоуровневый сервис, который координирует работу с обновлениями.
/// В будущем здесь можно добавить дополнительную бизнес-логику, аналитику и т.д.
class UpdatesService {
  final UpdatesRepository _repository;

  UpdatesService({required UpdatesRepository repository})
    : _repository = repository;

  /// Проверяет наличие обновлений
  Future<UpdateInfo?> checkForUpdates({bool forceCheck = false}) async {
    return await _repository.checkForUpdates(forceCheck: forceCheck);
  }

  /// Получает список всех релизов с пагинацией
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

  /// Получает конкретный релиз по тегу
  Future<UpdateInfo?> getReleaseByTag(String tag) async {
    return await _repository.getReleaseByTag(tag);
  }

  /// Загружает обновление с прогрессом
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

  /// Отмечает версию как пропущенную
  Future<void> skipVersion(String version) async {
    await _repository.skipVersion(version);
  }

  /// Сбрасывает пропущенную версию
  Future<void> clearSkippedVersion() async {
    await _repository.clearSkippedVersion();
  }

  /// Получает время последней проверки
  Future<DateTime?> getLastCheckTime() async {
    return await _repository.getLastCheckTime();
  }

  /// Очищает старые загруженные установщики
  Future<void> cleanupOldDownloads({
    Duration maxAge = const Duration(days: 7),
  }) async {
    await _repository.cleanupOldDownloads(maxAge: maxAge);
  }

  /// Освобождение ресурсов
  void dispose() {
    _repository.dispose();
  }
}
