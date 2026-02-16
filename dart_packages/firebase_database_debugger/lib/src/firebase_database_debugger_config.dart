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

  factory FirebaseDatabaseDebuggerConfig.defaults() {
    return FirebaseDatabaseDebuggerConfig(
      debuggerBaseUrl: 'http://localhost:9092',
      enabled: true,
    );
  }

  final String debuggerBaseUrl;
  final String databaseUrl;
  final bool enabled;
  final String? adminToken;
  final String captureId;

  /// Группировка сессий по пути:
  /// - `-1` (по умолчанию): отдельная сессия на полный путь (`/a/b/c`)
  /// - `0`: одна сессия на всю базу (в UI будет путь `/`)
  /// - `N>=1`: сессия на префикс пути глубины N (например, N=2: `/a/b/...` → `/a/b`)
  final int sessionPathDepth;

  /// Если `true`, добавляем короткий run-id в `session.id`, чтобы не конфликтовать
  /// с закрытыми сессиями при следующем запуске приложения.
  final bool includeRunIdInSessionId;
  final Duration flushInterval;
  final int maxBatchFrames;
  final int previewBodyThresholdBytes;
}
