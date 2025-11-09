// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'match_rules.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_MatchRules _$MatchRulesFromJson(Map<String, dynamic> json) => _MatchRules(
  methods:
      (json['methods'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
  pathPattern: json['pathPattern'] as String?,
  hostPattern: json['hostPattern'] as String?,
  patternType:
      $enumDecodeNullable(_$PatternTypeEnumMap, json['patternType']) ??
      PatternType.wildcard,
);

Map<String, dynamic> _$MatchRulesToJson(_MatchRules instance) =>
    <String, dynamic>{
      'methods': instance.methods,
      'pathPattern': instance.pathPattern,
      'hostPattern': instance.hostPattern,
      'patternType': _$PatternTypeEnumMap[instance.patternType]!,
    };

const _$PatternTypeEnumMap = {
  PatternType.exact: 'exact',
  PatternType.prefix: 'prefix',
  PatternType.wildcard: 'wildcard',
};
