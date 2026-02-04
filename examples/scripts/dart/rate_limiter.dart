// Example Dart script for rate limiting
// Run: dart run rate_limiter.dart
// This script works via subprocess executor (not WASM)

import 'dart:convert';
import 'dart:io';

class HTTPRequest {
  final String method;
  final String url;
  final Map<String, List<String>> headers;
  final List<int> body;

  HTTPRequest({
    required this.method,
    required this.url,
    required this.headers,
    required this.body,
  });

  factory HTTPRequest.fromJson(Map<String, dynamic> json) {
    return HTTPRequest(
      method: json['method'] as String,
      url: json['url'] as String,
      headers: (json['headers'] as Map<String, dynamic>).map(
        (key, value) => MapEntry(key, List<String>.from(value as List)),
      ),
      body: List<int>.from(json['body'] as List),
    );
  }

  Map<String, dynamic> toJson() {
    return {'method': method, 'url': url, 'headers': headers, 'body': body};
  }
}

class HTTPResponse {
  final int status;
  final Map<String, List<String>> headers;
  final List<int> body;

  HTTPResponse({
    required this.status,
    required this.headers,
    required this.body,
  });

  Map<String, dynamic> toJson() {
    return {'status': status, 'headers': headers, 'body': body};
  }
}

// Simple in-memory rate limiter
class RateLimiter {
  final Map<String, List<DateTime>> _requests = {};
  final int maxRequests;
  final Duration window;

  RateLimiter({required this.maxRequests, required this.window});

  bool isAllowed(String clientIP) {
    final now = DateTime.now();
    final requests = _requests[clientIP] ?? [];

    // Remove old requests
    requests.removeWhere((time) => now.difference(time) > window);

    if (requests.length >= maxRequests) {
      return false;
    }

    requests.add(now);
    _requests[clientIP] = requests;
    return true;
  }

  int getRemainingRequests(String clientIP) {
    final requests = _requests[clientIP] ?? [];
    return maxRequests - requests.length;
  }
}

void main() {
  // 10 requests per minute
  final rateLimiter = RateLimiter(
    maxRequests: 10,
    window: Duration(minutes: 1),
  );

  // Read JSON-RPC requests from stdin
  stdin.transform(utf8.decoder).transform(const LineSplitter()).listen((line) {
    try {
      final rpcRequest = jsonDecode(line);
      final params = rpcRequest['params'];

      if (params['request'] != null) {
        final req = HTTPRequest.fromJson(params['request']);

        // Extract client IP
        final clientIP =
            params['session']?['clientAddr'] as String? ?? 'unknown';

        stderr.writeln(
          '[Dart] Processing request from $clientIP to ${req.url}',
        );

        // Check rate limit
        if (!rateLimiter.isAllowed(clientIP)) {
          // Block request - return 429 Too Many Requests
          final response = HTTPResponse(
            status: 429,
            headers: {
              'Content-Type': ['application/json'],
              'Retry-After': ['60'],
              'X-RateLimit-Limit': ['${rateLimiter.maxRequests}'],
              'X-RateLimit-Remaining': ['0'],
            },
            body: utf8.encode(
              jsonEncode({
                'error': 'Rate limit exceeded',
                'message': 'Too many requests. Please try again later.',
                'limit': rateLimiter.maxRequests,
                'window': '1 minute',
              }),
            ),
          );

          // JSON-RPC response
          final rpcResponse = {
            'jsonrpc': '2.0',
            'id': rpcRequest['id'],
            'result': response.toJson(),
          };

          stdout.writeln(jsonEncode(rpcResponse));
          return;
        }

        // Rate limit OK - add headers and pass through
        req.headers['X-RateLimit-Limit'] = ['${rateLimiter.maxRequests}'];
        req.headers['X-RateLimit-Remaining'] = [
          '${rateLimiter.getRemainingRequests(clientIP)}',
        ];

        // JSON-RPC response with modified request
        final rpcResponse = {
          'jsonrpc': '2.0',
          'id': rpcRequest['id'],
          'result': req.toJson(),
        };

        stdout.writeln(jsonEncode(rpcResponse));
      }
    } catch (e) {
      stderr.writeln('[Dart] Error: $e');
    }
  });
}
