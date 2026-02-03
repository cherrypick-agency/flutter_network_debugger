import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';

class MappingRuleEditorMethodPresetChips extends StatefulWidget {
  const MappingRuleEditorMethodPresetChips({
    super.key,
    required this.controller,
    required this.enabled,
    required this.onChanged,
  });

  final TextEditingController controller;
  final bool enabled;
  final ValueChanged<String> onChanged;

  static const _presets = <String>[
    'GET',
    'POST',
    'PUT',
    'PATCH',
    'DELETE',
    'HEAD',
    'OPTIONS',
  ];

  @override
  State<MappingRuleEditorMethodPresetChips> createState() =>
      _MappingRuleEditorMethodPresetChipsState();
}

class _MappingRuleEditorMethodPresetChipsState
    extends State<MappingRuleEditorMethodPresetChips> {
  @override
  void initState() {
    super.initState();
    widget.controller.addListener(_onTextChanged);
  }

  @override
  void didUpdateWidget(covariant MappingRuleEditorMethodPresetChips oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.controller == widget.controller) return;
    oldWidget.controller.removeListener(_onTextChanged);
    widget.controller.addListener(_onTextChanged);
  }

  void _onTextChanged() {
    if (!mounted) return;
    setState(() {});
  }

  @override
  void dispose() {
    widget.controller.removeListener(_onTextChanged);
    super.dispose();
  }

  Set<String> _current() {
    return widget.controller.text
        .split(',')
        .map((e) => e.trim().toUpperCase())
        .where((e) => e.isNotEmpty)
        .toSet();
  }

  @override
  Widget build(BuildContext context) {
    final current = _current();
    final scheme = Theme.of(context).colorScheme;
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        for (final m in MappingRuleEditorMethodPresetChips._presets)
          FilterChip(
            selected: current.contains(m),
            showCheckmark: true,
            checkmarkColor: scheme.onSecondaryContainer,
            selectedColor: scheme.secondaryContainer,
            backgroundColor: scheme.surfaceContainerHighest,
            side: BorderSide(color: scheme.outlineVariant),
            onSelected: !widget.enabled
                ? null
                : (v) {
                    final next = _current();
                    if (v) {
                      next.add(m);
                    } else {
                      next.remove(m);
                    }
                    final text = next.toList()..sort();
                    widget.controller.text = text.join(',');
                    widget.onChanged(widget.controller.text);
                  },
            label: Text(m, style: context.appText.body),
          ),
      ],
    );
  }
}
