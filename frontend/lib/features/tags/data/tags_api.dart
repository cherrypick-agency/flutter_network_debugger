import 'package:app_http_client/application/app_http_client.dart';
import 'package:app_http_client/application/app_http_exception.dart';
import 'package:app_http_client/application/server_error.dart';

import '../../../services/prefs.dart';

class TagsApi {
  TagsApi(this._http);
  final AppHttpClient _http;

  Future<Map<String, String>> _hdrs() async {
    final token = await PrefsService().loadAdminToken();
    return token.isEmpty ? const {} : <String, String>{'X-Admin-Token': token};
  }

  // Predefined Tags
  Future<List<dynamic>> listPredefinedTags() async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/tags/predefined',
      headers: await _hdrs(),
    );
    return (resp.data?['items'] as List<dynamic>?) ?? <dynamic>[];
  }

  Future<void> createPredefinedTag(Map<String, dynamic> body) async {
    await _http.post<void>(
      path: '/_api/v1/tags/predefined',
      body: body,
      headers: await _hdrs(),
    );
  }

  Future<void> deletePredefinedTag(String id) async {
    await _http.delete(
      path: '/_api/v1/tags/predefined/$id',
      headers: await _hdrs(),
    );
  }

  // Session Tags
  Future<List<dynamic>> getSessionTags(String sessionId) async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/sessions/$sessionId/tags',
      headers: await _hdrs(),
    );
    return (resp.data?['items'] as List<dynamic>?) ?? <dynamic>[];
  }

  Future<void> addSessionTag(String sessionId, String tagName) async {
    try {
      await _http.post<void>(
        path: '/_api/v1/sessions/$sessionId/tags',
        body: {'tagName': tagName},
        headers: await _hdrs(),
      );
    } on AppHttpServerException catch (e) {
      // Fallback for environments without session-endpoint: use bulk API
      if (e.code == ServerErrorCode.notFound) {
        await bulkTags({
          'operation': 'add',
          'sessionIds': [sessionId],
          'tagNames': [tagName],
        });
        return;
      }
      rethrow;
    } on AppHttpException catch (e) {
      if ((e.response?.statusCode ?? 0) == 404) {
        await bulkTags({
          'operation': 'add',
          'sessionIds': [sessionId],
          'tagNames': [tagName],
        });
        return;
      }
      rethrow;
    }
  }

  Future<void> removeSessionTag(String sessionId, String tagName) async {
    try {
      await _http.delete(
        path: '/_api/v1/sessions/$sessionId/tags/$tagName',
        headers: await _hdrs(),
      );
    } on AppHttpServerException catch (e) {
      if (e.code == ServerErrorCode.notFound) {
        await bulkTags({
          'operation': 'remove',
          'sessionIds': [sessionId],
          'tagNames': [tagName],
        });
        return;
      }
      rethrow;
    } on AppHttpException catch (e) {
      if ((e.response?.statusCode ?? 0) == 404) {
        await bulkTags({
          'operation': 'remove',
          'sessionIds': [sessionId],
          'tagNames': [tagName],
        });
        return;
      }
      rethrow;
    }
  }

  Future<void> bulkTags(Map<String, dynamic> body) async {
    await _http.post<void>(
      path: '/_api/v1/tags/bulk',
      body: body,
      headers: await _hdrs(),
    );
  }

  // Session Annotations
  Future<List<dynamic>> getSessionAnnotations(String sessionId) async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/sessions/$sessionId/annotations',
      headers: await _hdrs(),
    );
    return (resp.data?['items'] as List<dynamic>?) ?? <dynamic>[];
  }

  Future<void> upsertSessionAnnotation(
    String sessionId,
    String key,
    String value,
  ) async {
    await _http.post<void>(
      path: '/_api/v1/sessions/$sessionId/annotations',
      body: {'key': key, 'value': value},
      headers: await _hdrs(),
    );
  }

  Future<void> deleteSessionAnnotation(String sessionId, String key) async {
    await _http.delete(
      path: '/_api/v1/sessions/$sessionId/annotations/$key',
      headers: await _hdrs(),
    );
  }
}
