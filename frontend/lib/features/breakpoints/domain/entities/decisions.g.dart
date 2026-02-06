// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'decisions.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RequestDecision _$RequestDecisionFromJson(Map<String, dynamic> json) =>
    _RequestDecision(
      action: json['action'] as String,
      method: json['method'] as String?,
      url: json['url'] as String?,
      headers: (json['headers'] as Map<String, dynamic>?)?.map(
        (k, e) =>
            MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
      ),
      bodyBase64: json['bodyBase64'] as String?,
    );

Map<String, dynamic> _$RequestDecisionToJson(_RequestDecision instance) =>
    <String, dynamic>{
      'action': instance.action,
      'method': instance.method,
      'url': instance.url,
      'headers': instance.headers,
      'bodyBase64': instance.bodyBase64,
    };

_ResponseDecision _$ResponseDecisionFromJson(Map<String, dynamic> json) =>
    _ResponseDecision(
      action: json['action'] as String,
      status: (json['status'] as num?)?.toInt(),
      headers: (json['headers'] as Map<String, dynamic>?)?.map(
        (k, e) =>
            MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
      ),
      bodyBase64: json['bodyBase64'] as String?,
    );

Map<String, dynamic> _$ResponseDecisionToJson(_ResponseDecision instance) =>
    <String, dynamic>{
      'action': instance.action,
      'status': instance.status,
      'headers': instance.headers,
      'bodyBase64': instance.bodyBase64,
    };
