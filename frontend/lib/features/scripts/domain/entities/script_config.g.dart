// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'script_config.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ScriptConfig _$ScriptConfigFromJson(Map<String, dynamic> json) =>
    _ScriptConfig(
      timeoutMs: (json['timeoutMs'] as num?)?.toInt(),
      memoryLimitMB: (json['memoryLimitMB'] as num?)?.toInt(),
      allowedHosts:
          (json['allowedHosts'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          const [],
    );

Map<String, dynamic> _$ScriptConfigToJson(_ScriptConfig instance) =>
    <String, dynamic>{
      'timeoutMs': instance.timeoutMs,
      'memoryLimitMB': instance.memoryLimitMB,
      'allowedHosts': instance.allowedHosts,
    };
