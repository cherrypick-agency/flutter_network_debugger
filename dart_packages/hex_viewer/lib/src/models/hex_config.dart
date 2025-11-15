import 'byte_color_scheme.dart';

/// Configuration for hex viewer display and behavior.
///
/// This class controls all visual and interactive aspects of the [HexViewer] widget,
/// including layout, colors, and enabled features.
///
/// Example:
/// ```dart
/// const config = HexConfig(
///   bytesPerLine: 16,
///   groupSize: 8,
///   showOffset: true,
///   showAscii: true,
///   enableSelection: true,
///   enableCopy: true,
///   colorScheme: ByteColorScheme.fromTheme(theme),
/// );
/// ```
///
/// See also:
/// - [HexViewer] which uses this configuration
/// - [ByteColorScheme] for color customization
class HexConfig {
  /// Number of bytes displayed per line.
  ///
  /// Common values:
  /// - 8: Compact view for narrow screens
  /// - 16: Standard hex editor layout (default)
  /// - 32: Wide view for large displays
  ///
  /// Must be positive. Defaults to 16.
  final int bytesPerLine;

  /// Size of byte groups for visual separation.
  ///
  /// Controls how bytes are grouped with separators:
  /// - 0: No grouping (`48 65 6C 6C 6F`)
  /// - 4: Four-byte groups (`48656C6C | 6F20576F`)
  /// - 8: Eight-byte groups (default, `48 65 6C 6C 6F 20 57 6F | 72 6C 64`)
  ///
  /// Must be 0 or less than [bytesPerLine]. Defaults to 8.
  ///
  /// Example with groupSize=4 and bytesPerLine=16:
  /// ```
  /// 48656C6C | 6F20576F | 726C6421 | 0A000000
  /// ```
  final int groupSize;

  /// Whether to display the offset column on the left.
  ///
  /// When true, shows addresses like `00000000`, `00000010`, etc.
  /// Defaults to true.
  final bool showOffset;

  /// Whether to display the ASCII column on the right.
  ///
  /// When true, shows ASCII representation of bytes (printable characters
  /// or `·` for non-printable). Defaults to true.
  final bool showAscii;

  /// Whether to enable interactive byte selection.
  ///
  /// When true, users can:
  /// - Click bytes to select them
  /// - Shift+click to extend selection
  /// - See selection info in status bar
  ///
  /// Defaults to true.
  final bool enableSelection;

  /// Whether to enable copy to clipboard functionality.
  ///
  /// When true, users can copy selected bytes in:
  /// - Hex format (`48 65 6C 6C 6F`)
  /// - Text format (`Hello`)
  /// - Binary format (raw bytes)
  ///
  /// Requires [enableSelection] to be true. Defaults to true.
  final bool enableCopy;

  /// Color scheme for different byte types.
  ///
  /// Controls the visual appearance of:
  /// - Null bytes (0x00)
  /// - Printable ASCII (0x20-0x7E)
  /// - Control characters (0x01-0x1F)
  /// - Extended ASCII (0x80+)
  ///
  /// Use [ByteColorScheme.fromTheme] to automatically adapt to the app theme.
  /// Defaults to [ByteColorScheme.standard].
  final ByteColorScheme colorScheme;

  /// Creates a hex viewer configuration.
  ///
  /// All parameters are optional with sensible defaults suitable for most use cases.
  ///
  /// Throws [AssertionError] if:
  /// - [bytesPerLine] is not positive
  /// - [groupSize] is negative
  /// - [groupSize] exceeds [bytesPerLine]
  const HexConfig({
    this.bytesPerLine = 16,
    this.groupSize = 8,
    this.showOffset = true,
    this.showAscii = true,
    this.enableSelection = true,
    this.enableCopy = true,
    this.colorScheme = const ByteColorScheme.standard(),
  }) : assert(bytesPerLine > 0, 'bytesPerLine must be positive'),
       assert(groupSize >= 0, 'groupSize must be non-negative'),
       assert(
         groupSize == 0 || groupSize <= bytesPerLine,
         'groupSize cannot exceed bytesPerLine',
       );

  /// Creates a copy of this configuration with the given fields replaced.
  ///
  /// Example:
  /// ```dart
  /// final newConfig = config.copyWith(
  ///   bytesPerLine: 32,
  ///   groupSize: 16,
  /// );
  /// ```
  HexConfig copyWith({
    int? bytesPerLine,
    int? groupSize,
    bool? showOffset,
    bool? showAscii,
    bool? enableSelection,
    bool? enableCopy,
    ByteColorScheme? colorScheme,
  }) {
    return HexConfig(
      bytesPerLine: bytesPerLine ?? this.bytesPerLine,
      groupSize: groupSize ?? this.groupSize,
      showOffset: showOffset ?? this.showOffset,
      showAscii: showAscii ?? this.showAscii,
      enableSelection: enableSelection ?? this.enableSelection,
      enableCopy: enableCopy ?? this.enableCopy,
      colorScheme: colorScheme ?? this.colorScheme,
    );
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is HexConfig &&
        other.bytesPerLine == bytesPerLine &&
        other.groupSize == groupSize &&
        other.showOffset == showOffset &&
        other.showAscii == showAscii &&
        other.enableSelection == enableSelection &&
        other.enableCopy == enableCopy &&
        other.colorScheme == colorScheme;
  }

  @override
  int get hashCode {
    return Object.hash(
      bytesPerLine,
      groupSize,
      showOffset,
      showAscii,
      enableSelection,
      enableCopy,
      colorScheme,
    );
  }
}
