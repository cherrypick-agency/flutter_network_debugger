import 'package:app_http_client/app_http_client.dart';
import 'package:app_http_client/application/tokens_storage.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../domain/tokens_data.dart';
import 'app_http_client.dart';
import 'app_http_client_token_refresher.dart';
import 'app_http_exception.dart';

class AppHttpClientTokenRefresherImpl implements AppHttpClientTokenRefresher {
  final OnTokenRefreshedType? onTokenRefreshed;

  const AppHttpClientTokenRefresherImpl(this.onTokenRefreshed);

  @override
  Future<void> refresh(AppHttpClient httpClient) async {
    final dioForInternalRequests =
        Dio(BaseOptions(baseUrl: httpClient.defaultHost));

    try {
      final refreshToken = httpClient.refreshToken;
      
      if (kDebugMode) {
        print('🔄 Attempting to refresh token with: $refreshToken');
        print('🔄 Refresh URL: ${httpClient.defaultHost}${httpClient.refreshPath}');
      }
      
      // Запрос на обновление токена
      final response = await dioForInternalRequests.post(
        '${httpClient.defaultHost}${httpClient.refreshPath}',
        data: {
          TokensStorageKeys.refreshToken: refreshToken,
        },
      );
      final data = response.data;

      dioForInternalRequests.close();

      final accessToken = data[TokensStorageKeys.token] as String? ??
          (data['meta']?[TokensStorageKeys.token] as String?);
      final newRefreshToken = data[TokensStorageKeys.refreshToken] as String? ??
          (data['meta']?[TokensStorageKeys.refreshToken] as String?);

      if (kDebugMode) {
        print('✅ Token refresh successful');
      }

      onTokenRefreshed?.call(TokensData(
        accessToken: accessToken,
        refreshToken: newRefreshToken,
      ),);

      await httpClient.rememberTokens(
        accessToken,
        newRefreshToken,
      );
    } catch (e) {
      if (kDebugMode) {
        print('❌ Token refresh failed: $e');
      }
      
      if (e is DioException) {
        final statusCode = e.response?.statusCode;
        
        // НЕ разлогиниваем на 404 (эндпоинт может отсутствовать в окружении)
        // Разлогиниваем только на 401/403      TODO  || statusCode == 404)?
        if (statusCode == 401 || statusCode == 403) {
          if (kDebugMode) {
            print('🚨 Token refresh failed with status $statusCode - clearing tokens');
          }
          
          await httpClient.clearTokens();

          throw AppHttp401Exception(AppHttpException(
            requestOptions: e.requestOptions,
            error: e.error,
            response: e.response,
          ),);
        }
      }

      // Пробрасываем дальше для обработки на уровне запроса/UX
      rethrow;
    }
  }
}
