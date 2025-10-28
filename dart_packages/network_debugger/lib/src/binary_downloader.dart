import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:archive/archive.dart';
import 'package:path/path.dart' as p;
import 'retry_helper.dart';
import 'checksum_validator.dart';
import 'logger.dart';

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
  final Logger _logger = Logger('BinaryDownloader');

  BinaryDownloader({
    http.Client? client,
    RetryHelper? retryHelper,
    ChecksumValidator? checksumValidator,
  })  : _client = client,
        retryHelper = retryHelper ?? RetryHelper(),
        checksumValidator =
            checksumValidator ?? ChecksumValidator(client: client);

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
    _logger.debug('Starting download: $url');
    _logger.debug('Destination: $destinationPath');

    final uri = Uri.parse(url);
    final request = http.Request('GET', uri);

    final response = await client.send(request).timeout(
      timeout,
      onTimeout: () {
        _logger.error('Download timeout after ${timeout.inMinutes} minutes');
        throw DownloadException(
          'Download timeout after ${timeout.inMinutes} minutes',
        );
      },
    );

    if (response.statusCode != 200) {
      _logger.error('Download failed with status ${response.statusCode}');
      throw DownloadException(
        'Failed to download from $url: ${response.statusCode} ${response.reasonPhrase}',
      );
    }

    final file = File(destinationPath);
    await file.parent.create(recursive: true);

    final total = response.contentLength ?? 0;
    var received = 0;

    if (total > 0) {
      _logger.info('Downloading ${_formatBytes(total)}');
    } else {
      _logger.info('Downloading (size unknown)');
    }

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
      _logger.error('Download incomplete: $received/$total bytes');
      throw DownloadException(
        'Download incomplete: received $received bytes, expected $total bytes',
      );
    }

    _logger.info('Download complete: ${_formatBytes(received)}');
  }

  /// Extracts an archive file (tar.gz or zip) to a destination directory.
  Future<String> extractArchive(
    String archivePath,
    String destinationDir, {
    String? expectedBinaryName,
  }) async {
    _logger.debug('Extracting archive: $archivePath');
    _logger.debug('Destination dir: $destinationDir');

    final file = File(archivePath);
    if (!await file.exists()) {
      _logger.error('Archive file not found: $archivePath');
      throw DownloadException('Archive file not found: $archivePath');
    }

    final bytes = await file.readAsBytes();
    _logger.debug('Archive size: ${_formatBytes(bytes.length)}');

    final archive = _decodeArchive(archivePath, bytes);
    _logger.info('Extracting ${archive.length} files from archive');

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
        if (!Platform.isWindows &&
            (expectedBinaryName == null || filename == expectedBinaryName)) {
          await Process.run('chmod', ['+x', outputPath]);
          binaryPath = outputPath;
        } else if (Platform.isWindows &&
            (filename.endsWith('.exe') ||
                (expectedBinaryName != null &&
                    filename == expectedBinaryName))) {
          binaryPath = outputPath;
        }
      }
    }

    if (binaryPath == null) {
      _logger
          .error('Binary not found in archive. Expected: $expectedBinaryName');
      throw DownloadException(
        'Binary not found in archive. Expected: $expectedBinaryName',
      );
    }

    _logger.info('Extraction complete. Binary: $binaryPath');
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
    _logger.info('Starting download and extract from: $url');

    final archiveName = p.basename(Uri.parse(url).path);
    final archivePath = p.join(cacheDir, archiveName);

    // Download
    _logger.debug('Downloading to: $archivePath');
    await downloadFile(
      url,
      archivePath,
      onProgress: onProgress,
      onRetry: onRetry,
      timeout: timeout,
    );

    // Validate checksum if not skipped
    if (!skipChecksumValidation && availableAssetUrls != null) {
      _logger.debug('Validating checksum...');
      try {
        final validated = await checksumValidator.tryValidateFromGitHubAssets(
          archivePath,
          url,
          availableAssetUrls,
        );

        if (validated) {
          _logger.info('Checksum validation: PASSED');
        } else {
          _logger
              .warning('Checksum validation: SKIPPED (no .sha256 file found)');
        }

        if (onChecksum != null) {
          // Get actual checksum for callback
          String? checksum;
          if (validated) {
            checksum = await checksumValidator.computeFileChecksum(archivePath);
          }
          onChecksum(validated, checksum);
        }

        if (!validated &&
            availableAssetUrls.any((u) => u.contains('.sha256'))) {
          // Checksum file exists but validation failed - this is an error
          _logger.error('Checksum validation FAILED');
          throw DownloadException(
            'Checksum validation failed for $archiveName. File may be corrupted or tampered with.',
          );
        }
      } on ChecksumValidationException {
        // Clean up downloaded file on checksum failure
        _logger.error('Checksum validation exception - cleaning up');
        await File(archivePath).delete();
        rethrow;
      }
    } else {
      _logger.warning('Checksum validation SKIPPED by user request');
    }

    // Extract
    final binaryPath = await extractArchive(
      archivePath,
      cacheDir,
      expectedBinaryName: expectedBinaryName,
    );

    // Clean up archive
    _logger.debug('Cleaning up archive file');
    await File(archivePath).delete();

    _logger.info('Download and extract complete');
    return binaryPath;
  }

  /// Formats bytes into human-readable string.
  String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }
}

/// Exception thrown when download or extraction fails.
class DownloadException implements Exception {
  final String message;

  DownloadException(this.message);

  @override
  String toString() => 'DownloadException: $message';
}
