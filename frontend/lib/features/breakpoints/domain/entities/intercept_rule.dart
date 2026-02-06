import 'package:freezed_annotation/freezed_annotation.dart';

part 'intercept_rule.freezed.dart';
part 'intercept_rule.g.dart';

@freezed
sealed class RuleStringMatch with _$RuleStringMatch {
  const RuleStringMatch._();

  const factory RuleStringMatch({
    @Default('') String equals,
    @Default('') String prefix,
    @Default('') String suffix,
    @Default('') String contains,
    @Default([]) List<String> anyOf,
    @Default('') String regex,
  }) = _RuleStringMatch;

  factory RuleStringMatch.fromJson(Map<String, dynamic> json) =>
      _$RuleStringMatchFromJson(json);

  bool get isEmpty =>
      equals.isEmpty &&
      prefix.isEmpty &&
      suffix.isEmpty &&
      contains.isEmpty &&
      anyOf.isEmpty &&
      regex.isEmpty;
}

@freezed
sealed class RuleHeaderMatch with _$RuleHeaderMatch {
  const RuleHeaderMatch._();

  const factory RuleHeaderMatch({
    @Default(RuleStringMatch()) RuleStringMatch name,
    @Default(RuleStringMatch()) RuleStringMatch value,
  }) = _RuleHeaderMatch;

  factory RuleHeaderMatch.fromJson(Map<String, dynamic> json) =>
      _$RuleHeaderMatchFromJson(json);

  bool get isEmpty => name.isEmpty && value.isEmpty;
}

@freezed
sealed class RuleStatusMatch with _$RuleStatusMatch {
  const RuleStatusMatch._();

  const factory RuleStatusMatch({
    @Default([]) @JsonKey(name: 'equals') List<int> statusEquals,
    @Default(0) int from,
    @Default(0) int to,
    @Default(false) bool is4xx,
    @Default(false) bool is5xx,
  }) = _RuleStatusMatch;

  factory RuleStatusMatch.fromJson(Map<String, dynamic> json) =>
      _$RuleStatusMatchFromJson(json);

  bool get isEmpty =>
      statusEquals.isEmpty && from == 0 && to == 0 && !is4xx && !is5xx;
}

@freezed
sealed class InterceptWhen with _$InterceptWhen {
  const InterceptWhen._();

  const factory InterceptWhen({
    @Default([]) List<String> method,
    @Default([]) List<String> scheme,
    @JsonKey(includeIfNull: false) RuleStringMatch? host,
    @JsonKey(includeIfNull: false) RuleStringMatch? port,
    @JsonKey(includeIfNull: false) RuleStringMatch? path,
    @JsonKey(includeIfNull: false) RuleStringMatch? contentType,
    @JsonKey(name: 'responseStatus', includeIfNull: false)
    RuleStatusMatch? responseStatus,
    @JsonKey(includeIfNull: false) RuleHeaderMatch? header,
    @JsonKey(includeIfNull: false) String? bodyContains,
  }) = _InterceptWhen;

  factory InterceptWhen.fromJson(Map<String, dynamic> json) =>
      _$InterceptWhenFromJson(json);
}

@freezed
sealed class InterceptRule with _$InterceptRule {
  const InterceptRule._();

  const factory InterceptRule({
    @Default('') String id,
    @Default(true) bool enabled,
    @Default(10) int priority,
    @Default('both') String action,
    @Default(false) bool once,
    @Default(false) bool stopProcessing,
    @Default(InterceptWhen()) InterceptWhen when,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _InterceptRule;

  factory InterceptRule.fromJson(Map<String, dynamic> json) =>
      _$InterceptRuleFromJson(json);
}
