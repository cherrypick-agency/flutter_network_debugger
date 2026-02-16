/// Конфигурация [FirebaseDatabaseDebugger].
///
/// Управляет подключением к дебаггеру, группировкой сессий и батчингом фреймов.
class FirebaseDatabaseDebuggerConfig {
  FirebaseDatabaseDebuggerConfig({
    required this.debuggerBaseUrl,
    this.databaseUrl = 'https://firebase.local',
    this.enabled = true,
    this.adminToken,
    this.captureId = 'current',
    int sessionPathDepth = -1,
    bool includeRunIdInSessionId = true,
    Duration? flushInterval,
    int maxBatchFrames = 100,
    int previewBodyThresholdBytes = 16 * 1024,
  })  : flushInterval = flushInterval ?? const Duration(milliseconds: 200),
        sessionPathDepth = sessionPathDepth.clamp(-1, 20),
        includeRunIdInSessionId = includeRunIdInSessionId,
        maxBatchFrames = maxBatchFrames.clamp(1, 200),
        previewBodyThresholdBytes = previewBodyThresholdBytes.clamp(
          1024,
          256 * 1024,
        );

  /// Конфигурация с дефолтными значениями для локальной разработки.
  factory FirebaseDatabaseDebuggerConfig.defaults() {
    return FirebaseDatabaseDebuggerConfig(
      debuggerBaseUrl: 'http://localhost:9092',
      enabled: true,
    );
  }

  /// Базовый URL дебаггера, куда отправляются фреймы.
  ///
  /// Например, `http://localhost:9092`.
  final String debuggerBaseUrl;

  /// URL Firebase Realtime Database, используется для формирования target сессии.
  ///
  /// По умолчанию `https://firebase.local`.
  final String databaseUrl;

  /// Включён ли сбор и отправка отладочных данных.
  ///
  /// Если `false`, все вызовы `logOperation` игнорируются.
  final bool enabled;

  /// Токен авторизации для ingest API. Передаётся в заголовке `X-Admin-Token`.
  final String? adminToken;

  /// Идентификатор захвата (capture) в дебаггере.
  ///
  /// По умолчанию `current`.
  final String captureId;

  /// Группировка сессий по глубине пути:
  /// - `-1` (по умолчанию): отдельная сессия на полный путь (`/a/b/c`)
  /// - `0`: одна сессия на всю базу (в UI будет путь `/`)
  /// - `N>=1`: сессия на префикс пути глубины N (например, N=2: `/a/b/...` → `/a/b`)
  final int sessionPathDepth;

  /// Если `true`, добавляем короткий run-id в `session.id`, чтобы не конфликтовать
  /// с закрытыми сессиями при следующем запуске приложения.
  final bool includeRunIdInSessionId;

  /// Интервал автоматического flush буферизованных фреймов.
  ///
  /// По умолчанию 200 мс.
  final Duration flushInterval;

  /// Максимальное количество фреймов в одном батче перед принудительным flush.
  ///
  /// Допустимый диапазон: 1–200.
  final int maxBatchFrames;

  /// Порог размера превью (в байтах), после которого тело выносится в отдельное
  /// поле `body` с base64-кодировкой.
  ///
  /// Допустимый диапазон: 1 КБ – 256 КБ.
  final int previewBodyThresholdBytes;
}
