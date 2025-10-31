import 'dart:convert';
import 'content_detector.dart';

/// Base64 content detector with auto-decode capability
class Base64ContentDetector implements ContentDetector {
  @override
  bool detect(String content, {String? contentType}) {
    if (content.isEmpty || content.length < 16) return false;

    // Skip if it's already detected as another format
    final ct = (contentType ?? '').toLowerCase();
    if (ct.contains('json') ||
        ct.contains('html') ||
        ct.contains('xml') ||
        ct.contains('javascript')) {
      return false;
    }

    // Check Base64 pattern (RFC 4648)
    final trimmed = content.trim();
    final base64Pattern = RegExp(r'^[A-Za-z0-9+/]*={0,2}$');

    if (!base64Pattern.hasMatch(trimmed)) return false;

    // Length must be multiple of 4 for valid Base64
    if (trimmed.length % 4 != 0) return false;

    // Try to decode to verify it's valid Base64
    try {
      base64Decode(trimmed);
      return true;
    } catch (_) {
      return false;
    }
  }

  /// Attempt to decode Base64 content to UTF-8 string
  String? tryDecode(String content) {
    try {
      final decoded = base64Decode(content.trim());
      return utf8.decode(decoded, allowMalformed: true);
    } catch (_) {
      return null;
    }
  }

  /// Decode and return metadata about decoded content
  DecodeResult decode(String content) {
    try {
      final decoded = base64Decode(content.trim());
      final asString = utf8.decode(decoded, allowMalformed: true);

      // Try to detect if decoded content is JSON
      bool isJson = false;
      try {
        jsonDecode(asString);
        isJson = true;
      } catch (_) {}

      return DecodeResult(
        success: true,
        decoded: asString,
        decodedBytes: decoded,
        isJson: isJson,
        originalSize: content.length,
        decodedSize: decoded.length,
      );
    } catch (e) {
      return DecodeResult(
        success: false,
        error: e.toString(),
        originalSize: content.length,
      );
    }
  }

  @override
  String get displayName => 'Base64';

  @override
  int get priority => 5; // Lower priority - check after JSON/HTML/XML
}

/// Result of Base64 decode operation
class DecodeResult {
  final bool success;
  final String? decoded;
  final List<int>? decodedBytes;
  final bool isJson;
  final String? error;
  final int originalSize;
  final int? decodedSize;

  const DecodeResult({
    required this.success,
    this.decoded,
    this.decodedBytes,
    this.isJson = false,
    this.error,
    required this.originalSize,
    this.decodedSize,
  });

  double? get compressionRatio {
    if (decodedSize == null || originalSize == 0) return null;
    return decodedSize! / originalSize;
  }
}
