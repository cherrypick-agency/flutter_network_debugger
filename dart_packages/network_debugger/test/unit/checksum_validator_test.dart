import 'dart:io';
import 'package:test/test.dart';
import 'package:network_debugger/src/checksum_validator.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  group('ChecksumValidator', () {
    late Directory tempDir;
    late String testFilePath;

    setUp(() async {
      // Create temporary directory for test files
      tempDir = await Directory.systemTemp.createTemp('checksum_test_');
      testFilePath = '${tempDir.path}/test_file.txt';
    });

    tearDown(() async {
      // Clean up temporary directory
      if (await tempDir.exists()) {
        await tempDir.delete(recursive: true);
      }
    });

    test('computeFileChecksum returns correct SHA256 hash', () async {
      final file = File(testFilePath);
      await file.writeAsString('Hello, World!');

      final validator = ChecksumValidator();
      final checksum = await validator.computeFileChecksum(testFilePath);

      // Known SHA256 hash for "Hello, World!"
      expect(
        checksum,
        equals(
          'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f',
        ),
      );
    });

    test('computeFileChecksum throws when file not found', () async {
      final validator = ChecksumValidator();

      expect(
        () => validator.computeFileChecksum('/nonexistent/file.txt'),
        throwsA(isA<ChecksumValidationException>()),
      );
    });

    test('validateFile returns true for matching checksum', () async {
      final file = File(testFilePath);
      await file.writeAsString('Hello, World!');

      final validator = ChecksumValidator();
      final expectedChecksum =
          'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f';

      final isValid = await validator.validateFile(
        testFilePath,
        expectedChecksum,
      );

      expect(isValid, isTrue);
    });

    test('validateFile returns false for non-matching checksum', () async {
      final file = File(testFilePath);
      await file.writeAsString('Hello, World!');

      final validator = ChecksumValidator();
      final wrongChecksum =
          '0000000000000000000000000000000000000000000000000000000000000000';

      final isValid = await validator.validateFile(
        testFilePath,
        wrongChecksum,
      );

      expect(isValid, isFalse);
    });

    test('validateFile handles uppercase checksums', () async {
      final file = File(testFilePath);
      await file.writeAsString('Hello, World!');

      final validator = ChecksumValidator();
      final expectedChecksum =
          'DFFD6021BB2BD5B0AF676290809EC3A53191DD81C7F70A4B28688A362182986F';

      final isValid = await validator.validateFile(
        testFilePath,
        expectedChecksum,
      );

      expect(isValid, isTrue);
    });

    test('downloadChecksumFile parses checksum correctly', () async {
      final mockClient = MockClient((request) async {
        return http.Response(
          'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f',
          200,
        );
      });

      final validator = ChecksumValidator(client: mockClient);
      final checksum = await validator.downloadChecksumFile(
        'https://example.com/file.sha256',
      );

      expect(
        checksum,
        equals(
          'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f',
        ),
      );
    });

    test('downloadChecksumFile parses checksum with filename', () async {
      final mockClient = MockClient((request) async {
        return http.Response(
          'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f  file.tar.gz',
          200,
        );
      });

      final validator = ChecksumValidator(client: mockClient);
      final checksum = await validator.downloadChecksumFile(
        'https://example.com/file.sha256',
      );

      expect(
        checksum,
        equals(
          'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f',
        ),
      );
    });

    test('downloadChecksumFile throws on invalid checksum format', () async {
      final mockClient = MockClient((request) async {
        return http.Response('invalid_checksum', 200);
      });

      final validator = ChecksumValidator(client: mockClient);

      expect(
        () => validator.downloadChecksumFile('https://example.com/file.sha256'),
        throwsA(isA<ChecksumValidationException>()),
      );
    });

    test('downloadChecksumFile throws on HTTP error', () async {
      final mockClient = MockClient((request) async {
        return http.Response('Not Found', 404);
      });

      final validator = ChecksumValidator(client: mockClient);

      expect(
        () => validator.downloadChecksumFile('https://example.com/file.sha256'),
        throwsA(isA<ChecksumValidationException>()),
      );
    });

    test('downloadChecksumFile throws on empty response', () async {
      final mockClient = MockClient((request) async {
        return http.Response('', 200);
      });

      final validator = ChecksumValidator(client: mockClient);

      expect(
        () => validator.downloadChecksumFile('https://example.com/file.sha256'),
        throwsA(isA<ChecksumValidationException>()),
      );
    });
  });
}
