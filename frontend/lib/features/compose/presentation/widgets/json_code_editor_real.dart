// ignore_for_file: uri_does_not_exist, undefined_identifier, undefined_class, undefined_method
import 'package:flutter/material.dart';
import 'package:flutter_code_editor/flutter_code_editor.dart';
import 'package:highlight/languages/json.dart' as hl_json;

class JsonCodeEditor extends StatelessWidget {
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
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('JSON body'),
        const SizedBox(height: 6),
        Expanded(
          child: CodeTheme(
            data: CodeThemeData(styles: const {}),
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
                controller: CodeController(
                  text: controller.text,
                  language: hl_json.json,
                ),
                onChanged: (s) {
                  controller.text = s;
                  onChanged(s);
                },
              ),
            ),
          ),
        ),
        if (errorText != null)
          Padding(
            padding: const EdgeInsets.only(top: 6),
            child: Text(
              errorText!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ),
      ],
    );
  }
}
