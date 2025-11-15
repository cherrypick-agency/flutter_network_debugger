import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/byte_selection.dart';
import '../models/byte_color_scheme.dart';
import '../models/hex_config.dart';
import '../models/hex_line.dart';
import '../formatting/hex_formatter.dart';
import '../utils/column_width_calculator.dart';

/// Single row in hex viewer displaying one line of data
class HexRow extends StatefulWidget {
  final HexLine line;
  final HexConfig config;
  final ByteSelection? selection;
  final TextStyle monoStyle;
  final void Function(int offset, {bool isShiftHeld})? onBytePressed;

  const HexRow({
    super.key,
    required this.line,
    required this.config,
    required this.selection,
    required this.monoStyle,
    this.onBytePressed,
  });

  @override
  State<HexRow> createState() => _HexRowState();
}

class _HexRowState extends State<HexRow> {
  // Track which byte is hovered (null if none)
  int? _hoveredOffset;

  @override
  Widget build(BuildContext context) {
    // Calculate dynamic flex ratio based on bytes per line and grouping
    final (hexFlex, asciiFlex) = ColumnWidthCalculator.calculateFlexRatio(
      bytesPerLine: widget.config.bytesPerLine,
      groupSize: widget.config.groupSize,
    );

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Offset column
          if (widget.config.showOffset)
            SizedBox(
              width: 100,
              child: SelectableText(
                HexFormatter.formatOffset(widget.line.offset),
                style: widget.monoStyle.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
              ),
            ),

          if (widget.config.showOffset) const SizedBox(width: 16),

          // Hex bytes
          Expanded(flex: hexFlex, child: _buildHexBytes(context)),

          const SizedBox(width: 16),

          // ASCII column
          if (widget.config.showAscii)
            Expanded(
              flex: asciiFlex,
              child: SelectableText(
                widget.line.asciiString,
                style: widget.monoStyle,
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildHexBytes(BuildContext context) {
    final widgets = <Widget>[];
    int groupByteCount = 0;

    for (int i = 0; i < widget.line.bytes.length; i++) {
      final byte = widget.line.bytes[i];
      final globalOffset = widget.line.offset + i;
      final isSelected = widget.selection?.contains(globalOffset) ?? false;
      final isHovered = _hoveredOffset == globalOffset;

      widgets.add(
        _HexByte(
          key: ValueKey(globalOffset),
          byte: byte,
          offset: globalOffset,
          isSelected: isSelected,
          isHovered: isHovered,
          colorScheme: widget.config.colorScheme,
          monoStyle: widget.monoStyle,
          onPressed: (isShiftHeld) {
            widget.onBytePressed?.call(globalOffset, isShiftHeld: isShiftHeld);
          },
          onHoverChanged: (hovered) {
            setState(() {
              _hoveredOffset = hovered ? globalOffset : null;
            });
          },
        ),
      );

      groupByteCount++;

      // Add spacing
      if (i < widget.line.bytes.length - 1) {
        // Check if we need group separator
        if (widget.config.groupSize > 0 &&
            groupByteCount == widget.config.groupSize) {
          widgets.add(
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: Text(
                '|',
                style: widget.monoStyle.copyWith(
                  color: Theme.of(context).dividerColor,
                ),
              ),
            ),
          );
          groupByteCount = 0;
        } else {
          widgets.add(const SizedBox(width: 4));
        }
      }
    }

    return Wrap(
      crossAxisAlignment: WrapCrossAlignment.center,
      children: widgets,
    );
  }
}

/// Individual byte display widget (now stateless for better performance)
class _HexByte extends StatelessWidget {
  final int byte;
  final int offset;
  final bool isSelected;
  final bool isHovered;
  final ByteColorScheme colorScheme;
  final TextStyle monoStyle;
  final void Function(bool isShiftHeld) onPressed;
  final void Function(bool hovered) onHoverChanged;

  const _HexByte({
    super.key,
    required this.byte,
    required this.offset,
    required this.isSelected,
    required this.isHovered,
    required this.colorScheme,
    required this.monoStyle,
    required this.onPressed,
    required this.onHoverChanged,
  });

  @override
  Widget build(BuildContext context) {
    final byteColor = colorScheme.getColorForByte(byte);

    return MouseRegion(
      onEnter: (_) => onHoverChanged(true),
      onExit: (_) => onHoverChanged(false),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () {
          final isShiftHeld = HardwareKeyboard.instance.isShiftPressed;
          onPressed(isShiftHeld);
        },
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 2, vertical: 1),
          decoration: BoxDecoration(
            color: isSelected
                ? colorScheme.selectedBackground
                : isHovered
                ? Theme.of(context).colorScheme.primary.withValues(alpha: 0.1)
                : null,
            borderRadius: BorderRadius.circular(2),
          ),
          child: Text(
            HexFormatter.formatByte(byte),
            style: monoStyle.copyWith(
              color: isSelected ? colorScheme.selectedForeground : byteColor,
            ),
          ),
        ),
      ),
    );
  }
}
