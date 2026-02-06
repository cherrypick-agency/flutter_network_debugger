import 'package:freezed_annotation/freezed_annotation.dart';

part 'decisions.freezed.dart';
part 'decisions.g.dart';

@freezed
sealed class RequestDecision with _$RequestDecision {
  const RequestDecision._();

  const factory RequestDecision({
    required String action,
    String? method,
    String? url,
    Map<String, List<String>>? headers,
    String? bodyBase64,
  }) = _RequestDecision;

  factory RequestDecision.fromJson(Map<String, dynamic> json) =>
      _$RequestDecisionFromJson(json);
}

@freezed
sealed class ResponseDecision with _$ResponseDecision {
  const ResponseDecision._();

  const factory ResponseDecision({
    required String action,
    int? status,
    Map<String, List<String>>? headers,
    String? bodyBase64,
  }) = _ResponseDecision;

  factory ResponseDecision.fromJson(Map<String, dynamic> json) =>
      _$ResponseDecisionFromJson(json);
}
