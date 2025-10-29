import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../app_http_client.dart';
import '../app_http_client_token_refresher.dart';
import '../app_http_exception.dart';
import '../http_method.dart';

class DioRefreshTokenInterceptor extends QueuedInterceptor {
  final AppHttpClient _httpClient;
  final AppHttpClientTokenRefresher _refresher;

  DioRefreshTokenInterceptor({
    required AppHttpClient httpClient,
    required AppHttpClientTokenRefresher refresher,
  })  : _httpClient = httpClient,
        _refresher = refresher;

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response == null) {
      return handler.next(err);
    } else if (err.response!.statusCode == 401) {
      if (kDebugMode) {
        print('🔄 Intercepted 401 error - attempting token refresh');
      }

      try {
        await _refresher.refresh(_httpClient);

        if (kDebugMode) {
          print('✅ Token refresh successful - retrying original request');
        }
      } catch (e) {
        if (kDebugMode) {
          print('❌ Token refresh failed: $e');
        }

        if (e is AppHttp401Exception) {
          // Пробрасываем AppHttp401Exception дальше для глобальной обработки
          return handler.reject(DioException(
            requestOptions: err.requestOptions,
            error: e,
            response: err.response,
            type: DioExceptionType.badResponse,
          ));
        } else if (e is DioException) {
          return handler.next(err);
        }

        rethrow;
      }

      try {
        // If refreshed, then retry
        return handler.resolve(await _retry(err.requestOptions));
      } catch (e) {
        if (e is DioException) {
          // Если будет плохая связь, то следующий интерцептор обработает для ретрая
          return handler.next(err);
        }

        rethrow;
      }
    }

    handler.next(err);
  }

  Future<Response<dynamic>> _retry(RequestOptions requestOptions) async {
    // Use new token
    requestOptions.headers[HttpHeaders.authorizationHeader] =
        'Bearer ${_httpClient.token}';

    final withProtocol = requestOptions.uri.host.startsWith('http');
    final host = '${withProtocol ? '' : 'https://'}${requestOptions.uri.host}';

    return _httpClient.performRequest(
      host: host,
      path: requestOptions.uri.path,
      query: requestOptions.queryParameters,
      body: requestOptions.data as Map<String, dynamic>?,
      headers: requestOptions.headers,
      method: HttpMethod.values.byName(requestOptions.method.toLowerCase()),
    );
  }
}
