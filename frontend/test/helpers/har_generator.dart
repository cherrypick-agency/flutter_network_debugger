/// Helper class to generate HAR (HTTP Archive) test data
///
/// This utility makes it easy to create HAR files for testing import functionality.
/// Supports both HTTP and WebSocket entries with configurable options.
class HARGenerator {
  /// Creates a complete HAR JSON structure
  static Map<String, dynamic> createHAR({
    required List<Map<String, dynamic>> entries,
    String version = '1.2',
    String creatorName = 'test-harness',
    String creatorVersion = '1.0',
  }) {
    return {
      'log': {
        'version': version,
        'creator': {'name': creatorName, 'version': creatorVersion},
        'entries': entries,
      },
    };
  }

  /// Creates a simple HTTP entry
  static Map<String, dynamic> createHTTPEntry({
    required String method,
    required String url,
    required int status,
    String statusText = 'OK',
    String? startedDateTime,
    double time = 100.0,
    Map<String, String>? requestHeaders,
    Map<String, String>? responseHeaders,
    String? responseBody,
    String? contentType,
  }) {
    final started = startedDateTime ?? DateTime.now().toUtc().toIso8601String();
    final respHeaders = <Map<String, String>>[];
    if (responseHeaders != null) {
      responseHeaders.forEach((name, value) {
        respHeaders.add({'name': name, 'value': value});
      });
    }
    if (contentType != null &&
        !respHeaders.any((h) => h['name'] == 'Content-Type')) {
      respHeaders.add({'name': 'Content-Type', 'value': contentType});
    }

    final reqHeaders = <Map<String, String>>[];
    if (requestHeaders != null) {
      requestHeaders.forEach((name, value) {
        reqHeaders.add({'name': name, 'value': value});
      });
    }

    return {
      'startedDateTime': started,
      'time': time,
      'request': {
        'method': method,
        'url': url,
        'httpVersion': 'HTTP/1.1',
        'headers': reqHeaders,
        'queryString': <Map<String, String>>[],
        'cookies': <Map<String, dynamic>>[],
        'headersSize': -1,
        'bodySize': 0,
      },
      'response': {
        'status': status,
        'statusText': statusText,
        'httpVersion': 'HTTP/1.1',
        'headers': respHeaders,
        'cookies': <Map<String, dynamic>>[],
        'content': {
          'size': responseBody?.length ?? 0,
          'mimeType': contentType ?? '',
          if (responseBody != null) 'text': responseBody,
        },
        'redirectURL': '',
        'headersSize': -1,
        'bodySize': responseBody?.length ?? 0,
      },
      'cache': <String, dynamic>{},
      'timings': {'send': 1, 'wait': time - 2, 'receive': 1},
    };
  }

  /// Creates a WebSocket entry with messages
  static Map<String, dynamic> createWebSocketEntry({
    required String url,
    required List<Map<String, dynamic>> messages,
    String? startedDateTime,
    double time = 50.0,
  }) {
    final started = startedDateTime ?? DateTime.now().toUtc().toIso8601String();

    return {
      'startedDateTime': started,
      'time': time,
      'request': {
        'method': 'GET',
        'url': url,
        'httpVersion': 'HTTP/1.1',
        'headers': [
          {'name': 'Upgrade', 'value': 'websocket'},
          {'name': 'Connection', 'value': 'Upgrade'},
        ],
        'queryString': <Map<String, String>>[],
        'cookies': <Map<String, dynamic>>[],
        'headersSize': -1,
        'bodySize': 0,
      },
      'response': {
        'status': 101,
        'statusText': 'Switching Protocols',
        'httpVersion': 'HTTP/1.1',
        'headers': [
          {'name': 'Upgrade', 'value': 'websocket'},
          {'name': 'Connection', 'value': 'Upgrade'},
        ],
        'cookies': <Map<String, dynamic>>[],
        'content': {'size': 0, 'mimeType': ''},
        'redirectURL': '',
        'headersSize': -1,
        'bodySize': 0,
      },
      'cache': <String, dynamic>{},
      'timings': {'send': 0, 'wait': time, 'receive': 0},
      '_webSocketMessages': messages,
    };
  }

  /// Creates a WebSocket message
  static Map<String, dynamic> createWebSocketMessage({
    required String type, // 'send' or 'receive'
    required String data,
    double? time,
    int opcode = 1, // 1 = text, 2 = binary
  }) {
    final timestamp = time ?? DateTime.now().millisecondsSinceEpoch / 1000.0;

    return {'type': type, 'time': timestamp, 'opcode': opcode, 'data': data};
  }

  /// Creates a HAR with a single HTTP GET request
  static Map<String, dynamic> createSimpleHTTP({
    String url = 'https://api.example.com/test',
    int status = 200,
    String? responseBody,
  }) {
    return createHAR(
      entries: [
        createHTTPEntry(
          method: 'GET',
          url: url,
          status: status,
          responseBody: responseBody ?? '{"message":"success"}',
          contentType: 'application/json',
        ),
      ],
    );
  }

  /// Creates a HAR with multiple HTTP entries
  static Map<String, dynamic> createMultipleHTTP({
    required List<Map<String, dynamic>> requests,
  }) {
    final entries = requests.map((req) {
      return createHTTPEntry(
        method: req['method'] as String? ?? 'GET',
        url: req['url'] as String,
        status: req['status'] as int? ?? 200,
        responseBody: req['responseBody'] as String?,
        contentType: req['contentType'] as String?,
      );
    }).toList();

    return createHAR(entries: entries);
  }

  /// Creates a HAR with a WebSocket connection
  static Map<String, dynamic> createSimpleWebSocket({
    String url = 'wss://ws.example.com/socket',
    int messageCount = 4,
  }) {
    final messages = <Map<String, dynamic>>[];
    final baseTime = DateTime.now().millisecondsSinceEpoch / 1000.0;

    for (var i = 0; i < messageCount; i++) {
      messages.add(
        createWebSocketMessage(
          type: i % 2 == 0 ? 'send' : 'receive',
          data: 'Message $i',
          time: baseTime + i * 0.5,
        ),
      );
    }

    return createHAR(
      entries: [createWebSocketEntry(url: url, messages: messages)],
    );
  }

  /// Creates a HAR with invalid version for testing error handling
  static Map<String, dynamic> createInvalidVersion() {
    return createHAR(
      version: '2.0', // Invalid version
      entries: [
        createHTTPEntry(
          method: 'GET',
          url: 'https://api.example.com/test',
          status: 200,
        ),
      ],
    );
  }

  /// Creates malformed JSON (as String, not Map) for testing error handling
  static String createMalformedJSON() {
    return '{invalid json';
  }
}
