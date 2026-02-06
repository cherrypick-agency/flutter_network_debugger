import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../../../../core/di/di.dart';
import '../../application/stores/mapping_config_store.dart';
import '../../application/stores/mapping_store.dart';
import '../../domain/mapping_rule.dart';
import '../../domain/repositories/mapping_repository.dart';
import 'package:app_http_client/application/app_http_exception.dart';
import '../../../../core/notifications/notifications_service.dart';
import 'mapping_rule_editor/advanced_options_tile.dart';
import 'mapping_rule_editor/editor_actions_row.dart';
import 'mapping_rule_editor/editor_basics_row.dart';
import 'mapping_rule_editor/editor_header.dart';
import 'mapping_rule_editor/editor_methods_row.dart';
import 'mapping_rule_editor/editor_patterns_row.dart';
import 'mapping_rule_editor/mapping_rule_error_banner.dart';
import 'mapping_rule_editor/mapping_rule_local_section.dart';
import 'mapping_rule_editor/mapping_rule_remote_section.dart';
import 'mapping_rule_editor/section_card.dart';

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
  bool _uploading = false;

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
        child: SingleChildScrollView(
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
              MappingRuleEditorHeader(
                isNew: widget.initial == null,
                saving: _saving,
                onClose: () => Navigator.of(context).pop(),
              ),
              const SizedBox(height: 10),
              MappingRuleEditorSectionCard(
                title: 'Rule',
                subtitle: 'Basic settings and priority.',
                child: MappingRuleEditorBasicsRow(
                  kind: _kind,
                  priority: _priority,
                  enabled: _enabled,
                  saving: _saving,
                  kindErrorText: _errorTextFor('kind'),
                  priorityErrorText: _errorTextFor('priority'),
                  onKindChanged: (v) {
                    setState(() => _kind = v);
                    _clearServerErrorIfAny();
                  },
                  onPriorityChanged: (n) {
                    setState(() => _priority = n);
                    _clearServerErrorIfAny();
                  },
                  onEnabledChanged: (v) => setState(() => _enabled = v),
                ),
              ),
              const SizedBox(height: 12),
              MappingRuleEditorSectionCard(
                title: 'Match',
                subtitle:
                    'Which requests/responses this rule matches. Empty fields match more broadly.',
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    MappingRuleEditorPatternsRow(
                      hostPatternController: _hostPattern,
                      pathPatternController: _pathPattern,
                      saving: _saving,
                      hostErrorText: _errorTextFor('hostPattern'),
                      pathErrorText: _errorTextFor('pathPattern'),
                      onChanged: (_) => _clearServerErrorIfAny(),
                    ),
                    const SizedBox(height: 10),
                    MappingRuleEditorMethodsRow(
                      methodsController: _methods,
                      patternType: _patternType,
                      saving: _saving,
                      methodsErrorText: _errorTextFor('methods'),
                      patternTypeErrorText: _errorTextFor('patternType'),
                      onMethodsChanged: (_) => _clearServerErrorIfAny(),
                      onPatternTypeChanged: (v) {
                        setState(() => _patternType = v);
                        _clearServerErrorIfAny();
                      },
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 12),
              MappingRuleEditorSectionCard(
                title: 'Action',
                subtitle: _kind == 'local'
                    ? 'Serve a local file or an uploaded blob.'
                    : 'Proxy to another URL using a template.',
                child: _kind == 'local'
                    ? MappingRuleLocalSection(
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
                        statusOverrideErrorText: _errorTextFor(
                          'statusOverride',
                        ),
                        contentTypeOverrideErrorText: _errorTextFor(
                          'contentTypeOverride',
                        ),
                        onStatusOverrideChanged: (_) =>
                            _clearServerErrorIfAny(),
                        onContentTypeOverrideChanged: (_) =>
                            _clearServerErrorIfAny(),
                      )
                    : MappingRuleRemoteSection(
                        targetURLTemplateController: _targetURLTemplate,
                        saving: _saving,
                        preserveHost: _preserveHost,
                        onPreserveHostChanged: (v) =>
                            setState(() => _preserveHost = v),
                        targetURLTemplateErrorText: _errorTextFor(
                          'targetURLTemplate',
                        ),
                        onTargetURLTemplateChanged: (_) =>
                            _clearServerErrorIfAny(),
                      ),
              ),
              const SizedBox(height: 12),
              MappingRuleEditorSectionCard(
                title: 'Advanced',
                subtitle: 'Usually you can leave this as-is.',
                child: MappingRuleEditorAdvancedOptionsTile(
                  stopProcessing: _stopProcessing,
                  saving: _saving,
                  onStopProcessingChanged: (v) =>
                      setState(() => _stopProcessing = v),
                ),
              ),
              const SizedBox(height: 16),
              MappingRuleEditorActionsRow(
                saving: _saving,
                onCancel: () => Navigator.of(context).pop(),
                onSave: _save,
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _pickAndUpload() async {
    if (_uploading) return;
    setState(() => _uploading = true);
    try {
      final res = await FilePicker.platform.pickFiles(
        withReadStream: !kIsWeb,
        withData: kIsWeb,
      );
      final picked = res?.files.first;
      if (picked == null || picked.name.isEmpty) return;

      var maxMB = 20;
      try {
        maxMB = sl<MappingConfigStore>().config?.uploadMaxMB ?? 20;
      } catch (_) {}
      final maxBytes = maxMB * 1024 * 1024;
      if (picked.size > maxBytes) {
        sl<NotificationsService>().error(
          'File is too large',
          'Max allowed size is ${maxMB}MB',
        );
        return;
      }
      final repo = sl<MappingRepository>();
      Map<String, dynamic> m;
      if (!kIsWeb && picked.readStream != null) {
        m = await repo.uploadStream(
          fileName: picked.name,
          stream: picked.readStream!,
          length: picked.size,
        );
      } else if (picked.bytes != null) {
        m = await repo.uploadFile(picked.name, picked.bytes as Uint8List);
      } else {
        sl<NotificationsService>().error('Upload failed', 'Cannot read file');
        return;
      }
      if (!mounted) return;
      setState(() {
        final blob = (m['blobPath'] ?? '').toString();
        _blobPath = blob.isEmpty ? null : blob;
        // Backend does not allow both filePath and blobPath at the same time.
        // If a blob was uploaded, clear filePath so the rule passes validation.
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
        if (!mounted) return;
        setState(() {
          _serverErrorCode = e.serverError.code;
          _serverErrorMessage = e.messageFromServer;
          _serverErrorField = e.serverError.details?['field']?.toString();
        });
        return;
      }
      sl<NotificationsService>().error('Upload failed', e.toString());
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  String? _validateLocally() {
    if (_patternType == 'regex') {
      final host = _hostPattern.text.trim();
      if (host.isNotEmpty) {
        try {
          RegExp(host);
        } catch (e) {
          return 'Invalid host regex: ${e.toString().replaceAll('FormatException: ', '')}';
        }
      }
      final path = _pathPattern.text.trim();
      if (path.isNotEmpty) {
        try {
          RegExp(path);
        } catch (e) {
          return 'Invalid path regex: ${e.toString().replaceAll('FormatException: ', '')}';
        }
      }
    }
    if (_kind == 'remote' && _targetURLTemplate.text.trim().isEmpty) {
      return 'Target URL template is required for remote rules';
    }
    return null;
  }

  Future<void> _save() async {
    if (_saving) return;
    final localError = _validateLocally();
    if (localError != null) {
      setState(() {
        _serverErrorCode = 'VALIDATION';
        _serverErrorMessage = localError;
        _serverErrorField = null;
      });
      return;
    }
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

      final store = sl<MappingStore>();
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
