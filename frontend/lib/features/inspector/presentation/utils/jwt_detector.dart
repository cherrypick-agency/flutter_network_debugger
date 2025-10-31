import 'dart:convert';
import 'content_detector.dart';

/// JWT (JSON Web Token) content detector
class JwtContentDetector implements ContentDetector {
  @override
  bool detect(String content, {String? contentType}) {
    if (content.isEmpty) return false;

    final trimmed = content.trim();
    final parts = trimmed.split('.');

    // JWT must have exactly 3 parts: header.payload.signature
    if (parts.length != 3) return false;

    // Each part should be non-empty
    if (parts.any((p) => p.isEmpty)) return false;

    // Try to decode first two parts as Base64 URL encoded JSON
    try {
      final header = _decodeJwtPart(parts[0]);
      final payload = _decodeJwtPart(parts[1]);

      // Header and payload must be valid JSON
      jsonDecode(header);
      jsonDecode(payload);

      return true;
    } catch (_) {
      return false;
    }
  }

  /// Parse JWT into its components
  JwtData? parse(String token) {
    try {
      final trimmed = token.trim();
      final parts = trimmed.split('.');

      if (parts.length != 3) return null;

      final headerRaw = _decodeJwtPart(parts[0]);
      final payloadRaw = _decodeJwtPart(parts[1]);
      final signature = parts[2];

      // Parse header and payload as JSON
      final headerJson = jsonDecode(headerRaw) as Map<String, dynamic>;
      final payloadJson = jsonDecode(payloadRaw) as Map<String, dynamic>;

      return JwtData(
        header: headerJson,
        payload: payloadJson,
        signature: signature,
        headerRaw: headerRaw,
        payloadRaw: payloadRaw,
        algorithm: headerJson['alg']?.toString(),
        tokenType: headerJson['typ']?.toString(),
        issuedAt: _extractTimestamp(payloadJson['iat']),
        expiresAt: _extractTimestamp(payloadJson['exp']),
        notBefore: _extractTimestamp(payloadJson['nbf']),
        issuer: payloadJson['iss']?.toString(),
        subject: payloadJson['sub']?.toString(),
        audience: payloadJson['aud'],
      );
    } catch (_) {
      return null;
    }
  }

  /// Decode a JWT part (Base64 URL encoded)
  String _decodeJwtPart(String part) {
    // JWT uses Base64 URL encoding (RFC 4648 Section 5)
    // We need to normalize it to standard Base64
    String normalized = part.replaceAll('-', '+').replaceAll('_', '/');

    // Add padding if needed
    switch (normalized.length % 4) {
      case 2:
        normalized += '==';
        break;
      case 3:
        normalized += '=';
        break;
    }

    final decoded = base64Decode(normalized);
    return utf8.decode(decoded);
  }

  DateTime? _extractTimestamp(dynamic value) {
    if (value == null) return null;
    if (value is int) {
      return DateTime.fromMillisecondsSinceEpoch(value * 1000);
    }
    if (value is String) {
      final parsed = int.tryParse(value);
      if (parsed != null) {
        return DateTime.fromMillisecondsSinceEpoch(parsed * 1000);
      }
    }
    return null;
  }

  @override
  String get displayName => 'JWT';

  @override
  int get priority => 15; // High priority - check before Base64
}

/// Parsed JWT data with all fields
class JwtData {
  final Map<String, dynamic> header;
  final Map<String, dynamic> payload;
  final String signature;
  final String headerRaw;
  final String payloadRaw;

  // Standard JWT header fields
  final String? algorithm;
  final String? tokenType;

  // Standard JWT payload fields
  final DateTime? issuedAt;
  final DateTime? expiresAt;
  final DateTime? notBefore;
  final String? issuer;
  final String? subject;
  final dynamic audience;

  const JwtData({
    required this.header,
    required this.payload,
    required this.signature,
    required this.headerRaw,
    required this.payloadRaw,
    this.algorithm,
    this.tokenType,
    this.issuedAt,
    this.expiresAt,
    this.notBefore,
    this.issuer,
    this.subject,
    this.audience,
  });

  /// Check if token is currently valid (not expired and after notBefore)
  bool get isValid {
    final now = DateTime.now();

    if (expiresAt != null && now.isAfter(expiresAt!)) {
      return false;
    }

    if (notBefore != null && now.isBefore(notBefore!)) {
      return false;
    }

    return true;
  }

  /// Get time until expiration (null if no expiration or already expired)
  Duration? get timeUntilExpiration {
    if (expiresAt == null) return null;
    final now = DateTime.now();
    if (now.isAfter(expiresAt!)) return null;
    return expiresAt!.difference(now);
  }

  /// Pretty-print JSON for display
  String get headerFormatted => _formatJson(header);
  String get payloadFormatted => _formatJson(payload);

  String _formatJson(Map<String, dynamic> json) {
    const encoder = JsonEncoder.withIndent('  ');
    return encoder.convert(json);
  }
}
