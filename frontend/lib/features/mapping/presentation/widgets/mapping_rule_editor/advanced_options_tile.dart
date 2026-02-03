import 'package:flutter/material.dart';

class MappingRuleEditorAdvancedOptionsTile extends StatelessWidget {
  const MappingRuleEditorAdvancedOptionsTile({
    super.key,
    required this.stopProcessing,
    required this.saving,
    required this.onStopProcessingChanged,
  });

  final bool stopProcessing;
  final bool saving;
  final ValueChanged<bool> onStopProcessingChanged;

  @override
  Widget build(BuildContext context) {
    return ExpansionTile(
      title: const Text('Advanced'),
      subtitle: const Text('Rarely needed, but sometimes helpful'),
      childrenPadding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
      children: [
        SwitchListTile(
          dense: true,
          contentPadding: EdgeInsets.zero,
          title: const Text('Stop processing'),
          subtitle: const Text(
            'If enabled, the first matching rule stops further processing.',
          ),
          value: stopProcessing,
          onChanged: saving ? null : onStopProcessingChanged,
        ),
      ],
    );
  }
}
