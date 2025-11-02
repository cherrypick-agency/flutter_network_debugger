import 'dart:async';
import 'package:dio/dio.dart';
import '../entities/update_info.dart';
import '../../presentation/widgets/download_progress_dialog.dart';

/// Repository интерфейс для работы с обновлениями
abstract class UpdatesRepository {
  /// Проверяет наличие обновлений через GitHub Releases API
  Future<UpdateInfo?> checkForUpdates({bool forceCheck = false});

  /// Получает список всех релизов из GitHub с пагинацией
  Future<List<UpdateInfo>> getAllReleases({
    int page = 1,
    int perPage = 10,
    bool includePrerelease = true,
  });

  /// Получает конкретный релиз по тегу (версии)
  Future<UpdateInfo?> getReleaseByTag(String tag);

  /// Загружает обновление с прогрессом
  /// Возвращает путь к загруженному файлу или null при ошибке
  Future<String?> downloadUpdate(
    UpdateInfo updateInfo, {
    required StreamController<DownloadProgress> progressController,
    CancelToken? cancelToken,
    int maxRetries = 3,
  });

  /// Отмечает версию как пропущенную
  Future<void> skipVersion(String version);

  /// Сбрасывает пропущенную версию
  Future<void> clearSkippedVersion();

  /// Получает время последней проверки
  Future<DateTime?> getLastCheckTime();

  /// Очищает старые загруженные установщики
  Future<void> cleanupOldDownloads({Duration maxAge = const Duration(days: 7)});

  /// Освобождение ресурсов
  void dispose();
}
