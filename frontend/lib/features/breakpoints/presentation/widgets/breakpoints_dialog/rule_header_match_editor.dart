import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_rule.dart';
import 'rule_string_match_editor.dart';

class RuleHeaderMatchEditor extends StatelessWidget {
  const RuleHeaderMatchEditor({
    required this.value,
    required this.onChanged,
    super.key,
  });

  final RuleHeaderMatch? value;
  final ValueChanged<RuleHeaderMatch?> onChanged;

  RuleHeaderMatch? _build({
    required RuleStringMatch? name,
    required RuleStringMatch? val,
  }) {
    if (name == null) return null;
    return RuleHeaderMatch(name: name, value: val);
  }

  @override
  Widget build(BuildContext context) {
    final name = value?.name;
    final val = value?.value;

    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: context.appColors.border),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text('Header match', style: context.appText.subtitle),
                ),
                if (value != null)
                  TextButton.icon(
                    onPressed: () => onChanged(null),
                    icon: const Icon(Icons.clear),
                    label: const Text('Clear'),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            RuleStringMatchEditor(
              label: 'Name',
              value: name,
              onChanged: (n) => onChanged(_build(name: n, val: val)),
            ),
            const SizedBox(height: 12),
            RuleStringMatchEditor(
              label: 'Value (optional)',
              value: val,
              onChanged: (v) => onChanged(_build(name: name, val: v)),
            ),
          ],
        ),
      ),
    );
  }
}
