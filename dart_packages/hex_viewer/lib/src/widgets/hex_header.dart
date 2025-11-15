import 'package:flutter/material.dart';
import '../models/hex_config.dart';
import '../formatting/hex_formatter.dart';
import '../utils/column_width_calculator.dart';

/// Header row showing column labels (OFFSET | HEX | ASCII)
class HexHeader extends StatelessWidget {
  final HexConfig config;
  final TextStyle monoStyle;

  const HexHeader({super.key, required this.config, required this.monoStyle});

  @override
  Widget build(BuildContext context) {
    final headerStyle = monoStyle.copyWith(
      fontWeight: FontWeight.w600,
      color: Theme.of(context).colorScheme.onSurfaceVariant,
    );

    // Calculate dynamic flex ratio based on bytes per line and grouping
    final (hexFlex, asciiFlex) = ColumnWidthCalculator.calculateFlexRatio(
      bytesPerLine: config.bytesPerLine,
      groupSize: config.groupSize,
    );

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      color: Theme.of(context).colorScheme.surfaceContainerHighest,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Offset column header
          if (config.showOffset)
            SizedBox(width: 100, child: Text('OFFSET', style: headerStyle)),

          if (config.showOffset) const SizedBox(width: 16),

          // Hex bytes header (show byte positions 00-0F or 00-07 depending on bytesPerLine)
          Expanded(
            flex: hexFlex,
            child: Text(_buildHexHeader(), style: headerStyle),
          ),

          const SizedBox(width: 16),

          // ASCII column header
          if (config.showAscii)
            Expanded(
              flex: asciiFlex,
              child: Text('ASCII', style: headerStyle),
            ),
        ],
      ),
    );
  }

  String _buildHexHeader() {
    final parts = <String>[];
    for (int i = 0; i < config.bytesPerLine; i++) {
      parts.add(HexFormatter.formatByte(i));

      // Add group separator
      if (config.groupSize > 0 &&
          (i + 1) % config.groupSize == 0 &&
          i < config.bytesPerLine - 1) {
        parts.add('|');
      }
    }
    return parts.join(' ');
  }
}
