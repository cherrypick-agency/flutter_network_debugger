import 'package:flutter/foundation.dart';

@immutable
class InterceptConfig {
  const InterceptConfig({
    required this.enabled,
    required this.requests,
    required this.responses,
    required this.timeoutMs,
    required this.queueMax,
    required this.bodyMaxBytes,
    required this.reencode,
    required this.overflow,
  });
  final bool enabled;
  final bool requests;
  final bool responses;
  final int timeoutMs;
  final int queueMax;
  final int bodyMaxBytes;
  final bool reencode;
  final String overflow; // auto-continue-oldest | drop-new
}
