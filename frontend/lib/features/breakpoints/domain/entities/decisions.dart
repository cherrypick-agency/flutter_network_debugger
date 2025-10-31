import 'package:flutter/foundation.dart';

@immutable
class RequestDecision {
  const RequestDecision({
    required this.action, // continue | drop
    this.method,
    this.url,
    this.headers,
    this.bodyBase64,
  });
  final String action;
  final String? method;
  final String? url;
  final Map<String, List<String>>? headers;
  final String? bodyBase64;
}

@immutable
class ResponseDecision {
  const ResponseDecision({
    required this.action, // continue
    this.status,
    this.headers,
    this.bodyBase64,
  });
  final String action;
  final int? status;
  final Map<String, List<String>>? headers;
  final String? bodyBase64;
}
