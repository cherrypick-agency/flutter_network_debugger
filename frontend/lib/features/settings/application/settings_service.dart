import 'package:app_http_client/application/app_http_client.dart';
import '../../../core/di/di.dart';
import '../../../services/prefs.dart';

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
}
