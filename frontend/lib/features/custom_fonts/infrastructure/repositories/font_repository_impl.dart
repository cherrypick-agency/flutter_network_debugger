import 'dart:convert';
import 'dart:typed_data';
import 'package:shared_preferences/shared_preferences.dart';
import '../../domain/entities/custom_font.dart';
import '../../domain/repositories/font_repository.dart';
import '../persistence/font_storage_adapter.dart';

/// Implementation of FontRepository
/// Following Dependency Inversion Principle - depends on abstractions
class FontRepositoryImpl implements FontRepository {
  final FontStorageAdapter storageAdapter;
  static const String _metadataKey = 'custom_font_metadata';

  FontRepositoryImpl({required this.storageAdapter});

  @override
  Future<Uint8List?> loadFontData(String fontId) async {
    return await storageAdapter.loadFontFile(fontId);
  }

  @override
  Future<void> saveFontData(String fontId, Uint8List data) async {
    await storageAdapter.saveFontFile(fontId, data);
  }

  @override
  Future<CustomFont?> getCustomFont() async {
    final prefs = await SharedPreferences.getInstance();
    final jsonString = prefs.getString(_metadataKey);
    if (jsonString == null) {
      return null;
    }

    try {
      final json = jsonDecode(jsonString) as Map<String, dynamic>;
      return CustomFont(
        id: json['id'] as String,
        name: json['name'] as String,
        familyName: json['familyName'] as String,
        uploadedAt: DateTime.parse(json['uploadedAt'] as String),
        sizeInBytes: json['sizeInBytes'] as int,
      );
    } catch (e) {
      // Invalid metadata, clean up
      await prefs.remove(_metadataKey);
      return null;
    }
  }

  @override
  Future<void> saveCustomFont(CustomFont font) async {
    final prefs = await SharedPreferences.getInstance();
    final json = {
      'id': font.id,
      'name': font.name,
      'familyName': font.familyName,
      'uploadedAt': font.uploadedAt.toIso8601String(),
      'sizeInBytes': font.sizeInBytes,
    };
    await prefs.setString(_metadataKey, jsonEncode(json));
  }

  @override
  Future<void> removeCustomFont() async {
    final font = await getCustomFont();
    if (font != null) {
      await storageAdapter.deleteFontFile(font.id);
    }

    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_metadataKey);
  }

  @override
  Future<bool> hasCustomFont() async {
    final font = await getCustomFont();
    return font != null;
  }
}
