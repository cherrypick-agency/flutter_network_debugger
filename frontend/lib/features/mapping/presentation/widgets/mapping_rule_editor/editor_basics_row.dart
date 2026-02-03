import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class MappingRuleEditorBasicsRow extends StatelessWidget {
  const MappingRuleEditorBasicsRow({
    super.key,
    required this.kind,
    required this.priority,
    required this.enabled,
    required this.saving,
    required this.kindErrorText,
    required this.priorityErrorText,
    required this.onKindChanged,
    required this.onPriorityChanged,
    required this.onEnabledChanged,
  });

  final String kind;
  final int priority;
  final bool enabled;
  final bool saving;

  final String? kindErrorText;
  final String? priorityErrorText;

  final ValueChanged<String> onKindChanged;
  final ValueChanged<int> onPriorityChanged;
  final ValueChanged<bool> onEnabledChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        SizedBox(
          width: 220,
          child: SegmentedButton<String>(
            showSelectedIcon: false,
            segments: const [
              ButtonSegment<String>(
                value: 'local',
                label: Text('Local'),
                icon: Icon(Icons.insert_drive_file_outlined),
              ),
              ButtonSegment<String>(
                value: 'remote',
                label: Text('Remote'),
                icon: Icon(Icons.public),
              ),
            ],
            selected: {kind},
            onSelectionChanged: saving ? null : (v) => onKindChanged(v.first),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: TextFormField(
            enabled: !saving,
            decoration: InputDecoration(
              labelText: 'Priority',
              errorText: priorityErrorText,
            ),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            initialValue: priority.toString(),
            onChanged: (v) {
              final n = int.tryParse(v.trim());
              if (n != null) onPriorityChanged(n);
            },
          ),
        ),
        const SizedBox(width: 12),
        SizedBox(
          width: 160,
          child: SwitchListTile(
            contentPadding: EdgeInsets.zero,
            dense: true,
            title: const Text('Enabled'),
            value: enabled,
            onChanged: saving ? null : onEnabledChanged,
          ),
        ),
      ],
    );
  }
}
