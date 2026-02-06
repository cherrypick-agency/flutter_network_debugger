import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

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
  late final FocusNode _priorityFocus;
  late int _priorityLastValid;

  @override
  void initState() {
    super.initState();
    _priorityCtrl = TextEditingController(
      text: widget.rule.priority.toString(),
    );
    _priorityFocus = FocusNode();
    _priorityLastValid = widget.rule.priority;

    _priorityFocus.addListener(() {
      if (_priorityFocus.hasFocus) return;
      final n = int.tryParse(_priorityCtrl.text.trim());
      if (n != null) {
        _priorityLastValid = n;
        return;
      }
      _priorityCtrl.text = _priorityLastValid.toString();
      _emit(priority: _priorityLastValid);
      setState(() {});
    });
  }

  @override
  void didUpdateWidget(covariant InterceptRuleRow oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!_priorityFocus.hasFocus &&
        _priorityCtrl.text != widget.rule.priority.toString()) {
      _priorityCtrl.text = widget.rule.priority.toString();
      _priorityLastValid = widget.rule.priority;
    }
  }

  @override
  void dispose() {
    _priorityCtrl.dispose();
    _priorityFocus.dispose();
    super.dispose();
  }

  void _emit({bool? enabled, int? priority, String? action}) {
    widget.onChanged(
      widget.rule.copyWith(
        enabled: enabled ?? widget.rule.enabled,
        priority: priority ?? widget.rule.priority,
        action: action ?? widget.rule.action,
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
              focusNode: _priorityFocus,
              decoration: const InputDecoration(labelText: 'Priority'),
              keyboardType: TextInputType.number,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              onChanged: (v) {
                final n = int.tryParse(v.trim());
                if (n == null) return;
                _priorityLastValid = n;
                _emit(priority: n);
              },
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 140,
            child: DropdownButtonFormField<String>(
              key: ValueKey('action-${widget.rule.action}'),
              initialValue: widget.rule.action,
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
          const Spacer(),
          widget.trailing,
        ],
      ),
    );
  }
}
