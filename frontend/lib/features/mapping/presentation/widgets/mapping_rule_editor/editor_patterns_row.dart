import 'package:flutter/material.dart';

class MappingRuleEditorPatternsRow extends StatelessWidget {
  const MappingRuleEditorPatternsRow({
    super.key,
    required this.hostPatternController,
    required this.pathPatternController,
    required this.saving,
    required this.hostErrorText,
    required this.pathErrorText,
    required this.onChanged,
  });

  final TextEditingController hostPatternController;
  final TextEditingController pathPatternController;
  final bool saving;

  final String? hostErrorText;
  final String? pathErrorText;

  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: hostPatternController,
            enabled: !saving,
            decoration: InputDecoration(
              labelText: 'Host pattern (glob/regex)',
              errorText: hostErrorText,
            ),
            onChanged: onChanged,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: TextField(
            controller: pathPatternController,
            enabled: !saving,
            decoration: InputDecoration(
              labelText: 'Path pattern (glob/regex)',
              errorText: pathErrorText,
            ),
            onChanged: onChanged,
          ),
        ),
      ],
    );
  }
}
