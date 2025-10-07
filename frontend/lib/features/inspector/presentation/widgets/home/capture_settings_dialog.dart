import 'package:flutter/material.dart';
import '../../../application/stores/home_ui_store.dart';
import '../../../../../core/di/di.dart';

class CaptureSettingsDialog extends StatefulWidget {
  const CaptureSettingsDialog({
    super.key,
    required this.initialRecording,
    required this.initialScope,
    required this.initialIncludePaused,
  });
  final bool initialRecording;
  final String initialScope; // 'current' | 'all'
  final bool initialIncludePaused;

  @override
  State<CaptureSettingsDialog> createState() => _CaptureSettingsDialogState();
}

class _CaptureSettingsDialogState extends State<CaptureSettingsDialog> {
  late bool _recording;
  late String _scope;
  late bool _includePaused;

  @override
  void initState() {
    super.initState();
    _recording = widget.initialRecording;
    _scope = widget.initialScope;
    _includePaused = widget.initialIncludePaused;
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Capture settings'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Recording toggle
          Row(
            children: [
              Switch(
                value: _recording,
                onChanged: (v) => setState(() => _recording = v),
              ),
              const SizedBox(width: 8),
              Text(_recording ? 'Recording' : 'Paused'),
            ],
          ),
          const SizedBox(height: 8),
          // Scope
          Row(
            children: [
              const Text('Scope: '),
              const SizedBox(width: 8),
              DropdownButton<String>(
                value: _scope,
                items: const [
                  DropdownMenuItem(value: 'current', child: Text('Current')),
                  DropdownMenuItem(value: 'all', child: Text('All')),
                ],
                onChanged: (v) => setState(() => _scope = v ?? _scope),
              ),
            ],
          ),
          const SizedBox(height: 8),
          // Include paused
          CheckboxListTile(
            dense: true,
            contentPadding: EdgeInsets.zero,
            title: const Text('Include paused'),
            value: _includePaused,
            onChanged:
                (v) => setState(() => _includePaused = v ?? _includePaused),
          ),
          const SizedBox(height: 8),
          // Captures list placeholder (MVP)
          FutureBuilder<Map<String, dynamic>>(
            future: _loadCaptures(),
            builder: (context, snap) {
              final items = (snap.data?['items'] as List?) ?? const [];
              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Available captures',
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 6),
                  if (items.isEmpty)
                    const Text('—', style: TextStyle(color: Colors.grey))
                  else
                    ...items.map((e) => Text('• id: ${e['id']}')),
                ],
              );
            },
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(false),
          child: const Text('Cancel'),
        ),
        ElevatedButton(onPressed: _apply, child: const Text('Apply')),
      ],
    );
  }

  Future<Map<String, dynamic>> _loadCaptures() async {
    try {
      final client = sl.get<Object>() as dynamic;
      final res = await client.get(path: '/_api/v1/captures');
      Map<String, dynamic> data =
          (res.data is Map<String, dynamic>)
              ? (res.data as Map<String, dynamic>)
              : <String, dynamic>{'items': []};
      final items = (data['items'] as List?) ?? const [];
      if (items.isEmpty) {
        // fallback to current capture
        try {
          final res2 = await client.get(path: '/_api/v1/capture');
          final d2 =
              (res2.data is Map<String, dynamic>)
                  ? (res2.data as Map<String, dynamic>)
                  : const <String, dynamic>{};
          final cur = d2['current'];
          if (cur is int) {
            data = {
              'items': [
                {'id': cur},
              ],
            };
          }
        } catch (_) {}
      }
      return data;
    } catch (_) {
      // On error also try direct fallback to current capture
      try {
        final client = sl.get<Object>() as dynamic;
        final res2 = await client.get(path: '/_api/v1/capture');
        final d2 =
            (res2.data is Map<String, dynamic>)
                ? (res2.data as Map<String, dynamic>)
                : const <String, dynamic>{};
        final cur = d2['current'];
        if (cur is int) {
          return {
            'items': [
              {'id': cur},
            ],
          };
        }
      } catch (_) {}
      return <String, dynamic>{'items': []};
    }
  }

  Future<void> _apply() async {
    // Persist recording toggle via backend
    try {
      final client = sl.get<Object>() as dynamic;
      await client.post(
        path: '/_api/v1/capture',
        body: {'action': _recording ? 'start' : 'stop'},
      );
    } catch (_) {}

    // Update UI store
    final ui = sl<HomeUiStore>();
    ui.setIsRecording(_recording);
    ui.setCaptureScope(_scope);
    ui.setIncludePaused(_includePaused);

    if (mounted) Navigator.of(context).pop(true);
  }
}
