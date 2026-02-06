import 'package:freezed_annotation/freezed_annotation.dart';

part 'intercept_config.freezed.dart';
part 'intercept_config.g.dart';

@freezed
sealed class InterceptConfig with _$InterceptConfig {
  const InterceptConfig._();

  const factory InterceptConfig({
    @Default(false) bool enabled,
    @Default(true) bool requests,
    @Default(true) bool responses,
    @Default(60000) int timeoutMs,
    @Default(200) int queueMax,
    @Default(1048576) int bodyMaxBytes,
    @Default(true) bool reencode,
    @Default('auto-continue-oldest') String overflow,
  }) = _InterceptConfig;

  factory InterceptConfig.fromJson(Map<String, dynamic> json) =>
      _$InterceptConfigFromJson(json);
}
