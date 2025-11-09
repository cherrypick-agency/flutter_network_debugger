import 'package:app_http_client/application/app_http_client.dart';

class PerformanceApi {
  PerformanceApi(this._http);
  final AppHttpClient _http;

  Future<Map<String, dynamic>> getOverview(DateTime from, DateTime to) async {
    final resp = await _http.get<Map<String, dynamic>>(
      path: '/_api/v1/performance/overview',
      query: {'from': from.toIso8601String(), 'to': to.toIso8601String()},
    );
    return resp.data ?? {};
  }
}
