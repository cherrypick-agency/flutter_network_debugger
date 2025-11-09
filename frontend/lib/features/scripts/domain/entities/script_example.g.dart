// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'script_example.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_ScriptExample _$ScriptExampleFromJson(Map<String, dynamic> json) =>
    _ScriptExample(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String,
      language: json['language'] as String,
      difficulty: $enumDecode(_$ExampleDifficultyEnumMap, json['difficulty']),
      category: $enumDecode(_$ExampleCategoryEnumMap, json['category']),
      triggerType: json['triggerType'] as String,
      sourceCode: json['sourceCode'] as String,
    );

Map<String, dynamic> _$ScriptExampleToJson(_ScriptExample instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'description': instance.description,
      'language': instance.language,
      'difficulty': _$ExampleDifficultyEnumMap[instance.difficulty]!,
      'category': _$ExampleCategoryEnumMap[instance.category]!,
      'triggerType': instance.triggerType,
      'sourceCode': instance.sourceCode,
    };

const _$ExampleDifficultyEnumMap = {
  ExampleDifficulty.beginner: 'beginner',
  ExampleDifficulty.intermediate: 'intermediate',
  ExampleDifficulty.advanced: 'advanced',
};

const _$ExampleCategoryEnumMap = {
  ExampleCategory.requestModification: 'request_modification',
  ExampleCategory.responseModification: 'response_modification',
  ExampleCategory.both: 'both',
};
