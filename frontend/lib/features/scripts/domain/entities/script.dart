import 'package:freezed_annotation/freezed_annotation.dart';
import 'match_rules.dart';
import 'script_config.dart';

part 'script.freezed.dart';
part 'script.g.dart';

/// Script runtime type
enum ScriptRuntime {
  @JsonValue('extism')
  extism, // WASM runtime (Rust, Go, JavaScript)

  @JsonValue('dart')
  dart, // Dart subprocess runtime
}

/// Script trigger type - when the script should execute
enum TriggerType {
  @JsonValue('request')
  request, // Before proxying request

  @JsonValue('response')
  response, // After receiving response

  @JsonValue('both')
  both, // Both request and response
}

/// Main Script entity
/// Represents a script that can modify HTTP requests/responses
@freezed
sealed class Script with _$Script {
  const Script._();

  const factory Script({
    required String id,
    required String name,
    String? description,
    required ScriptRuntime runtime,
    required String code,
    required String language,
    required TriggerType triggerType,
    @Default(10) int priority,
    @Default(true) bool enabled,
    MatchRules? matchRules,
    ScriptConfig? config,
    DateTime? createdAt,
    DateTime? updatedAt,
    // Compilation fields
    String? sourceCode,
    Map<String, String>? dependencies,
    String? compilationStatus,
    String? compilationError,
    DateTime? lastCompiledAt,
  }) = _Script;

  factory Script.fromJson(Map<String, dynamic> json) => _$ScriptFromJson(json);
}

/// Extension for Script business logic
extension ScriptX on Script {
  /// Check if script matches a request
  bool matchesRequest({
    required String method,
    required String url,
    String? host,
  }) {
    if (matchRules == null) return true; // No rules = match everything

    // Check HTTP method
    if (matchRules!.methods.isNotEmpty) {
      if (!matchRules!.methods.contains(method.toUpperCase())) {
        return false;
      }
    }

    // Check path pattern
    if (matchRules!.pathPattern != null &&
        matchRules!.pathPattern!.isNotEmpty) {
      if (!_matchesPattern(
        url,
        matchRules!.pathPattern!,
        matchRules!.patternType,
      )) {
        return false;
      }
    }

    // Check host pattern
    if (matchRules!.hostPattern != null &&
        matchRules!.hostPattern!.isNotEmpty &&
        host != null) {
      if (!_matchesPattern(
        host,
        matchRules!.hostPattern!,
        matchRules!.patternType,
      )) {
        return false;
      }
    }

    return true;
  }

  bool _matchesPattern(String value, String pattern, PatternType type) {
    switch (type) {
      case PatternType.exact:
        return value == pattern;
      case PatternType.prefix:
        return value.startsWith(pattern);
      case PatternType.wildcard:
        // Simple wildcard: /api/* matches /api/test, /api/users, etc.
        if (pattern.endsWith('*')) {
          final prefix = pattern.substring(0, pattern.length - 1);
          return value.startsWith(prefix);
        }
        return value == pattern;
    }
  }

  /// Get display name for runtime
  String get runtimeDisplay {
    switch (runtime) {
      case ScriptRuntime.extism:
        return 'Extism (WASM)';
      case ScriptRuntime.dart:
        return 'Dart Subprocess';
    }
  }

  /// Get display name for language with icon
  String get languageDisplay {
    switch (language.toLowerCase()) {
      case 'rust':
        return '🦀 Rust';
      case 'javascript':
      case 'js':
        return '🟨 JavaScript';
      case 'typescript':
      case 'ts':
        return '🔷 TypeScript';
      case 'go':
        return '🐹 Go';
      case 'dart':
        return '🎯 Dart';
      case 'python':
      case 'py':
        return '🐍 Python';
      default:
        return language;
    }
  }

  /// Check if script is a request interceptor
  bool get isRequestInterceptor =>
      triggerType == TriggerType.request || triggerType == TriggerType.both;

  /// Check if script is a response interceptor
  bool get isResponseInterceptor =>
      triggerType == TriggerType.response || triggerType == TriggerType.both;
}
