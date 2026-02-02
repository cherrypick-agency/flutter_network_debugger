import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';

class StringListEditor extends StatefulWidget {
  const StringListEditor({
    required this.label,
    required this.value,
    required this.onChanged,
    this.hint,
    this.normalize,
    super.key,
  });

  final String label;
  final String? hint;
  final List<String> value;
  final ValueChanged<List<String>> onChanged;
  final String Function(String)? normalize;

  @override
  State<StringListEditor> createState() => _StringListEditorState();
}

class _StringListEditorState extends State<StringListEditor> {
  late final TextEditingController _ctrl;
  late final FocusNode _focus;

  @override
  void initState() {
    super.initState();
    _ctrl = TextEditingController(text: widget.value.join(','));
    _focus = FocusNode();
  }

  @override
  void didUpdateWidget(covariant StringListEditor oldWidget) {
    super.didUpdateWidget(oldWidget);
    final next = widget.value.join(',');
    if (!_focus.hasFocus && _ctrl.text != next) _ctrl.text = next;
  }

  @override
  void dispose() {
    _ctrl.dispose();
    _focus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: _ctrl,
      focusNode: _focus,
      decoration: InputDecoration(
        labelText: widget.label,
        hintText: widget.hint,
        suffixIcon: widget.value.isEmpty
            ? null
            : IconButton(
                tooltip: 'Clear',
                onPressed: () {
                  _ctrl.text = '';
                  widget.onChanged(const <String>[]);
                },
                icon: const Icon(Icons.clear),
              ),
      ),
      onChanged: (v) {
        final parts = v
            .split(',')
            .map((e) => e.trim())
            .where((e) => e.isNotEmpty)
            .map((e) => widget.normalize?.call(e) ?? e)
            .where((e) => e.isNotEmpty)
            .toList(growable: false);
        widget.onChanged(parts);
      },
      style: context.appText.body,
    );
  }
}
