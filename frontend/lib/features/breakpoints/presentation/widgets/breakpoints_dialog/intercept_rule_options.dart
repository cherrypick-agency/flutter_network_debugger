import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_rule.dart';

class InterceptRuleOptions extends StatelessWidget {
  const InterceptRuleOptions({
    required this.rule,
    required this.onChanged,
    super.key,
  });

  final InterceptRule rule;
  final ValueChanged<InterceptRule> onChanged;

  @override
  Widget build(BuildContext context) {
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
            Text('Options', style: context.appText.subtitle),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                FilterChip(
                  selected: rule.once,
                  label: const Text('Once'),
                  onSelected: (v) => onChanged(rule.copyWith(once: v)),
                ),
                FilterChip(
                  selected: rule.stopProcessing,
                  label: const Text('Stop processing'),
                  onSelected: (v) => onChanged(rule.copyWith(stopProcessing: v)),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              'Once: after first match, the rule is disabled.\n'
              'Stop processing: do not evaluate rules below after a match.',
              style: context.appText.body.copyWith(
                color: context.appColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }

}
