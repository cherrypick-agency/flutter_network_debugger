import 'package:freezed_annotation/freezed_annotation.dart';

part 'intercept_item.freezed.dart';
part 'intercept_item.g.dart';

@freezed
sealed class HTTPRequestSnapshot with _$HTTPRequestSnapshot {
  const HTTPRequestSnapshot._();

  const factory HTTPRequestSnapshot({
    @Default('') String method,
    @Default('') String url,
    @Default({}) Map<String, List<String>> headers,
    String? bodyBase64,
    @Default(false) bool bodyTruncated,
    String? contentType,
  }) = _HTTPRequestSnapshot;

  factory HTTPRequestSnapshot.fromJson(Map<String, dynamic> json) =>
      _$HTTPRequestSnapshotFromJson(json);
}

@freezed
sealed class HTTPResponseSnapshot with _$HTTPResponseSnapshot {
  const HTTPResponseSnapshot._();

  const factory HTTPResponseSnapshot({
    @Default(0) int status,
    @Default({}) Map<String, List<String>> headers,
    String? bodyBase64,
    @Default(false) bool bodyTruncated,
    String? contentType,
  }) = _HTTPResponseSnapshot;

  factory HTTPResponseSnapshot.fromJson(Map<String, dynamic> json) =>
      _$HTTPResponseSnapshotFromJson(json);
}

@freezed
sealed class InterceptItem with _$InterceptItem {
  const InterceptItem._();

  const factory InterceptItem({
    required String id,
    required DateTime createdAt,
    required DateTime deadline,
    required String direction,
    required String sessionId,
    @Default('PENDING') String state,
    String? ruleId,
    HTTPRequestSnapshot? req,
    HTTPResponseSnapshot? res,
  }) = _InterceptItem;

  factory InterceptItem.fromJson(Map<String, dynamic> json) =>
      _$InterceptItemFromJson(json);
}
