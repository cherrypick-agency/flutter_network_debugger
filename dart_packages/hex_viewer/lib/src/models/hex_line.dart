/// Represents a single line in the hex viewer
class HexLine {
  /// Offset of the first byte in this line
  final int offset;

  /// Raw bytes for this line
  final List<int> bytes;

  /// Formatted hex string (e.g., "48 65 6C 6C 6F | 20 57 6F 72")
  final String hexString;

  /// ASCII representation (e.g., "Hello Wor")
  final String asciiString;

  const HexLine({
    required this.offset,
    required this.bytes,
    required this.hexString,
    required this.asciiString,
  });

  @override
  String toString() {
    return 'HexLine(offset: 0x${offset.toRadixString(16)}, bytes: ${bytes.length})';
  }
}
