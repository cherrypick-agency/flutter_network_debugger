import 'package:app_http_client/application/app_http_client.dart';

import '../domain/models.dart';

class ComposeRepository {
  final AppHttpClient _http;
  ComposeRepository(this._http);

  Future<Map<String, dynamic>> send(ComposeTemplateDTO tpl) async {
    final resp = await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/compose/send',
      body: {'template': tpl.toJson()},
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> library() async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/compose/library',
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> config() async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/compose/config',
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> upsertRequest(ComposeTemplateDTO tpl) async {
    final resp = await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/compose/library/requests',
      body: tpl.toJson(),
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> deleteRequest(String id) async {
    final resp = await _http.delete<Map<String, dynamic>>(
      path: '/_api/v1/compose/library/requests/$id',
      body: const <String, dynamic>{},
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> history({int limit = 20}) async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/compose/history',
      query: {'limit': '$limit'},
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<void> moveRequest({
    required String id,
    required String collectionId,
    required String folderId,
    int index = -1,
  }) async {
    await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/compose/library/requests_move/$id',
      body: {
        'collectionId': collectionId,
        'folderId': folderId,
        'index': index,
      },
    );
  }

  Future<Map<String, dynamic>> upsertCollection(
    Map<String, dynamic> collection,
  ) async {
    final resp = await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/compose/library/collections',
      body: collection,
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> deleteCollection(String id) async {
    final resp = await _http.delete<Map<String, dynamic>>(
      path: '/_api/v1/compose/library/collections/$id',
      body: const <String, dynamic>{},
    );
    return resp.data ?? <String, dynamic>{};
  }
}
