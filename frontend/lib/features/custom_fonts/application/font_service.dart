import 'package:flutter/foundation.dart';
import '../domain/entities/custom_font.dart';
import '../domain/repositories/font_repository.dart';
import '../domain/usecases/load_custom_font.dart';
import '../domain/usecases/save_custom_font.dart';
import '../domain/usecases/remove_custom_font.dart';
import '../infrastructure/persistence/font_storage_factory.dart';
import '../infrastructure/repositories/font_repository_impl.dart';

/// Application service for managing custom fonts
/// Following Facade Pattern and Single Responsibility Principle
class FontService {
  static FontService? _instance;
  late final FontRepository _repository;
  late final LoadCustomFontUseCase _loadUseCase;
  late final SaveCustomFontUseCase _saveUseCase;
  late final RemoveCustomFontUseCase _removeUseCase;

  /// Current loaded custom font (null if using system font)
  final ValueNotifier<CustomFont?> currentFont = ValueNotifier(null);

  /// Factory constructor for singleton
  factory FontService() {
    _instance ??= FontService._internal();
    return _instance!;
  }

  FontService._internal() {
    _repository = FontRepositoryImpl(
      storageAdapter: FontStorageFactory.create(),
    );
    _loadUseCase = LoadCustomFontUseCase(repository: _repository);
    _saveUseCase = SaveCustomFontUseCase(repository: _repository);
    _removeUseCase = RemoveCustomFontUseCase(repository: _repository);
  }

  /// Initialize service and load custom font if exists
  Future<void> initialize() async {
    try {
      final font = await _loadUseCase.execute();
      currentFont.value = font;
    } catch (e) {
      debugPrint('Failed to load custom font: $e');
      currentFont.value = null;
    }
  }

  /// Save and load a new custom font
  /// Returns the saved font metadata
  Future<CustomFont> saveCustomFont({
    required Uint8List fontData,
    required String fileName,
    required String familyName,
  }) async {
    // Remove old font if exists
    if (currentFont.value != null) {
      await _removeUseCase.execute();
    }

    // Save new font
    final font = await _saveUseCase.execute(
      fontData: fontData,
      fileName: fileName,
      familyName: familyName,
    );

    currentFont.value = font;
    return font;
  }

  /// Remove custom font and revert to system font
  Future<void> removeCustomFont() async {
    await _removeUseCase.execute();
    currentFont.value = null;
  }

  /// Check if a custom font is currently loaded
  bool hasCustomFont() {
    return currentFont.value != null;
  }

  /// Get the current font family name
  /// Returns null if using system font
  String? getCurrentFontFamily() {
    return currentFont.value?.familyName;
  }

  /// Dispose resources
  /// Note: This method is not called in practice as FontService is a singleton
  /// that lives for the entire application lifetime. The ValueNotifier will be
  /// disposed when the app terminates.
  void dispose() {
    currentFont.dispose();
  }
}
