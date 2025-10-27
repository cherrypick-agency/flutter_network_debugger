import 'dart:io';
import 'package:crypto/crypto.dart';
import 'package:http/http.dart' as http;

/// Validates file checksums using SHA256.
class ChecksumValidator {
  final http.Client? _client;

  ChecksumValidator({http.Client? client}) : _client = client;

  http.Client get client => _client ?? http.Client();

  /// Computes SHA256 checksum of a file.
  ///
  /// Returns the checksum as a lowercase hexadecimal string.
  Future<String> computeFileChecksum(String filePath) async {
    final file = File(filePath);
    if (!await file.exists()) {
      throw ChecksumValidationException('File not found: $filePath');
    }

    try {
      final bytes = await file.readAsBytes();
      final digest = sha256.convert(bytes);
      return digest.toString();
    } catch (e) {
      throw ChecksumValidationException(
        'Failed to compute checksum for $filePath: $e',
      );
    }
  }

  /// Validates file checksum against expected value.
  ///
  /// Returns true if checksums match, false otherwise.
  Future<bool> validateFile(String filePath, String expectedChecksum) async {
    final actualChecksum = await computeFileChecksum(filePath);
    final expected = expectedChecksum.trim().toLowerCase();
    return actualChecksum == expected;
  }

  /// Downloads checksum file from URL.
  ///
  /// Many GitHub releases include `.sha256` files alongside binaries.
  /// This method downloads such a file and extracts the checksum.
  ///
  /// Expected format:
  /// - Single line with checksum: `abc123...`
  /// - Or with filename: `abc123...  filename.tar.gz`
  Future<String> downloadChecksumFile(String checksumUrl) async {
    try {
      final response = await client.get(Uri.parse(checksumUrl)).timeout(
        const Duration(seconds: 30),
        onTimeout: () {
          throw ChecksumValidationException(
            'Timeout downloading checksum from $checksumUrl',
          );
        },
      );

      if (response.statusCode != 200) {
        throw ChecksumValidationException(
          'Failed to download checksum: ${response.statusCode} ${response.reasonPhrase}',
        );
      }

      // Parse checksum from response
      final content = response.body.trim();

      // Handle empty response
      if (content.isEmpty) {
        throw ChecksumValidationException(
          'Empty checksum file from $checksumUrl',
        );
      }

      // Extract checksum (handle both "checksum" and "checksum  filename" formats)
      final parts = content.split(RegExp(r'\s+'));
      final checksum = parts[0].toLowerCase();

      // Validate checksum format (SHA256 is 64 hex characters)
      if (!RegExp(r'^[a-f0-9]{64}$').hasMatch(checksum)) {
        throw ChecksumValidationException(
          'Invalid checksum format: $checksum',
        );
      }

      return checksum;
    } catch (e) {
      if (e is ChecksumValidationException) rethrow;
      throw ChecksumValidationException(
        'Failed to download or parse checksum from $checksumUrl: $e',
      );
    }
  }

  /// Validates file against checksum URL.
  ///
  /// Convenience method that downloads the checksum file and validates.
  Future<bool> validateFileFromUrl(
    String filePath,
    String checksumUrl,
  ) async {
    final expectedChecksum = await downloadChecksumFile(checksumUrl);
    return validateFile(filePath, expectedChecksum);
  }

  /// Tries to find and validate checksum from GitHub release assets.
  ///
  /// Looks for common checksum file patterns:
  /// - `filename.sha256`
  /// - `filename.tar.gz.sha256`
  /// - `SHA256SUMS`
  ///
  /// Returns true if checksum found and valid, false if checksum not found.
  /// Throws if checksum found but validation fails.
  Future<bool> tryValidateFromGitHubAssets(
    String filePath,
    String archiveUrl,
    List<String> availableAssetUrls,
  ) async {
    // Try common checksum file patterns
    final checksumPatterns = [
      '$archiveUrl.sha256',
      '${archiveUrl.replaceAll('.tar.gz', '')}.sha256',
      '${archiveUrl.replaceAll('.zip', '')}.sha256',
    ];

    // Find matching checksum URL
    String? checksumUrl;
    for (final pattern in checksumPatterns) {
      if (availableAssetUrls.any((url) => url == pattern)) {
        checksumUrl = pattern;
        break;
      }
    }

    // If no checksum file found, return false (not an error)
    if (checksumUrl == null) {
      return false;
    }

    // Validate using found checksum
    try {
      final isValid = await validateFileFromUrl(filePath, checksumUrl);
      if (!isValid) {
        throw ChecksumValidationException(
          'Checksum validation failed for $filePath',
        );
      }
      return true;
    } catch (e) {
      if (e is ChecksumValidationException) rethrow;
      throw ChecksumValidationException(
        'Error validating checksum: $e',
      );
    }
  }
}

/// Exception thrown when checksum validation fails.
class ChecksumValidationException implements Exception {
  final String message;

  ChecksumValidationException(this.message);

  @override
  String toString() => 'ChecksumValidationException: $message';
}
