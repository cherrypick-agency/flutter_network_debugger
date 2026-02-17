import 'package:firebase_database_debugger/src/platform_stub.dart'
    if (dart.library.io) 'package:firebase_database_debugger/src/platform_io.dart'
    as platform;

const String _kDefineFirebaseDebuggerBaseUrl =
    String.fromEnvironment('FIREBASE_DEBUGGER_BASE_URL');
const String _kDefineFirebaseDatabaseDebuggerBaseUrl =
    String.fromEnvironment('FIREBASE_DATABASE_DEBUGGER_BASE_URL');
const String _kDefineFirebaseDebuggerEnabled =
    String.fromEnvironment('FIREBASE_DEBUGGER_ENABLED');
const String _kDefineFirebaseDatabaseDebuggerEnabled =
    String.fromEnvironment('FIREBASE_DATABASE_DEBUGGER_ENABLED');

/// Configuration for [FirebaseDatabaseDebugger].
///
/// Controls connection to the debugger, session grouping, and frame batching.
class FirebaseDatabaseDebuggerConfig {
  FirebaseDatabaseDebuggerConfig({
    String? debuggerBaseUrl,
    this.databaseUrl = 'https://firebase.local',
    this.enabled = true,
    this.adminToken,
    this.captureId = 'current',
    int sessionPathDepth = -1,
    bool includeRunIdInSessionId = true,
    Duration? flushInterval,
    int maxBatchFrames = 100,
    int previewBodyThresholdBytes = 16 * 1024,
  })  : debuggerBaseUrl = _normalizeDebuggerBaseUrl(
          debuggerBaseUrl ?? _defaultDebuggerBaseUrl(),
        ),
        flushInterval = flushInterval ?? const Duration(milliseconds: 200),
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
      debuggerBaseUrl: _defaultDebuggerBaseUrlFromDefine(),
      enabled: _defaultEnabledFromDefine(),
    );
  }

  /// Base URL of the debugger where frames are sent.
  ///
  /// Defaults to:
  /// - `http://10.0.2.2:9092` on Android emulator
  /// - `http://localhost:9092` on other platforms
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

  static String _defaultDebuggerBaseUrl() {
    return platform.isAndroid
        ? 'http://10.0.2.2:9092'
        : 'http://localhost:9092';
  }

  static String _defaultDebuggerBaseUrlFromDefine() {
    final fromDefine = _firstNonEmpty([
      _kDefineFirebaseDatabaseDebuggerBaseUrl,
      _kDefineFirebaseDebuggerBaseUrl,
    ]);
    return fromDefine ?? _defaultDebuggerBaseUrl();
  }

  static bool _defaultEnabledFromDefine() {
    final fromDefine = _firstNonEmpty([
      _kDefineFirebaseDatabaseDebuggerEnabled,
      _kDefineFirebaseDebuggerEnabled,
    ]);
    if (fromDefine == null) return true;
    final v = fromDefine.trim().toLowerCase();
    return v == '1' || v == 'true' || v == 'yes' || v == 'on';
  }

  static String? _firstNonEmpty(List<String?> values) {
    for (final value in values) {
      if (value != null && value.trim().isNotEmpty) return value;
    }
    return null;
  }

  static String _normalizeDebuggerBaseUrl(String value) {
    var normalized = value.trim();
    if (normalized.isEmpty) return normalized;
    if (!normalized.contains('://')) {
      normalized = 'http://$normalized';
    }
    if (normalized.endsWith('/')) {
      normalized = normalized.substring(0, normalized.length - 1);
    }
    return normalized;
  }
}
