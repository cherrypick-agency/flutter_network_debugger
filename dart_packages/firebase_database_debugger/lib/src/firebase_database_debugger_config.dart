/// Configuration for [FirebaseDatabaseDebugger].
///
/// Controls connection to the debugger, session grouping, and frame batching.
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

  /// Configuration with default values for local development.
  factory FirebaseDatabaseDebuggerConfig.defaults() {
    return FirebaseDatabaseDebuggerConfig(
      debuggerBaseUrl: 'http://localhost:9092',
      enabled: true,
    );
  }

  /// Base URL of the debugger where frames are sent.
  ///
  /// E.g. `http://localhost:9092`.
  final String debuggerBaseUrl;

  /// Firebase Realtime Database URL, used to form the session target.
  ///
  /// Defaults to `https://firebase.local`.
  final String databaseUrl;

  /// Whether debug data collection and sending is enabled.
  ///
  /// If `false`, all `logOperation` calls are ignored.
  final bool enabled;

  /// Authorization token for the ingest API. Sent in the `X-Admin-Token` header.
  final String? adminToken;

  /// Capture ID in the debugger.
  ///
  /// Defaults to `current`.
  final String captureId;

  /// Session grouping by path depth:
  /// - `-1` (default): one session per full path (`/a/b/c`)
  /// - `0`: single session for the entire database (UI shows path `/`)
  /// - `N>=1`: session for path prefix of depth N (e.g. N=2: `/a/b/...` → `/a/b`)
  final int sessionPathDepth;

  /// If `true`, a short run-id is added to `session.id` to avoid conflicts
  /// with closed sessions on the next app launch.
  final bool includeRunIdInSessionId;

  /// Interval for automatic flush of buffered frames.
  ///
  /// Defaults to 200 ms.
  final Duration flushInterval;

  /// Maximum number of frames in one batch before forcing a flush.
  ///
  /// Valid range: 1–200.
  final int maxBatchFrames;

  /// Preview size threshold (in bytes) above which the body is moved to a
  /// separate `body` field with base64 encoding.
  ///
  /// Valid range: 1 KB – 256 KB.
  final int previewBodyThresholdBytes;
}
