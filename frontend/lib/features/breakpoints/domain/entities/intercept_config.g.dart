// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'intercept_config.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_InterceptConfig _$InterceptConfigFromJson(Map<String, dynamic> json) =>
    _InterceptConfig(
      enabled: json['enabled'] as bool? ?? false,
      requests: json['requests'] as bool? ?? true,
      responses: json['responses'] as bool? ?? true,
      timeoutMs: (json['timeoutMs'] as num?)?.toInt() ?? 60000,
      queueMax: (json['queueMax'] as num?)?.toInt() ?? 200,
      bodyMaxBytes: (json['bodyMaxBytes'] as num?)?.toInt() ?? 1048576,
      reencode: json['reencode'] as bool? ?? true,
      overflow: json['overflow'] as String? ?? 'auto-continue-oldest',
    );

Map<String, dynamic> _$InterceptConfigToJson(_InterceptConfig instance) =>
    <String, dynamic>{
      'enabled': instance.enabled,
      'requests': instance.requests,
      'responses': instance.responses,
      'timeoutMs': instance.timeoutMs,
      'queueMax': instance.queueMax,
      'bodyMaxBytes': instance.bodyMaxBytes,
      'reencode': instance.reencode,
      'overflow': instance.overflow,
    };
