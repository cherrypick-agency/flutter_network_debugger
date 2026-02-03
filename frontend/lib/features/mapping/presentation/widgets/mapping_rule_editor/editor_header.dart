import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';

class MappingRuleEditorHeader extends StatelessWidget {
  const MappingRuleEditorHeader({
    super.key,
    required this.isNew,
    required this.saving,
    required this.onClose,
  });

  final bool isNew;
  final bool saving;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Text(
            isNew ? 'New Mapping Rule' : 'Edit Mapping Rule',
            style: context.appText.title,
          ),
        ),
        IconButton(
          onPressed: saving ? null : onClose,
          icon: const Icon(Icons.close),
        ),
      ],
    );
  }
}
