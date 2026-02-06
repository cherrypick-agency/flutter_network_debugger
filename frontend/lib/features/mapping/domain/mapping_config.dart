import 'package:freezed_annotation/freezed_annotation.dart';

import '../../../core/utils/json_cast.dart';

part 'mapping_config.freezed.dart';

@Freezed(toJson: false, fromJson: false)
sealed class MappingConfig with _$MappingConfig {
  const MappingConfig._();

  const factory MappingConfig({
    @Default(true) bool enabled,
    @Default(20) int uploadMaxMB,
  }) = _MappingConfig;

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
