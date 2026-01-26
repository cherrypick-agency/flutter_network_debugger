import 'package:shared_preferences/shared_preferences.dart';
import 'dart:convert';

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
  static const _keyRespDelayEnabled = 'resp_delay_enabled';
  static const _keyRespDelayValue =
      'resp_delay_value'; // "1500" или "1000-3000"
  static const _keyIsRecording = 'is_recording';
  static const _keyRecentWindowEnabled = 'recent_window_enabled';
  static const _keyRecentWindowMinutes = 'recent_window_minutes';
  static const _keyAdminToken = 'admin_token';
  static const _keyFontScale = 'font_scale';
  static const _keyHighlightTheme = 'highlight_theme';
  static const _keySelectedTags = 'selected_tags';

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
    bool? respDelayEnabled,
    String? respDelayValue,
    String? selectedTags,
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
    if (httpMinDurationMs != null)
      await p.setInt(_keyHttpMinDur, httpMinDurationMs);
    if (groupBy != null) await p.setString(_keyGroupBy, groupBy);
    if (headerKey != null) await p.setString(_keyHeaderKey, headerKey);
    if (headerVal != null) await p.setString(_keyHeaderVal, headerVal);
    if (respDelayEnabled != null)
      await p.setBool(_keyRespDelayEnabled, respDelayEnabled);
    if (respDelayValue != null)
      await p.setString(_keyRespDelayValue, respDelayValue);
    if (selectedTags != null) await p.setString(_keySelectedTags, selectedTags);
  }

  Future<Map<String, String>> load() async {
    final p = await SharedPreferences.getInstance();
    return {
      'baseUrl': p.getString(_keyBaseUrl) ?? 'http://localhost:9092',
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
      'respDelayEnabled': (p.getBool(_keyRespDelayEnabled) ?? false).toString(),
      'respDelayValue': p.getString(_keyRespDelayValue) ?? '',
      // recent window (stringified for legacy map)
      'recentWindowEnabled': (p.getBool(_keyRecentWindowEnabled) ?? false)
          .toString(),
      'recentWindowMinutes': (p.getInt(_keyRecentWindowMinutes) ?? 5)
          .toString(),
      'adminToken': p.getString(_keyAdminToken) ?? '',
      'fontScale': (p.getDouble(_keyFontScale) ?? 1.0).toString(),
      'selectedTags': p.getString(_keySelectedTags) ?? '',
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
    try {
      return DateTime.parse(s).toUtc();
    } catch (_) {
      return null;
    }
  }
}

extension PrefsServiceMonitor on PrefsService {
  Future<void> saveMonitorLog(List<String> items) async {
    final p = await SharedPreferences.getInstance();
    final trimmed = items.length > 500 ? items.sublist(0, 500) : items;
    await p.setString(
      PrefsService._keyMonitorLog,
      trimmed.join('\n'),
    ); // компактно
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

extension PrefsServiceFont on PrefsService {
  Future<void> saveFontScale(double v) async {
    final p = await SharedPreferences.getInstance();
    await p.setDouble(PrefsService._keyFontScale, v);
  }

  Future<double> loadFontScale() async {
    final p = await SharedPreferences.getInstance();
    return p.getDouble(PrefsService._keyFontScale) ?? 1.0;
  }
}

extension PrefsServiceHighlightTheme on PrefsService {
  Future<void> saveHighlightTheme(String themeKey) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(PrefsService._keyHighlightTheme, themeKey);
  }

  Future<String?> loadHighlightTheme() async {
    final p = await SharedPreferences.getInstance();
    return p.getString(PrefsService._keyHighlightTheme);
  }
}

extension PrefsServiceRecording on PrefsService {
  Future<void> saveIsRecording(bool isRecording) async {
    final p = await SharedPreferences.getInstance();
    await p.setBool(PrefsService._keyIsRecording, isRecording);
  }

