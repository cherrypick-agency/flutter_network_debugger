import 'dart:convert';

/// Firebase RTDB событие, распарсенное из preview JSON.
class FirebaseEventData {
  final String op;
  final String path;
  final bool ok;
  final String? error;
  final String? query;
  final dynamic payload;

  const FirebaseEventData({
    required this.op,
    required this.path,
    required this.ok,
    this.error,
    this.query,
    this.payload,
  });

  /// Пробуем распарсить preview как Firebase event.
  /// Возвращает null если это не Firebase формат.
  static FirebaseEventData? tryParse(String preview) {
    final t = preview.trim();
    if (t.isEmpty) return null;
    if (!(t.startsWith('{') || t.startsWith('['))) return null;
    try {
      final decoded = jsonDecode(t);
      if (decoded is! Map<String, dynamic>) return null;
      if (decoded['type'] != 'firebase_database') return null;
      return FirebaseEventData(
        op: (decoded['op'] ?? '').toString(),
        path: (decoded['path'] ?? '').toString(),
        ok: decoded['ok'] == true,
        error: decoded['error']?.toString(),
        query: decoded['query']?.toString(),
        payload: decoded['payload'],
      );
    } catch (_) {
      return null;
    }
  }
}
