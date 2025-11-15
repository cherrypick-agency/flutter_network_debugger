import 'dart:math';

/// Represents a selection of bytes in the hex viewer.
///
/// This class describes a range of selected bytes with inclusive start and end offsets.
/// It provides utilities for querying selection properties and extracting selected bytes.
///
/// Example:
/// ```dart
/// // Single byte selection
/// final selection = ByteSelection.single(10);
/// print(selection.length); // 1
///
/// // Range selection
/// final range = ByteSelection.range(5, 15);
/// print(range.length); // 11 bytes (5-15 inclusive)
/// print(range.contains(10)); // true
///
/// // Get selected bytes
/// final bytes = range.getBytes(myData);
/// ```
///
/// See also:
/// - [HexViewer.onSelectionChanged] which provides this object
class ByteSelection {
  /// Start offset of the selection (inclusive).
  ///
  /// This is the index of the first selected byte in the data array.
  /// For a single-byte selection, [startOffset] equals [endOffset].
  final int startOffset;

  /// End offset of the selection (inclusive).
  ///
  /// This is the index of the last selected byte in the data array.
  /// For a single-byte selection, [endOffset] equals [startOffset].
  final int endOffset;

  /// Creates a byte selection with the given offsets.
  ///
  /// Both offsets are inclusive. For selecting bytes 5 through 10,
  /// use `ByteSelection(startOffset: 5, endOffset: 10)`.
  ///
  /// Consider using [ByteSelection.single] or [ByteSelection.range]
  /// factory constructors for more convenient creation.
  const ByteSelection({required this.startOffset, required this.endOffset});

  /// Number of bytes in the selection.
  ///
  /// Returns the count of selected bytes. For inclusive range [0, 9],
  /// returns 10. Guaranteed to be non-negative.
  ///
  /// Example:
  /// ```dart
  /// ByteSelection(startOffset: 5, endOffset: 9).length; // 5 bytes
  /// ByteSelection.single(10).length; // 1 byte
  /// ```
  int get length =>
      (endOffset - startOffset + 1).clamp(0, double.infinity).toInt();

  /// Convenience getter for start offset.
  ///
  /// Alias for [startOffset] for more natural usage:
  /// ```dart
  /// print('Selected from ${selection.start} to ${selection.end}');
  /// ```
  int get start => startOffset;

  /// Convenience getter for end offset.
  ///
  /// Alias for [endOffset] for more natural usage:
  /// ```dart
  /// print('Selected from ${selection.start} to ${selection.end}');
  /// ```
  int get end => endOffset;

  /// Checks if the given offset is within this selection.
  ///
  /// Returns true if [offset] is between [startOffset] and [endOffset] (inclusive).
  ///
  /// Example:
  /// ```dart
  /// final sel = ByteSelection.range(5, 10);
  /// sel.contains(7);  // true
  /// sel.contains(5);  // true (inclusive)
  /// sel.contains(10); // true (inclusive)
  /// sel.contains(11); // false
  /// ```
  bool contains(int offset) {
    return offset >= startOffset && offset <= endOffset;
  }

  /// Extracts the selected bytes from the given data.
  ///
  /// Returns a sublist of [data] containing only the selected bytes.
  /// If the selection is out of bounds, returns an empty list or truncates
  /// to available data.
  ///
  /// Example:
  /// ```dart
  /// final data = [0x48, 0x65, 0x6C, 0x6C, 0x6F]; // "Hello"
  /// final sel = ByteSelection.range(0, 2);
  /// final bytes = sel.getBytes(data); // [0x48, 0x65, 0x6C]
  /// ```
  List<int> getBytes(List<int> data) {
    if (startOffset < 0 || startOffset >= data.length) {
      return [];
    }
    // Use min to avoid integer overflow when endOffset is near int.maxValue
    final end = min(endOffset + 1, data.length);
    return data.sublist(startOffset, end);
  }

  /// Creates a selection of a single byte at the given offset.
  ///
  /// This is equivalent to `ByteSelection(startOffset: offset, endOffset: offset)`.
  ///
  /// Example:
  /// ```dart
  /// final sel = ByteSelection.single(42);
  /// print(sel.length); // 1
  /// print(sel.contains(42)); // true
  /// ```
  factory ByteSelection.single(int offset) {
    return ByteSelection(startOffset: offset, endOffset: offset);
  }

  /// Creates a selection from a range of bytes.
  ///
  /// Automatically orders the offsets so [startOffset] ≤ [endOffset],
  /// regardless of the order of [start] and [end] parameters.
  ///
  /// Example:
  /// ```dart
  /// // Both create the same selection (5-10):
  /// final sel1 = ByteSelection.range(5, 10);
  /// final sel2 = ByteSelection.range(10, 5);
  /// print(sel1 == sel2); // true
  /// ```
  factory ByteSelection.range(int start, int end) {
    final actualStart = start < end ? start : end;
    final actualEnd = start < end ? end : start;
    return ByteSelection(startOffset: actualStart, endOffset: actualEnd);
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is ByteSelection &&
        other.startOffset == startOffset &&
        other.endOffset == endOffset;
  }

  @override
  int get hashCode => Object.hash(startOffset, endOffset);

  @override
  String toString() {
    if (startOffset == endOffset) {
      return 'ByteSelection(offset: 0x${startOffset.toRadixString(16)})';
    }
    return 'ByteSelection(0x${startOffset.toRadixString(16)} - 0x${endOffset.toRadixString(16)}, $length bytes)';
  }
}
