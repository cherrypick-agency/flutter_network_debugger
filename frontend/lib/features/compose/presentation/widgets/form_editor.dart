import 'package:flutter/material.dart';

import '../../domain/models.dart';

class FormEditor extends StatefulWidget {
  const FormEditor({super.key, required this.items, this.onChanged});
  final List<ComposeFormFieldDTO> items;
  final ValueChanged<List<ComposeFormFieldDTO>>? onChanged;

  @override
  State<FormEditor> createState() => _FormEditorState();
}

class _FormEditorState extends State<FormEditor> {
  Set<String> _dupKeys = <String>{};
  @override
  Widget build(BuildContext context) {
    _recalcDuplicateKeys();
    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        children: [
          Expanded(
            child: ListView.builder(
              itemCount: widget.items.length,
              itemBuilder: (context, index) {
                final item = widget.items[index];
                final k = TextEditingController(text: item.key);
                final v = TextEditingController(text: item.value);
                final keyEmpty = item.key.trim().isEmpty;
                final isDup =
                    item.key.isNotEmpty && _dupKeys.contains(item.key);
                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 4),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: k,
                          decoration: InputDecoration(
                            labelText: 'Key',
                            errorText: keyEmpty
                                ? 'Empty key'
                                : (isDup ? 'Duplicate key' : null),
                          ),
                          onChanged: (s) {
                            widget.items[index] = ComposeFormFieldDTO(
                              key: s,
                              value: item.value,
                            );
                            setState(() {});
                            widget.onChanged?.call(widget.items);
                          },
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: TextField(
                          controller: v,
                          decoration: const InputDecoration(labelText: 'Value'),
                          onChanged: (s) {
                            widget.items[index] = ComposeFormFieldDTO(
                              key: item.key,
                              value: s,
                            );
                            widget.onChanged?.call(widget.items);
                          },
                        ),
                      ),
                      IconButton(
                        onPressed: () {
                          setState(() => widget.items.removeAt(index));
                          widget.onChanged?.call(widget.items);
                        },
                        icon: const Icon(Icons.delete_outline),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              onPressed: () {
                setState(
                  () => widget.items.add(
                    const ComposeFormFieldDTO(key: '', value: ''),
                  ),
                );
                widget.onChanged?.call(widget.items);
              },
              icon: const Icon(Icons.add),
              label: const Text('Add field'),
            ),
          ),
        ],
      ),
    );
  }

  void _recalcDuplicateKeys() {
    final seen = <String, int>{};
    for (final i in widget.items) {
      if (i.key.isEmpty) continue;
      seen[i.key] = (seen[i.key] ?? 0) + 1;
    }
    _dupKeys = seen.entries.where((e) => e.value > 1).map((e) => e.key).toSet();
  }
}
