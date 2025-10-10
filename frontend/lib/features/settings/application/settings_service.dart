import 'package:app_http_client/application/app_http_client.dart';
import '../../../services/prefs.dart';
import '../../../core/di/di.dart';
import '../../inspector/application/services/recent_window_service.dart';

class SettingsService {
  Future<void> syncPrefsToBackend() async {
    try {
      final data = await PrefsService().load();
      final enabled = (data['respDelayEnabled'] ?? 'false') == 'true';
      final value = (data['respDelayValue'] ?? '').toString().trim();
      final api = sl<AppHttpClient>();
      await api.post(
        path: '/_api/v1/settings',
        body: {
          'responseDelay': {'enabled': enabled, 'value': value},
        },
      );
    } catch (_) {
      // тихо игнорируем: баннер покажет проблемы соединения
    }
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

  Future<void> saveRecentWindow({
    required bool enabled,
    required int minutes,
  }) async {
    // Сохраняем и отдаём в сервис, чтобы он обновил since и дёрнул перезагрузку
    await PrefsService().saveRecentWindow(enabled: enabled, minutes: minutes);
    sl<RecentWindowService>().apply(enabled: enabled, minutes: minutes);
  }
}
