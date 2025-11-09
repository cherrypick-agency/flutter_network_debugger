import 'package:app_http_client/application/app_http_client.dart';

class ProcessService {
  final AppHttpClient _api;

  ProcessService(this._api);

  /// Получить текущую конфигурацию детекции процессов
  Future<ProcessConfig> fetchConfig() async {
    final res = await _api.get(path: '/_api/v1/process/config');
    return ProcessConfig.fromJson(res.data as Map<String, dynamic>);
  }

  /// Сохранить конфигурацию детекции процессов
  Future<void> saveConfig(ProcessConfig config) async {
    await _api.post(path: '/_api/v1/process/config', body: config.toJson());
  }

  /// Проверить статус helper tool
  Future<HelperStatus> checkHelperStatus() async {
    final res = await _api.get(path: '/_api/v1/process/helper/status');
    return HelperStatus.fromJson(res.data as Map<String, dynamic>);
  }

  /// Установить helper tool (требует admin пароль)
  Future<void> installHelper() async {
    await _api.post(path: '/_api/v1/process/helper/install');
  }
}

/// Конфигурация детекции процессов
class ProcessConfig {
  final bool enabled;
  final bool useHelperTool;
  final bool helperInstalled;
  final bool cacheEnabled;
  final int cacheTtl;
  final bool fallbackEnabled;

  ProcessConfig({
    required this.enabled,
    required this.useHelperTool,
    required this.helperInstalled,
    required this.cacheEnabled,
    required this.cacheTtl,
    required this.fallbackEnabled,
  });

  factory ProcessConfig.fromJson(Map<String, dynamic> json) {
    return ProcessConfig(
      enabled: json['enabled'] as bool? ?? true,
      useHelperTool: json['useHelperTool'] as bool? ?? false,
      helperInstalled: json['helperInstalled'] as bool? ?? false,
      cacheEnabled: json['cacheEnabled'] as bool? ?? true,
      cacheTtl: json['cacheTtl'] as int? ?? 300,
      fallbackEnabled: json['fallbackEnabled'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toJson() => {
    'enabled': enabled,
    'useHelperTool': useHelperTool,
    'helperInstalled': helperInstalled,
    'cacheEnabled': cacheEnabled,
    'cacheTtl': cacheTtl,
    'fallbackEnabled': fallbackEnabled,
  };

  ProcessConfig copyWith({
    bool? enabled,
    bool? useHelperTool,
    bool? helperInstalled,
    bool? cacheEnabled,
    int? cacheTtl,
    bool? fallbackEnabled,
  }) {
    return ProcessConfig(
      enabled: enabled ?? this.enabled,
      useHelperTool: useHelperTool ?? this.useHelperTool,
      helperInstalled: helperInstalled ?? this.helperInstalled,
      cacheEnabled: cacheEnabled ?? this.cacheEnabled,
      cacheTtl: cacheTtl ?? this.cacheTtl,
      fallbackEnabled: fallbackEnabled ?? this.fallbackEnabled,
    );
  }
}

/// Статус helper tool
class HelperStatus {
  final bool running;
  final bool installed;
  final String version;

  HelperStatus({
    required this.running,
    required this.installed,
    required this.version,
  });

  factory HelperStatus.fromJson(Map<String, dynamic> json) {
    return HelperStatus(
      running: json['running'] as bool? ?? false,
      installed: json['installed'] as bool? ?? false,
      version: json['version'] as String? ?? '',
    );
  }
}
