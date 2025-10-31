import 'dart:convert';

/// Простая эвристика для распознавания GraphQL в тексте запроса.
/// Стараемся угадать как "чистый" GraphQL, так и JSON-обёртку вида
/// { query: "...", variables: {...}, operationName: "..." }.
class GraphqlLanguageDetector {
  /// Возвращает true, если текст с высокой вероятностью содержит GraphQL.
  static bool isLikelyGraphql(String source) {
    final trimmed = source.trim();
    if (trimmed.isEmpty) return false;

    // 1) Попробуем разобрать как JSON и достать поле `query`
    final queryFromJson = _extractQueryFromJson(trimmed);
    if (queryFromJson != null) {
      return _looksLikeGraphqlQuery(queryFromJson);
    }

    // Если это валидный JSON, но без поля `query` — это точно не GraphQL
    try {
      final decoded = json.decode(trimmed);
      if (decoded is Map || decoded is List) {
        return false;
      }
    } catch (_) {
      // не JSON — продолжаем эвристику
    }

    // 2) Эвристика для "чистого" GraphQL
    if (_looksLikeGraphqlQuery(trimmed)) return true;

    return false;
  }

  /// Если это JSON-обёртка GraphQL, вернёт строку запроса.
  static String? extractGraphqlQuery(String source) {
    return _extractQueryFromJson(source.trim());
  }

  /// Какой язык подсветки выбрать для highlight
  /// Возвращает 'graphql' если угадывается GraphQL, иначе 'json'.
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
      // не JSON — ок, идём дальше по эвристикам
    }
    return null;
  }

  static bool _looksLikeGraphqlQuery(String s) {
    // ключевые слова GraphQL
    final kw = RegExp(r'\b(query|mutation|subscription|fragment)\b');
    if (kw.hasMatch(s)) return true;

    // наличие блока {...} без явных JSON-ключей
    final hasBraces = RegExp(r'\{[\s\S]*\}', multiLine: true).hasMatch(s);
    if (hasBraces) {
      // если похоже на JSON-ключи: "key":, то это не GraphQL
      final jsonLike = RegExp(r'"\s*[a-zA-Z0-9_]+\s*"\s*:');
      if (!jsonLike.hasMatch(s)) return true;
    }

    // встречаются директивы @include/@skip — тоже признак
    final directives = RegExp(r'@[a-zA-Z_][a-zA-Z0-9_]*');
    if (directives.hasMatch(s)) return true;

    return false;
  }
}
