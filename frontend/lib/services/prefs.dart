import 'package:shared_preferences/shared_preferences.dart';

class PrefsService {
  static const _keyBaseUrl = 'base_url';
  static const _keyTarget = 'target_ws';
  static const _keyQ = 'sessions_q';
  static const _keyTargetFilter = 'sessions_target_filter';
  static const _keyOpcode = 'frames_opcode';
  static const _keyDirection = 'frames_direction';
  static const _keyNamespace = 'events_namespace';
  static const _keyHttpMethod = 'http_method_filter';
  static const _keyHttpStatus = 'http_status_filter';
  static const _keyHttpMime = 'http_mime_filter';
  static const _keyHttpMinDur = 'http_min_duration';
  static const _keyGroupBy = 'sessions_group_by';
  static const _keyHeaderKey = 'sessions_header_key';
  static const _keyHeaderVal = 'sessions_header_val';
  static const _keyMonitorLog = 'monitor_log_json';
  static const _keyThemeMode = 'theme_mode';
  static const _keySinceTs = 'clear_since_ts';

  Future<void> save({
    required String baseUrl,
    required String targetWs,
    String? q,
    String? targetFilter,
    String? opcode,
    String? direction,
    String? namespace,
    String? httpMethod,
    String? httpStatus,
    String? httpMime,
    int? httpMinDurationMs,
    String? groupBy,
    String? headerKey,
    String? headerVal,
  }) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(_keyBaseUrl, baseUrl);
    await p.setString(_keyTarget, targetWs);
    if (q != null) await p.setString(_keyQ, q);
    if (targetFilter != null) await p.setString(_keyTargetFilter, targetFilter);
    if (opcode != null) await p.setString(_keyOpcode, opcode);
    if (direction != null) await p.setString(_keyDirection, direction);
    if (namespace != null) await p.setString(_keyNamespace, namespace);
    if (httpMethod != null) await p.setString(_keyHttpMethod, httpMethod);
    if (httpStatus != null) await p.setString(_keyHttpStatus, httpStatus);
    if (httpMime != null) await p.setString(_keyHttpMime, httpMime);
    if (httpMinDurationMs != null) await p.setInt(_keyHttpMinDur, httpMinDurationMs);
    if (groupBy != null) await p.setString(_keyGroupBy, groupBy);
    if (headerKey != null) await p.setString(_keyHeaderKey, headerKey);
    if (headerVal != null) await p.setString(_keyHeaderVal, headerVal);
  }

  Future<Map<String, String>> load() async {
    final p = await SharedPreferences.getInstance();
    return {
      'baseUrl': p.getString(_keyBaseUrl) ?? 'http://localhost:9091',
      'targetWs': p.getString(_keyTarget) ?? 'ws://echo.websocket.events',
      'q': p.getString(_keyQ) ?? '',
      'targetFilter': p.getString(_keyTargetFilter) ?? '',
      'opcode': p.getString(_keyOpcode) ?? 'all',
      'direction': p.getString(_keyDirection) ?? 'all',
      'namespace': p.getString(_keyNamespace) ?? '',
      'httpMethod': p.getString(_keyHttpMethod) ?? 'any',
      'httpStatus': p.getString(_keyHttpStatus) ?? 'any',
      'httpMime': p.getString(_keyHttpMime) ?? '',
      'httpMinDuration': (p.getInt(_keyHttpMinDur) ?? 0).toString(),
      'groupBy': p.getString(_keyGroupBy) ?? 'none',
      'headerKey': p.getString(_keyHeaderKey) ?? '',
      'headerVal': p.getString(_keyHeaderVal) ?? '',
    };
  }

  Future<void> saveSince(DateTime tsUtc) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(_keySinceTs, tsUtc.toIso8601String());
  }

  Future<DateTime?> loadSince() async {
    final p = await SharedPreferences.getInstance();
    final s = p.getString(_keySinceTs);
    if (s == null || s.isEmpty) return null;
    try { return DateTime.parse(s).toUtc(); } catch (_) { return null; }
  }
}

extension PrefsServiceMonitor on PrefsService {
  Future<void> saveMonitorLog(List<String> items) async {
    final p = await SharedPreferences.getInstance();
    final trimmed = items.length > 500 ? items.sublist(0, 500) : items;
    await p.setString(PrefsService._keyMonitorLog, trimmed.join('\n')); // компактно
  }

  Future<List<String>> loadMonitorLog() async {
    final p = await SharedPreferences.getInstance();
    final raw = p.getString(PrefsService._keyMonitorLog);
    if (raw == null || raw.isEmpty) return <String>[];
    return raw.split('\n');
  }
}

extension PrefsServiceTheme on PrefsService {
  Future<void> saveThemeModeString(String mode) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(PrefsService._keyThemeMode, mode);
  }

  Future<String> loadThemeModeString() async {
    final p = await SharedPreferences.getInstance();
    return p.getString(PrefsService._keyThemeMode) ?? 'system';
  }
}
