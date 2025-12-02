import 'dart:convert';

import '../models.dart';

/// Use case для конвертации сырых данных HTTP запроса в ComposeTemplateDTO.
/// Извлекает method, url, headers, query params, body и определяет тип авторизации.
class ConvertRequestToTemplateUseCase {
  /// Конвертирует данные запроса из Inspector в шаблон для Compose.
  ///
  /// [requestData] - Map с данными запроса (url, method, headers, body и т.д.)
  /// [headersRaw] - опционально, сырые заголовки без маскирования
  ComposeTemplateDTO call({
    required Map<String, dynamic> requestData,
    Map<String, String>? headersRaw,
  }) {
    final method = (requestData['method'] ?? 'GET').toString().toUpperCase();
    final rawUrl = (requestData['url'] ?? '').toString();
    final uri = Uri.tryParse(rawUrl);

    // Извлекаем URL без query params (они пойдут отдельно)
    final urlWithoutQuery = uri != null ? _buildUrlWithoutQuery(uri) : rawUrl;

    // Извлекаем query параметры из URL
    final queryParams = _extractQueryParams(uri);

    // Обрабатываем заголовки
    final rawHeaders = _parseHeaders(requestData['headers']);
    // Предпочитаем сырые заголовки если доступны
    final effectiveHeaders = headersRaw != null && headersRaw.isNotEmpty
        ? headersRaw
        : rawHeaders;

    // Определяем авторизацию из заголовков
    final auth = _extractAuth(effectiveHeaders);

    // Фильтруем заголовки, убирая те что относятся к авторизации
    final filteredHeaders = _filterAuthHeaders(effectiveHeaders, auth);

    // Обрабатываем body
    final bodyData = _extractBody(requestData, effectiveHeaders);

    return ComposeTemplateDTO(
      id: 'from-inspector-${DateTime.now().microsecondsSinceEpoch}',
      name: _generateName(method, uri),
      method: method,
      url: urlWithoutQuery,
      headers: filteredHeaders.entries
          .map((e) => ComposeHeaderDTO(key: e.key, value: e.value))
          .toList(),
      query: queryParams
          .map((e) => ComposeQueryDTO(key: e.key, value: e.value))
          .toList(),
      body: bodyData,
      auth: auth,
    );
  }

  /// Строит URL без query параметров
  String _buildUrlWithoutQuery(Uri uri) {
    final sb = StringBuffer();
    sb.write(uri.scheme);
    sb.write('://');
    sb.write(uri.host);
    if (uri.hasPort &&
        !((uri.scheme == 'http' && uri.port == 80) ||
            (uri.scheme == 'https' && uri.port == 443))) {
      sb.write(':${uri.port}');
    }
    sb.write(uri.path.isEmpty ? '/' : uri.path);
    return sb.toString();
  }

  /// Извлекает query параметры из URI
  List<_KvEntry> _extractQueryParams(Uri? uri) {
    if (uri == null) return [];
    final result = <_KvEntry>[];
    uri.queryParametersAll.forEach((key, values) {
      for (final value in values) {
        result.add(_KvEntry(key, value));
      }
    });
    return result;
  }

  /// Парсит заголовки из разных форматов
  Map<String, String> _parseHeaders(dynamic headers) {
    if (headers == null) return {};
    if (headers is Map) {
      return headers.map((k, v) => MapEntry(k.toString(), v.toString()));
    }
    return {};
  }

