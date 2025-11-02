import 'package:flutter/material.dart';

import 'kv.dart';

class KeyValueEditor extends StatefulWidget {
  const KeyValueEditor({
    required this.items,
    required this.onChanged,
    required this.labelKey,
    super.key,
  });

  final List<KvPair> items;
  final ValueChanged<List<KvPair>> onChanged;
  final String labelKey;

  @override
  State<KeyValueEditor> createState() => _KeyValueEditorState();
}

class _KeyValueEditorState extends State<KeyValueEditor> {
  // Простой показ ошибок: пустые ключи и дубликаты
  Set<String> _dupKeys = <String>{};

  @override
  Widget build(BuildContext context) {
    _recalcDuplicateKeys();
    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        children: [
          // Список делаем shrinkWrap, чтобы блок растягивался по содержимому,
          // а прокрутка была общей у страницы, а не локальной у списка
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: widget.items.length,
            itemBuilder: (context, index) {
              final item = widget.items[index];
              final keyCtrl = TextEditingController(text: item.key);
              final valCtrl = TextEditingController(text: item.value);

              final isKeyEmpty = item.key.trim().isEmpty;
              final isDup = item.key.isNotEmpty && _dupKeys.contains(item.key);

              return Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: keyCtrl,
                        decoration: InputDecoration(
                          labelText: '${widget.labelKey} key',
                          errorText:
                              isKeyEmpty
                                  ? 'Пустой ключ'
                                  : (isDup ? 'Дубликат ключа' : null),
                        ),
                        onChanged: (v) {
                          widget.items[index] = widget.items[index].copyWith(
                            key: v,
                          );
                          setState(() {});
                          widget.onChanged(widget.items);
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: TextField(
                        controller: valCtrl,
                        decoration: const InputDecoration(labelText: 'Value'),
                        onChanged: (v) {
                          widget.items[index] = widget.items[index].copyWith(
                            value: v,
                          );
                          widget.onChanged(widget.items);
                        },
                      ),
                    ),
                    IconButton(
                      onPressed: () {
                        setState(() => widget.items.removeAt(index));
                        widget.onChanged(widget.items);
                      },
                      icon: const Icon(Icons.delete_outline),
                    ),
                  ],
                ),
              );
            },
          ),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton.icon(
              onPressed: () {
                setState(() => widget.items.add(const KvPair('', '')));
                widget.onChanged(widget.items);
              },
              icon: const Icon(Icons.add),
              label: Text('Add ${widget.labelKey.toLowerCase()}'),
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
