import 'package:flutter/material.dart';

/// Color scheme for syntax highlighting different byte types in the hex viewer.
///
/// This class defines colors for various byte categories to provide visual
/// cues about data content:
/// - Null bytes (0x00) - typically shown in gray
/// - Printable ASCII (0x20-0x7E) - regular text characters
/// - Control characters (0x01-0x1F) - shown in red to highlight special bytes
/// - Extended ASCII (0x80+) - shown in blue for non-standard characters
///
/// Example:
/// ```dart
/// // Use predefined schemes:
/// const lightScheme = ByteColorScheme.standard();
/// const darkScheme = ByteColorScheme.dark();
///
/// // Auto-adapt to app theme:
/// final scheme = ByteColorScheme.fromTheme(Theme.of(context));
///
/// // Custom colors:
/// const custom = ByteColorScheme(
///   nullByte: Colors.grey,
///   printableAscii: Colors.black,
///   controlChar: Colors.red,
///   extendedAscii: Colors.blue,
///   selectedBackground: Colors.blue,
///   selectedForeground: Colors.white,
/// );
/// ```
///
/// See also:
/// - [HexConfig.colorScheme] where this is used
class ByteColorScheme {
  /// Color for null bytes (0x00).
  ///
  /// Null bytes are often significant in binary data (string terminators,
  /// padding, etc.) and are highlighted differently. Typically gray.
  final Color nullByte;

  /// Color for printable ASCII characters (0x20-0x7E).
  ///
  /// This includes standard text characters: letters, digits, punctuation.
  /// Typically the default text color (black or white depending on theme).
  final Color printableAscii;

  /// Color for control characters (0x01-0x1F).
  ///
  /// Control characters like newline, tab, escape, etc. Shown in red
  /// to draw attention to special/non-printable bytes.
  final Color controlChar;

  /// Color for extended ASCII and high bytes (0x80+).
  ///
  /// Bytes outside standard ASCII range, including UTF-8 continuation bytes,
  /// special characters, or binary data. Typically blue.
  final Color extendedAscii;

  /// Background color for selected bytes.
  ///
  /// Used to highlight bytes when the user selects them. Typically a
  /// blue shade that contrasts with the foreground.
  final Color selectedBackground;

  /// Text color for selected bytes.
  ///
  /// Used for the hex and ASCII text of selected bytes. Must provide
  /// sufficient contrast against [selectedBackground]. Typically white.
  final Color selectedForeground;

  /// Creates a custom color scheme with explicit colors for each byte type.
  ///
  /// All parameters are required. Consider using [ByteColorScheme.standard],
  /// [ByteColorScheme.dark], or [ByteColorScheme.fromTheme] for predefined schemes.
  const ByteColorScheme({
    required this.nullByte,
    required this.printableAscii,
    required this.controlChar,
    required this.extendedAscii,
    required this.selectedBackground,
    required this.selectedForeground,
  });

  /// Standard color scheme optimized for light mode.
  ///
  /// Uses black text on white background with colored highlights:
  /// - Null bytes: Gray
  /// - Printable ASCII: Black
  /// - Control chars: Red (#D32F2F)
  /// - Extended ASCII: Blue (#1976D2)
  /// - Selection: Blue background (#2196F3) with white text
  const ByteColorScheme.standard()
    : nullByte = Colors.grey,
      printableAscii = Colors.black,
      controlChar = const Color(0xFFD32F2F),
      extendedAscii = const Color(0xFF1976D2),
      selectedBackground = const Color(0xFF2196F3),
      selectedForeground = Colors.white;

  /// Dark theme color scheme optimized for dark mode.
  ///
  /// Uses light text on dark background with colored highlights:
  /// - Null bytes: Medium gray (#757575)
  /// - Printable ASCII: Light gray (#E0E0E0)
  /// - Control chars: Light red (#EF5350)
  /// - Extended ASCII: Light blue (#64B5F6)
  /// - Selection: Dark blue background (#1976D2) with white text
  const ByteColorScheme.dark()
    : nullByte = const Color(0xFF757575),
      printableAscii = const Color(0xFFE0E0E0),
      controlChar = const Color(0xFFEF5350),
      extendedAscii = const Color(0xFF64B5F6),
      selectedBackground = const Color(0xFF1976D2),
      selectedForeground = Colors.white;

  /// Creates a color scheme that automatically adapts to the app's theme.
  ///
  /// Returns [ByteColorScheme.dark] for dark themes and [ByteColorScheme.standard]
  /// for light themes based on [ThemeData.brightness].
  ///
  /// Example:
  /// ```dart
  /// HexViewer(
  ///   data: myData,
  ///   config: HexConfig(
  ///     colorScheme: ByteColorScheme.fromTheme(Theme.of(context)),
  ///   ),
  /// )
  /// ```
  factory ByteColorScheme.fromTheme(ThemeData theme) {
    final isDark = theme.brightness == Brightness.dark;
    return isDark
        ? const ByteColorScheme.dark()
        : const ByteColorScheme.standard();
  }

  /// Returns the appropriate color for a specific byte value.
  ///
  /// Uses the byte value to determine which color category it falls into:
  /// - 0x00: [nullByte]
  /// - 0x20-0x7E: [printableAscii]
  /// - 0x01-0x1F: [controlChar]
  /// - 0x80+: [extendedAscii]
  ///
  /// Example:
  /// ```dart
  /// final scheme = ByteColorScheme.standard();
  /// final color = scheme.getColorForByte(0x41); // Returns printableAscii (for 'A')
  /// ```
  Color getColorForByte(int byte) {
    if (byte == 0x00) return nullByte;
    if (byte >= 0x20 && byte <= 0x7E) return printableAscii;
    if (byte >= 0x01 && byte <= 0x1F) return controlChar;
    return extendedAscii;
  }
}
