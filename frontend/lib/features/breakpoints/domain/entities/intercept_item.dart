import 'package:flutter/foundation.dart';

@immutable
class InterceptItem {
  const InterceptItem({
    required this.id,
    required this.createdAt,
    required this.deadline,
    required this.direction,
    required this.sessionId,
    required this.state,
    this.ruleId,
    this.req,
    this.res,
  });

  final String id;
  final DateTime createdAt;
  final DateTime deadline;
  final String direction; // 'request' | 'response'
  final String sessionId;
  final String state; // PENDING | APPLIED | CANCELED | TIMED_OUT
  final String? ruleId;
  final HTTPRequestSnapshot? req;
  final HTTPResponseSnapshot? res;
}

@immutable
class HTTPRequestSnapshot {
  const HTTPRequestSnapshot({
    required this.method,
    required this.url,
    required this.headers,
    this.bodyBase64,
    this.bodyTruncated = false,
    this.contentType,
  });
  final String method;
  final String url;
  final Map<String, List<String>> headers;
  final String? bodyBase64;
  final bool bodyTruncated;
  final String? contentType;
}

@immutable
class HTTPResponseSnapshot {
  const HTTPResponseSnapshot({
    required this.status,
    required this.headers,
    this.bodyBase64,
    this.bodyTruncated = false,
    this.contentType,
  });
  final int status;
  final Map<String, List<String>> headers;
  final String? bodyBase64;
  final bool bodyTruncated;
  final String? contentType;
}
