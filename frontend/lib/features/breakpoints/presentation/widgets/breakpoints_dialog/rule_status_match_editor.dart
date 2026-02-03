import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_rule.dart';

class RuleStatusMatchEditor extends StatefulWidget {
  const RuleStatusMatchEditor({
    required this.value,
    required this.onChanged,
    super.key,
  });

  final RuleStatusMatch? value;
  final ValueChanged<RuleStatusMatch?> onChanged;

  @override
  State<RuleStatusMatchEditor> createState() => _RuleStatusMatchEditorState();
}

class _RuleStatusMatchEditorState extends State<RuleStatusMatchEditor> {
  late final TextEditingController _equalsCtrl;
  late final TextEditingController _fromCtrl;
  late final TextEditingController _toCtrl;
  late final FocusNode _equalsFocus;
  late final FocusNode _fromFocus;
  late final FocusNode _toFocus;
  String? _rangeError;

  @override
  void initState() {
    super.initState();
    _equalsCtrl = TextEditingController(
      text: (widget.value?.equals ?? const <int>[]).join(','),
    );
    _fromCtrl = TextEditingController(
      text: widget.value?.from?.toString() ?? '',
    );
    _toCtrl = TextEditingController(text: widget.value?.to?.toString() ?? '');
    _equalsFocus = FocusNode();
    _fromFocus = FocusNode();
    _toFocus = FocusNode();
    _revalidate();
  }

  @override
  void didUpdateWidget(covariant RuleStatusMatchEditor oldWidget) {
    super.didUpdateWidget(oldWidget);
    final v = widget.value;
    final equals = (v?.equals ?? const <int>[]).join(',');
    if (!_equalsFocus.hasFocus && _equalsCtrl.text != equals) {
      _equalsCtrl.text = equals;
    }
    final from = v?.from?.toString() ?? '';
    if (!_fromFocus.hasFocus && _fromCtrl.text != from) _fromCtrl.text = from;
    final to = v?.to?.toString() ?? '';
    if (!_toFocus.hasFocus && _toCtrl.text != to) _toCtrl.text = to;
    _revalidate();
  }

  @override
  void dispose() {
    _equalsCtrl.dispose();
    _fromCtrl.dispose();
    _toCtrl.dispose();
    _equalsFocus.dispose();
    _fromFocus.dispose();
    _toFocus.dispose();
    super.dispose();
  }

  List<int> _parseEquals(String v) {
    final out = <int>[];
    for (final part in v.split(',')) {
      final n = int.tryParse(part.trim());
      if (n == null) continue;
      if (n < 100 || n > 599) continue;
      out.add(n);
    }
    return out;
  }

  RuleStatusMatch? _build({
    required List<int> equals,
    required int? from,
    required int? to,
    required bool is4xx,
    required bool is5xx,
  }) {
    if (equals.isEmpty && from == null && to == null && !is4xx && !is5xx) {
      return null;
    }
    return RuleStatusMatch(
      equals: equals,
      from: from,
      to: to,
      is4xx: is4xx,
      is5xx: is5xx,
    );
  }

  void _emit() {
    final v = widget.value;
    final equals = _parseEquals(_equalsCtrl.text);
    final from0 = int.tryParse(_fromCtrl.text.trim());
    final to0 = int.tryParse(_toCtrl.text.trim());
    final from = (from0 == null || from0 == 0) ? null : from0;
    final to = (to0 == null || to0 == 0) ? null : to0;
    final is4xx = v?.is4xx ?? false;
    final is5xx = v?.is5xx ?? false;
    _revalidate();
    widget.onChanged(
      _build(equals: equals, from: from, to: to, is4xx: is4xx, is5xx: is5xx),
    );
  }

  void _revalidate() {
    final from = int.tryParse(_fromCtrl.text.trim());
    final to = int.tryParse(_toCtrl.text.trim());
    if (from != null && to != null && from > 0 && to > 0 && from > to) {
      final next = '"From" cannot be greater than "To"';
      if (_rangeError != next) setState(() => _rangeError = next);
      return;
    }
    if (_rangeError != null) setState(() => _rangeError = null);
  }

  @override
  Widget build(BuildContext context) {
    final v = widget.value;
    final is4xx = v?.is4xx ?? false;
    final is5xx = v?.is5xx ?? false;

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
                  child: Text(
                    'Response status',
                    style: context.appText.subtitle,
                  ),
                ),
                if (widget.value != null)
                  TextButton.icon(
                    onPressed: () {
                      _equalsCtrl.text = '';
                      _fromCtrl.text = '';
                      _toCtrl.text = '';
                      setState(() => _rangeError = null);
                      widget.onChanged(null);
                    },
                    icon: const Icon(Icons.clear),
                    label: const Text('Clear'),
                  ),
              ],
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _equalsCtrl,
              focusNode: _equalsFocus,
              decoration: const InputDecoration(
                labelText: 'Equals',
                hintText: '200,201,404',
              ),
              keyboardType: TextInputType.number,
              onChanged: (_) => _emit(),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _fromCtrl,
                    focusNode: _fromFocus,
                    decoration: const InputDecoration(labelText: 'From'),
                    keyboardType: TextInputType.number,
                    onChanged: (_) => _emit(),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: _toCtrl,
                    focusNode: _toFocus,
                    decoration: const InputDecoration(labelText: 'To'),
                    keyboardType: TextInputType.number,
                    onChanged: (_) => _emit(),
                  ),
                ),
              ],
            ),
            if (_rangeError != null) ...[
              const SizedBox(height: 6),
              Text(
                _rangeError!,
                style: context.appText.body.copyWith(
                  color: context.appColors.danger,
                ),
              ),
            ],
            const SizedBox(height: 8),
            Wrap(
              spacing: 12,
              children: [
                FilterChip(
                  selected: is4xx,
                  label: const Text('4xx'),
                  onSelected: (b) {
                    final equals = _parseEquals(_equalsCtrl.text);
                    final from0 = int.tryParse(_fromCtrl.text.trim());
                    final to0 = int.tryParse(_toCtrl.text.trim());
                    final from = (from0 == null || from0 == 0) ? null : from0;
                    final to = (to0 == null || to0 == 0) ? null : to0;
                    widget.onChanged(
                      _build(
                        equals: equals,
                        from: from,
                        to: to,
                        is4xx: b,
                        is5xx: is5xx,
                      ),
                    );
                  },
                ),
                FilterChip(
                  selected: is5xx,
                  label: const Text('5xx'),
                  onSelected: (b) {
                    final equals = _parseEquals(_equalsCtrl.text);
                    final from0 = int.tryParse(_fromCtrl.text.trim());
                    final to0 = int.tryParse(_toCtrl.text.trim());
                    final from = (from0 == null || from0 == 0) ? null : from0;
                    final to = (to0 == null || to0 == 0) ? null : to0;
                    widget.onChanged(
                      _build(
                        equals: equals,
                        from: from,
                        to: to,
                        is4xx: is4xx,
                        is5xx: b,
                      ),
                    );
                  },
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
