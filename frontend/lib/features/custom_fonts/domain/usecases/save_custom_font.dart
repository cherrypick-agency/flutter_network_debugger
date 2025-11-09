import 'package:flutter/services.dart';
import 'package:uuid/uuid.dart';
import '../entities/custom_font.dart';
import '../repositories/font_repository.dart';

/// Use case for saving and registering a custom font
/// Follows Single Responsibility Principle
class SaveCustomFontUseCase {
  static const _uuid = Uuid();
  final FontRepository repository;

  SaveCustomFontUseCase({required this.repository});

  /// Execute the use case
  /// [fontData] - raw font file bytes
  /// [fileName] - original file name
  /// [familyName] - font family name to register in Flutter
  Future<CustomFont> execute({
    required Uint8List fontData,
    required String fileName,
    required String familyName,
  }) async {
    // Validate font data
    if (fontData.isEmpty) {
      throw ArgumentError('Font data cannot be empty');
    }

    if (familyName.isEmpty) {
      throw ArgumentError('Font family name cannot be empty');
    }

    // Create font entity
    final customFont = CustomFont(
      id: _uuid.v4(),
      name: fileName,
      familyName: familyName,
      uploadedAt: DateTime.now(),
      sizeInBytes: fontData.length,
    );

    try {
      // Save font data and metadata
      await repository.saveFontData(customFont.id, fontData);
      await repository.saveCustomFont(customFont);

      // Register font with Flutter
      await _registerFont(familyName, fontData);

      return customFont;
    } catch (e) {
      // Rollback on error - clean up saved data
      try {
        await repository.removeCustomFont();
      } catch (_) {
        // Ignore cleanup errors
      }
      rethrow;
    }
  }

  Future<void> _registerFont(String familyName, Uint8List fontData) async {
    final loader = FontLoader(familyName);
    loader.addFont(Future.value(ByteData.sublistView(fontData)));
    await loader.load();
  }
}
