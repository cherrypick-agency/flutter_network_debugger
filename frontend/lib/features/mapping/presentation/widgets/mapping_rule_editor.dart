import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter/services.dart';

import '../../application/stores/mapping_config_store.dart';
import '../../application/stores/mapping_store.dart';
import '../../data/mapping_api.dart';
import '../../domain/mapping_rule.dart';
import '../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart' as app_http;
import 'package:app_http_client/application/app_http_exception.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../theme/context_ext.dart';
import 'mapping_rule_editor/mapping_rule_error_banner.dart';
import 'mapping_rule_editor/mapping_rule_local_section.dart';
import 'mapping_rule_editor/mapping_rule_remote_section.dart';

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

  String? _serverErrorCode;
  String? _serverErrorMessage;
  String? _serverErrorField;

  void _clearServerErrorIfAny() {
    if (_serverErrorMessage == null) return;
    setState(() {
      _serverErrorCode = null;
      _serverErrorMessage = null;
      _serverErrorField = null;
    });
  }

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
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 720),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_serverErrorMessage != null) ...[
              MappingRuleErrorBanner(
                title: _serverErrorCode ?? 'Error',
                message: _serverErrorMessage!,
                onClose: () => setState(() {
                  _serverErrorCode = null;
                  _serverErrorMessage = null;
                  _serverErrorField = null;
                }),
              ),
              const SizedBox(height: 10),
            ],
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
                            _clearServerErrorIfAny();
                          },
                    decoration: InputDecoration(
                      labelText: 'Kind',
                      errorText: _errorTextFor('kind'),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    enabled: !_saving,
                    decoration: InputDecoration(
                      labelText: 'Priority',
                      errorText: _errorTextFor('priority'),
                    ),
                    keyboardType: TextInputType.number,
                    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                    initialValue: _priority.toString(),
                    onChanged: (v) {
                      final n = int.tryParse(v.trim());
                      if (n != null) setState(() => _priority = n);
                      _clearServerErrorIfAny();
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
                    decoration: InputDecoration(
                      labelText: 'Host pattern (glob/regex)',
                      errorText: _errorTextFor('hostPattern'),
                    ),
                    onChanged: (_) => _clearServerErrorIfAny(),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: _pathPattern,
                    enabled: !_saving,
                    decoration: InputDecoration(
                      labelText: 'Path pattern (glob/regex)',
                      errorText: _errorTextFor('pathPattern'),
                    ),
                    onChanged: (_) => _clearServerErrorIfAny(),
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
                    decoration: InputDecoration(
                      labelText: 'HTTP methods (comma-separated)',
                      hintText: 'GET,POST',
                      errorText: _errorTextFor('methods'),
                    ),
                    onChanged: (_) => _clearServerErrorIfAny(),
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
                        : (v) {
                            setState(() => _patternType = v ?? 'glob');
                            _clearServerErrorIfAny();
                          },
                    // если пользователь меняет тип сопоставления, старый server error уже не актуален
                    // (например, invalid regex)
                    decoration: InputDecoration(
                      labelText: 'Pattern type',
                      errorText: _errorTextFor('patternType'),
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
            if (_kind == 'local')
              MappingRuleLocalSection(
                filePathController: _filePath,
                statusOverrideController: _statusOverride,
                contentTypeOverrideController: _contentTypeOverride,
                saving: _saving,
                blobPath: _blobPath,
                onPickAndUpload: _pickAndUpload,
                onClearUploaded: () {
                  setState(() => _blobPath = null);
                  _clearServerErrorIfAny();
                },
                onFilePathChanged: (_) {
                  setState(() => _blobPath = null);
                  _clearServerErrorIfAny();
                },
                filePathErrorText: _errorTextFor('filePath'),
                blobPathErrorText: _errorTextFor('blobPath'),
                statusOverrideErrorText: _errorTextFor('statusOverride'),
                contentTypeOverrideErrorText: _errorTextFor(
                  'contentTypeOverride',
                ),
                onStatusOverrideChanged: (_) => _clearServerErrorIfAny(),
                onContentTypeOverrideChanged: (_) => _clearServerErrorIfAny(),
              )
            else
              MappingRuleRemoteSection(
                targetURLTemplateController: _targetURLTemplate,
                saving: _saving,
                preserveHost: _preserveHost,
                onPreserveHostChanged: (v) => setState(() => _preserveHost = v),
                targetURLTemplateErrorText: _errorTextFor('targetURLTemplate'),
                onTargetURLTemplateChanged: (_) => _clearServerErrorIfAny(),
              ),
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

  Future<void> _pickAndUpload() async {
    try {
      final res = await FilePicker.platform.pickFiles(
        withReadStream: !kIsWeb,
        withData: kIsWeb,
      );
      final picked = res?.files.first;
      if (picked == null || picked.name.isEmpty) return;

      var maxMB = 20;
      try {
        maxMB = context.read<MappingConfigStore>().config?.uploadMaxMB ?? 20;
      } catch (_) {}
      final maxBytes = maxMB * 1024 * 1024;
      if (picked.size > maxBytes) {
        sl<NotificationsService>().error(
          'File is too large',
          'Max allowed size is ${maxMB}MB',
        );
        return;
      }
      final http = sl<app_http.AppHttpClient>();
      final api = MappingApi(http);
      Map<String, dynamic> m;
      if (!kIsWeb && picked.readStream != null) {
        m = await api.uploadStream(
          fileName: picked.name,
          stream: picked.readStream!,
          length: picked.size,
        );
      } else if (picked.bytes != null) {
        m = await api.uploadBytes(picked.name, picked.bytes as Uint8List);
      } else {
        sl<NotificationsService>().error('Upload failed', 'Cannot read file');
        return;
      }
      setState(() {
        final blob = (m['blobPath'] ?? '').toString();
        _blobPath = blob.isEmpty ? null : blob;
        // На бэке запрещено одновременно filePath и blobPath.
        // Если загрузили blob — очищаем filePath, чтобы правило точно прошло валидацию.
        if (_filePath.text.trim().isNotEmpty) {
          _filePath.text = '';
        }
        if ((m['contentType'] ?? '').toString().isNotEmpty &&
            _contentTypeOverride.text.isEmpty) {
          _contentTypeOverride.text = m['contentType'].toString();
        }
        _serverErrorCode = null;
        _serverErrorMessage = null;
        _serverErrorField = null;
      });
    } catch (e) {
      if (e is AppHttpServerException) {
        setState(() {
          _serverErrorCode = e.serverError.code;
          _serverErrorMessage = e.messageFromServer;
          _serverErrorField = e.serverError.details?['field']?.toString();
        });
        return;
      }
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
      if (mounted) {
        setState(() {
          _serverErrorCode = null;
          _serverErrorMessage = null;
          _serverErrorField = null;
        });
      }
      if (mounted) Navigator.of(context).pop(true);
    } catch (e) {
      if (e is AppHttpServerException) {
        setState(() {
          _serverErrorCode = e.serverError.code;
          _serverErrorMessage = e.messageFromServer;
          _serverErrorField = e.serverError.details?['field']?.toString();
        });
        return;
      }
      sl<NotificationsService>().error('Save failed', e.toString());
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  String? _errorTextFor(String field) {
    final msg = _serverErrorMessage;
    final f = _serverErrorField;
    if (msg == null || f == null) return null;
    if (f == field) return msg;
    if (f.contains(field)) return msg;
    return null;
  }
}
