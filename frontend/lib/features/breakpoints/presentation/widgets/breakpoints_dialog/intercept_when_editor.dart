import 'package:flutter/material.dart';

import '../../../../../theme/context_ext.dart';
import '../../../domain/entities/intercept_rule.dart';
import 'rule_header_match_editor.dart';
import 'rule_status_match_editor.dart';
import 'rule_string_match_editor.dart';
import 'string_list_editor.dart';

class InterceptWhenEditor extends StatelessWidget {
  const InterceptWhenEditor({
    required this.value,
    required this.onChanged,
    super.key,
  });

  final InterceptWhen value;
  final ValueChanged<InterceptWhen> onChanged;

  static final Object _unset = Object();

  void _emit({
    List<String>? method,
    List<String>? scheme,
    Object? host = _unset,
    Object? port = _unset,
    Object? path = _unset,
    Object? contentType = _unset,
    Object? responseStatus = _unset,
    Object? header = _unset,
    Object? bodyContains = _unset,
  }) {
    final next = InterceptWhen(
      method: method ?? value.method,
      scheme: scheme ?? value.scheme,
      host: host == _unset ? value.host : host as RuleStringMatch?,
      port: port == _unset ? value.port : port as RuleStringMatch?,
      path: path == _unset ? value.path : path as RuleStringMatch?,
      contentType: contentType == _unset
          ? value.contentType
          : contentType as RuleStringMatch?,
      responseStatus: responseStatus == _unset
          ? value.responseStatus
          : responseStatus as RuleStatusMatch?,
      header: header == _unset ? value.header : header as RuleHeaderMatch?,
      bodyContains: bodyContains == _unset
          ? value.bodyContains
          : bodyContains as String?,
    );
    onChanged(next);
  }

  @override
  Widget build(BuildContext context) {
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
            Text('When', style: context.appText.subtitle),
            const SizedBox(height: 8),
            StringListEditor(
              label: 'Methods',
              hint: 'GET,POST',
              value: value.method,
              normalize: (s) => s.trim().toUpperCase(),
              onChanged: (v) => _emit(method: v),
            ),
            const SizedBox(height: 12),
            StringListEditor(
              label: 'Scheme',
              hint: 'http,https',
              value: value.scheme,
              normalize: (s) => s.trim().toLowerCase(),
              onChanged: (v) => _emit(scheme: v),
            ),
            const SizedBox(height: 12),
            RuleStringMatchEditor(
              label: 'Host',
              value: value.host,
              onChanged: (v) => _emit(host: v),
            ),
            const SizedBox(height: 12),
            RuleStringMatchEditor(
              label: 'Port',
              value: value.port,
              onChanged: (v) => _emit(port: v),
            ),
            const SizedBox(height: 12),
            RuleStringMatchEditor(
              label: 'Path',
              value: value.path,
              onChanged: (v) => _emit(path: v),
            ),
            const SizedBox(height: 12),
            RuleStringMatchEditor(
              label: 'Content-Type',
              value: value.contentType,
              onChanged: (v) => _emit(contentType: v),
            ),
            const SizedBox(height: 12),
            RuleStatusMatchEditor(
              value: value.responseStatus,
              onChanged: (v) => _emit(responseStatus: v),
            ),
            const SizedBox(height: 12),
            RuleHeaderMatchEditor(
              value: value.header,
              onChanged: (v) => _emit(header: v),
            ),
            const SizedBox(height: 12),
            _BodyContainsField(
              value: value.bodyContains,
              onChanged: (v) => _emit(bodyContains: v),
            ),
            const SizedBox(height: 4),
            Text(
              'Условия применяются через AND (в MVP). Пустое поле = не фильтруем.',
              style: context.appText.body.copyWith(
                color: context.appColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _BodyContainsField extends StatefulWidget {
  const _BodyContainsField({required this.value, required this.onChanged});

  @override
  State<_BodyContainsField> createState() => _BodyContainsFieldState();
  final String? value;
  final ValueChanged<String?> onChanged;
}

class _BodyContainsFieldState extends State<_BodyContainsField> {
  late final TextEditingController _ctrl;

  @override
  void initState() {
    super.initState();
    _ctrl = TextEditingController(text: widget.value ?? '');
  }

  @override
  void didUpdateWidget(covariant _BodyContainsField oldWidget) {
    super.didUpdateWidget(oldWidget);
    final next = widget.value ?? '';
    if (_ctrl.text != next) _ctrl.text = next;
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: _ctrl,
      decoration: InputDecoration(
        labelText: 'Body contains',
        hintText: 'substring',
        suffixIcon: (_ctrl.text.trim().isEmpty)
            ? null
            : IconButton(
                tooltip: 'Clear',
                onPressed: () {
                  _ctrl.text = '';
                  widget.onChanged(null);
                },
                icon: const Icon(Icons.clear),
              ),
      ),
      onChanged: (v) => widget.onChanged(v.trim().isEmpty ? null : v),
    );
  }
}
