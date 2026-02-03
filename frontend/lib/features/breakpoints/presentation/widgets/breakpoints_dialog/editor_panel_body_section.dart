import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../../compose/presentation/widgets/body_editor.dart';
import 'editor_panel_content_chips.dart';

class EditorPanelBodySection extends StatelessWidget {
  const EditorPanelBodySection({
    super.key,
    required this.contentType,
    required this.reencode,
    required this.isBinary,
    required this.isTruncated,
    required this.mode,
    required this.onModeChanged,
    required this.rawController,
    required this.jsonController,
  });

  final String? contentType;
  final bool reencode;
  final bool isBinary;
  final bool isTruncated;

  final String mode;
  final ValueChanged<String> onModeChanged;
  final TextEditingController rawController;
  final TextEditingController jsonController;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: context.appColors.border),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(8),
            child: EditorPanelContentChips(
              contentType: contentType,
              reencode: reencode,
              isBinary: isBinary,
              isTruncated: isTruncated,
            ),
          ),
          if (isBinary || isTruncated)
            Padding(
              padding: const EdgeInsets.all(8),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Wrap(
                  spacing: 8,
                  children: [
                    if (isBinary)
                      const Chip(label: Text('Binary body — editing disabled')),
                    if (isTruncated)
                      const Chip(label: Text('Truncated preview')),
                  ],
                ),
              ),
            ),
          Expanded(
            child: (isBinary || isTruncated)
                ? const Center(child: Text('Body preview is not editable'))
                : BodyEditor(
                    mode: mode,
                    onModeChanged: onModeChanged,
                    rawCtrl: rawController,
                    jsonCtrl: jsonController,
                    form: const [],
                    multipart: const [],
                    allowedModes: const ['raw', 'json'],
                  ),
          ),
        ],
      ),
    );
  }
}
