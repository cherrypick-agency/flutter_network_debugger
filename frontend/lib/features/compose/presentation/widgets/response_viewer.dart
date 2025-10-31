import 'dart:convert';

import 'package:flutter/material.dart';

class ResponseViewer extends StatefulWidget {
  final Map<String, dynamic>? data;
  final void Function(String sessionId)? onOpenSession;
  const ResponseViewer({super.key, required this.data, this.onOpenSession});
  @override
  State<ResponseViewer> createState() => _ResponseViewerState();
}

class _ResponseViewerState extends State<ResponseViewer> {
  int _tab = 0; // 0: Pretty, 1: Raw
  bool _wrap = true;
  @override
  Widget build(BuildContext context) {
    if (widget.data == null) {
      return const Center(child: Text('No response yet'));
    }
    final isError = widget.data!['error'] == true;
    final friendlyMessage = widget.data!['message']?.toString();
    final tx = widget.data!['transaction'] as Map<String, dynamic>?;
    final status = tx?['status']?.toString() ?? '—';
    final sessionId = widget.data!['sessionId']?.toString();
    final bodyBase64 = widget.data!['respBody'];
    String raw = '';
    String pretty = '';
    if (bodyBase64 is String) {
      try {
        raw = utf8.decode(base64.decode(bodyBase64));
        pretty = _tryPretty(raw);
      } catch (_) {}
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Text(
                'Status: $status',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Spacer(),
              if (sessionId != null && sessionId.isNotEmpty)
                TextButton.icon(
                  onPressed: () {
                    widget.onOpenSession?.call(sessionId);
                  },
                  icon: const Icon(Icons.open_in_new),
                  label: const Text('Open session'),
                ),
              const SizedBox(width: 8),
              Row(
                children: [
                  const Text('Wrap'),
                  Switch(
                    value: _wrap,
                    onChanged: (v) {
                      setState(() => _wrap = v);
                    },
                  ),
                ],
              ),
              const SizedBox(width: 8),
              SegmentedButton<int>(
                segments: const [
                  ButtonSegment(value: 0, label: Text('Pretty')),
                  ButtonSegment(value: 1, label: Text('Raw')),
                ],
                selected: {_tab},
                onSelectionChanged: (s) {
                  setState(() => _tab = s.first);
                },
              ),
            ],
          ),
        ),
        if (isError && (friendlyMessage != null && friendlyMessage.isNotEmpty))
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 0, 12, 8),
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                border: Border.all(color: Theme.of(context).colorScheme.error),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                friendlyMessage,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Theme.of(context).colorScheme.error,
                ),
              ),
            ),
          ),
        const Divider(height: 1),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                border: Border.all(
                  color: Theme.of(context).colorScheme.outlineVariant,
                ),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Builder(
                builder: (context) {
                  final text =
                      _tab == 0 ? (pretty.isEmpty ? raw : pretty) : raw;
                  final content = SelectableText(text);
                  if (_wrap) {
                    return SingleChildScrollView(child: content);
                  } else {
                    return SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: content,
                    );
                  }
                },
              ),
            ),
          ),
        ),
      ],
    );
  }

  String _tryPretty(String s) {
    try {
      final j = jsonDecode(s);
      return const JsonEncoder.withIndent('  ').convert(j);
    } catch (_) {
      return s;
    }
  }
}
