import 'package:freezed_annotation/freezed_annotation.dart';

part 'script_example.freezed.dart';
part 'script_example.g.dart';

/// Script example difficulty level
enum ExampleDifficulty {
  @JsonValue('beginner')
  beginner,

  @JsonValue('intermediate')
  intermediate,

  @JsonValue('advanced')
  advanced,
}

/// Script example category
enum ExampleCategory {
  @JsonValue('request_modification')
  requestModification,

  @JsonValue('response_modification')
  responseModification,

  @JsonValue('both')
  both,
}

/// Script Example entity
/// Represents a pre-built example script from the library
@freezed
sealed class ScriptExample with _$ScriptExample {
  const ScriptExample._();

  const factory ScriptExample({
    required String id,
    required String name,
    required String description,
    required String language,
    required ExampleDifficulty difficulty,
    required ExampleCategory category,
    required String triggerType,
    required String sourceCode,
  }) = _ScriptExample;

  factory ScriptExample.fromJson(Map<String, dynamic> json) =>
      _$ScriptExampleFromJson(json);
}
