import 'package:freezed_annotation/freezed_annotation.dart';

part 'match_rules.freezed.dart';
part 'match_rules.g.dart';

/// Pattern matching type
enum PatternType {
  @JsonValue('exact')
  exact, // Exact string match

  @JsonValue('prefix')
  prefix, // Starts with

  @JsonValue('wildcard')
  wildcard, // Simple wildcard with * (e.g., /api/*)

  @JsonValue('regex')
  regex,
}

/// Match rules for filtering when script should execute
@freezed
sealed class MatchRules with _$MatchRules {
  const MatchRules._();

  const factory MatchRules({
    @Default([]) List<String> methods,
    String? pathPattern,
    String? hostPattern,
    @Default(PatternType.wildcard) PatternType patternType,
  }) = _MatchRules;

  factory MatchRules.fromJson(Map<String, dynamic> json) =>
      _$MatchRulesFromJson(json);
}

/// Extension for MatchRules helpers
extension MatchRulesX on MatchRules {
  /// Check if rules are empty (match everything)
  bool get isEmpty =>
      methods.isEmpty &&
      (pathPattern == null || pathPattern!.isEmpty) &&
      (hostPattern == null || hostPattern!.isEmpty);

  /// Get human-readable description
  String get description {
    if (isEmpty) return 'Applies to all requests';

    final parts = <String>[];

    if (methods.isNotEmpty) {
      parts.add('Methods: ${methods.join(', ')}');
    }

    if (pathPattern != null && pathPattern!.isNotEmpty) {
      parts.add('Path: $pathPattern');
    }

    if (hostPattern != null && hostPattern!.isNotEmpty) {
      parts.add('Host: $hostPattern');
    }

    return parts.join(' • ');
  }

  /// Get example URLs that would match
  List<String> get exampleMatches {
    if (isEmpty) return ['Any URL'];

    final examples = <String>[];
    final method = methods.isNotEmpty ? methods.first : 'GET';
    final host = hostPattern ?? 'example.com';

    if (pathPattern != null && pathPattern!.isNotEmpty) {
      if (patternType == PatternType.wildcard && pathPattern!.endsWith('*')) {
        final prefix = pathPattern!.substring(0, pathPattern!.length - 1);
        examples.add('$method https://$host${prefix}test');
        examples.add('$method https://$host${prefix}users/123');
      } else {
        examples.add('$method https://$host$pathPattern');
      }
    } else {
      examples.add('$method https://$host/any/path');
    }

    if (patternType == PatternType.regex && pathPattern != null) {
      examples.add('$method https://$host/matching-path (regex: $pathPattern)');
    }

    return examples;
  }

  /// Get example URLs that would NOT match
  List<String> get exampleNonMatches {
    if (isEmpty) return [];

    final examples = <String>[];

    // Different method
    if (methods.isNotEmpty && methods.length == 1) {
      final otherMethod = methods.first == 'GET' ? 'POST' : 'GET';
      final host = hostPattern ?? 'example.com';
      final path = pathPattern ?? '/api/test';
      examples.add('$otherMethod https://$host$path (method mismatch)');
    }

    // Different path
    if (pathPattern != null && pathPattern!.isNotEmpty) {
      final method = methods.isNotEmpty ? methods.first : 'GET';
      final host = hostPattern ?? 'example.com';

      if (patternType == PatternType.wildcard && pathPattern!.endsWith('*')) {
        final prefix = pathPattern!.substring(0, pathPattern!.length - 1);
        examples.add(
          '$method https://$host/other/path (path doesn\'t start with $prefix)',
        );
      } else if (patternType == PatternType.prefix) {
        examples.add(
          '$method https://$host/other$pathPattern (different prefix)',
        );
      } else if (patternType == PatternType.regex) {
        examples.add('$method https://$host/non-matching (regex mismatch)');
      } else {
        examples.add(
          '$method https://$host$pathPattern/extra (not exact match)',
        );
      }
    }

    return examples;
  }
}
