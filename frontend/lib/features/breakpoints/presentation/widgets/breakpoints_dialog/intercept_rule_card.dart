import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_rule.dart';
import 'intercept_rule_row.dart';
import 'intercept_rule_options.dart';
import 'intercept_when_editor.dart';

class InterceptRuleCard extends StatefulWidget {
  const InterceptRuleCard({
    required this.rule,
    required this.onChanged,
    required this.onDelete,
    super.key,
  });

  final InterceptRule rule;
  final ValueChanged<InterceptRule> onChanged;
  final VoidCallback onDelete;

  @override
  State<InterceptRuleCard> createState() => _InterceptRuleCardState();
}

class _InterceptRuleCardState extends State<InterceptRuleCard> {
  bool _expanded = false;

  void _emitOptions(InterceptRule next) {
    widget.onChanged(next);
  }

  void _emitWhen(InterceptWhen w) {
    widget.onChanged(
      InterceptRule(
        id: widget.rule.id,
        enabled: widget.rule.enabled,
        priority: widget.rule.priority,
        action: widget.rule.action,
        once: widget.rule.once,
        stopProcessing: widget.rule.stopProcessing,
        when: w,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: context.appColors.border),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        child: Column(
          children: [
            InterceptRuleRow(
              rule: widget.rule,
              onChanged: widget.onChanged,
              onDelete: widget.onDelete,
              trailing: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    tooltip: _expanded ? 'Hide conditions' : 'Edit conditions',
                    onPressed: () => setState(() => _expanded = !_expanded),
                    icon: Icon(
                      _expanded ? Icons.expand_less : Icons.expand_more,
                    ),
                  ),
                  IconButton(
                    onPressed: widget.onDelete,
                    icon: const Icon(Icons.delete_outline),
                    tooltip: 'Delete',
                  ),
                ],
              ),
            ),
            if (_expanded) ...[
              const Divider(height: 1),
              const SizedBox(height: 8),
              InterceptRuleOptions(rule: widget.rule, onChanged: _emitOptions),
              const SizedBox(height: 12),
              InterceptWhenEditor(
                value: widget.rule.when,
                onChanged: _emitWhen,
              ),
              const SizedBox(height: 8),
            ],
          ],
        ),
      ),
    );
  }
}
