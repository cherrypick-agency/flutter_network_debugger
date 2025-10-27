import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:archive/archive.dart';
import 'package:path/path.dart' as p;

/// Callback for download progress updates.
typedef ProgressCallback = void Function(int received, int total);

/// Downloads and extracts binary files.
class BinaryDownloader {
  final http.Client? _client;

  BinaryDownloader({http.Client? client}) : _client = client;

  http.Client get client => _client ?? http.Client();

  /// Downloads a file from [url] to [destinationPath] with optional progress tracking.
  Future<void> downloadFile(
    String url,
    String destinationPath, {
    ProgressCallback? onProgress,
  }) async {
    final uri = Uri.parse(url);
    final request = http.Request('GET', uri);
    final response = await client.send(request);

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
  }) async {
    final archiveName = p.basename(Uri.parse(url).path);
    final archivePath = p.join(cacheDir, archiveName);

    // Download
    await downloadFile(url, archivePath, onProgress: onProgress);

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
