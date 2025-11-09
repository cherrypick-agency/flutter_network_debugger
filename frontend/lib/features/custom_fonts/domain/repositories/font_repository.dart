import 'dart:typed_data';
import '../entities/custom_font.dart';

/// Repository interface for custom font operations (Domain Layer)
/// Following Dependency Inversion Principle - domain defines the contract
abstract class FontRepository {
  /// Load font data from storage
  Future<Uint8List?> loadFontData(String fontId);

  /// Save font data to storage
  Future<void> saveFontData(String fontId, Uint8List data);

  /// Get metadata about the currently loaded custom font
  Future<CustomFont?> getCustomFont();

  /// Save custom font metadata
  Future<void> saveCustomFont(CustomFont font);

  /// Remove custom font and its data
  Future<void> removeCustomFont();

  /// Check if a custom font is currently loaded
  Future<bool> hasCustomFont();
}
