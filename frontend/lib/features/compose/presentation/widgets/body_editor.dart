import 'dart:convert';

import 'package:flutter/material.dart';
import 'json_code_editor_stub.dart'
    if (dart.library.html) 'json_code_editor_real.dart'
    if (dart.library.io) 'json_code_editor_real.dart';

import '../../domain/models.dart';
import 'form_editor.dart';
import 'multipart_editor.dart';

class BodyEditor extends StatefulWidget {
  const BodyEditor({
    super.key,
    required this.mode,
    required this.onModeChanged,
    required this.rawCtrl,
    required this.jsonCtrl,
    required this.form,
    required this.multipart,
    this.onFormChanged,
    this.onMultipartChanged,
    this.maxUploadMB,
    this.allowedModes = const ['raw', 'json', 'form', 'multipart'],
  });

  final String mode; // raw|json|form|multipart
  final ValueChanged<String> onModeChanged;
  final TextEditingController rawCtrl;
  final TextEditingController jsonCtrl;
  final List<ComposeFormFieldDTO> form;
  final List<ComposePartDTO> multipart;
  final ValueChanged<List<ComposeFormFieldDTO>>? onFormChanged;
  final ValueChanged<List<ComposePartDTO>>? onMultipartChanged;
  final int? maxUploadMB;
  final List<String> allowedModes;

  @override
  State<BodyEditor> createState() => _BodyEditorState();
}

class _BodyEditorState extends State<BodyEditor> {
  String? _jsonError; // сообщение об ошибке json (если есть)

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(12),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final hasBoundedHeight = constraints.maxHeight.isFinite;
          Widget editor() {
            final mode = widget.allowedModes.contains(widget.mode)
                ? widget.mode
                : widget.allowedModes.first;
            switch (mode) {
              case 'json':
                return JsonCodeEditor(
                  controller: widget.jsonCtrl,
                  errorText: _jsonError,
                  onChanged: (s) {
                    _validateJson(s);
                  },
                );
              case 'form':
                return FormEditor(
                  items: widget.form,
                  onChanged: widget.onFormChanged,
                );
              case 'multipart':
                return MultipartEditor(
                  items: widget.multipart,
                  maxUploadMB: widget.maxUploadMB ?? 10,
                  onChanged: widget.onMultipartChanged,
                );
              case 'raw':
              default:
                return TextFormField(
                  controller: widget.rawCtrl,
                  maxLines: null,
                  expands: true,
                  decoration: const InputDecoration(labelText: 'Raw body'),
                );
            }
          }

          final header = Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              DropdownButtonFormField<String>(
                value: widget.allowedModes.contains(widget.mode)
                    ? widget.mode
                    : widget.allowedModes.first,
                onChanged: (v) =>
                    setState(() => widget.onModeChanged(v ?? 'raw')),
                isDense: true,
                isExpanded: true,
                iconSize: 16,
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(fontSize: 12),
                items: widget.allowedModes
                    .map((m) => DropdownMenuItem(value: m, child: Text(m)))
                    .toList(),
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  if (widget.mode == 'json')
                    OutlinedButton(
                      onPressed: () {
                        final sample = const JsonEncoder.withIndent('  ')
                            .convert({
                              'name': 'John',
                              'email': 'john@example.com',
                              'active': true,
                            });
                        widget.jsonCtrl.text = sample;
                        setState(() => _jsonError = null);
                      },
                      child: const Text('Вставить JSON пример'),
                    ),
                ],
              ),
              const SizedBox(height: 8),
            ],
          );

          if (hasBoundedHeight) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                header,
                Expanded(child: editor()),
              ],
            );
          }

          // Используем Flexible для динамической высоты по контенту
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              header,
              Flexible(child: editor()),
            ],
          );
        },
      ),
    );
  }

  void _validateJson(String s) {
    try {
      if (s.trim().isEmpty) {
        setState(() => _jsonError = 'Пустой JSON');
        return;
      }
      // Если парсится — всё ок
      jsonDecode(s);
      setState(() => _jsonError = null);
    } catch (e) {
      setState(() => _jsonError = 'Ошибка JSON: ${e.toString()}');
    }
  }
}
