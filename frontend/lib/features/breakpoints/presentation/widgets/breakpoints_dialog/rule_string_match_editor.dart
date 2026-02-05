import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_rule.dart';

enum _MatchKind { equals, prefix, suffix, contains, regex, anyOf }

class RuleStringMatchEditor extends StatefulWidget {
  const RuleStringMatchEditor({
    required this.label,
    required this.value,
    required this.onChanged,
    super.key,
  });

  final String label;
  final RuleStringMatch? value;
  final ValueChanged<RuleStringMatch?> onChanged;

  @override
  State<RuleStringMatchEditor> createState() => _RuleStringMatchEditorState();
}

class _RuleStringMatchEditorState extends State<RuleStringMatchEditor> {
  late final TextEditingController _ctrl;
  late final FocusNode _focus;
  _MatchKind _kind = _MatchKind.contains;
  String? _regexError;
  String? _regexWarning;

  @override
  void initState() {
    super.initState();
    _ctrl = TextEditingController(text: _textFromValue(widget.value));
    _focus = FocusNode();
    _kind = _kindFromValue(widget.value) ?? _MatchKind.contains;
    _revalidate();
  }

  @override
  void didUpdateWidget(covariant RuleStringMatchEditor oldWidget) {
    super.didUpdateWidget(oldWidget);
    final nextText = _textFromValue(widget.value);
    if (!_focus.hasFocus && _ctrl.text != nextText) _ctrl.text = nextText;
    final nextKind = _kindFromValue(widget.value);
    if (nextKind != null && nextKind != _kind) _kind = nextKind;
    _revalidate();
  }

  @override
  void dispose() {
    _ctrl.dispose();
    _focus.dispose();
    super.dispose();
  }

  _MatchKind? _kindFromValue(RuleStringMatch? v) {
    if (v == null) return null;
    if ((v.equals ?? '').isNotEmpty) return _MatchKind.equals;
    if ((v.prefix ?? '').isNotEmpty) return _MatchKind.prefix;
    if ((v.suffix ?? '').isNotEmpty) return _MatchKind.suffix;
    if ((v.contains ?? '').isNotEmpty) return _MatchKind.contains;
    if ((v.regex ?? '').isNotEmpty) return _MatchKind.regex;
    if (v.anyOf.isNotEmpty) return _MatchKind.anyOf;
    return null;
  }

  String _textFromValue(RuleStringMatch? v) {
    if (v == null) return '';
    return switch (_kindFromValue(v)) {
      _MatchKind.equals => v.equals ?? '',
      _MatchKind.prefix => v.prefix ?? '',
      _MatchKind.suffix => v.suffix ?? '',
      _MatchKind.contains => v.contains ?? '',
      _MatchKind.regex => v.regex ?? '',
      _MatchKind.anyOf => v.anyOf.join(','),
      null => '',
    };
  }

  RuleStringMatch? _buildValue(_MatchKind kind, String raw) {
    if (raw.trim().isEmpty) return null;
    final t = raw;
    return switch (kind) {
      _MatchKind.equals => RuleStringMatch(equals: t),
      _MatchKind.prefix => RuleStringMatch(prefix: t),
      _MatchKind.suffix => RuleStringMatch(suffix: t),
      _MatchKind.contains => RuleStringMatch(contains: t),
      _MatchKind.regex => RuleStringMatch(regex: t),
      _MatchKind.anyOf => () {
        final list = t
            .split(',')
            .map((e) => e.trim())
            .where((e) => e.isNotEmpty)
            .toList(growable: false);
        if (list.isEmpty) return null;
        return RuleStringMatch(anyOf: list);
      }(),
    };
  }

  void _revalidate() {
    if (_kind != _MatchKind.regex) {
      if (_regexError != null || _regexWarning != null) {
        setState(() {
          _regexError = null;
          _regexWarning = null;
        });
      }
      return;
    }
    final raw = _ctrl.text;
    if (raw.trim().isEmpty) {
      if (_regexError != null || _regexWarning != null) {
        setState(() {
          _regexError = null;
          _regexWarning = null;
        });
      }
      return;
    }
    try {
      RegExp(raw);
      final warn = _re2Warning(raw);
      if (_regexError != null || _regexWarning != warn) {
        setState(() {
          _regexError = null;
          _regexWarning = warn;
        });
      }
    } catch (e) {
      final next = 'Regex looks invalid: $e';
      if (_regexError != next || _regexWarning != null) {
        setState(() {
          _regexError = next;
          _regexWarning = null;
        });
      }
    }
  }

  String? _re2Warning(String raw) {
    // Backend uses Go/RE2. These constructs are not supported there.
    // We don't try to fully validate RE2, just provide quick hints.
    final s = raw;
    if (s.contains('(?<=')) return 'RE2 does not support lookbehind (?<=...)';
    if (s.contains('(?<!')) return 'RE2 does not support lookbehind (?<!...)';
    if (RegExp(r'\\[1-9]').hasMatch(s)) {
      return 'RE2 does not support backreferences (\\1, \\2, …)';
    }
    if (s.contains('(?>')) return 'RE2 does not support atomic groups (?>...)';
    if (RegExp(r'\(\?\(.+\)').hasMatch(s)) {
      return 'RE2 does not support conditional groups (?(...))';
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          flex: 2,
          child: DropdownButtonFormField<_MatchKind>(
            key: ValueKey('matchKind-$_kind'),
            initialValue: _kind,
            decoration: InputDecoration(labelText: widget.label),
            items: const [
              DropdownMenuItem(
                value: _MatchKind.contains,
                child: Text('contains'),
              ),
              DropdownMenuItem(value: _MatchKind.equals, child: Text('equals')),
              DropdownMenuItem(value: _MatchKind.prefix, child: Text('prefix')),
              DropdownMenuItem(value: _MatchKind.suffix, child: Text('suffix')),
              DropdownMenuItem(value: _MatchKind.regex, child: Text('regex')),
              DropdownMenuItem(value: _MatchKind.anyOf, child: Text('anyOf')),
            ],
            onChanged: (v) {
              if (v == null) return;
              setState(() => _kind = v);
              _revalidate();
              widget.onChanged(_buildValue(v, _ctrl.text));
            },
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          flex: 3,
          child: TextField(
            controller: _ctrl,
            focusNode: _focus,
            decoration: InputDecoration(
              labelText: 'Value',
              hintText: _kind == _MatchKind.anyOf ? 'a,b,c' : null,
              errorText: _regexError,
              helperText: _kind == _MatchKind.regex
                  ? (_regexWarning ??
                        'Backend uses RE2; if the regex is invalid, the rule will not match.')
                  : null,
              suffixIcon: (_ctrl.text.trim().isEmpty)
                  ? null
                  : IconButton(
                      tooltip: 'Clear',
                      onPressed: () {
                        _ctrl.text = '';
                        _revalidate();
                        widget.onChanged(null);
                      },
                      icon: const Icon(Icons.clear),
                    ),
            ),
            onChanged: (v) {
              _revalidate();
              widget.onChanged(_buildValue(_kind, v));
            },
            style: context.appText.body,
          ),
        ),
      ],
    );
  }
}
