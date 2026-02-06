/// Shared ZIP validation utility
/// Constants synced with backend: internal/infrastructure/httpapi/script_handlers.go
class ZipValidator {
  static const maxCompressedSizeBytes = 10 * 1024 * 1024; // 10MB (multipart form limit)
  static const maxUncompressedSizeBytes = 5 * 1024 * 1024; // 5MB (backend uncompressed limit)
  static const maxSingleFileSizeBytes = 500 * 1024; // 500KB per file in ZIP

  /// Validate ZIP file bytes (size and magic number)
  static void validate(List<int> bytes) {
    if (bytes.length > maxCompressedSizeBytes) {
      throw ArgumentError('File is too large. Maximum compressed size is 10MB.');
    }
    if (bytes.length < 4 || bytes[0] != 0x50 || bytes[1] != 0x4B) {
      throw ArgumentError('Invalid ZIP file. Please select a valid ZIP archive.');
    }
  }
}
