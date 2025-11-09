import 'package:freezed_annotation/freezed_annotation.dart';

part 'script_config.freezed.dart';
part 'script_config.g.dart';

/// Script execution configuration and limits
@freezed
sealed class ScriptConfig with _$ScriptConfig {
  const ScriptConfig._();

  const factory ScriptConfig({
    int? timeoutMs,
    int? memoryLimitMB,
    @Default([]) List<String> allowedHosts,
  }) = _ScriptConfig;

  factory ScriptConfig.fromJson(Map<String, dynamic> json) =>
      _$ScriptConfigFromJson(json);
}

/// Extension for ScriptConfig helpers
extension ScriptConfigX on ScriptConfig {
  /// Get default config values
  static ScriptConfig get defaults => const ScriptConfig(
    timeoutMs: 5000, // 5 seconds
    memoryLimitMB: 128, // 128 MB
  );

  /// Check if config has custom values
  bool get hasCustomValues =>
      timeoutMs != null || memoryLimitMB != null || allowedHosts.isNotEmpty;

  /// Get human-readable timeout description
  String get timeoutDescription {
    if (timeoutMs == null) return 'Default (5s)';
    if (timeoutMs! < 1000) return '${timeoutMs}ms';
    return '${(timeoutMs! / 1000).toStringAsFixed(1)}s';
  }

  /// Get human-readable memory limit description
  String get memoryDescription {
    if (memoryLimitMB == null) return 'Default (128MB)';
    if (memoryLimitMB! >= 1024) {
      return '${(memoryLimitMB! / 1024).toStringAsFixed(1)}GB';
    }
    return '${memoryLimitMB}MB';
  }
}