  Future<bool> loadIsRecording() async {
    final p = await SharedPreferences.getInstance();
    return p.getBool(PrefsService._keyIsRecording) ?? true;
  }
}

extension PrefsServiceRecentWindow on PrefsService {
  Future<void> saveRecentWindow({
    required bool enabled,
    required int minutes,
  }) async {
    final p = await SharedPreferences.getInstance();
    await p.setBool(PrefsService._keyRecentWindowEnabled, enabled);
    await p.setInt(
      PrefsService._keyRecentWindowMinutes,
      minutes.clamp(1, 1440),
    );
  }

  Future<bool> loadRecentWindowEnabled() async {
    final p = await SharedPreferences.getInstance();
    return p.getBool(PrefsService._keyRecentWindowEnabled) ?? false;
  }

  Future<int> loadRecentWindowMinutes() async {
    final p = await SharedPreferences.getInstance();
    return p.getInt(PrefsService._keyRecentWindowMinutes) ?? 5;
  }
}

// Compose drafts persistence
extension PrefsServiceCompose on PrefsService {
  static const _composeDraftKey = 'compose_draft_json';
  static String _composeDraftKeyFor(String id) => 'compose_draft_json_$id';

  Future<void> saveComposeDraft(Map<String, dynamic> draft) async {
    try {
      final p = await SharedPreferences.getInstance();
      await p.setString(_composeDraftKey, jsonEncode(draft));
    } catch (_) {}
  }

  Future<Map<String, dynamic>?> loadComposeDraft() async {
    try {
      final p = await SharedPreferences.getInstance();
      final raw = p.getString(_composeDraftKey);
      if (raw == null || raw.isEmpty) return null;
      final m = jsonDecode(raw);
      if (m is Map<String, dynamic>) return m;
      return Map<String, dynamic>.from(m as Map);
    } catch (_) {
      return null;
    }
  }

  Future<void> clearComposeDraft() async {
    try {
      final p = await SharedPreferences.getInstance();
      await p.remove(_composeDraftKey);
    } catch (_) {}
  }

  Future<void> saveComposeDraftFor(
    String id,
    Map<String, dynamic> draft,
  ) async {
    try {
      final p = await SharedPreferences.getInstance();
      await p.setString(_composeDraftKeyFor(id), jsonEncode(draft));
    } catch (_) {}
  }

  Future<Map<String, dynamic>?> loadComposeDraftFor(String id) async {
    try {
      final p = await SharedPreferences.getInstance();
      final raw = p.getString(_composeDraftKeyFor(id));
      if (raw == null || raw.isEmpty) return null;
      final m = jsonDecode(raw);
      if (m is Map<String, dynamic>) return m;
      return Map<String, dynamic>.from(m as Map);
    } catch (_) {
      return null;
    }
  }

  Future<void> clearComposeDraftFor(String id) async {
    try {
      final p = await SharedPreferences.getInstance();
      await p.remove(_composeDraftKeyFor(id));
    } catch (_) {}
  }
}

extension PrefsServiceAdmin on PrefsService {
  Future<void> saveAdminToken(String token) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(PrefsService._keyAdminToken, token.trim());
  }

  Future<String> loadAdminToken() async {
    final p = await SharedPreferences.getInstance();
    return p.getString(PrefsService._keyAdminToken) ?? '';
  }
}

extension PrefsServiceVisualDensity on PrefsService {
  static const _keyVisualDensity = 'visual_density';

  /// Сохраняет уровень компактности UI: 'standard', 'comfortable', 'compact'
  Future<void> saveVisualDensity(String density) async {
    final p = await SharedPreferences.getInstance();
    await p.setString(_keyVisualDensity, density);
  }

  /// Загружает уровень компактности UI
  Future<String> loadVisualDensity() async {
    final p = await SharedPreferences.getInstance();
    return p.getString(_keyVisualDensity) ?? 'standard';
  }
}
