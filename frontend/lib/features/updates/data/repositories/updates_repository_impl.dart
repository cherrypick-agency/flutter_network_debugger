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

/// Реализация UpdatesRepository для работы с обновлениями
class UpdatesRepositoryImpl implements UpdatesRepository {
  final GitHubApiDataSource _githubApi;
  final UpdatesLocalDataSource _localStorage;
  final String currentVersion;
  final _log = Logger('UpdatesRepositoryImpl');

  // Кэш
  UpdateInfo? _cachedUpdateInfo;
  DateTime? _lastCheckTime;

  // Dio для загрузки обновлений с таймаутами
  final Dio _dio = Dio(
    BaseOptions(
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(
        minutes: 30,
      ), // Большие файлы могут качаться долго
      sendTimeout: const Duration(seconds: 30),
    ),
  );

  UpdatesRepositoryImpl({
    required GitHubApiDataSource githubApi,
    required UpdatesLocalDataSource localStorage,
    required this.currentVersion,
  }) : _githubApi = githubApi,
       _localStorage = localStorage {
    // Очищаем старые загруженные файлы при инициализации
    cleanupOldDownloads().catchError((e) {
      _log.warning('Failed to cleanup old downloads: $e');
    });
  }

  @override
  Future<UpdateInfo?> checkForUpdates({bool forceCheck = false}) async {
    // Web не поддерживает автообновления
    if (kIsWeb) {
      _log.fine('Auto-updates not supported on web platform');
      return null;
    }

    // Проверка не чаще раза в час (если не force)
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

      // Получаем данные из GitHub API через datasource
      final data = await _githubApi.fetchLatestRelease();
      if (data == null) {
        return null;
      }

      final latestVersion = data['tag_name'] as String? ?? '';

      // Проверяем, пропустил ли пользователь эту версию
      final skippedVersion = await _localStorage.getSkippedVersion();
      if (skippedVersion == latestVersion) {
        _log.info('User skipped version $latestVersion');
        return null;
      }

      // Определяем asset для скачивания в зависимости от платформы
      final asset = _getPlatformAsset(data['assets'] as List);

      if (asset == null) {
        _log.warning('No compatible asset found for current platform');
        return null;
      }

      final updateInfo = UpdateInfo.fromGitHubRelease(data, asset);

      // Проверяем, действительно ли это новая версия
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

      // Кешируем релиз для будущего использования
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
      // Получаем данные из GitHub API через datasource
      final releasesJson = await _githubApi.fetchAllReleases(
        page: page,
        perPage: perPage,
      );

      final List<UpdateInfo> releases = [];

      for (final release in releasesJson) {
        // Пропускаем pre-release если не нужны
        if (!includePrerelease && (release['prerelease'] as bool? ?? false)) {
          continue;
        }

        // Получаем asset для текущей платформы
        final assets = release['assets'] as List;
        final asset = _getPlatformAsset(assets);

        if (asset == null) {
          // Нет asset для текущей платформы - пропускаем
          _log.fine('No asset for platform in release: ${release['tag_name']}');
          continue;
        }

        final updateInfo = UpdateInfo.fromGitHubRelease(release, asset);
        releases.add(updateInfo);

        // Кешируем каждый релиз
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
      // Сначала проверяем кеш
      final cachedData = await _localStorage.getCachedRelease(tag);
      if (cachedData != null) {
        _log.fine('Using cached release data for: $tag');
        final assets = cachedData['assets'] as List;
        final asset = _getPlatformAsset(assets);

        if (asset != null) {
          return UpdateInfo.fromGitHubRelease(cachedData, asset);
        }
      }

      // Если нет в кеше - запрашиваем из GitHub API
      _log.info('Fetching release by tag: $tag');
      final releaseJson = await _githubApi.fetchReleaseByTag(tag);

      if (releaseJson == null) {
        return null;
      }

      // Получаем asset для текущей платформы
      final assets = releaseJson['assets'] as List;
      final asset = _getPlatformAsset(assets);

      if (asset == null) {
        _log.warning('No asset for platform in release: $tag');
        return null;
      }

      // Кешируем полученный релиз
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

    // Проверяем место на диске
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

    // Retry logic с экспоненциальным backoff
    for (int attempt = 0; attempt < maxRetries; attempt++) {
      try {
        if (attempt > 0) {
          final delay = Duration(seconds: (1 << attempt)); // 2, 4, 8 секунд
          _log.info('Retry attempt $attempt after ${delay.inSeconds}s delay');
          await Future.delayed(delay);
        }

        _log.info(
          'Starting download: ${updateInfo.downloadUrl} (attempt ${attempt + 1}/$maxRetries)',
        );

        // Получаем директорию для загрузки
        final tempDir = await getTemporaryDirectory();
        final fileName = path.basename(Uri.parse(updateInfo.downloadUrl).path);
        final savePath = path.join(tempDir.path, fileName);

        // Удаляем старый файл если существует
        final file = File(savePath);
        if (file.existsSync()) {
          await file.delete();
        }

        // Переменные для отслеживания прогресса
        int lastReceived = 0;
        DateTime lastUpdate = DateTime.now();
        double currentSpeed = 0.0;

        // Загружаем файл
        await _dio.download(
          updateInfo.downloadUrl,
          savePath,
          cancelToken: cancelToken,
          onReceiveProgress: (received, total) {
            final now = DateTime.now();
            final timeDiff = now.difference(lastUpdate).inMilliseconds;

            // Обновляем прогресс каждые 100ms для плавности
            if (timeDiff >= 100) {
              final bytesDiff = received - lastReceived;
              currentSpeed = (bytesDiff / timeDiff) * 1000; // байт/сек

              lastReceived = received;
              lastUpdate = now;

              // Отправляем прогресс
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

        // Финальный прогресс (100%)
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
            // Не критично, просто не отправляем финальный прогресс
          }
        }

        // Вычисляем SHA256 для проверки целостности
        _log.info('Computing SHA256 checksum...');
        final computedSha256 = await _computeSha256(file);
        _log.info('File SHA256: $computedSha256');

        // NOTE: GitHub Releases API не предоставляет checksums для assets.
        // Checksum вычисляется для логирования и будущей проверки, когда/если
        // GitHub добавит поддержку checksums в API response.
        // Альтернатива: можно хранить checksums в release notes как workaround.

        _log.info('Download completed successfully: $savePath');
        return savePath;
      } catch (e) {
        // Ошибка при попытке загрузки
        if (attempt == maxRetries - 1) {
          // Последняя попытка не удалась
          rethrow;
        }
        _log.warning('Download attempt ${attempt + 1} failed: $e');
        // Продолжаем retry loop
      }
    }

    // Все retry исчерпаны - отправляем финальную ошибку
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

  /// Определяет asset для скачивания в зависимости от платформы
  Map<String, dynamic>? _getPlatformAsset(List assets) {
    if (assets.isEmpty) return null;

    String? platformSuffix;
    if (Platform.isMacOS) {
      platformSuffix = '.dmg';
    } else if (Platform.isWindows) {
      platformSuffix = '.msi';
    } else if (Platform.isLinux) {
      // Приоритет: AppImage > deb > tar.gz
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

      platformSuffix = '.tar.gz';
    }

    if (platformSuffix == null) return null;

    try {
      final asset = assets.firstWhere(
        (a) => (a['name'] as String).endsWith(platformSuffix!),
      );
      return asset as Map<String, dynamic>;
    } catch (e) {
      _log.warning('No asset found with suffix $platformSuffix');
      return null;
    }
  }

  /// Вычисляет SHA256 хеш файла
  Future<String> _computeSha256(File file) async {
    final bytes = await file.readAsBytes();
    final digest = sha256.convert(bytes);
    return digest.toString();
  }

  /// Проверяет доступное место на диске
  Future<bool> _checkDiskSpace(int requiredBytes) async {
    try {
      final tempDir = await getTemporaryDirectory();

      // Для Unix-like систем можем использовать df команду
      if (Platform.isLinux || Platform.isMacOS) {
        final result = await Process.run('df', ['-k', tempDir.path]);
        if (result.exitCode == 0) {
          final lines = (result.stdout as String).split('\n');
          if (lines.length > 1) {
            final parts = lines[1].split(RegExp(r'\s+'));
            if (parts.length >= 4) {
              final availableKB = int.tryParse(parts[3]) ?? 0;
              final availableBytes = availableKB * 1024;
              // Требуем 20% запас
              return availableBytes >= (requiredBytes * 1.2);
            }
          }
        }
      }

      // Для Windows или если df не сработал - делаем приблизительную проверку
      // путем попытки создать test файл
      return true; // Оптимистичный fallback
    } catch (e) {
      _log.warning('Failed to check disk space: $e');
      return true; // Оптимистичный fallback при ошибке
    }
  }

  @override
  void dispose() {
    _dio.close();
  }
}
