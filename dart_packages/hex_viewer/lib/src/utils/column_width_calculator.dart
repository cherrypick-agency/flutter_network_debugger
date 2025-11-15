/// Utility for calculating optimal column width ratios in hex viewer.
///
/// This class calculates the flex ratio between hex and ASCII columns
/// based on the actual content requirements (bytes per line and grouping).
class ColumnWidthCalculator {
  /// Calculates optimal flex ratio for hex and ASCII columns.
  ///
  /// Returns a tuple of (hexFlex, asciiFlex) that can be used with [Expanded]
  /// widgets to properly distribute space between columns.
  ///
  /// The calculation is based on actual content requirements:
  /// - Hex column: Each byte needs 3 character widths (2 hex chars + 1 space)
  /// - ASCII column: Each byte needs 1 character width
  /// - Group separators ' | ' add 3 characters each
  /// - Spaces at end of groups are subtracted
  ///
  /// Example:
  /// ```dart
  /// final (hexFlex, asciiFlex) = ColumnWidthCalculator.calculateFlexRatio(
  ///   bytesPerLine: 16,
  ///   groupSize: 8,
  /// );
  /// // Returns (49, 16) for 16 bytes with 8-byte grouping (~3.06:1 ratio)
  /// ```
  static (int hexFlex, int asciiFlex) calculateFlexRatio({
    required int bytesPerLine,
    required int groupSize,
  }) {
    // Calculate relative width needs in character units

    // Hex column: each byte displayed as 2 hex chars + 1 space = 3 characters
    double hexWidth = bytesPerLine * 3.0;

    // Adjust for group separators
    if (groupSize > 0 && bytesPerLine > groupSize) {
      final numGroups = (bytesPerLine / groupSize).ceil();

      // Subtract spaces not needed at end of each group
      hexWidth -= numGroups;

      // Add width for group separators: ' | ' (3 characters each)
      final numSeparators = numGroups - 1;
      hexWidth += numSeparators * 3.0;
    }

    // ASCII column: each byte needs exactly 1 character width
    double asciiWidth = bytesPerLine.toDouble();

    // Convert to integer flex values
    final hexFlex = hexWidth.round();
    final asciiFlex = asciiWidth.round();

    // Simplify ratio by finding GCD
    final gcd = _gcd(hexFlex, asciiFlex);

    return (hexFlex ~/ gcd, asciiFlex ~/ gcd);
  }

  /// Calculates greatest common divisor for simplifying flex ratios.
  static int _gcd(int a, int b) {
    while (b != 0) {
      final temp = b;
      b = a % b;
      a = temp;
    }
    return a;
  }
}