  /// Определяет тип авторизации по заголовкам
  ComposeAuthDTO? _extractAuth(Map<String, String> headers) {
    // Ищем Authorization header (case-insensitive)
    String? authValue;
    for (final entry in headers.entries) {
      if (entry.key.toLowerCase() == 'authorization') {
        authValue = entry.value;
        break;
      }
    }

    if (authValue != null && authValue.isNotEmpty) {
      // Bearer token
      if (authValue.toLowerCase().startsWith('bearer ')) {
        final token = authValue.substring(7).trim();
        return ComposeAuthDTO(type: 'bearer', bearerToken: token);
      }

      // Basic auth
      if (authValue.toLowerCase().startsWith('basic ')) {
        final encoded = authValue.substring(6).trim();
        try {
          final decoded = utf8.decode(base64Decode(encoded));
          final parts = decoded.split(':');
          if (parts.length >= 2) {
            return ComposeAuthDTO(
              type: 'basic',
              username: parts[0],
              password: parts.sublist(1).join(':'),
            );
          }
        } catch (_) {
          // Невалидный base64, пропускаем
        }
      }
    }

    // Ищем API Key в известных заголовках
    final apiKeyHeaders = ['x-api-key', 'api-key', 'apikey'];
    for (final entry in headers.entries) {
      final lowerKey = entry.key.toLowerCase();
      if (apiKeyHeaders.contains(lowerKey) && entry.value.isNotEmpty) {
        return ComposeAuthDTO(
          type: 'apiKey',
          apiKey: entry.value,
          apiKeyHeader: entry.key,
        );
      }
    }

    return null;
  }

  /// Фильтрует заголовки, убирая относящиеся к авторизации
  Map<String, String> _filterAuthHeaders(
    Map<String, String> headers,
    ComposeAuthDTO? auth,
  ) {
    if (auth == null) return headers;

    final result = <String, String>{};
    for (final entry in headers.entries) {
      final lowerKey = entry.key.toLowerCase();

      // Пропускаем Authorization header если есть bearer/basic auth
      if (lowerKey == 'authorization' &&
          (auth.type == 'bearer' || auth.type == 'basic')) {
        continue;
      }

      // Пропускаем API Key header
      if (auth.type == 'apiKey' &&
          lowerKey == (auth.apiKeyHeader?.toLowerCase() ?? 'x-api-key')) {
        continue;
      }

      result[entry.key] = entry.value;
    }
    return result;
  }

  /// Извлекает и определяет тип body
  ComposeBodyDTO _extractBody(
    Map<String, dynamic> requestData,
    Map<String, String> headers,
  ) {
    final body = (requestData['body'] ?? '').toString();
    final contentType = _findContentType(headers);

    // Определяем mode по Content-Type
    if (contentType.contains('application/json')) {
      return ComposeBodyDTO(
        mode: 'json',
        json: body.isNotEmpty ? body : '{\n  \n}',
      );
    }

    if (contentType.contains('application/x-www-form-urlencoded')) {
      // Парсим form data из body
      final formFields = _parseFormBody(body);
      return ComposeBodyDTO(mode: 'form', form: formFields);
    }

    if (contentType.contains('multipart/form-data')) {
      // Multipart сложно восстановить из строки, используем raw
      return ComposeBodyDTO(mode: 'raw', raw: body);
    }

    // По умолчанию raw
    return ComposeBodyDTO(mode: 'raw', raw: body);
  }

  /// Находит Content-Type header
  String _findContentType(Map<String, String> headers) {
    for (final entry in headers.entries) {
      if (entry.key.toLowerCase() == 'content-type') {
        return entry.value.toLowerCase();
      }
    }
    return '';
  }

  /// Парсит form-urlencoded body в список полей
  List<ComposeFormFieldDTO> _parseFormBody(String body) {
    if (body.isEmpty) return [];
    final result = <ComposeFormFieldDTO>[];
    final pairs = body.split('&');
    for (final pair in pairs) {
      final idx = pair.indexOf('=');
      if (idx > 0) {
        final key = Uri.decodeComponent(pair.substring(0, idx));
        final value = Uri.decodeComponent(pair.substring(idx + 1));
        result.add(ComposeFormFieldDTO(key: key, value: value));
      } else if (pair.isNotEmpty) {
        result.add(
          ComposeFormFieldDTO(key: Uri.decodeComponent(pair), value: ''),
        );
      }
    }
    return result;
  }

  /// Генерирует имя для запроса
  String _generateName(String method, Uri? uri) {
    if (uri == null) return '$method Request';
    final path = uri.path.isEmpty ? '/' : uri.path;
    // Берём последний сегмент пути
    final segments = path.split('/').where((s) => s.isNotEmpty).toList();
    if (segments.isEmpty) return '$method ${uri.host}';
    return '$method /${segments.last}';
  }
}

/// Вспомогательный класс для key-value пар
class _KvEntry {
  final String key;
  final String value;
  _KvEntry(this.key, this.value);
}
