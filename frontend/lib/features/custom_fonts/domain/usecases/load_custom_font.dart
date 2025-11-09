import 'package:flutter/services.dart';
import '../entities/custom_font.dart';
import '../repositories/font_repository.dart';

/// Use case for loading and registering a custom font
/// Follows Single Responsibility Principle
class LoadCustomFontUseCase {
  final FontRepository repository;

  LoadCustomFontUseCase({required this.repository});

  /// Execute the use case
  /// Returns the loaded font metadata if successful, null otherwise
  Future<CustomFont?> execute() async {
    try {
      final fontMetadata = await repository.getCustomFont();
      if (fontMetadata == null) {
        return null;
      }

      final fontData = await repository.loadFontData(fontMetadata.id);
      if (fontData == null) {
        // Metadata exists but data is missing - clean up
        await repository.removeCustomFont();
        return null;
      }

      // Register font with Flutter
      await _registerFont(fontMetadata.familyName, fontData);

      return fontMetadata;
    } catch (e) {
      // Clean up on error
      await repository.removeCustomFont();
      rethrow;
    }
  }

  Future<void> _registerFont(String familyName, Uint8List fontData) async {
    final loader = FontLoader(familyName);
    loader.addFont(Future.value(ByteData.sublistView(fontData)));
    await loader.load();
  }
}
