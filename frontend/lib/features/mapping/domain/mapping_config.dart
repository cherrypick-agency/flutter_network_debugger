import '../../../core/utils/json_cast.dart';

class MappingConfig {
  const MappingConfig({required this.enabled, required this.uploadMaxMB});
  final bool enabled;
  final int uploadMaxMB;

  Map<String, dynamic> toJson() => {
    'enabled': enabled,
    'uploadMaxMB': uploadMaxMB,
  };

  static MappingConfig fromJson(Map<String, dynamic> j) => MappingConfig(
    enabled: JsonCast.asBool(j['enabled'], fallback: true),
    uploadMaxMB: () {
      final v = JsonCast.asInt(j['uploadMaxMB'], fallback: 20);
      return v > 0 ? v : 20;
    }(),
  );
}
