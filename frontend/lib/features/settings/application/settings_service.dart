import 'package:app_http_client/application/app_http_client.dart';
import '../../../services/prefs.dart';
import '../../../core/di/di.dart';
import '../../inspector/application/services/recent_window_service.dart';

class SettingsService {
  Future<Map<String, dynamic>> fetchRuntime() async {
    final api = sl<AppHttpClient>();
    final res = await api.get(path: '/_api/v1/settings');
    return (res.data as Map).cast<String, dynamic>();
  }

  Future<void> saveHighlightTheme(String themeKey) async {
    final api = sl<AppHttpClient>();
    await api.post(
      path: '/_api/v1/settings',
      body: {'highlightTheme': themeKey},
    );
  }

  Future<void> saveHighlightThemeDual({
    required String light,
    required String dark,
  }) async {
    final api = sl<AppHttpClient>();
    await api.post(
      path: '/_api/v1/settings',
      body: {'highlightThemeLight': light, 'highlightThemeDark': dark},
    );
  }

  Future<void> saveHighlightThemeLight(String themeKey) async {
    final api = sl<AppHttpClient>();
    await api.post(
      path: '/_api/v1/settings',
      body: {'highlightThemeLight': themeKey},
    );
  }

  Future<void> saveHighlightThemeDark(String themeKey) async {
    final api = sl<AppHttpClient>();
    await api.post(
      path: '/_api/v1/settings',
      body: {'highlightThemeDark': themeKey},
    );
  }

  Future<void> syncPrefsToBackend() async {
    try {
      final data = await PrefsService().load();
      final enabled = (data['respDelayEnabled'] ?? 'false') == 'true';
      final value = (data['respDelayValue'] ?? '').toString().trim();
      final fontScale =
          double.tryParse((data['fontScale'] ?? '1.0').toString()) ?? 1.0;
      final api = sl<AppHttpClient>();
      await api.post(
        path: '/_api/v1/settings',
        body: {
          'responseDelay': {'enabled': enabled, 'value': value},
          'fontScale': fontScale,
        },
      );
    } catch (_) {
      // silently ignore: banner will show connection problems
    }
  }

  // ---- Proxy runtime (HTTP forward + SOCKS) ----
  Future<Map<String, dynamic>> fetchProxyConfig() async {
    final api = sl<AppHttpClient>();
    final res = await api.get(path: '/_api/v1/proxy/config');
    return (res.data as Map).cast<String, dynamic>();
  }

  Future<Map<String, dynamic>> saveProxyConfig({
    required bool forwardEnabled,
    required int forwardPort,
    required bool socksEnabled,
    required int socksPort,
    String authMode = 'none',
    String user = '',
    String pass = '',
  }) async {
    final api = sl<AppHttpClient>();
    final body = {
      'forward': {'enabled': forwardEnabled, 'port': forwardPort},
      'socks': {
        'enabled': socksEnabled,
        'port': socksPort,
        'authMode': authMode,
        if (user.isNotEmpty) 'user': user,
        if (pass.isNotEmpty) 'pass': pass,
      },
    };
    final res = await api.post(path: '/_api/v1/proxy/config', body: body);
    return (res.data as Map).cast<String, dynamic>();
  }

  Future<void> saveResponseDelay({
    required bool enabled,
    required String value,
  }) async {
    await PrefsService().save(
      baseUrl: sl<AppHttpClient>().defaultHost,
      targetWs: '',
      respDelayEnabled: enabled,
      respDelayValue: value.trim(),
    );
    await syncPrefsToBackend();
  }

  Future<void> saveFontScale(double v) async {
    await PrefsService().saveFontScale(v);
    await syncPrefsToBackend();
  }

  Future<void> saveRecentWindow({
    required bool enabled,
    required int minutes,
  }) async {
    // Save and pass to service so it updates since and triggers reload
    await PrefsService().saveRecentWindow(enabled: enabled, minutes: minutes);
    sl<RecentWindowService>().apply(enabled: enabled, minutes: minutes);
  }
}
