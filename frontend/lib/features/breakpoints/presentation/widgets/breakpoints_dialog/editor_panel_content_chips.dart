import 'package:flutter/material.dart';

class EditorPanelContentChips extends StatelessWidget {
  const EditorPanelContentChips({
    super.key,
    required this.contentType,
    required this.reencode,
    required this.isBinary,
    required this.isTruncated,
  });

  final String? contentType;
  final bool reencode;
  final bool isBinary;
  final bool isTruncated;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      children: [
        if (contentType != null && contentType!.isNotEmpty)
          Chip(label: Text(contentType!.toLowerCase())),
        if (reencode) const Chip(label: Text('Re-encode')),
        if (isBinary) const Chip(label: Text('Binary')),
        if (isTruncated) const Chip(label: Text('Truncated')),
      ],
    );
  }
}
