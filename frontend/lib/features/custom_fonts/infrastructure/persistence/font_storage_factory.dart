import 'package:flutter/foundation.dart' show kIsWeb;
import 'font_storage_adapter.dart';
import 'font_storage_file.dart';
import 'font_storage_web.dart';

/// Factory for creating platform-specific storage adapters
/// Following Factory Pattern and Open/Closed Principle
class FontStorageFactory {
  /// Create appropriate storage adapter based on current platform
  static FontStorageAdapter create() {
    if (kIsWeb) {
      return FontStorageWeb();
    } else {
      return FontStorageFile();
    }
  }
}
