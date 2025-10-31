import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/inspector/presentation/utils/jwt_detector.dart';
import 'package:frontend/features/inspector/presentation/utils/base64_detector.dart';
import 'package:frontend/features/inspector/presentation/utils/content_detector.dart';

void main() {
  group('JwtContentDetector', () {
    final detector = JwtContentDetector();

    test('detects valid JWT token', () {
      const token =
          'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';

      expect(detector.detect(token), isTrue);
    });

    test('does not detect invalid JWT token', () {
      expect(detector.detect('not.a.jwt'), isFalse);
      expect(detector.detect('eyJhbGci'), isFalse);
      expect(detector.detect(''), isFalse);
    });

    test('parses JWT token correctly', () {
      const token =
          'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';

      final jwt = detector.parse(token);

      expect(jwt, isNotNull);
      expect(jwt!.algorithm, equals('HS256'));
      expect(jwt.tokenType, equals('JWT'));
      expect(jwt.payload['sub'], equals('1234567890'));
      expect(jwt.payload['name'], equals('John Doe'));
    });
  });

  group('Base64ContentDetector', () {
    final detector = Base64ContentDetector();

    test('detects valid Base64', () {
      const base64 = 'SGVsbG8gV29ybGQh'; // "Hello World!"

      expect(detector.detect(base64), isTrue);
    });

    test('does not detect invalid Base64', () {
      expect(detector.detect('not base64!'), isFalse);
      expect(detector.detect('short'), isFalse);
      expect(detector.detect(''), isFalse);
    });

    test('decodes Base64 correctly', () {
      const base64 = 'SGVsbG8gV29ybGQh'; // "Hello World!"

      final result = detector.decode(base64);

      expect(result.success, isTrue);
      expect(result.decoded, equals('Hello World!'));
    });

    test('detects JSON in decoded Base64', () {
      // Base64 of '{"message":"Hello"}'
      const base64 = 'eyJtZXNzYWdlIjoiSGVsbG8ifQ==';

      final result = detector.decode(base64);

      expect(result.success, isTrue);
      expect(result.isJson, isTrue);
      expect(result.decoded, contains('message'));
    });
  });

  group('JsonContentDetector', () {
    final detector = JsonContentDetector();

    test('detects valid JSON', () {
      expect(detector.detect('{"key": "value"}'), isTrue);
      expect(detector.detect('[1, 2, 3]'), isTrue);
    });

    test('does not detect invalid JSON', () {
      expect(detector.detect('not json'), isFalse);
      expect(detector.detect('{invalid}'), isFalse);
    });

    test('detects JSON by content type', () {
      expect(detector.detect('text', contentType: 'application/json'), isTrue);
      expect(
        detector.detect('text', contentType: 'application/vnd.api+json'),
        isTrue,
      );
    });
  });

  group('HtmlContentDetector', () {
    final detector = HtmlContentDetector();

    test('detects HTML by content', () {
      expect(detector.detect('<!DOCTYPE html>'), isTrue);
      expect(detector.detect('<html><body>Test</body></html>'), isTrue);
    });

    test('detects HTML by content type', () {
      expect(detector.detect('text', contentType: 'text/html'), isTrue);
    });

    test('does not detect non-HTML', () {
      expect(detector.detect('plain text'), isFalse);
    });
  });

  group('ContentDetectorRegistry', () {
    test('detects content in priority order', () {
      final registry = ContentDetectorRegistry();
      registry.registerAll([
        JwtContentDetector(),
        JsonContentDetector(),
        Base64ContentDetector(),
      ]);

      // JWT should be detected first (higher priority)
      const jwtToken =
          'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';

      final jwtResult = registry.detect(jwtToken);
      expect(jwtResult.isDetected, isTrue);
      expect(jwtResult.detector, isA<JwtContentDetector>());

      // JSON should be detected
      final jsonResult = registry.detect('{"key": "value"}');
      expect(jsonResult.isDetected, isTrue);
      expect(jsonResult.detector, isA<JsonContentDetector>());
    });
  });
}
