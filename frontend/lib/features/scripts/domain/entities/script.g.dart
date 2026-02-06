// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'script.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Script _$ScriptFromJson(Map<String, dynamic> json) => _Script(
  id: json['id'] as String,
  name: json['name'] as String,
  description: json['description'] as String?,
  runtime: $enumDecode(_$ScriptRuntimeEnumMap, json['runtime']),
  code: json['code'] as String,
  language: json['language'] as String,
  triggerType: $enumDecode(_$TriggerTypeEnumMap, json['triggerType']),
  priority: (json['priority'] as num?)?.toInt() ?? 10,
  enabled: json['enabled'] as bool? ?? true,
  matchRules: json['matchRules'] == null
      ? null
      : MatchRules.fromJson(json['matchRules'] as Map<String, dynamic>),
  config: json['config'] == null
      ? null
      : ScriptConfig.fromJson(json['config'] as Map<String, dynamic>),
  createdAt: json['createdAt'] == null
      ? null
      : DateTime.parse(json['createdAt'] as String),
  updatedAt: json['updatedAt'] == null
      ? null
      : DateTime.parse(json['updatedAt'] as String),
  sourceCode: json['sourceCode'] as String?,
  dependencies: (json['dependencies'] as Map<String, dynamic>?)?.map(
    (k, e) => MapEntry(k, e as String),
  ),
  compilationStatus: $enumDecodeNullable(
    _$CompilationStatusEnumMap,
    json['compilationStatus'],
  ),
  compilationError: json['compilationError'] as String?,
  lastCompiledAt: json['lastCompiledAt'] == null
      ? null
      : DateTime.parse(json['lastCompiledAt'] as String),
  validationStatus: $enumDecodeNullable(
    _$ValidationStatusEnumMap,
    json['validationStatus'],
  ),
  validationError: json['validationError'] as String?,
);

Map<String, dynamic> _$ScriptToJson(_Script instance) => <String, dynamic>{
  'id': instance.id,
  'name': instance.name,
  'description': instance.description,
  'runtime': _$ScriptRuntimeEnumMap[instance.runtime]!,
  'code': instance.code,
  'language': instance.language,
  'triggerType': _$TriggerTypeEnumMap[instance.triggerType]!,
  'priority': instance.priority,
  'enabled': instance.enabled,
  'matchRules': instance.matchRules,
  'config': instance.config,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
  'sourceCode': instance.sourceCode,
  'dependencies': instance.dependencies,
  'compilationStatus': _$CompilationStatusEnumMap[instance.compilationStatus],
  'compilationError': instance.compilationError,
  'lastCompiledAt': instance.lastCompiledAt?.toIso8601String(),
  'validationStatus': _$ValidationStatusEnumMap[instance.validationStatus],
  'validationError': instance.validationError,
};

const _$ScriptRuntimeEnumMap = {
  ScriptRuntime.extism: 'extism',
  ScriptRuntime.dart: 'dart',
};

const _$TriggerTypeEnumMap = {
  TriggerType.request: 'request',
  TriggerType.response: 'response',
  TriggerType.both: 'both',
};

const _$CompilationStatusEnumMap = {
  CompilationStatus.notCompiled: 'not_compiled',
  CompilationStatus.pending: 'pending',
  CompilationStatus.compiling: 'compiling',
  CompilationStatus.success: 'success',
  CompilationStatus.error: 'error',
};

const _$ValidationStatusEnumMap = {
  ValidationStatus.notValidated: 'not_validated',
  ValidationStatus.valid: 'valid',
  ValidationStatus.invalid: 'invalid',
};
