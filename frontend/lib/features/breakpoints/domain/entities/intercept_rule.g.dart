// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'intercept_rule.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RuleStringMatch _$RuleStringMatchFromJson(Map<String, dynamic> json) =>
    _RuleStringMatch(
      equals: json['equals'] as String? ?? '',
      prefix: json['prefix'] as String? ?? '',
      suffix: json['suffix'] as String? ?? '',
      contains: json['contains'] as String? ?? '',
      anyOf:
          (json['anyOf'] as List<dynamic>?)?.map((e) => e as String).toList() ??
          const [],
      regex: json['regex'] as String? ?? '',
    );

Map<String, dynamic> _$RuleStringMatchToJson(_RuleStringMatch instance) =>
    <String, dynamic>{
      'equals': instance.equals,
      'prefix': instance.prefix,
      'suffix': instance.suffix,
      'contains': instance.contains,
      'anyOf': instance.anyOf,
      'regex': instance.regex,
    };

_RuleHeaderMatch _$RuleHeaderMatchFromJson(Map<String, dynamic> json) =>
    _RuleHeaderMatch(
      name: json['name'] == null
          ? const RuleStringMatch()
          : RuleStringMatch.fromJson(json['name'] as Map<String, dynamic>),
      value: json['value'] == null
          ? const RuleStringMatch()
          : RuleStringMatch.fromJson(json['value'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$RuleHeaderMatchToJson(_RuleHeaderMatch instance) =>
    <String, dynamic>{'name': instance.name, 'value': instance.value};

_RuleStatusMatch _$RuleStatusMatchFromJson(Map<String, dynamic> json) =>
    _RuleStatusMatch(
      statusEquals:
          (json['equals'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          const [],
      from: (json['from'] as num?)?.toInt() ?? 0,
      to: (json['to'] as num?)?.toInt() ?? 0,
      is4xx: json['is4xx'] as bool? ?? false,
      is5xx: json['is5xx'] as bool? ?? false,
    );

Map<String, dynamic> _$RuleStatusMatchToJson(_RuleStatusMatch instance) =>
    <String, dynamic>{
      'equals': instance.statusEquals,
      'from': instance.from,
      'to': instance.to,
      'is4xx': instance.is4xx,
      'is5xx': instance.is5xx,
    };

_InterceptWhen _$InterceptWhenFromJson(
  Map<String, dynamic> json,
) => _InterceptWhen(
  method:
      (json['method'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
  scheme:
      (json['scheme'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
  host: json['host'] == null
      ? null
      : RuleStringMatch.fromJson(json['host'] as Map<String, dynamic>),
  port: json['port'] == null
      ? null
      : RuleStringMatch.fromJson(json['port'] as Map<String, dynamic>),
  path: json['path'] == null
      ? null
      : RuleStringMatch.fromJson(json['path'] as Map<String, dynamic>),
  contentType: json['contentType'] == null
      ? null
      : RuleStringMatch.fromJson(json['contentType'] as Map<String, dynamic>),
  responseStatus: json['responseStatus'] == null
      ? null
      : RuleStatusMatch.fromJson(
          json['responseStatus'] as Map<String, dynamic>,
        ),
  header: json['header'] == null
      ? null
      : RuleHeaderMatch.fromJson(json['header'] as Map<String, dynamic>),
  bodyContains: json['bodyContains'] as String?,
);

Map<String, dynamic> _$InterceptWhenToJson(_InterceptWhen instance) =>
    <String, dynamic>{
      'method': instance.method,
      'scheme': instance.scheme,
      'host': ?instance.host,
      'port': ?instance.port,
      'path': ?instance.path,
      'contentType': ?instance.contentType,
      'responseStatus': ?instance.responseStatus,
      'header': ?instance.header,
      'bodyContains': ?instance.bodyContains,
    };

_InterceptRule _$InterceptRuleFromJson(Map<String, dynamic> json) =>
    _InterceptRule(
      id: json['id'] as String? ?? '',
      enabled: json['enabled'] as bool? ?? true,
      priority: (json['priority'] as num?)?.toInt() ?? 10,
      action: json['action'] as String? ?? 'both',
      once: json['once'] as bool? ?? false,
      stopProcessing: json['stopProcessing'] as bool? ?? false,
      when: json['when'] == null
          ? const InterceptWhen()
          : InterceptWhen.fromJson(json['when'] as Map<String, dynamic>),
      createdAt: json['createdAt'] == null
          ? null
          : DateTime.parse(json['createdAt'] as String),
      updatedAt: json['updatedAt'] == null
          ? null
          : DateTime.parse(json['updatedAt'] as String),
    );

Map<String, dynamic> _$InterceptRuleToJson(_InterceptRule instance) =>
    <String, dynamic>{
      'id': instance.id,
      'enabled': instance.enabled,
      'priority': instance.priority,
      'action': instance.action,
      'once': instance.once,
      'stopProcessing': instance.stopProcessing,
      'when': instance.when,
      'createdAt': instance.createdAt?.toIso8601String(),
      'updatedAt': instance.updatedAt?.toIso8601String(),
    };
