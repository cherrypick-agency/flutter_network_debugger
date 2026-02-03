import 'package:flutter/material.dart';

import 'method_preset_chips.dart';

class MappingRuleEditorMethodsRow extends StatelessWidget {
  const MappingRuleEditorMethodsRow({
    super.key,
    required this.methodsController,
    required this.patternType,
    required this.saving,
    required this.methodsErrorText,
    required this.patternTypeErrorText,
    required this.onMethodsChanged,
    required this.onPatternTypeChanged,
  });

  final TextEditingController methodsController;
  final String patternType;
  final bool saving;

  final String? methodsErrorText;
  final String? patternTypeErrorText;

  final ValueChanged<String> onMethodsChanged;
  final ValueChanged<String> onPatternTypeChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: methodsController,
                enabled: !saving,
                decoration: InputDecoration(
                  labelText: 'HTTP methods',
                  hintText: 'empty = any; e.g. GET,POST',
                  errorText: methodsErrorText,
                ),
                onChanged: onMethodsChanged,
              ),
            ),
            const SizedBox(width: 12),
            SizedBox(
              width: 160,
              child: DropdownButtonFormField<String>(
                value: patternType,
                isDense: true,
                items: const [
                  DropdownMenuItem(value: 'glob', child: Text('glob')),
                  DropdownMenuItem(value: 'regex', child: Text('regex')),
                ],
                onChanged: saving
                    ? null
                    : (v) {
                        onPatternTypeChanged(v ?? 'glob');
                      },
                decoration: InputDecoration(
                  labelText: 'Pattern type',
                  errorText: patternTypeErrorText,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        MappingRuleEditorMethodPresetChips(
          controller: methodsController,
          enabled: !saving,
          onChanged: onMethodsChanged,
        ),
      ],
    );
  }
}
