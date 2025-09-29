import 'dart:async';
import 'dart:io';
import 'dart:convert';

import '../../app_http_client.dart';
import 'package:dio/dio.dart';

import '../app_http_client.dart';
import '../app_http_client_token_refresher.dart';
import '../app_http_client_token_refresher_impl.dart';
import '../app_http_exception.dart';
import '../server_error.dart';
import '../http_method.dart';
import '../tokens_storage.dart';
import 'dio_exception_adapter.dart';
import 'dio_refresh_token_interceptor.dart';

class DioHttpClientImpl implements AppHttpClient {
  final TokensStorage _tokensStorage;
  final BaseUrlType _defaultHost;
  final Dio _dio;
  final String _refreshPath;
  final DioExceptionAdapter _exceptionAdapter;
  final AppHttpClientTokenRefresher _tokenRefresher;
  final Transformer? _transformer;
  final Iterable<Interceptor> _interceptors;

  DioHttpClientImpl({
    required BaseUrlType defaultHost,
    Dio? dio,
    TokensStorage? tokensStorage,
    DioExceptionAdapter? exceptionAdapter,
    AppHttpClientTokenRefresher? tokenRefresher,
    OnTokenRefreshedType? onTokenRefreshed,
    Transformer? transformer,
    Iterable<Interceptor> interceptors = const Iterable.empty(),
    String refreshPath = '/refresh_token',
  })  : _defaultHost = defaultHost,
        _dio = dio ??
            Dio(
              BaseOptions(
                  // validateStatus: (int? status) {
                  //   // Fix https://github.com/flutterchina/dio/issues/995#issuecomment-739902537
                  //   return status != null && status >= 100 && status <= 400;
                  // },
                  ),
            ),
        _transformer = transformer,
        _interceptors = interceptors,
        _tokensStorage = tokensStorage ?? TokensStorageImpl(),
        _exceptionAdapter = exceptionAdapter ?? const DioExceptionAdapter(),
        _tokenRefresher =
            tokenRefresher ?? AppHttpClientTokenRefresherImpl(onTokenRefreshed),
        _refreshPath = refreshPath;

  @override
  String get defaultHost => _defaultHost();

  @override
  String? get token => _tokensStorage.token;

  @override
  String? get refreshToken => _tokensStorage.refreshToken;

  @override
  String get refreshPath => _refreshPath;

  @override
  Future<void> init() async {
    // See https://stackoverflow.com/a/62911616/5286034
    final transformer = _transformer;
    if (transformer != null) _dio.transformer = transformer;
    _dio.interceptors.addAll([
      DioRefreshTokenInterceptor(
        httpClient: this,
        refresher: _tokenRefresher,
      ),
      // RetryInterceptor(dio: _dio),
      ..._interceptors,
    ]);
    // ..httpClientAdapter = Http2Adapter(
    //   ConnectionManager(
    //     idleTimeout: 10000,
    //     // Ignore bad certificate
    //     onClientCreate: (_, config) => config.onBadCertificate = (_) => true,
    //   ),
    // );
    await _tokensStorage.init();
  }

  @override
  Future<void> refresh() => _tokenRefresher.refresh(this);

  // @throws AppHttpException if http error
  @override
  Future<Response<T>> performRequest<T>({
    String? host,
    String path = '',
    Map<String, dynamic>? query,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? body,
    required HttpMethod method,
    Map<String, String>? fields,
  }) async {
    final requestHost = host ?? defaultHost;
    final haveContentTypeHeader = headers?.keys.any(
          (key) =>
              key.toLowerCase() == HttpHeaders.contentTypeHeader.toLowerCase(),
        ) ??
        false;
    final internalHeaders = <String, dynamic>{
      if (!haveContentTypeHeader)
        HttpHeaders.contentTypeHeader:
            ContentType('application', 'json', charset: 'utf-8').toString(),
      if (token != null && token!.isNotEmpty)
        // На имени HttpHeaders.authorizationHeader завязан другой код!
        HttpHeaders.authorizationHeader: 'Bearer $token',
      HttpHeaders.acceptHeader:
          ContentType('application', 'json', charset: 'utf-8').toString(),
      ...(headers ?? {}),
    };

    late Response<T> response;

    try {
      response = await _dio.request<T>(
        requestHost + path,
        data: body,
        queryParameters: query,
        options: Options(
          method: _getHttpMethodString(method),
          headers: internalHeaders,
        ),
      );
    } catch (e) {
      if (e is DioException && e.response != null) {
        var exception = AppHttpException(
          requestOptions: e.requestOptions,
          error: e.error,
          response: e.response,
          type: _exceptionAdapter.adapt(e.type),
        );

        // Проверяем, есть ли AppHttp401Exception в error
        if (e.error is AppHttp401Exception) {
          throw e.error as AppHttp401Exception;
        } else if (e.response!.statusCode == 401) {
          exception = AppHttp401Exception(exception);
          throw exception;
        }

        // Пытаемся распарсить стандартную ошибку бэка: { error: { code, message, details } }
        try {
          final data = e.response!.data;
          final Map<String, dynamic> body = (data is Map<String, dynamic>)
              ? data
              : (data is String)
                  ? (jsonDecode(data) as Map<String, dynamic>)
                  : <String, dynamic>{};
          final err = (body['error'] as Map?)?.cast<String, dynamic>();
          if (err != null) {
            final payload = ServerErrorPayload(
              code: (err['code'] ?? '').toString(),
              message: (err['message'] ?? '').toString(),
              details: (err['details'] as Map?)?.cast<String, dynamic>(),
            );
            throw AppHttpServerException(exception, payload);
          }
        } catch (_) {
          // ignore parse issues, fallthrough to generic
        }

        throw exception;
      }

      rethrow;
    }

    await _rememberTokensFromResponse(response);

    return response;
  }

