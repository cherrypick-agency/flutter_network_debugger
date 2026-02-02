import 'package:app_http_client/application/app_http_client.dart';
import '../../../services/prefs.dart';

class BreakpointsApi {
  BreakpointsApi(this._http);
  final AppHttpClient _http;

  Future<Map<String, dynamic>> getConfig() async {
    final hdrs = await _hdrs();
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/intercept/config',
      headers: hdrs,
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<void> setConfig(Map<String, dynamic> body) async {
    final hdrs = await _hdrs();
    await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/intercept/config',
      body: body,
      headers: hdrs,
    );
  }

  Future<List<dynamic>> listRules() async {
    final hdrs = await _hdrs();
    final resp = await _http.get<List<dynamic>>(
      path: '/_api/v1/intercept/rules',
      headers: hdrs,
    );
    return resp.data ?? <dynamic>[];
  }

  Future<void> replaceRules(List<dynamic> rules) async {
    final hdrs = await _hdrs();
    // Бэк ожидает массив правил, а не объект.
    await _http.post<void>(
      path: '/_api/v1/intercept/rules',
      headers: hdrs,
      body: rules,
    );
  }

  Future<List<dynamic>> listPending({int? limit}) async {
    final hdrs = await _hdrs();
    final resp = await _http.get<List<dynamic>>(
      path: '/_api/v1/intercept/pending',
      query: {if (limit != null) 'limit': limit.toString()},
      headers: hdrs,
    );
    return resp.data ?? <dynamic>[];
  }

  Future<Map<String, dynamic>> getItem(String id) async {
    final hdrs = await _hdrs();
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/intercept/items/$id',
      headers: hdrs,
    );
    return resp.data ?? <String, dynamic>{};
  }

  Future<void> continueItem(String id, Map<String, dynamic> body) async {
    final hdrs = await _hdrs();
    await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/intercept/items/$id/continue',
      body: body,
      headers: hdrs,
    );
  }

  Future<void> cancel(String id) async {
    final hdrs = await _hdrs();
    await _http.post<Map<String, dynamic>>(
      path: '/_api/v1/intercept/items/$id/cancel',
      body: <String, dynamic>{},
      headers: hdrs,
    );
  }

  Future<Map<String, String>?> _hdrs() async {
    try {
      final tok = await PrefsService().loadAdminToken();
      if (tok.isNotEmpty) return {'X-Admin-Token': tok};
    } catch (_) {}
    return null;
  }
}
