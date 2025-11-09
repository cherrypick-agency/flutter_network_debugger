// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'script_test_result.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TestRequest _$TestRequestFromJson(Map<String, dynamic> json) => _TestRequest(
  method: json['method'] as String,
  url: json['url'] as String,
  headers:
      (json['headers'] as Map<String, dynamic>?)?.map(
        (k, e) =>
            MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
      ) ??
      const {},
  body: json['body'] as String?,
);

Map<String, dynamic> _$TestRequestToJson(_TestRequest instance) =>
    <String, dynamic>{
      'method': instance.method,
      'url': instance.url,
      'headers': instance.headers,
      'body': instance.body,
    };

_ModifiedHTTP _$ModifiedHTTPFromJson(Map<String, dynamic> json) =>
    _ModifiedHTTP(
      method: json['method'] as String?,
      url: json['url'] as String?,
      headers: (json['headers'] as Map<String, dynamic>?)?.map(
        (k, e) =>
            MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
      ),
      body: json['body'] as String?,
      status: (json['status'] as num?)?.toInt(),
    );

Map<String, dynamic> _$ModifiedHTTPToJson(_ModifiedHTTP instance) =>
    <String, dynamic>{
      'method': instance.method,
      'url': instance.url,
      'headers': instance.headers,
      'body': instance.body,
      'status': instance.status,
    };

_ScriptTestResult _$ScriptTestResultFromJson(
  Map<String, dynamic> json,
) => _ScriptTestResult(
  success: json['success'] as bool,
  error: json['error'] as String?,
  modifiedRequest: json['modifiedRequest'] == null
      ? null
      : ModifiedHTTP.fromJson(json['modifiedRequest'] as Map<String, dynamic>),
  modifiedResponse: json['modifiedResponse'] == null
      ? null
      : ModifiedHTTP.fromJson(json['modifiedResponse'] as Map<String, dynamic>),
  logs:
      (json['logs'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
  durationMs: (json['durationMs'] as num?)?.toInt(),
);

Map<String, dynamic> _$ScriptTestResultToJson(_ScriptTestResult instance) =>
    <String, dynamic>{
      'success': instance.success,
      'error': instance.error,
      'modifiedRequest': instance.modifiedRequest,
      'modifiedResponse': instance.modifiedResponse,
      'logs': instance.logs,
      'durationMs': instance.durationMs,
    };