  @override
  Future<Response<T>> get<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
  }) {
    return performRequest<T>(
      method: HttpMethod.get,
      host: host,
      path: path,
      query: query,
      headers: headers,
    );
  }

  @override
  Future<Response<T>> post<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Map<String, dynamic>? body,
  }) {
    return performRequest<T>(
      method: HttpMethod.post,
      host: host,
      path: path,
      query: query,
      headers: headers,
      body: body,
    );
  }

  @override
  Future<Response<T>> put<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Map<String, dynamic>? body,
  }) {
    return performRequest<T>(
      method: HttpMethod.put,
      host: host,
      path: path,
      query: query,
      headers: headers,
      body: body,
    );
  }

  @override
  Future<Response<T>> delete<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Map<String, dynamic>? body,
  }) {
    return performRequest<T>(
      method: HttpMethod.delete,
      host: host,
      path: path,
      query: query,
      headers: headers,
      body: body,
    );
  }

  @override
  Future<Response<T>> patch<T>({
    String? host,
    String path = '',
    Map<String, String>? query,
    Map<String, String>? headers,
    Map<String, dynamic>? body,
  }) {
    return performRequest<T>(
      method: HttpMethod.patch,
      host: host,
      path: path,
      query: query,
      headers: headers,
      body: body,
    );
  }

  @override
  Future<Response<T>> filesPost<T>({
    String? host,
    String path = '',
    Map<String, String>? headers,
    Map<String, String>? query,
    required files,
    Map<String, String>? fields,
  }) {
    return performRequest<T>(
      method: HttpMethod.filePost,
      host: host,
      path: path,
      headers: headers,
      query: query,
      fields: fields,
    );
  }

  @override
  Future<Response<T>> filesPatch<T>({
    String? host,
    String path = '',
    Map<String, String>? headers,
    Map<String, String>? query,
    required files,
    Map<String, String>? fields,
  }) {
    return performRequest<T>(
      method: HttpMethod.filePatch,
      host: host,  
      path: path,
      headers: headers,
      query: query,
      fields: fields,
    );
  }

  @override
  Future<void> clearTokens() => _tokensStorage.clear();

  /// Преобразует HttpMethod в строку HTTP метода
  String _getHttpMethodString(HttpMethod method) {
    switch (method) {
      case HttpMethod.get:
        return 'GET';
      case HttpMethod.post:
        return 'POST';
      case HttpMethod.put:
        return 'PUT';
      case HttpMethod.delete:
        return 'DELETE';
      case HttpMethod.patch:
        return 'PATCH';
      case HttpMethod.filePost:
        return 'POST';
      case HttpMethod.filePatch:
        return 'PATCH';
    }
  }

  @override
  Future<void> rememberTokens(
    String? token,
    String? refreshToken,
  ) =>
      _tokensStorage.save(
        token,
        refreshToken,
      );

  Future<void> _rememberTokensFromResponse(Response value) async {
    try {
      final data = value.data;

      // Токены могут приходить как в корне, так и внутри meta
      final accessToken = data[TokensStorageKeys.token] as String? ??
          (data['meta']?[TokensStorageKeys.token] as String?);
      final refreshToken = data[TokensStorageKeys.refreshToken] as String? ??
          (data['meta']?[TokensStorageKeys.refreshToken] as String?);

      await rememberTokens(accessToken, refreshToken);
    } catch (_) {}
  }
}
