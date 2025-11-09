import 'dart:io';
import 'dart:typed_data';
import 'package:path_provider/path_provider.dart';
import 'font_storage_adapter.dart';

/// File system based storage adapter for mobile and desktop platforms
/// Following Single Responsibility Principle
class FontStorageFile implements FontStorageAdapter {
  static const String _fontsDirName = 'custom_fonts';

  @override
  Future<void> saveFontFile(String fontId, Uint8List data) async {
    final file = await _getFontFile(fontId);
    await file.parent.create(recursive: true);
    await file.writeAsBytes(data);
  }

  @override
  Future<Uint8List?> loadFontFile(String fontId) async {
    final file = await _getFontFile(fontId);
    if (!await file.exists()) {
      return null;
    }
    return await file.readAsBytes();
  }

  @override
  Future<void> deleteFontFile(String fontId) async {
    final file = await _getFontFile(fontId);
    if (await file.exists()) {
      await file.delete();
    }
  }

  @override
  Future<bool> fontFileExists(String fontId) async {
    final file = await _getFontFile(fontId);
    return await file.exists();
  }

  @override
  Future<String?> getFontPath(String fontId) async {
    final file = await _getFontFile(fontId);
    return file.path;
  }

  Future<File> _getFontFile(String fontId) async {
    final appDir = await getApplicationDocumentsDirectory();
    final fontsDir = Directory('${appDir.path}/$_fontsDirName');
    return File('${fontsDir.path}/$fontId.ttf');
  }
}
