import 'package:flutter/material.dart';

import '../../../domain/entities/intercept_rule.dart';

class InterceptRuleRow extends StatefulWidget {
  const InterceptRuleRow({
    required this.rule,
    required this.onChanged,
    required this.onDelete,
    required this.trailing,
    super.key,
  });

  final InterceptRule rule;
  final ValueChanged<InterceptRule> onChanged;
  final VoidCallback onDelete;
  final Widget trailing;

  @override
  State<InterceptRuleRow> createState() => _InterceptRuleRowState();
}

class _InterceptRuleRowState extends State<InterceptRuleRow> {
  late final TextEditingController _priorityCtrl;

  @override
  void initState() {
    super.initState();
    _priorityCtrl = TextEditingController(
      text: widget.rule.priority.toString(),
    );
  }

  @override
  void didUpdateWidget(covariant InterceptRuleRow oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_priorityCtrl.text != widget.rule.priority.toString()) {
      _priorityCtrl.text = widget.rule.priority.toString();
    }
  }

  @override
  void dispose() {
    _priorityCtrl.dispose();
    super.dispose();
  }

  void _emit({
    bool? enabled,
    int? priority,
    String? action,
    bool? once,
    bool? stopProcessing,
  }) {
    widget.onChanged(
      InterceptRule(
        id: widget.rule.id,
        enabled: enabled ?? widget.rule.enabled,
        priority: priority ?? widget.rule.priority,
        action: action ?? widget.rule.action,
        once: once ?? widget.rule.once,
        stopProcessing: stopProcessing ?? widget.rule.stopProcessing,
        when: widget.rule.when,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Checkbox(
            value: widget.rule.enabled,
            onChanged: (v) => _emit(enabled: v ?? widget.rule.enabled),
          ),
          SizedBox(
            width: 90,
            child: TextField(
              controller: _priorityCtrl,
              decoration: const InputDecoration(labelText: 'Priority'),
              keyboardType: TextInputType.number,
              onChanged: (v) {
                final n = int.tryParse(v.trim());
                if (n == null) return;
                _emit(priority: n);
              },
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 140,
            child: DropdownButtonFormField<String>(
              value: widget.rule.action,
              decoration: const InputDecoration(labelText: 'Action'),
              onChanged: (v) => _emit(action: v ?? widget.rule.action),
              items: const [
                'request',
                'response',
                'both',
              ].map((e) => DropdownMenuItem(value: e, child: Text(e))).toList(),
            ),
          ),
          const SizedBox(width: 8),
          FilterChip(
            selected: widget.rule.once,
            label: const Text('Once'),
            onSelected: (v) => _emit(once: v),
          ),
          const SizedBox(width: 8),
          FilterChip(
            selected: widget.rule.stopProcessing,
            label: const Text('Stop'),
            onSelected: (v) => _emit(stopProcessing: v),
          ),
          const Spacer(),
          widget.trailing,
        ],
      ),
    );
  }
}
