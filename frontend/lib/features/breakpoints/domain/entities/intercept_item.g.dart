// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'intercept_item.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_HTTPRequestSnapshot _$HTTPRequestSnapshotFromJson(Map<String, dynamic> json) =>
    _HTTPRequestSnapshot(
      method: json['method'] as String? ?? '',
      url: json['url'] as String? ?? '',
      headers:
          (json['headers'] as Map<String, dynamic>?)?.map(
            (k, e) => MapEntry(
              k,
              (e as List<dynamic>).map((e) => e as String).toList(),
            ),
          ) ??
          const {},
      bodyBase64: json['bodyBase64'] as String?,
      bodyTruncated: json['bodyTruncated'] as bool? ?? false,
      contentType: json['contentType'] as String?,
    );

Map<String, dynamic> _$HTTPRequestSnapshotToJson(
  _HTTPRequestSnapshot instance,
) => <String, dynamic>{
  'method': instance.method,
  'url': instance.url,
  'headers': instance.headers,
  'bodyBase64': instance.bodyBase64,
  'bodyTruncated': instance.bodyTruncated,
  'contentType': instance.contentType,
};

_HTTPResponseSnapshot _$HTTPResponseSnapshotFromJson(
  Map<String, dynamic> json,
) => _HTTPResponseSnapshot(
  status: (json['status'] as num?)?.toInt() ?? 0,
  headers:
      (json['headers'] as Map<String, dynamic>?)?.map(
        (k, e) =>
            MapEntry(k, (e as List<dynamic>).map((e) => e as String).toList()),
      ) ??
      const {},
  bodyBase64: json['bodyBase64'] as String?,
  bodyTruncated: json['bodyTruncated'] as bool? ?? false,
  contentType: json['contentType'] as String?,
);

Map<String, dynamic> _$HTTPResponseSnapshotToJson(
  _HTTPResponseSnapshot instance,
) => <String, dynamic>{
  'status': instance.status,
  'headers': instance.headers,
  'bodyBase64': instance.bodyBase64,
  'bodyTruncated': instance.bodyTruncated,
  'contentType': instance.contentType,
};

_InterceptItem _$InterceptItemFromJson(Map<String, dynamic> json) =>
    _InterceptItem(
      id: json['id'] as String,
      createdAt: DateTime.parse(json['createdAt'] as String),
      deadline: DateTime.parse(json['deadline'] as String),
      direction: json['direction'] as String,
      sessionId: json['sessionId'] as String,
      state: json['state'] as String? ?? 'PENDING',
      ruleId: json['ruleId'] as String?,
      req: json['req'] == null
          ? null
          : HTTPRequestSnapshot.fromJson(json['req'] as Map<String, dynamic>),
      res: json['res'] == null
          ? null
          : HTTPResponseSnapshot.fromJson(json['res'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$InterceptItemToJson(_InterceptItem instance) =>
    <String, dynamic>{
      'id': instance.id,
      'createdAt': instance.createdAt.toIso8601String(),
      'deadline': instance.deadline.toIso8601String(),
      'direction': instance.direction,
      'sessionId': instance.sessionId,
      'state': instance.state,
      'ruleId': instance.ruleId,
      'req': instance.req,
      'res': instance.res,
    };
