import '../models/hex_config.dart';
import '../models/hex_line.dart';

/// Formats binary data as hexadecimal strings
class HexFormatter {
  /// Format a single byte as hex (e.g., 0x0A -> "0A")
  static String formatByte(int byte) {
    return byte.toRadixString(16).padLeft(2, '0').toUpperCase();
  }

  /// Format an offset as hex (e.g., 256 -> "00000100")
  static String formatOffset(int offset, {int width = 8}) {
    return offset.toRadixString(16).padLeft(width, '0').toUpperCase();
  }

  /// Convert byte to ASCII character or placeholder
  /// Returns middle dot (·) for non-printable characters
  static String formatAscii(int byte) {
    if (byte >= 0x20 && byte <= 0x7E) {
      return String.fromCharCode(byte);
    }
    return '·';
  }

  /// Format a complete line with optional grouping
  static HexLine formatLine(List<int> data, int offset, HexConfig config) {
    final end = (offset + config.bytesPerLine).clamp(0, data.length);
    if (offset >= data.length || offset < 0) {
      return HexLine(
        offset: offset,
        bytes: const [],
        hexString: '',
        asciiString: '',
      );
    }

    final bytes = data.sublist(offset, end);
    final hexParts = <String>[];
    final asciiParts = <String>[];

    for (int i = 0; i < bytes.length; i++) {
      hexParts.add(formatByte(bytes[i]));
      asciiParts.add(formatAscii(bytes[i]));

      // Add group separator (| between groups)
      if (config.groupSize > 0 &&
          (i + 1) % config.groupSize == 0 &&
          i < bytes.length - 1) {
        hexParts.add('|');
      }
    }

    return HexLine(
      offset: offset,
      bytes: bytes,
      hexString: hexParts.join(' '),
      asciiString: asciiParts.join(),
    );
  }

  /// Calculate total number of lines for given data
  static int calculateLineCount(int dataLength, int bytesPerLine) {
    if (dataLength == 0 || bytesPerLine <= 0) return 0;
    return (dataLength / bytesPerLine).ceil();
  }

  /// Format selection as hex string
  static String formatSelectionAsHex(List<int> bytes) {
    return bytes.map(formatByte).join(' ');
  }

  /// Format selection as raw binary (for clipboard)
  static String formatSelectionAsBinary(List<int> bytes) {
    return String.fromCharCodes(bytes);
  }

  /// Format selection as ASCII text (non-printable -> .)
  static String formatSelectionAsText(List<int> bytes) {
    return bytes.map(formatAscii).join();
  }
}
