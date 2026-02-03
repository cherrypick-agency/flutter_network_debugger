/// Утилиты для безопасного приведения значений из JSON.
///
/// В реальных ответах бэка часто встречаются "плавающие" типы (0/1 вместо bool,
/// числа как строки и т.п.). Эти методы помогают не падать на `as T`.
class JsonCast {
  static bool asBool(dynamic v, {required bool fallback}) {
    if (v is bool) return v;
    if (v is num) return v != 0;
    if (v is String) {
      final s = v.trim().toLowerCase();
      if (s == 'true' || s == '1' || s == 'yes' || s == 'y') return true;
      if (s == 'false' || s == '0' || s == 'no' || s == 'n') return false;
    }
    return fallback;
  }

  static int asInt(dynamic v, {required int fallback}) {
    if (v is int) return v;
    if (v is num) return v.toInt();
    if (v is String) return int.tryParse(v.trim()) ?? fallback;
    return fallback;
  }

  static String asString(dynamic v, {required String fallback}) {
    if (v == null) return fallback;
    if (v is String) return v;
    return v.toString();
  }

  static String? asTrimmedStringOrNull(dynamic v) {
    if (v == null) return null;
    final s = v.toString().trim();
    return s.isEmpty ? null : s;
  }

  static List<String> asStringList(dynamic v) {
    if (v is List) {
      return v
          .map((e) => (e ?? '').toString().trim())
          .where((e) => e.isNotEmpty)
          .toList(growable: false);
    }
    if (v is String) {
      final s = v.trim();
      return s.isEmpty ? const <String>[] : <String>[s];
    }
    return const <String>[];
  }

  static Map<String, dynamic> asMap(dynamic v) {
    if (v is Map<String, dynamic>) return v;
    if (v is Map) {
      return v.map((k, val) => MapEntry(k.toString(), val));
    }
    return const <String, dynamic>{};
  }

  static List<dynamic> asList(dynamic v) {
    if (v is List) return v;
    return const <dynamic>[];
  }
}
