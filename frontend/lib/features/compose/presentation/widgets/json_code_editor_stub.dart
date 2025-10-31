import 'package:flutter/material.dart';

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
    return TextFormField(
      controller: controller,
      maxLines: null,
      expands: true,
      decoration: InputDecoration(
        labelText: 'JSON body',
        errorText: errorText,
        border: const OutlineInputBorder(),
        isDense: true,
      ),
      onChanged: onChanged,
    );
  }
}
