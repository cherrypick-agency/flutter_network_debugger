import 'package:flutter/foundation.dart';
import '../../application/font_service.dart';
import '../../domain/entities/custom_font.dart';

/// Provider for custom font UI state management
/// Following MVVM pattern and Separation of Concerns
class FontProvider extends ChangeNotifier {
  final FontService _fontService = FontService();

  bool _isLoading = false;
  String? _error;

  bool get isLoading => _isLoading;
  String? get error => _error;
  CustomFont? get currentFont => _fontService.currentFont.value;
  bool get hasCustomFont => _fontService.hasCustomFont();
  String? get fontFamily => _fontService.getCurrentFontFamily();

  /// Pending font for preview (not yet saved)
  CustomFont? get pendingFont => _fontService.pendingFont;

  /// Check if font marked for removal
  bool get hasPendingRemove => _fontService.hasPendingRemove;

  /// Check if there are pending changes
  bool get hasPendingChanges => _fontService.hasPendingChanges;

  /// Effective font family for preview (pending or current)
  String? get previewFontFamily {
    if (_fontService.hasPendingRemove) return null;
    return _fontService.pendingFont?.familyName ?? currentFont?.familyName;
  }

  /// Effective font for display (pending or current)
  CustomFont? get effectiveFont {
    if (_fontService.hasPendingRemove) return null;
    return _fontService.pendingFont ?? currentFont;
  }

  /// Check if effectively has a custom font (for preview)
  bool get effectivelyHasCustomFont {
    if (_fontService.hasPendingRemove) return false;
    return _fontService.pendingFont != null || hasCustomFont;
  }

  FontProvider() {
    _fontService.currentFont.addListener(_onFontChanged);
  }

  void _onFontChanged() {
    notifyListeners();
  }

  /// Load font for preview (not saved yet)
  Future<void> loadFontForPreview({
    required Uint8List fontData,
    required String fileName,
    required String familyName,
  }) async {
    _setLoading(true);
    _clearError();

    try {
      await _fontService.loadFontForPreview(
        fontData: fontData,
        fileName: fileName,
        familyName: familyName,
      );
      notifyListeners();
    } catch (e) {
      _setError('Failed to load font: $e');
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  /// Mark font for removal (will be removed on commit)
  void markFontForRemoval() {
    _fontService.markFontForRemoval();
    notifyListeners();
  }

  /// Commit pending changes (save or remove font)
  Future<void> commitPendingChanges() async {
    _setLoading(true);
    _clearError();

    try {
      await _fontService.commitPendingChanges();
      notifyListeners();
    } catch (e) {
      _setError('Failed to save font changes: $e');
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  /// Cancel pending changes
  void cancelPendingChanges() {
    _fontService.cancelPendingChanges();
    notifyListeners();
  }

  void _setLoading(bool value) {
    _isLoading = value;
    notifyListeners();
  }

  void _setError(String message) {
    _error = message;
    notifyListeners();
  }

  void _clearError() {
    _error = null;
    notifyListeners();
  }

  @override
  void dispose() {
    _fontService.currentFont.removeListener(_onFontChanged);
    super.dispose();
  }
}
