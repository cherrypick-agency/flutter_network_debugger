import 'dart:convert';

/// Simple heuristic for recognizing GraphQL in request text.
/// Tries to detect both "pure" GraphQL and JSON wrapper of the form
/// { query: "...", variables: {...}, operationName: "..." }.
class GraphqlLanguageDetector {
  /// Returns true if the text is highly likely to contain GraphQL.
  static bool isLikelyGraphql(String source) {
    final trimmed = source.trim();
    if (trimmed.isEmpty) return false;

    // 1) Try to parse as JSON and extract the `query` field
    final queryFromJson = _extractQueryFromJson(trimmed);
    if (queryFromJson != null) {
      return _looksLikeGraphqlQuery(queryFromJson);
    }

    // If this is valid JSON but without `query` field — it's definitely not GraphQL
    try {
      final decoded = json.decode(trimmed);
      if (decoded is Map || decoded is List) {
        return false;
      }
    } catch (_) {
      // not JSON — continue with heuristics
    }

    // 2) Heuristic for "pure" GraphQL
    if (_looksLikeGraphqlQuery(trimmed)) return true;

    return false;
  }

  /// If this is a JSON wrapper for GraphQL, returns the query string.
  static String? extractGraphqlQuery(String source) {
    return _extractQueryFromJson(source.trim());
  }

  /// Which highlight language to choose
  /// Returns 'graphql' if GraphQL is detected, otherwise 'json'.
  static String detectHighlightLanguage(String source) {
    return isLikelyGraphql(source) ? 'graphql' : 'json';
  }

  static String? _extractQueryFromJson(String s) {
    try {
      final decoded = json.decode(s);
      if (decoded is Map<String, dynamic>) {
        final q = decoded['query'];
        if (q is String && q.trim().isNotEmpty) return q;
      }
    } catch (_) {
      // not JSON — ok, continue with heuristics
    }
    return null;
  }

  static bool _looksLikeGraphqlQuery(String s) {
    // GraphQL keywords
    final kw = RegExp(r'\b(query|mutation|subscription|fragment)\b');
    if (kw.hasMatch(s)) return true;

    // presence of {...} block without explicit JSON keys
    final hasBraces = RegExp(r'\{[\s\S]*\}', multiLine: true).hasMatch(s);
    if (hasBraces) {
      // if it looks like JSON keys: "key":, then it's not GraphQL
      final jsonLike = RegExp(r'"\s*[a-zA-Z0-9_]+\s*"\s*:');
      if (!jsonLike.hasMatch(s)) return true;
    }

    // directives @include/@skip are found — also a sign
    final directives = RegExp(r'@[a-zA-Z_][a-zA-Z0-9_]*');
    if (directives.hasMatch(s)) return true;

    return false;
  }
}
