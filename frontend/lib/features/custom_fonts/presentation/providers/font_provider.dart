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

  FontProvider() {
    // Listen to font changes from service
    _fontService.currentFont.addListener(_onFontChanged);
  }

  void _onFontChanged() {
    notifyListeners();
  }

  /// Load a custom font from file
  Future<void> loadCustomFont({
    required Uint8List fontData,
    required String fileName,
    required String familyName,
  }) async {
    _setLoading(true);
    _clearError();

    try {
      await _fontService.saveCustomFont(
        fontData: fontData,
        fileName: fileName,
        familyName: familyName,
      );
    } catch (e) {
      _setError('Failed to load font: $e');
      rethrow;
    } finally {
      _setLoading(false);
    }
  }

  /// Remove custom font and revert to system font
  Future<void> removeCustomFont() async {
    _setLoading(true);
    _clearError();

    try {
      await _fontService.removeCustomFont();
    } catch (e) {
      _setError('Failed to remove font: $e');
      rethrow;
    } finally {
      _setLoading(false);
    }
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
