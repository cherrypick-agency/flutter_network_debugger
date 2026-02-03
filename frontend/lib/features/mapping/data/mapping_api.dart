import 'dart:typed_data';
import 'package:app_http_client/application/app_http_client.dart';
import 'package:dio/dio.dart';
import '../../../services/prefs.dart';

class MappingApi {
  MappingApi(this._http);
  final AppHttpClient _http;

  Future<Map<String, String>?> _hdrs() async {
    try {
      final token = await PrefsService().loadAdminToken();
      if (token.isNotEmpty) return <String, String>{'X-Admin-Token': token};
    } catch (_) {}
    return null;
  }

  Future<Map<String, dynamic>> getConfig() async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/mapping/config',
      headers: await _hdrs(),
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<void> setConfig(Map<String, dynamic> body) async {
    await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/mapping/config',
      body: body,
      headers: await _hdrs(),
    );
  }

  Future<List<dynamic>> listRules() async {
    final resp = await _http.get<List<dynamic>>(
      path: '/_api/v1/mapping/rules',
      headers: await _hdrs(),
    );
    return resp.data ?? <dynamic>[];
  }

  Future<Map<String, dynamic>> upsertRule(Map<String, dynamic> body) async {
    final resp = await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/mapping/rules',
      body: body,
      headers: await _hdrs(),
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<void> reorder(List<String> ids) async {
    // Эндпоинт ждёт JSON-массив. AppHttpClient спокойно умеет отправлять List.
    await _http.post<void>(
      path: '/_api/v1/mapping/rules/reorder',
      body: ids,
      headers: await _hdrs(),
    );
  }

  Future<void> deleteRule(String id) async {
    await _http.delete(
      path: '/_api/v1/mapping/rules/$id',
      headers: await _hdrs(),
    );
  }

  Future<Map<String, dynamic>> upload(String fileName, List<int> bytes) async {
    return uploadBytes(fileName, bytes);
  }

  Future<Map<String, dynamic>> uploadBytes(
    String fileName,
    List<int> bytes,
  ) async {
    final headers = await _hdrs() ?? const <String, String>{};
    final form = FormData.fromMap({
      'file': MultipartFile.fromBytes(
        Uint8List.fromList(bytes),
        filename: fileName,
      ),
    });
    final resp = await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/mapping/upload',
      body: form,
      headers: headers,
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<Map<String, dynamic>> uploadStream({
    required String fileName,
    required Stream<List<int>> stream,
    required int length,
  }) async {
    final headers = await _hdrs() ?? const <String, String>{};
    final form = FormData.fromMap({
      'file': MultipartFile.fromStream(
        () => stream,
        length,
        filename: fileName,
      ),
    });
    final resp = await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/mapping/upload',
      body: form,
      headers: headers,
    );
    return resp.data ?? <String, dynamic>{};
  }
}
