import 'dart:convert';

import 'package:http/http.dart' as http;

import 'firebase_database_debugger_config.dart';
import 'models.dart';

/// HTTP client for sending debug data to the debugger ingest API.
///
/// Network errors are swallowed — the debug channel must not break the app.
class FirebaseIngestClient {
  /// Creates a client with the given configuration.
  ///
  /// If [httpClient] is not provided, a new `http.Client` is created and will
  /// be closed when [dispose] is called.
  FirebaseIngestClient({
    required FirebaseDatabaseDebuggerConfig config,
    http.Client? httpClient,
  })  : _config = config,
        _http = httpClient ?? http.Client(),
        _ownsHttp = httpClient == null;

  final FirebaseDatabaseDebuggerConfig _config;
  final http.Client _http;
  final bool _ownsHttp;

  /// Sends [request] to the `/_api/v1/ingest/firebase_database` endpoint.
  ///
  /// Timeout is 2 seconds. If the backend is unavailable or returns an error,
  /// the exception is suppressed.
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
      // If UI/backend is unavailable — don't disrupt the app.
      return;
    }
  }

  /// Closes the HTTP client if it was created by this class.
  void dispose() {
    if (_ownsHttp) {
      _http.close();
    }
  }
}
