import 'dart:convert';

import 'package:http/http.dart' as http;

import 'firebase_database_debugger_config.dart';
import 'models.dart';

/// HTTP-клиент для отправки отладочных данных в ingest API дебаггера.
///
/// Все ошибки сети глотаются — отладочный канал не должен ломать приложение.
class FirebaseIngestClient {
  /// Создаёт клиент с заданной конфигурацией.
  ///
  /// Если [httpClient] не передан, создаётся свой экземпляр `http.Client`,
  /// который будет закрыт при вызове [dispose].
  FirebaseIngestClient({
    required FirebaseDatabaseDebuggerConfig config,
    http.Client? httpClient,
  })  : _config = config,
        _http = httpClient ?? http.Client(),
        _ownsHttp = httpClient == null;

  final FirebaseDatabaseDebuggerConfig _config;
  final http.Client _http;
  final bool _ownsHttp;

  /// Отправляет [request] на endpoint `/_api/v1/ingest/firebase_database`.
  ///
  /// Таймаут — 2 секунды. Если бэкенд недоступен или вернул ошибку,
  /// исключение подавляется.
  Future<void> ingest(FirebaseIngestRequest request) async {
    final base = _config.debuggerBaseUrl.trim();
    if (base.isEmpty) return;
    try {
      final uri = Uri.parse(base).replace(
        path: '/_api/v1/ingest/firebase_database',
        query: null,
        queryParameters: null,
      );
      final headers = <String, String>{
        'content-type': 'application/json',
        if ((_config.adminToken ?? '').trim().isNotEmpty)
          'X-Admin-Token': _config.adminToken!.trim(),
      };
      final response = await _http
          .post(
            uri,
            headers: headers,
            body: jsonEncode(request.toJson()),
          )
          .timeout(const Duration(seconds: 2));
      if (response.statusCode == 204) return;
    } catch (_) {
      // Если UI/бэкенд недоступен — не мешаем приложению.
      return;
    }
  }

  /// Закрывает HTTP-клиент, если он был создан внутри этого класса.
  void dispose() {
    if (_ownsHttp) {
      _http.close();
    }
  }
}
