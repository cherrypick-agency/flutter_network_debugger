import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:archive/archive.dart';
import 'package:path/path.dart' as p;
import 'retry_helper.dart';
import 'checksum_validator.dart';

/// Callback for download progress updates.
typedef ProgressCallback = void Function(int received, int total);

/// Callback for retry attempts.
typedef RetryCallback = void Function(int attempt, Exception error);

/// Callback for checksum validation.
typedef ChecksumCallback = void Function(bool validated, String? checksum);

/// Downloads and extracts binary files.
class BinaryDownloader {
  final http.Client? _client;
  final RetryHelper retryHelper;
  final ChecksumValidator checksumValidator;

  BinaryDownloader({
    http.Client? client,
    RetryHelper? retryHelper,
    ChecksumValidator? checksumValidator,
  })  : _client = client,
        retryHelper = retryHelper ?? RetryHelper(),
        checksumValidator = checksumValidator ?? ChecksumValidator(client: client);

  http.Client get client => _client ?? http.Client();

  /// Downloads a file from [url] to [destinationPath] with optional progress tracking.
  Future<void> downloadFile(
    String url,
    String destinationPath, {
    ProgressCallback? onProgress,
    RetryCallback? onRetry,
    Duration timeout = const Duration(minutes: 5),
  }) async {
    // Wrap download in retry logic
    await retryHelper.executeNetworkOperation(
      () => _downloadFileInternal(
        url,
        destinationPath,
        onProgress: onProgress,
        timeout: timeout,
      ),
      onRetry: onRetry,
    );
  }

  /// Internal download implementation (can be retried).
  Future<void> _downloadFileInternal(
    String url,
    String destinationPath, {
    ProgressCallback? onProgress,
    required Duration timeout,
  }) async {
    final uri = Uri.parse(url);
    final request = http.Request('GET', uri);

    final response = await client.send(request).timeout(
      timeout,
      onTimeout: () {
        throw DownloadException(
          'Download timeout after ${timeout.inMinutes} minutes',
        );
      },
    );

    if (response.statusCode != 200) {
      throw DownloadException(
        'Failed to download from $url: ${response.statusCode} ${response.reasonPhrase}',
      );
    }

    final file = File(destinationPath);
    await file.parent.create(recursive: true);

    final total = response.contentLength ?? 0;
    var received = 0;

    final sink = file.openWrite();

    try {
      await for (final chunk in response.stream) {
        sink.add(chunk);
        received += chunk.length;

        if (onProgress != null && total > 0) {
          onProgress(received, total);
        }
      }
    } finally {
      await sink.close();
    }

    if (total > 0 && received != total) {
      throw DownloadException(
        'Download incomplete: received $received bytes, expected $total bytes',
      );
    }
  }

  /// Extracts an archive file (tar.gz or zip) to a destination directory.
  Future<String> extractArchive(
    String archivePath,
    String destinationDir, {
    String? expectedBinaryName,
  }) async {
    final file = File(archivePath);
    if (!await file.exists()) {
      throw DownloadException('Archive file not found: $archivePath');
    }

    final bytes = await file.readAsBytes();
    final archive = _decodeArchive(archivePath, bytes);

    await Directory(destinationDir).create(recursive: true);

    String? binaryPath;

    for (final file in archive) {
      if (file.isFile) {
        final filename = file.name;
        final outputPath = p.join(destinationDir, filename);

        final outputFile = File(outputPath);
        await outputFile.create(recursive: true);
        await outputFile.writeAsBytes(file.content as List<int>);

        // Make executable on Unix-like systems
        if (!Platform.isWindows && (expectedBinaryName == null || filename == expectedBinaryName)) {
          await Process.run('chmod', ['+x', outputPath]);
          binaryPath = outputPath;
        } else if (Platform.isWindows && filename.endsWith('.exe')) {
          binaryPath = outputPath;
        }
      }
    }

    if (binaryPath == null) {
      throw DownloadException(
        'Binary not found in archive. Expected: $expectedBinaryName',
      );
    }

    return binaryPath;
  }

  Archive _decodeArchive(String path, List<int> bytes) {
    if (path.endsWith('.tar.gz') || path.endsWith('.tgz')) {
      // First decompress gzip, then untar
      final gzipDecoder = GZipDecoder();
      final tarBytes = gzipDecoder.decodeBytes(bytes);
      return TarDecoder().decodeBytes(tarBytes);
    } else if (path.endsWith('.zip')) {
      return ZipDecoder().decodeBytes(bytes);
    } else {
      throw DownloadException(
        'Unsupported archive format: $path. Supported formats: .tar.gz, .tgz, .zip',
      );
    }
  }

  /// Downloads and extracts a binary in one operation.
  Future<String> downloadAndExtract(
    String url,
    String cacheDir, {
    String? expectedBinaryName,
    ProgressCallback? onProgress,
    RetryCallback? onRetry,
    ChecksumCallback? onChecksum,
    List<String>? availableAssetUrls,
    bool skipChecksumValidation = false,
    Duration timeout = const Duration(minutes: 5),
  }) async {
    final archiveName = p.basename(Uri.parse(url).path);
    final archivePath = p.join(cacheDir, archiveName);

    // Download
    await downloadFile(
      url,
      archivePath,
      onProgress: onProgress,
      onRetry: onRetry,
      timeout: timeout,
    );

    // Validate checksum if not skipped
    if (!skipChecksumValidation && availableAssetUrls != null) {
      try {
        final validated = await checksumValidator.tryValidateFromGitHubAssets(
          archivePath,
          url,
          availableAssetUrls,
        );

        if (onChecksum != null) {
          // Get actual checksum for callback
          String? checksum;
          if (validated) {
            checksum = await checksumValidator.computeFileChecksum(archivePath);
          }
          onChecksum(validated, checksum);
        }

        if (!validated && availableAssetUrls.any((u) => u.contains('.sha256'))) {
          // Checksum file exists but validation failed - this is an error
          throw DownloadException(
            'Checksum validation failed for $archiveName. File may be corrupted or tampered with.',
          );
        }
      } on ChecksumValidationException {
        // Clean up downloaded file on checksum failure
        await File(archivePath).delete();
        rethrow;
      }
    }

    // Extract
    final binaryPath = await extractArchive(
      archivePath,
      cacheDir,
      expectedBinaryName: expectedBinaryName,
    );

    // Clean up archive
    await File(archivePath).delete();

    return binaryPath;
  }
}

/// Exception thrown when download or extraction fails.
class DownloadException implements Exception {
  final String message;

  DownloadException(this.message);

  @override
  String toString() => 'DownloadException: $message';
}
