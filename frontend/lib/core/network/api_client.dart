import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:app_http_client/application/app_http_exception.dart';
import 'package:app_http_client/application/server_error.dart';

class ApiClient {
  ApiClient(String baseUrl)
      : _dio = Dio(BaseOptions(baseUrl: baseUrl, connectTimeout: const Duration(seconds: 10), receiveTimeout: const Duration(seconds: 20)));
  final Dio _dio;

  void updateBaseUrl(String baseUrl) {
    _dio.options.baseUrl = baseUrl;
  }

  Future<Map<String, dynamic>> getJson(String path, {Map<String, dynamic>? query}) async {
    try {
      final res = await _dio.get(path, queryParameters: query);
      return (res.data is Map<String, dynamic>) ? (res.data as Map<String, dynamic>) : jsonDecode(res.data.toString()) as Map<String, dynamic>;
    } on DioException catch (e) {
      // Пробрасываем как единый формат исключений для верхнего уровня
      final ex = AppHttpException(requestOptions: e.requestOptions, response: e.response);
      // Пытаемся получить стандартный payload
      try {
        final data = e.response?.data;
        final mp = (data is Map<String, dynamic>) ? data : jsonDecode(data.toString()) as Map<String, dynamic>;
        final err = (mp['error'] as Map?)?.cast<String, dynamic>();
        if (err != null) {
          throw AppHttpServerException(
            ex,
            ServerErrorPayload(
              code: (err['code'] ?? '').toString(),
              message: (err['message'] ?? '').toString(),
              details: (err['details'] as Map?)?.cast<String, dynamic>(),
            ),
          );
        }
      } catch (_) {}
      throw ex;
    }
  }

  Future<Response<dynamic>> getRaw(String path, {Map<String, dynamic>? query}) => _dio.get(path, queryParameters: query);
}


