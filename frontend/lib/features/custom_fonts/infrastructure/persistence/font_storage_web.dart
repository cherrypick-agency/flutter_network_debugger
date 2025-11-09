import 'dart:convert';
import 'dart:typed_data';
import 'package:shared_preferences/shared_preferences.dart';
import 'font_storage_adapter.dart';

/// Web-based storage adapter using SharedPreferences (IndexedDB)
/// Following Single Responsibility Principle
class FontStorageWeb implements FontStorageAdapter {
  static const String _keyPrefix = 'custom_font_';

  @override
  Future<void> saveFontFile(String fontId, Uint8List data) async {
    final prefs = await SharedPreferences.getInstance();
    final base64Data = base64Encode(data);

    try {
      await prefs.setString(_getKey(fontId), base64Data);
    } catch (e) {
      // Check for quota exceeded error
      final errorMsg = e.toString().toLowerCase();
      if (errorMsg.contains('quota') ||
          errorMsg.contains('storage') ||
          errorMsg.contains('exceeded')) {
        final sizeInMB = (data.length / (1024 * 1024)).toStringAsFixed(1);
        throw Exception(
          'Storage quota exceeded. Font file ($sizeInMB MB) is too large for web storage. '
          'Try using a smaller font file.',
        );
      }
      rethrow;
    }
  }

  @override
  Future<Uint8List?> loadFontFile(String fontId) async {
    final prefs = await SharedPreferences.getInstance();
    final base64Data = prefs.getString(_getKey(fontId));
    if (base64Data == null) {
      return null;
    }
    return base64Decode(base64Data);
  }

  @override
  Future<void> deleteFontFile(String fontId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_getKey(fontId));
  }

  @override
  Future<bool> fontFileExists(String fontId) async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.containsKey(_getKey(fontId));
  }

  @override
  Future<String?> getFontPath(String fontId) async {
    // Web doesn't have file paths
    return null;
  }

  String _getKey(String fontId) => '$_keyPrefix$fontId';
}
