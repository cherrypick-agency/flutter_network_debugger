import 'dart:typed_data';

/// Simple TTF/OTF font parser to extract font metadata
/// Supports reading font family name from the 'name' table
class TtfParser {
  /// Extract font family name from TTF/OTF font data
  /// Returns null if parsing fails
  static String? extractFontFamily(Uint8List fontData) {
    try {
      // Read offset table
      final sfntVersion = _readUint32(fontData, 0);

      // Check for valid font signature
      // TTF: 0x00010000, OTF: 'OTTO' (0x4F54544F), TTC: 'ttcf'
      if (sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F) {
        return null;
      }

      final numTables = _readUint16(fontData, 4);

      // Find the 'name' table
      int nameTableOffset = 0;
      int nameTableLength = 0;

      for (int i = 0; i < numTables; i++) {
        final tableEntryOffset = 12 + i * 16;
        final tag = _readTag(fontData, tableEntryOffset);

        if (tag == 'name') {
          nameTableOffset = _readUint32(fontData, tableEntryOffset + 8);
          nameTableLength = _readUint32(fontData, tableEntryOffset + 12);
          break;
        }
      }

      if (nameTableOffset == 0) {
        return null;
      }

      // Parse name table
      return _parseNameTable(fontData, nameTableOffset, nameTableLength);
    } catch (e) {
      return null;
    }
  }

  static String? _parseNameTable(Uint8List fontData, int offset, int length) {
    // final format = _readUint16(fontData, offset); // Not used for now
    final count = _readUint16(fontData, offset + 2);
    final stringOffset = _readUint16(fontData, offset + 4);

    String? fontFamily;

    // Look for name ID 1 (Font Family) or 16 (Typographic Family)
    for (int i = 0; i < count; i++) {
      final recordOffset = offset + 6 + i * 12;

      final platformID = _readUint16(fontData, recordOffset);
      final encodingID = _readUint16(fontData, recordOffset + 2);
      final nameID = _readUint16(fontData, recordOffset + 6);
      final nameLength = _readUint16(fontData, recordOffset + 8);
      final nameOffset = _readUint16(fontData, recordOffset + 10);

      // Look for Font Family name (ID 1) or Typographic Family (ID 16)
      // Prefer ID 16 if available, fallback to ID 1
      if (nameID == 1 || nameID == 16) {
        final actualOffset = offset + stringOffset + nameOffset;

        String? name;

        // Windows platform
        if (platformID == 3 && encodingID == 1) {
          // UTF-16 BE
          name = _readUtf16BE(fontData, actualOffset, nameLength);
        }
        // Macintosh platform
        else if (platformID == 1 && encodingID == 0) {
          // ASCII/MacRoman
          name = _readAscii(fontData, actualOffset, nameLength);
        }
        // Unicode platform
        else if (platformID == 0) {
          // UTF-16 BE
          name = _readUtf16BE(fontData, actualOffset, nameLength);
        }

        if (name != null && name.isNotEmpty) {
          // Prefer Typographic Family (ID 16) over Font Family (ID 1)
          if (nameID == 16) {
            return name;
          }
          fontFamily ??= name;
        }
      }
    }

    return fontFamily;
  }

  static int _readUint16(Uint8List data, int offset) {
    return (data[offset] << 8) | data[offset + 1];
  }

  static int _readUint32(Uint8List data, int offset) {
    return (data[offset] << 24) |
        (data[offset + 1] << 16) |
        (data[offset + 2] << 8) |
        data[offset + 3];
  }

  static String _readTag(Uint8List data, int offset) {
    return String.fromCharCodes(data.sublist(offset, offset + 4));
  }

  static String _readUtf16BE(Uint8List data, int offset, int length) {
    final chars = <int>[];
    for (int i = 0; i < length; i += 2) {
      final charCode = (data[offset + i] << 8) | data[offset + i + 1];
      chars.add(charCode);
    }
    return String.fromCharCodes(chars);
  }

  static String _readAscii(Uint8List data, int offset, int length) {
    return String.fromCharCodes(data.sublist(offset, offset + length));
  }
}
