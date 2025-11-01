import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';

import '../../domain/models.dart';
import 'dnd_widgets_stub.dart'
    if (dart.library.html) 'dnd_widgets_web.dart'
    if (dart.library.io) 'dnd_widgets_desktop.dart';

class MultipartEditor extends StatefulWidget {
  const MultipartEditor({
    super.key,
    required this.items,
    this.maxUploadMB = 10,
    this.onChanged,
  });

  final List<ComposePartDTO> items;
  final int maxUploadMB;
  final ValueChanged<List<ComposePartDTO>>? onChanged;

  @override
  State<MultipartEditor> createState() => _MultipartEditorState();
}

class _MultipartEditorState extends State<MultipartEditor> {
  bool _hovering = false; // подсветка дропа

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Padding(
      padding: const EdgeInsets.all(12),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Глобальная зона для массового дропа файлов — добавляет новые части
            SizedBox(
              height: 56,
              child: Stack(
                children: [
                  Container(
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: _hovering ? cs.primary : cs.outlineVariant,
                      ),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    alignment: Alignment.centerLeft,
                    padding: const EdgeInsets.symmetric(horizontal: 12),
                    child: Text(
                      _hovering
                          ? 'Отпустите файлы, чтобы добавить'
                          : 'Перетащите файлы сюда, чтобы добавить как части',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
                  DndDropArea(
                    onHoverChanged: (v) {
                      setState(() => _hovering = v);
                    },
                    onDrop: (bytes, filename) {
                      final b64 = base64Encode(bytes);
                      setState(() {
                        widget.items.add(
                          ComposePartDTO(
                            name: filename,
                            isFile: true,
                            fileName: filename,
                            base64: b64,
                          ),
                        );
                      });
                      widget.onChanged?.call(widget.items);
                    },
                    onDropMany: (files) {
                      setState(() {
                        for (final f in files) {
                          final b64 = base64Encode(f.bytes);
                          widget.items.add(
                            ComposePartDTO(
                              name: f.filename,
                              isFile: true,
                              fileName: f.filename,
                              base64: b64,
                            ),
                          );
                        }
                      });
                      widget.onChanged?.call(widget.items);
                    },
                  ),
                ],
              ),
            ),
            const SizedBox(height: 8),
            ListView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: widget.items.length,
              itemBuilder: (context, index) {
                final item = widget.items[index];
                final name = TextEditingController(text: item.name);
                final value = TextEditingController(text: item.value ?? '');
                final fileNameCtrl = TextEditingController(
                  text: item.fileName ?? '',
                );
                final b64Ctrl = TextEditingController(text: item.base64 ?? '');

                final fileTooLarge = _calcMB(item.base64) > widget.maxUploadMB;

                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 4),
                  child: Row(
                    children: [
                      SizedBox(
                        width: 96,
                        child: DropdownButtonFormField<bool>(
                          value: item.isFile,
                          onChanged: (v) {
                            setState(
                              () =>
                                  widget.items[index] = ComposePartDTO(
                                    name: item.name,
                                    isFile: v ?? false,
                                    fileName: item.fileName,
                                    contentType: item.contentType,
                                    base64: item.base64,
                                    value: item.value,
                                  ),
                            );
                            widget.onChanged?.call(widget.items);
                          },
                          isDense: true,
                          isExpanded: true,
                          iconSize: 16,
                          style: Theme.of(
                            context,
                          ).textTheme.bodyMedium?.copyWith(fontSize: 12),
                          items:
                              const [false, true]
                                  .map(
                                    (v) => DropdownMenuItem(
                                      value: v,
                                      child: Text(v ? 'File' : 'Text'),
                                    ),
                                  )
                                  .toList(),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: TextField(
                          controller: name,
                          decoration: const InputDecoration(labelText: 'Name'),
                          onChanged: (s) {
                            widget.items[index] = ComposePartDTO(
                              name: s,
                              isFile: item.isFile,
                              fileName: item.fileName,
                              contentType: item.contentType,
                              base64: item.base64,
                              value: item.value,
                            );
                            widget.onChanged?.call(widget.items);
                          },
                        ),
                      ),
                      const SizedBox(width: 8),
                      if (!item.isFile)
                        Expanded(
                          child: TextField(
                            controller: value,
                            decoration: const InputDecoration(
                              labelText: 'Value',
                            ),
                            onChanged: (s) {
                              widget.items[index] = ComposePartDTO(
                                name: item.name,
                                isFile: false,
                                value: s,
                              );
                              widget.onChanged?.call(widget.items);
                            },
                          ),
                        ),
                      if (item.isFile)
                        Expanded(
                          child: TextField(
                            decoration: const InputDecoration(
                              labelText: 'File name',
                            ),
                            controller: fileNameCtrl,
                            onChanged: (s) {
                              widget.items[index] = ComposePartDTO(
                                name: item.name,
                                isFile: true,
                                fileName: s,
                                contentType: item.contentType,
                                base64: item.base64,
                              );
                              widget.onChanged?.call(widget.items);
                            },
                          ),
                        ),
                      if (item.isFile) const SizedBox(width: 8),
                      if (item.isFile)
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              TextField(
                                decoration: InputDecoration(
                                  labelText: 'Base64',
                                  errorText:
                                      fileTooLarge
                                          ? 'Размер > ${widget.maxUploadMB}MB'
                                          : null,
                                ),
                                controller: b64Ctrl,
                                onChanged: (s) {
                                  widget.items[index] = ComposePartDTO(
                                    name: item.name,
                                    isFile: true,
                                    fileName: item.fileName,
                                    contentType: item.contentType,
                                    base64: s,
                                  );
                                  widget.onChanged?.call(widget.items);
                                },
                              ),
                              const SizedBox(height: 4),
                              Row(
                                children: [
                                  Text(
                                    '~ ${_calcMB(item.base64)}MB',
                                    style:
                                        Theme.of(context).textTheme.labelSmall,
                                  ),
                                  const SizedBox(width: 8),
                                  TextButton(
                                    onPressed: () {
                                      setState(() {
                                        widget.items[index] = ComposePartDTO(
                                          name: item.name,
                                          isFile: true,
                                        );
                                      });
                                      widget.onChanged?.call(widget.items);
                                    },
                                    child: const Text('Очистить файл'),
                                  ),
                                  const SizedBox(width: 8),
                                  TextButton(
                                    onPressed: () async {
                                      try {
                                        final res = await FilePicker.platform
                                            .pickFiles(withData: true);
                                        final picked = res?.files.first;
                                        if (picked?.bytes != null &&
                                            picked!.name.isNotEmpty) {
                                          final b64 = base64Encode(
                                            picked.bytes!,
                                          );
                                          setState(() {
                                            widget
                                                .items[index] = ComposePartDTO(
                                              name:
                                                  item.name.isEmpty
                                                      ? picked.name
                                                      : item.name,
                                              isFile: true,
                                              fileName: picked.name,
                                              base64: b64,
                                            );
                                          });
                                          widget.onChanged?.call(widget.items);
                                        }
                                      } catch (_) {}
                                    },
                                    child: const Text('Выбрать файл'),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                      if (item.isFile) const SizedBox(width: 8),
                      if (item.isFile)
                        SizedBox(
                          width: 160,
                          height: 36,
                          child: Stack(
                            children: [
                              Container(
                                decoration: BoxDecoration(
                                  border: Border.all(
                                    color:
                                        _hovering
                                            ? cs.primary
                                            : cs.outlineVariant,
                                  ),
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                alignment: Alignment.center,
                                child: Text(
                                  _hovering
                                      ? 'Отпустите для загрузки'
                                      : 'Drop file here',
                                  style: Theme.of(context).textTheme.bodySmall,
                                ),
                              ),
                              DndDropArea(
                                onHoverChanged: (v) {
                                  setState(() => _hovering = v);
                                },
                                onDrop: (bytes, filename) {
                                  final b64 = base64Encode(bytes);
                                  setState(
                                    () =>
                                        widget.items[index] = ComposePartDTO(
                                          name: item.name,
                                          isFile: true,
                                          fileName: filename,
                                          contentType: item.contentType,
                                          base64: b64,
                                        ),
                                  );
                                  widget.onChanged?.call(widget.items);
                                },
                                onDropMany: (files) {
                                  // заменим текущий и добавим остальные
                                  if (files.isEmpty) return;
                                  final first = files.first;
                                  final b64 = base64Encode(first.bytes);
                                  setState(
                                    () =>
                                        widget.items[index] = ComposePartDTO(
                                          name: item.name,
                                          isFile: true,
                                          fileName: first.filename,
                                          contentType: item.contentType,
                                          base64: b64,
                                        ),
                                  );
                                  for (var i = 1; i < files.length; i++) {
                                    final fi = files[i];
                                    final b = base64Encode(fi.bytes);
                                    widget.items.add(
                                      ComposePartDTO(
                                        name: fi.filename,
                                        isFile: true,
                                        fileName: fi.filename,
                                        base64: b,
                                      ),
                                    );
                                  }
                                  widget.onChanged?.call(widget.items);
                                },
                              ),
                            ],
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
            const SizedBox(height: 8),
            Text(
              'Лимит файла ~ ${widget.maxUploadMB}MB',
              style: Theme.of(context).textTheme.labelSmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 6),
            TextButton.icon(
              onPressed: () {
                setState(
                  () => widget.items.add(
                    const ComposePartDTO(name: '', isFile: false),
                  ),
                );
                widget.onChanged?.call(widget.items);
              },
              icon: const Icon(Icons.add),
              label: const Text('Add part'),
            ),
          ],
        ),
      ),
    );
  }

  int _calcMB(String? base64Str) {
    try {
      if (base64Str == null || base64Str.isEmpty) return 0;
      // на глаз: base64 ~ +33% к оригиналу, но нам хватает грубой оценки по длине
      final bytes = base64.decode(base64Str);
      return (bytes.length / (1024 * 1024)).ceil();
    } catch (_) {
      return 0;
    }
  }
}
