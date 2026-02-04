import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:uuid/uuid.dart';
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
  static const _uuid = Uuid();
  late final FontRepository _repository;
  late final LoadCustomFontUseCase _loadUseCase;
  late final SaveCustomFontUseCase _saveUseCase;
  late final RemoveCustomFontUseCase _removeUseCase;

  /// Current loaded custom font (null if using system font)
  final ValueNotifier<CustomFont?> currentFont = ValueNotifier(null);

  /// Pending font for preview (not yet saved)
  CustomFont? _pendingFont;
  Uint8List? _pendingFontData;
  bool _pendingRemove = false;

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

  /// Load font for preview only (register in memory, don't save to storage)
  /// Call commitPendingFont() to save permanently
  Future<CustomFont> loadFontForPreview({
    required Uint8List fontData,
    required String fileName,
    required String familyName,
  }) async {
    if (fontData.isEmpty) {
      throw ArgumentError('Font data cannot be empty');
    }
    if (familyName.isEmpty) {
      throw ArgumentError('Font family name cannot be empty');
    }

    // Create entity for preview
    final font = CustomFont(
      id: _uuid.v4(),
      name: fileName,
      familyName: familyName,
      uploadedAt: DateTime.now(),
      sizeInBytes: fontData.length,
    );

    // Register font in Flutter (memory only)
    final loader = FontLoader(familyName);
    loader.addFont(Future.value(ByteData.sublistView(fontData)));
    await loader.load();

    // Save as pending
    _pendingFont = font;
    _pendingFontData = fontData;
    _pendingRemove = false;

    return font;
  }

  /// Get pending font (for preview in settings)
  CustomFont? get pendingFont => _pendingFont;

  /// Check if there's a pending remove operation
  bool get hasPendingRemove => _pendingRemove;

  /// Check if there are any pending changes
  bool get hasPendingChanges => _pendingFont != null || _pendingRemove;

  /// Mark current font for removal (will be removed on commit)
  void markFontForRemoval() {
    _pendingRemove = true;
    _pendingFont = null;
    _pendingFontData = null;
  }

  /// Commit pending changes (save font or remove)
  Future<void> commitPendingChanges() async {
    if (_pendingRemove) {
      // Remove current font
      await _removeUseCase.execute();
      currentFont.value = null;
      _pendingRemove = false;
    } else if (_pendingFont != null && _pendingFontData != null) {
      // Remove old font if exists
      if (currentFont.value != null) {
        await _removeUseCase.execute();
      }
      // Save new font
      await _repository.saveFontData(_pendingFont!.id, _pendingFontData!);
      await _repository.saveCustomFont(_pendingFont!);
      currentFont.value = _pendingFont;
    }
    // Clear pending state
    _pendingFont = null;
    _pendingFontData = null;
  }

  /// Cancel pending changes
  void cancelPendingChanges() {
    _pendingFont = null;
    _pendingFontData = null;
    _pendingRemove = false;
  }

  /// Save and load a new custom font
  /// Returns the saved font metadata
  @Deprecated('Use loadFontForPreview + commitPendingChanges instead')
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
  @Deprecated('Use markFontForRemoval + commitPendingChanges instead')
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
