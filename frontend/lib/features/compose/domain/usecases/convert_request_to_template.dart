import 'dart:convert';

import '../models.dart';

/// Use case for converting raw HTTP request data to ComposeTemplateDTO.
/// Extracts method, url, headers, query params, body and determines auth type.
class ConvertRequestToTemplateUseCase {
  /// Converts request data from Inspector to Compose template.
  ///
  /// [requestData] - Map with request data (url, method, headers, body, etc.)
  /// [headersRaw] - optional, raw headers without masking
  ComposeTemplateDTO call({
    required Map<String, dynamic> requestData,
    Map<String, String>? headersRaw,
  }) {
    final method = (requestData['method'] ?? 'GET').toString().toUpperCase();
    final rawUrl = (requestData['url'] ?? '').toString();
    final uri = Uri.tryParse(rawUrl);

    // Extract URL without query params (they go separately)
    final urlWithoutQuery = uri != null ? _buildUrlWithoutQuery(uri) : rawUrl;

    // Extract query parameters from URL
    final queryParams = _extractQueryParams(uri);

    // Process headers
    final rawHeaders = _parseHeaders(requestData['headers']);
    // Prefer raw headers if available
    final effectiveHeaders = headersRaw != null && headersRaw.isNotEmpty
        ? headersRaw
        : rawHeaders;

    // Determine auth from headers
    final auth = _extractAuth(effectiveHeaders);

    // Filter headers, removing those related to auth
    final filteredHeaders = _filterAuthHeaders(effectiveHeaders, auth);

    // Process body
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

  /// Builds URL without query parameters
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

  /// Extracts query parameters from URI
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

  /// Parses headers from different formats
  Map<String, String> _parseHeaders(dynamic headers) {
    if (headers == null) return {};
    if (headers is Map) {
      return headers.map((k, v) => MapEntry(k.toString(), v.toString()));
    }
    return {};
  }

  /// Determines auth type from headers
  ComposeAuthDTO? _extractAuth(Map<String, String> headers) {
    // Find Authorization header (case-insensitive)
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
          // Invalid base64, skip
        }
      }
    }

    // Find API Key in known headers
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

  /// Filters headers, removing auth-related ones
  Map<String, String> _filterAuthHeaders(
    Map<String, String> headers,
    ComposeAuthDTO? auth,
  ) {
    if (auth == null) return headers;

    final result = <String, String>{};
    for (final entry in headers.entries) {
      final lowerKey = entry.key.toLowerCase();

      // Skip Authorization header if bearer/basic auth exists
      if (lowerKey == 'authorization' &&
          (auth.type == 'bearer' || auth.type == 'basic')) {
        continue;
      }

      // Skip API Key header
      if (auth.type == 'apiKey' &&
          lowerKey == (auth.apiKeyHeader?.toLowerCase() ?? 'x-api-key')) {
        continue;
      }

      result[entry.key] = entry.value;
    }
    return result;
  }

  /// Extracts and determines body type
  ComposeBodyDTO _extractBody(
    Map<String, dynamic> requestData,
    Map<String, String> headers,
  ) {
    final body = (requestData['body'] ?? '').toString();
    final contentType = _findContentType(headers);

    // Determine mode by Content-Type
    if (contentType.contains('application/json')) {
      return ComposeBodyDTO(
        mode: 'json',
        json: body.isNotEmpty ? body : '{\n  \n}',
      );
    }

    if (contentType.contains('application/x-www-form-urlencoded')) {
      // Parse form data from body
      final formFields = _parseFormBody(body);
      return ComposeBodyDTO(mode: 'form', form: formFields);
    }

    if (contentType.contains('multipart/form-data')) {
      // Multipart is hard to restore from string, use raw
      return ComposeBodyDTO(mode: 'raw', raw: body);
    }

    // Default to raw
    return ComposeBodyDTO(mode: 'raw', raw: body);
  }

  /// Finds Content-Type header
  String _findContentType(Map<String, String> headers) {
    for (final entry in headers.entries) {
      if (entry.key.toLowerCase() == 'content-type') {
        return entry.value.toLowerCase();
      }
    }
    return '';
  }

  /// Parses form-urlencoded body into list of fields
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

  /// Generates name for request
  String _generateName(String method, Uri? uri) {
    if (uri == null) return '$method Request';
    final path = uri.path.isEmpty ? '/' : uri.path;
    // Take last path segment
    final segments = path.split('/').where((s) => s.isNotEmpty).toList();
    if (segments.isEmpty) return '$method ${uri.host}';
    return '$method /${segments.last}';
  }
}

/// Helper class for key-value pairs
class _KvEntry {
  final String key;
  final String value;
  _KvEntry(this.key, this.value);
}
