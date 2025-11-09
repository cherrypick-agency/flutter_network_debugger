import 'package:freezed_annotation/freezed_annotation.dart';

part 'script_test_result.freezed.dart';
part 'script_test_result.g.dart';

/// Test request for script testing
@freezed
sealed class TestRequest with _$TestRequest {
  const TestRequest._();

  const factory TestRequest({
    required String method,
    required String url,
    @Default({}) Map<String, List<String>> headers,
    String? body,
  }) = _TestRequest;

  factory TestRequest.fromJson(Map<String, dynamic> json) =>
      _$TestRequestFromJson(json);
}

/// Modified HTTP request/response from script execution
@freezed
sealed class ModifiedHTTP with _$ModifiedHTTP {
  const ModifiedHTTP._();

  const factory ModifiedHTTP({
    String? method,
    String? url,
    Map<String, List<String>>? headers,
    String? body,
    int? status,
  }) = _ModifiedHTTP;

  factory ModifiedHTTP.fromJson(Map<String, dynamic> json) =>
      _$ModifiedHTTPFromJson(json);
}

/// Result of script test execution
@freezed
sealed class ScriptTestResult with _$ScriptTestResult {
  const ScriptTestResult._();

  const factory ScriptTestResult({
    required bool success,
    String? error,
    ModifiedHTTP? modifiedRequest,
    ModifiedHTTP? modifiedResponse,
    @Default([]) List<String> logs,
    int? durationMs,
  }) = _ScriptTestResult;

  factory ScriptTestResult.fromJson(Map<String, dynamic> json) =>
      _$ScriptTestResultFromJson(json);
}

/// Extension for test result helpers
extension ScriptTestResultX on ScriptTestResult {
  /// Check if there were any modifications
  bool get hasModifications =>
      modifiedRequest != null || modifiedResponse != null;

  /// Get duration in human-readable format
  String get durationFormatted {
    if (durationMs == null) return 'N/A';
    if (durationMs! < 1000) return '${durationMs}ms';
    return '${(durationMs! / 1000).toStringAsFixed(2)}s';
  }

  /// Get status text with emoji
  String get statusText {
    if (success) {
      return hasModifications
          ? '✅ Success (modified)'
          : '✅ Success (no changes)';
    }
    return '❌ Failed';
  }

  /// Get all logs as single string
  String get logsFormatted => logs.join('\n');
}
