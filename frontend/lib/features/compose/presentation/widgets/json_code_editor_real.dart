// ignore_for_file: uri_does_not_exist, undefined_identifier, undefined_class, undefined_method
import 'package:flutter/material.dart';
import 'package:flutter_code_editor/flutter_code_editor.dart';
import 'package:highlight/languages/json.dart' as hl_json;
import 'package:highlight_selectable/theme_map.dart';
import '../../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart';

class JsonCodeEditor extends StatefulWidget {
  const JsonCodeEditor({
    super.key,
    required this.controller,
    required this.onChanged,
    this.errorText,
  });

  final TextEditingController controller;
  final void Function(String) onChanged;
  final String? errorText;

  @override
  State<JsonCodeEditor> createState() => _JsonCodeEditorState();
}

class _JsonCodeEditorState extends State<JsonCodeEditor> {
  late final CodeController _codeController;
  bool _updatingFromExternal = false;
  bool _updatingExternal = false;
  String? _themeLight;
  String? _themeDark;

  @override
  void initState() {
    super.initState();
    _codeController = CodeController(
      text: widget.controller.text,
      language: hl_json.json,
    );
    _codeController.addListener(_onCodeChanged);
    widget.controller.addListener(_onExternalChanged);
    _fetchTheme();
  }

  @override
  void didUpdateWidget(covariant JsonCodeEditor oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.controller != widget.controller) {
      oldWidget.controller.removeListener(_onExternalChanged);
      widget.controller.addListener(_onExternalChanged);
      // sync current value
      _updatingFromExternal = true;
      _codeController.text = widget.controller.text;
      _updatingFromExternal = false;
    }
  }

  void _onCodeChanged() {
    if (_updatingFromExternal) return;
    final s = _codeController.text;
    if (widget.controller.text != s) {
      _updatingExternal = true;
      widget.controller.value = TextEditingValue(
        text: s,
        selection: TextSelection.collapsed(offset: s.length),
      );
      _updatingExternal = false;
    }
    widget.onChanged(s);
  }

  void _onExternalChanged() {
    if (_updatingExternal) return;
    final s = widget.controller.text;
    if (s != _codeController.text) {
      _updatingFromExternal = true;
      _codeController.text = s;
      _updatingFromExternal = false;
    }
  }

  @override
  void dispose() {
    _codeController.removeListener(_onCodeChanged);
    widget.controller.removeListener(_onExternalChanged);
    _codeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final def = isDark ? 'a11y-dark' : 'a11y-light';
    final key = isDark
        ? (_themeDark ?? _themeLight)
        : (_themeLight ?? _themeDark);
    final styles = themeMap[key ?? def] ?? themeMap[def]!;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('JSON body'),
        const SizedBox(height: 6),
        Expanded(
          child: CodeTheme(
            data: CodeThemeData(styles: styles),
            child: DecoratedBox(
              decoration: BoxDecoration(
                border: Border.all(
                  color: Theme.of(context).colorScheme.outlineVariant,
                ),
                borderRadius: BorderRadius.circular(8),
              ),
              child: CodeField(
                expands: true,
                wrap: true,
                lineNumbers: true,
                controller: _codeController,
              ),
            ),
          ),
        ),
        if (widget.errorText != null)
          Padding(
            padding: const EdgeInsets.only(top: 6),
            child: Text(
              widget.errorText!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
      ],
    );
  }

  Future<void> _fetchTheme() async {
    try {
      final api = sl<AppHttpClient>();
      final res = await api.get(path: '/_api/v1/settings');
      final data = (res.data as Map).cast<String, dynamic>();
      final light = (data['highlightThemeLight'] ?? '').toString();
      final dark = (data['highlightThemeDark'] ?? '').toString();
      final legacy = (data['highlightTheme'] ?? '').toString();
      setState(() {
        _themeLight = light.isNotEmpty
            ? light
            : (legacy.isNotEmpty ? legacy : null);
        _themeDark = dark.isNotEmpty
            ? dark
            : (legacy.isNotEmpty ? legacy : null);
      });
    } catch (_) {
      // ignore, use default themes
    }
  }
}
