class FirebaseDatabaseDebuggerConfig {
  FirebaseDatabaseDebuggerConfig({
    required this.debuggerBaseUrl,
    this.databaseUrl = 'https://firebase.local',
    this.enabled = true,
    this.adminToken,
    this.captureId = 'current',
    Duration? flushInterval,
    int maxBatchFrames = 100,
    int previewBodyThresholdBytes = 16 * 1024,
  })  : flushInterval = flushInterval ?? const Duration(milliseconds: 200),
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
  final Duration flushInterval;
  final int maxBatchFrames;
  final int previewBodyThresholdBytes;
}
