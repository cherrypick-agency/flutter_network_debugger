import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../theme/context_ext.dart';

class FormDataView extends StatelessWidget {
  const FormDataView({
    super.key,
    required this.form,
    required this.contentType,
    required this.rawBody,
  });

  // form – структура, которую отдаёт бэкенд в превью (см. buildHTTPRequestPreview)
  final Map<String, dynamic>? form;
  final String? contentType;
  final String? rawBody;

  bool get _isUrlEncodedCt {
    final ct = (contentType ?? '').toLowerCase();
    return ct.contains('application/x-www-form-urlencoded');
  }

  bool get _isMultipartCt {
    final ct = (contentType ?? '').toLowerCase();
    return ct.contains('multipart/form-data');
  }

  Map<String, dynamic>? _fallbackFromUrlEncoded(String? body) {
    if (body == null || body.isEmpty || !_isUrlEncodedCt) return null;
    // Разбираем как query-строку, поддержим повторяющиеся ключи
    try {
      final map = Uri(query: body).queryParametersAll;
      if (map.isEmpty) return null;
      final fields = <Map<String, dynamic>>[];
      map.forEach((k, values) {
        for (final v in values) {
          fields.add({'name': k, 'value': v});
        }
      });
      return {'type': 'urlencoded', 'fields': fields};
    } catch (_) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final effective = form ?? _fallbackFromUrlEncoded(rawBody);
    // Показываем только если реально есть что показать и это форма
    if (effective == null) return const SizedBox.shrink();
    final type = (effective['type'] ?? '').toString();
    if (!_isUrlEncodedCt && !_isMultipartCt && type.isEmpty) {
      return const SizedBox.shrink();
    }

    final fields =
        (effective['fields'] as List?)?.cast<Map<String, dynamic>>() ??
        const [];
    final files =
        (effective['files'] as List?)?.cast<Map<String, dynamic>>() ?? const [];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Form Data', style: context.appText.subtitle),
        const SizedBox(height: 6),
        if (fields.isNotEmpty) ...[
          _SectionTitle(text: 'Fields'),
          const SizedBox(height: 4),
          ...fields
              .map(
                (e) => _FieldRow(
                  name: e['name']?.toString() ?? '',
                  value: (e['value'] ?? e['valuePreview'] ?? '').toString(),
                  truncated: (e['truncated'] == true),
                ),
              )
              .toList(),
          const SizedBox(height: 8),
        ],
        if (files.isNotEmpty) ...[
          _SectionTitle(text: 'Files'),
          const SizedBox(height: 4),
          ...files.map((e) => _FileRow(item: e)).toList(),
          const SizedBox(height: 8),
        ],
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton.icon(
            onPressed: () {
              final jsonStr = const JsonEncoder.withIndent(
                '  ',
              ).convert(effective);
              Clipboard.setData(ClipboardData(text: jsonStr));
            },
            icon: const Icon(Icons.copy, size: 16),
            label: const Text('Copy as JSON'),
          ),
        ),
        const SizedBox(height: 8),
      ],
    );
  }
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle({required this.text});
  final String text;
  @override
  Widget build(BuildContext context) {
    return Text(text, style: Theme.of(context).textTheme.labelLarge);
  }
}

class _FieldRow extends StatelessWidget {
  const _FieldRow({
    required this.name,
    required this.value,
    required this.truncated,
  });
  final String name;
  final String value;
  final bool truncated;
  @override
  Widget build(BuildContext context) {
    final mono = context.appText.monospace;
    return Padding(
      padding: const EdgeInsets.only(bottom: 2),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('$name: ', style: mono.copyWith(fontWeight: FontWeight.w600)),
          Expanded(
            child: SelectableText.rich(
              TextSpan(
                children: [
                  TextSpan(text: value, style: mono),
                  if (truncated)
                    WidgetSpan(
                      alignment: PlaceholderAlignment.baseline,
                      baseline: TextBaseline.alphabetic,
                      child: Padding(
                        padding: const EdgeInsets.only(left: 6),
                        child: _Chip(text: 'truncated'),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _FileRow extends StatelessWidget {
  const _FileRow({required this.item});
  final Map<String, dynamic> item;
  @override
  Widget build(BuildContext context) {
    final mono = context.appText.monospace;
    final name = (item['name'] ?? '').toString();
    final filename = (item['filename'] ?? '').toString();
    final ct = (item['contentType'] ?? '').toString();
    final truncated = (item['truncated'] == true);
    final valuePreview = (item['valuePreview'] ?? '').toString();
    final previewSize =
        (item['previewSize'] is num)
            ? (item['previewSize'] as num).toInt()
            : null;
    final isImage = ct.toLowerCase().startsWith('image/');
    final isTextish =
        ct.isEmpty ||
        ct.toLowerCase().contains('text') ||
        ct.toLowerCase().contains('json') ||
        ct.toLowerCase().contains('+json') ||
        ct.toLowerCase().contains('xml') ||
        ct.toLowerCase().contains('csv');
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(isImage ? Icons.image : Icons.insert_drive_file, size: 16),
              const SizedBox(width: 6),
              Expanded(
                child: SelectableText(
                  '$name: $filename${ct.isNotEmpty ? ' ($ct)' : ''}' +
                      (previewSize != null
                          ? ' · ${_fmtSize(previewSize)}'
                          : ''),
                  style: mono,
                ),
              ),
              if (truncated) _Chip(text: 'truncated'),
            ],
          ),
          if (valuePreview.isNotEmpty && isTextish)
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: SelectableText(valuePreview, style: mono),
            ),
        ],
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({required this.text});
  final String text;
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceVariant,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(text, style: Theme.of(context).textTheme.labelSmall),
    );
  }
}

String _fmtSize(int bytes) {
  const units = ['B', 'KB', 'MB', 'GB'];
  double size = bytes.toDouble();
  int unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  if (unit == 0) return '${bytes} ${units[unit]}';
  return '${size.toStringAsFixed(1)} ${units[unit]}';
}
