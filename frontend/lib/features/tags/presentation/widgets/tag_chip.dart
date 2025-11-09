import 'package:flutter/material.dart';

class TagChip extends StatelessWidget {
  const TagChip({required this.label, this.color, this.onDeleted, super.key});

  final String label;
  final Color? color;
  final VoidCallback? onDeleted;

  @override
  Widget build(BuildContext context) {
    final chipColor = color ?? Colors.grey;

    return Chip(
      label: Text(
        label,
        style: TextStyle(color: _getTextColor(chipColor), fontSize: 12),
      ),
      backgroundColor: chipColor.withAlpha((0.2 * 255).round()),
      side: BorderSide(color: chipColor, width: 1),
      deleteIcon: onDeleted != null
          ? Icon(Icons.close, size: 16, color: _getTextColor(chipColor))
          : null,
      onDeleted: onDeleted,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }

  Color _getTextColor(Color background) {
    return background.computeLuminance() > 0.5 ? Colors.black87 : Colors.white;
  }
}
