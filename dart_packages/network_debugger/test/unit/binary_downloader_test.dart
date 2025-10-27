import 'dart:io';
import 'package:test/test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:archive/archive.dart';
import 'package:network_debugger/src/binary_downloader.dart';
import 'package:network_debugger/src/retry_helper.dart';
import 'package:network_debugger/src/checksum_validator.dart';
import 'package:network_debugger/src/logger.dart';

void main() {
  // Disable logging for tests
  setUpAll(() {
    Logger.level = LogLevel.none;
  });

  group('BinaryDownloader', () {
    late Directory tempDir;
    late String testFilePath;

    setUp(() async {
      // Create temporary directory for test files
      tempDir = await Directory.systemTemp.createTemp('downloader_test_');
      testFilePath = '${tempDir.path}/test_file.txt';
    });

    tearDown(() async {
      // Clean up temporary directory
      if (await tempDir.exists()) {
        await tempDir.delete(recursive: true);
      }
    });

    test('downloadFile succeeds with valid response', () async {
      final mockClient = MockClient((request) async {
        return http.Response('test content', 200, headers: {
          'content-length': '12',
        });
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
      );

      await downloader.downloadFile(
        'https://example.com/file.txt',
        testFilePath,
      );

      final file = File(testFilePath);
      expect(await file.exists(), isTrue);
      expect(await file.readAsString(), equals('test content'));
    });

    test('downloadFile calls progress callback', () async {
      final mockClient = MockClient((request) async {
        return http.Response('test content', 200, headers: {
          'content-length': '12',
        });
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
      );

      final progressCalls = <(int, int)>[];

      await downloader.downloadFile(
        'https://example.com/file.txt',
        testFilePath,
        onProgress: (received, total) {
          progressCalls.add((received, total));
        },
      );

      expect(progressCalls, isNotEmpty);
      expect(progressCalls.last.$1, equals(12)); // received
      expect(progressCalls.last.$2, equals(12)); // total
    });

    test('downloadFile retries on failure', () async {
      var attempts = 0;
      final mockClient = MockClient((request) async {
        attempts++;
        if (attempts < 2) {
          throw SocketException('Connection refused');
        }
        return http.Response('test content', 200, headers: {
          'content-length': '12',
        });
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(
            maxAttempts: 2,
            initialDelay: Duration(milliseconds: 10),
          ),
        ),
      );

      await downloader.downloadFile(
        'https://example.com/file.txt',
        testFilePath,
      );

      expect(attempts, equals(2));
      final file = File(testFilePath);
      expect(await file.exists(), isTrue);
    });

    test('downloadFile throws on HTTP error', () async {
      final mockClient = MockClient((request) async {
        return http.Response('Not Found', 404);
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
      );

      expect(
        () => downloader.downloadFile(
          'https://example.com/file.txt',
          testFilePath,
        ),
        throwsA(isA<DownloadException>()),
      );
    });

    test('downloadFile throws on timeout', () async {
      final mockClient = MockClient((request) async {
        await Future.delayed(const Duration(seconds: 10));
        return http.Response('test content', 200);
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
      );

      expect(
        () => downloader.downloadFile(
          'https://example.com/file.txt',
          testFilePath,
          timeout: const Duration(milliseconds: 100),
        ),
        throwsA(isA<DownloadException>()),
      );
    });

    test('extractArchive extracts tar.gz correctly', () async {
      // Create a simple tar.gz archive in memory
      final archive = Archive();
      final file = ArchiveFile('test_binary', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      archive.addFile(file);

      // Encode to tar
      final tarEncoder = TarEncoder();
      final tarBytes = tarEncoder.encode(archive);

      // Compress with gzip
      final gzipEncoder = GZipEncoder();
      final gzipBytes = gzipEncoder.encode(tarBytes);

      // Write to temp file
      final archivePath = '${tempDir.path}/test.tar.gz';
      await File(archivePath).writeAsBytes(gzipBytes!);

      final downloader = BinaryDownloader();

      final binaryPath = await downloader.extractArchive(
        archivePath,
        tempDir.path,
        expectedBinaryName: 'test_binary',
      );

      expect(binaryPath, endsWith('test_binary'));
      expect(await File(binaryPath).exists(), isTrue);
    });

    test('extractArchive extracts zip correctly', () async {
      // Create a simple zip archive in memory
      final archive = Archive();
      final file = ArchiveFile('test.exe', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      archive.addFile(file);

      // Encode to zip
      final zipEncoder = ZipEncoder();
      final zipBytes = zipEncoder.encode(archive);

      // Write to temp file
      final archivePath = '${tempDir.path}/test.zip';
      await File(archivePath).writeAsBytes(zipBytes!);

      final downloader = BinaryDownloader();

      final binaryPath = await downloader.extractArchive(
        archivePath,
        tempDir.path,
        expectedBinaryName: 'test.exe',
      );

      expect(binaryPath, endsWith('test.exe'));
      expect(await File(binaryPath).exists(), isTrue);
    });

    test('extractArchive throws when archive not found', () async {
      final downloader = BinaryDownloader();

      expect(
        () => downloader.extractArchive(
          '/nonexistent/archive.tar.gz',
          tempDir.path,
        ),
        throwsA(isA<DownloadException>()),
      );
    });

    test('extractArchive throws when binary not found in archive', () async {
      // Create archive without expected binary
      final archive = Archive();
      final file = ArchiveFile('wrong_file', 5, [1, 2, 3, 4, 5]);
      archive.addFile(file);

      final tarEncoder = TarEncoder();
      final tarBytes = tarEncoder.encode(archive);
      final gzipEncoder = GZipEncoder();
      final gzipBytes = gzipEncoder.encode(tarBytes);

      final archivePath = '${tempDir.path}/test.tar.gz';
      await File(archivePath).writeAsBytes(gzipBytes!);

      final downloader = BinaryDownloader();

      expect(
        () => downloader.extractArchive(
          archivePath,
          tempDir.path,
          expectedBinaryName: 'expected_binary',
        ),
        throwsA(isA<DownloadException>()),
      );
    });

    test('downloadAndExtract completes successfully', () async {
      // Create a simple tar.gz archive in memory
      final archive = Archive();
      final file = ArchiveFile('test_binary', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      archive.addFile(file);

      final tarEncoder = TarEncoder();
      final tarBytes = tarEncoder.encode(archive);
      final gzipEncoder = GZipEncoder();
      final gzipBytes = gzipEncoder.encode(tarBytes);

      final mockClient = MockClient((request) async {
        return http.Response.bytes(gzipBytes!, 200, headers: {
          'content-length': '${gzipBytes.length}',
        });
      });

      final mockChecksumValidator = ChecksumValidator(client: mockClient);

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
        checksumValidator: mockChecksumValidator,
      );

      final binaryPath = await downloader.downloadAndExtract(
        'https://example.com/archive.tar.gz',
        tempDir.path,
        expectedBinaryName: 'test_binary',
        skipChecksumValidation: true, // Skip checksum for this test
      );

      expect(binaryPath, endsWith('test_binary'));
      expect(await File(binaryPath).exists(), isTrue);
    });

    test('downloadAndExtract calls progress callback', () async {
      final archive = Archive();
      final file = ArchiveFile('test_binary', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      archive.addFile(file);

      final tarEncoder = TarEncoder();
      final tarBytes = tarEncoder.encode(archive);
      final gzipEncoder = GZipEncoder();
      final gzipBytes = gzipEncoder.encode(tarBytes);

      final mockClient = MockClient((request) async {
        return http.Response.bytes(gzipBytes!, 200, headers: {
          'content-length': '${gzipBytes.length}',
        });
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
      );

      final progressCalls = <(int, int)>[];

      await downloader.downloadAndExtract(
        'https://example.com/archive.tar.gz',
        tempDir.path,
        expectedBinaryName: 'test_binary',
        skipChecksumValidation: true,
        onProgress: (received, total) {
          progressCalls.add((received, total));
        },
      );

      expect(progressCalls, isNotEmpty);
    });

    test('downloadAndExtract cleans up archive after extraction', () async {
      final archive = Archive();
      final file = ArchiveFile('test_binary', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      archive.addFile(file);

      final tarEncoder = TarEncoder();
      final tarBytes = tarEncoder.encode(archive);
      final gzipEncoder = GZipEncoder();
      final gzipBytes = gzipEncoder.encode(tarBytes);

      final mockClient = MockClient((request) async {
        return http.Response.bytes(gzipBytes!, 200, headers: {
          'content-length': '${gzipBytes.length}',
        });
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(maxAttempts: 1),
        ),
      );

      await downloader.downloadAndExtract(
        'https://example.com/archive.tar.gz',
        tempDir.path,
        expectedBinaryName: 'test_binary',
        skipChecksumValidation: true,
      );

      // Archive file should be cleaned up
      final archivePath = '${tempDir.path}/archive.tar.gz';
      expect(await File(archivePath).exists(), isFalse);
    });

    test('downloadAndExtract calls retry callback on failure', () async {
      var attempts = 0;
      final archive = Archive();
      final file = ArchiveFile('test_binary', 10, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
      archive.addFile(file);

      final tarEncoder = TarEncoder();
      final tarBytes = tarEncoder.encode(archive);
      final gzipEncoder = GZipEncoder();
      final gzipBytes = gzipEncoder.encode(tarBytes);

      final mockClient = MockClient((request) async {
        attempts++;
        if (attempts < 2) {
          throw SocketException('Connection refused');
        }
        return http.Response.bytes(gzipBytes!, 200, headers: {
          'content-length': '${gzipBytes.length}',
        });
      });

      final downloader = BinaryDownloader(
        client: mockClient,
        retryHelper: RetryHelper(
          config: const RetryConfig(
            maxAttempts: 2,
            initialDelay: Duration(milliseconds: 10),
          ),
        ),
      );

      final retryCalls = <(int, Exception)>[];

      await downloader.downloadAndExtract(
        'https://example.com/archive.tar.gz',
        tempDir.path,
        expectedBinaryName: 'test_binary',
        skipChecksumValidation: true,
        onRetry: (attempt, error) {
          retryCalls.add((attempt, error));
        },
      );

      expect(retryCalls, hasLength(1));
      expect(retryCalls.first.$1, equals(1)); // First retry attempt
    });
  });
}
