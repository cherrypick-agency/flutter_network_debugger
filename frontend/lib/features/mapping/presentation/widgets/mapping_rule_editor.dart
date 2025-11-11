import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter/services.dart';

import '../../application/stores/mapping_store.dart';
import '../../data/mapping_api.dart';
import '../../data/mapping_repository_impl.dart';
import '../../domain/mapping_rule.dart';
import '../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart' as app_http;
import '../../../../core/notifications/notifications_service.dart';
import '../../../../theme/context_ext.dart';

class MappingRuleEditor extends StatefulWidget {
  const MappingRuleEditor({super.key, this.initial});

  final MappingRule? initial;

  @override
  State<MappingRuleEditor> createState() => _MappingRuleEditorState();
}

class _MappingRuleEditorState extends State<MappingRuleEditor> {
  late bool _enabled;
  late int _priority;
  late String _kind; // 'local' | 'remote'
  bool _stopProcessing = true;

  final TextEditingController _methods = TextEditingController();
  final TextEditingController _hostPattern = TextEditingController();
  final TextEditingController _pathPattern = TextEditingController();
  String _patternType = 'glob'; // or 'regex'

  // Local
  final TextEditingController _filePath = TextEditingController();
  String? _blobPath;
  final TextEditingController _statusOverride = TextEditingController();
  final TextEditingController _contentTypeOverride = TextEditingController();

