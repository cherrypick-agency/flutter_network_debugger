import 'dart:typed_data';

/// Abstract storage adapter for font files
/// Following Interface Segregation Principle and Strategy Pattern
abstract class FontStorageAdapter {
  /// Save font file data
  Future<void> saveFontFile(String fontId, Uint8List data);

  /// Load font file data
  Future<Uint8List?> loadFontFile(String fontId);

  /// Delete font file
  Future<void> deleteFontFile(String fontId);

  /// Check if font file exists
  Future<bool> fontFileExists(String fontId);

  /// Get storage path for font (may not be applicable for web)
  Future<String?> getFontPath(String fontId);
}