  // Remote
  final TextEditingController _targetURLTemplate = TextEditingController();
  bool _preserveHost = false;

  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final it = widget.initial;
    _enabled = it?.enabled ?? true;
    _priority = it?.priority ?? 100;
    _kind = it?.kind ?? 'local';
    _stopProcessing = it?.stopProcessing ?? true;
    _methods.text = (it?.methods ?? const <String>[]).join(',');
    _hostPattern.text = it?.hostPattern ?? '';
    _pathPattern.text = it?.pathPattern ?? '';
    _patternType = it?.patternType ?? 'glob';
    _filePath.text = it?.filePath ?? '';
    _blobPath = it?.blobPath;
    if (it?.statusOverride != null) {
      _statusOverride.text = (it!.statusOverride!).toString();
    }
    _contentTypeOverride.text = it?.contentTypeOverride ?? '';
    _targetURLTemplate.text = it?.targetURLTemplate ?? '';
    _preserveHost = it?.preserveHost ?? false;
  }

  @override
  void dispose() {
    _methods.dispose();
    _hostPattern.dispose();
    _pathPattern.dispose();
    _filePath.dispose();
    _statusOverride.dispose();
    _contentTypeOverride.dispose();
    _targetURLTemplate.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 720),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    widget.initial == null
                        ? 'New Mapping Rule'
                        : 'Edit Mapping Rule',
                    style: context.appText.title,
                  ),
                ),
                IconButton(
                  onPressed: _saving ? null : () => Navigator.of(context).pop(),
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                SizedBox(
                  width: 160,
                  child: DropdownButtonFormField<String>(
                    value: _kind,
                    isDense: true,
                    items: const [
                      DropdownMenuItem(value: 'local', child: Text('Local')),
                      DropdownMenuItem(value: 'remote', child: Text('Remote')),
                    ],
                    onChanged: _saving
                        ? null
                        : (v) {
                            if (v == null) return;
                            setState(() {
                              _kind = v;
                            });
                          },
                    decoration: const InputDecoration(labelText: 'Kind'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    enabled: !_saving,
                    decoration: const InputDecoration(labelText: 'Priority'),
                    keyboardType: TextInputType.number,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    initialValue: _priority.toString(),
                    onChanged: (v) {
                      final n = int.tryParse(v.trim());
                      if (n != null) setState(() => _priority = n);
                    },
                  ),
                ),
                const SizedBox(width: 12),
                SizedBox(
                  width: 160,
                  child: SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                    title: const Text('Enabled'),
                    value: _enabled,
                    onChanged: _saving
                        ? null
                        : (v) => setState(() => _enabled = v),
                  ),
                ),
              ],
            ),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _hostPattern,
                    enabled: !_saving,
                    decoration: const InputDecoration(
                      labelText: 'Host pattern (glob/regex)',
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: _pathPattern,
                    enabled: !_saving,
                    decoration: const InputDecoration(
                      labelText: 'Path pattern (glob/regex)',
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _methods,
                    enabled: !_saving,
                    decoration: const InputDecoration(
                      labelText: 'HTTP methods (comma-separated)',
                      hintText: 'GET,POST',
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                SizedBox(
                  width: 160,
                  child: DropdownButtonFormField<String>(
                    value: _patternType,
                    isDense: true,
                    items: const [
                      DropdownMenuItem(value: 'glob', child: Text('glob')),
                      DropdownMenuItem(value: 'regex', child: Text('regex')),
                    ],
                    onChanged: _saving
                        ? null
                        : (v) => setState(() => _patternType = v ?? 'glob'),
                    decoration: const InputDecoration(
                      labelText: 'Pattern type',
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                SizedBox(
                  width: 200,
                  child: SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                    title: const Text('Stop processing'),
                    value: _stopProcessing,
                    onChanged: _saving
                        ? null
                        : (v) => setState(() => _stopProcessing = v),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (_kind == 'local') _buildLocal(cs, tt) else _buildRemote(cs, tt),
            const SizedBox(height: 16),
            Row(
              children: [
                const Spacer(),
                TextButton(
                  onPressed: _saving ? null : () => Navigator.of(context).pop(),
                  child: const Text('Cancel'),
                ),
                const SizedBox(width: 8),
                ElevatedButton.icon(
                  onPressed: _saving ? null : _save,
                  icon: _saving
                      ? const SizedBox(
                          width: 14,
                          height: 14,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.save),
                  label: const Text('Save'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLocal(ColorScheme cs, TextTheme tt) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: _filePath,
                enabled: !_saving,
                decoration: const InputDecoration(
                  labelText: 'File path (optional)',
                  hintText: '/path/to/file',
                ),
                onChanged: (_) {
                  // Если вводим путь — стираем blobPath
                  setState(() => _blobPath = null);
                },
              ),
            ),
            const SizedBox(width: 12),
            SizedBox(
              width: 200,
              child: TextButton.icon(
                onPressed: _saving ? null : _pickAndUpload,
                icon: const Icon(Icons.file_upload),
                label: Text(_blobPath == null ? 'Upload file' : 'Re-upload'),
              ),
            ),
          ],
        ),
        if (_blobPath != null) ...[
          const SizedBox(height: 6),
          Row(
            children: [
              Expanded(
                child: Text(
                  'Uploaded: $_blobPath',
                  style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                tooltip: 'Clear uploaded',
                onPressed: _saving
                    ? null
                    : () => setState(() => _blobPath = null),
                icon: const Icon(Icons.clear),
              ),
            ],
          ),
        ],
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: _statusOverride,
                enabled: !_saving,
                keyboardType: TextInputType.number,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: const InputDecoration(
                  labelText: 'HTTP status override (optional)',
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: TextField(
                controller: _contentTypeOverride,
                enabled: !_saving,
                decoration: const InputDecoration(
                  labelText: 'Content-Type override (optional)',
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildRemote(ColorScheme cs, TextTheme tt) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(
          controller: _targetURLTemplate,
          enabled: !_saving,
          decoration: const InputDecoration(
            labelText: 'Target URL template',
            hintText: 'https://example.com/{path}',
          ),
        ),
        const SizedBox(height: 8),
        SwitchListTile(
          dense: true,
          contentPadding: EdgeInsets.zero,
          title: const Text('Preserve Host header'),
          value: _preserveHost,
          onChanged: _saving ? null : (v) => setState(() => _preserveHost = v),
        ),
      ],
    );
  }

  Future<void> _pickAndUpload() async {
    try {
      final res = await FilePicker.platform.pickFiles(withData: true);
      final picked = res?.files.first;
      if (picked?.bytes == null || (picked?.name.isEmpty ?? true)) return;
      final http = sl<app_http.AppHttpClient>();
      final repo = MappingRepositoryImpl(MappingApi(http));
      final m = await repo.uploadFile(picked!.name, picked.bytes as Uint8List);
      setState(() {
        _blobPath = (m['blobPath'] ?? '').toString();
        if ((m['contentType'] ?? '').toString().isNotEmpty &&
            _contentTypeOverride.text.isEmpty) {
          _contentTypeOverride.text = m['contentType'].toString();
        }
      });
    } catch (e) {
      sl<NotificationsService>().error('Upload failed', e.toString());
    }
  }

  Future<void> _save() async {
    if (_saving) return;
    setState(() => _saving = true);
    try {
      final methods = _methods.text
          .split(',')
          .map((e) => e.trim().toUpperCase())
          .where((e) => e.isNotEmpty)
          .toList();

      final rule = MappingRule(
        id: widget.initial?.id ?? '',
        enabled: _enabled,
        priority: _priority,
        kind: _kind,
        stopProcessing: _stopProcessing,
        methods: methods,
        hostPattern: _hostPattern.text.trim(),
        pathPattern: _pathPattern.text.trim(),
        patternType: _patternType,
        filePath: _kind == 'local' && _filePath.text.trim().isNotEmpty
            ? _filePath.text.trim()
            : null,
        blobPath: _kind == 'local' ? _blobPath : null,
        statusOverride:
            _kind == 'local' && _statusOverride.text.trim().isNotEmpty
            ? int.tryParse(_statusOverride.text.trim())
            : null,
        contentTypeOverride:
            _kind == 'local' && _contentTypeOverride.text.trim().isNotEmpty
            ? _contentTypeOverride.text.trim()
            : null,
        targetURLTemplate: _kind == 'remote'
            ? _targetURLTemplate.text.trim()
            : null,
        preserveHost: _kind == 'remote' ? _preserveHost : false,
      );

      final store = context.read<MappingStore>();
      await store.upsert(rule);
      if (mounted) Navigator.of(context).pop(true);
    } catch (e) {
      sl<NotificationsService>().error('Save failed', e.toString());
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }
}
