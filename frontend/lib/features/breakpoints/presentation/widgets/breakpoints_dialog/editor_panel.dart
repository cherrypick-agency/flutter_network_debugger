import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/hotkeys/hotkeys_service.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../../../theme/context_ext.dart';
import '../../../application/stores/breakpoints_store.dart';
import '../../../application/stores/intercept_editor_store.dart';
import '../../../application/stores/intercept_queue_store.dart';
import '../../../domain/entities/intercept_item.dart';
import '../../../../compose/presentation/widgets/body_editor.dart';
import '../../../../compose/presentation/widgets/key_value_editor.dart';
import '../../../../compose/presentation/widgets/kv.dart';

class EditorPanel extends StatefulWidget {
  const EditorPanel({super.key});

  @override
  State<EditorPanel> createState() => _EditorPanelState();
}

class _EditorPanelState extends State<EditorPanel> {
  final TextEditingController _methodCtrl = TextEditingController();
  final TextEditingController _urlCtrl = TextEditingController();
  final TextEditingController _statusCtrl = TextEditingController();
  final TextEditingController _rawCtrl = TextEditingController();
  final TextEditingController _jsonCtrl = TextEditingController();

  List<KvPair> _headers = const [];
  String _mode = 'raw';
  bool _isBinary = false;
  bool _isTruncated = false;
  String? _contentType;
  bool _submitting = false;
  String? _lastItemId;
  String? _lastSelectedId;
  InterceptItem? _lastItemRef;
  bool _suppressDirty = false;
  bool _headersDirty = false;
  bool _bodyDirty = false;

  @override
  void initState() {
    super.initState();
    _rawCtrl.addListener(_markBodyDirty);
    _jsonCtrl.addListener(_markBodyDirty);
  }

  @override
  void dispose() {
    _methodCtrl.dispose();
    _urlCtrl.dispose();
    _statusCtrl.dispose();
    _rawCtrl.dispose();
    _jsonCtrl.dispose();
    super.dispose();
  }

  void _markBodyDirty() {
    if (_suppressDirty) return;
    _bodyDirty = true;
  }

  void _populateFromItem(InterceptItem it) {
    _suppressDirty = true;
    final isReq = it.direction == 'request';

    if (isReq) {
      _methodCtrl.text = it.req?.method ?? '';
      _urlCtrl.text = it.req?.url ?? '';
      _headers = _fromHeaderMap(it.req?.headers ?? const {});
      _contentType = _firstHeader(_headers, 'Content-Type');

      final decoded = _safeB64Decode(it.req?.bodyBase64);
      final bodyStr = decoded ?? '';
      _jsonCtrl.text = _looksLikeJson(bodyStr) ? bodyStr : '';
      _rawCtrl.text = !_looksLikeJson(bodyStr) ? bodyStr : '';
      _mode = _looksLikeJson(bodyStr) ? 'json' : 'raw';

      _isBinary =
          (it.req?.bodyBase64 != null &&
          it.req!.bodyBase64!.isNotEmpty &&
          decoded == null);
      _isTruncated = it.req?.bodyTruncated == true;
      _headersDirty = false;
      _bodyDirty = false;
      _suppressDirty = false;
      return;
    }

    _statusCtrl.text = (it.res?.status ?? 0).toString();
    _headers = _fromHeaderMap(it.res?.headers ?? const {});
    _contentType = _firstHeader(_headers, 'Content-Type');

    final decoded = _safeB64Decode(it.res?.bodyBase64);
    final bodyStr = decoded ?? '';
    _jsonCtrl.text = _looksLikeJson(bodyStr) ? bodyStr : '';
    _rawCtrl.text = !_looksLikeJson(bodyStr) ? bodyStr : '';
    _mode = _looksLikeJson(bodyStr) ? 'json' : 'raw';

    _isBinary =
        (it.res?.bodyBase64 != null &&
        it.res!.bodyBase64!.isNotEmpty &&
        decoded == null);
    _isTruncated = it.res?.bodyTruncated == true;
    _headersDirty = false;
    _bodyDirty = false;
    _suppressDirty = false;
  }

  @override
  Widget build(BuildContext context) {
    final ed = context.watch<InterceptEditorStore>();
    final queue = context.watch<InterceptQueueStore>();
    final selectedId = queue.selected?.id;
    if (selectedId != _lastSelectedId) {
      _lastSelectedId = selectedId;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        context.read<InterceptEditorStore>().setItem(queue.selected);
      });
    }

    final it = ed.item;
    if (it == null) {
      return const Center(child: Text('Queue is empty'));
    }

    final canAutoSyncFromItem = !_submitting && !_headersDirty && !_bodyDirty;
    if (_lastItemId != it.id) {
      _populateFromItem(it);
      _lastItemId = it.id;
      _lastItemRef = it;
    } else if (canAutoSyncFromItem && !identical(it, _lastItemRef)) {
      // Item может обновиться (например, server refresh) без смены id.
      _populateFromItem(it);
      _lastItemRef = it;
    }

    final isReq = it.direction == 'request';
    final hk = sl<HotkeysService>();
    final bindings = hk.buildHandlers({
      'breakpoints.applyContinue': () => _applyAndContinue(ed, it),
      'breakpoints.applyContinue.ctrl': () => _applyAndContinue(ed, it),
      'breakpoints.cancel': () => ed.cancel(),
    });

    return CallbackShortcuts(
      bindings: bindings,
      child: Focus(
        autofocus: true,
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  if (isReq)
                    SizedBox(
                      width: 100,
                      child: TextField(
                        controller: _methodCtrl,
                        decoration: const InputDecoration(labelText: 'Method'),
                      ),
                    ),
                  if (isReq) const SizedBox(width: 8),
                  if (isReq)
                    Expanded(
                      child: TextField(
                        controller: _urlCtrl,
                        decoration: const InputDecoration(labelText: 'URL'),
                      ),
                    ),
                  if (!isReq)
                    SizedBox(
                      width: 120,
                      child: TextField(
                        controller: _statusCtrl,
                        decoration: const InputDecoration(labelText: 'Status'),
                        keyboardType: TextInputType.number,
                      ),
                    ),
                  if (!isReq) const SizedBox(width: 8),
                  if (!isReq)
                    Expanded(
                      child: Text(
                        it.res?.contentType ?? '',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: context.appText.body,
                      ),
                    ),
                  const SizedBox(width: 12),
                  TextButton(
                    onPressed: () =>
                        _addOrUpdateHeader('Authorization', 'Bearer '),
                    child: const Text('+Auth'),
                  ),
                  const SizedBox(width: 4),
                  TextButton(
                    onPressed: _ensureContentTypeForMode,
                    child: const Text('+JSON'),
                  ),
                  if (isReq)
                    FilledButton(
                      onPressed: _submitting ? null : () => _dropRequest(ed),
                      child: const Text('Drop'),
                    ),
                  const SizedBox(width: 8),
                  OutlinedButton(
                    onPressed: _submitting ? null : () => _cancelItem(ed),
                    child: const Text('Cancel'),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton(
                    onPressed: _submitting
                        ? null
                        : () => _applyAndContinue(ed, it),
                    child: const Text('Continue'),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Expanded(
                child: Row(
                  children: [
                    Expanded(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          border: Border.all(color: context.appColors.border),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: KeyValueEditor(
                          items: _headers,
                          onChanged: (v) {
                            setState(() {
                              _headers = v;
                              _headersDirty = true;
                              _updateContentInfo();
                            });
                          },
                          labelKey: 'Header',
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          border: Border.all(color: context.appColors.border),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Padding(
                              padding: const EdgeInsets.all(8),
                              child: Wrap(
                                spacing: 8,
                                children: [
                                  if (_contentType != null &&
                                      _contentType!.isNotEmpty)
                                    Chip(
                                      label: Text(_contentType!.toLowerCase()),
                                    ),
                                  if (context
                                          .read<BreakpointsStore>()
                                          .config
                                          ?.reencode ==
                                      true)
                                    const Chip(label: Text('Re-encode')),
                                  if (_isBinary)
                                    const Chip(label: Text('Binary')),
                                  if (_isTruncated)
                                    const Chip(label: Text('Truncated')),
                                ],
                              ),
                            ),
                            if (_isBinary || _isTruncated)
                              Padding(
                                padding: const EdgeInsets.all(8),
                                child: Align(
                                  alignment: Alignment.centerLeft,
                                  child: Wrap(
                                    spacing: 8,
                                    children: [
                                      if (_isBinary)
                                        const Chip(
                                          label: Text(
                                            'Binary body — editing disabled',
                                          ),
                                        ),
                                      if (_isTruncated)
                                        const Chip(
                                          label: Text('Truncated preview'),
                                        ),
                                    ],
                                  ),
                                ),
                              ),
                            Expanded(
                              child: (_isBinary || _isTruncated)
                                  ? const Center(
                                      child: Text(
                                        'Body preview is not editable',
                                      ),
                                    )
                                  : BodyEditor(
                                      mode: _mode,
                                      onModeChanged: (m) {
                                        setState(() => _mode = m);
                                        _ensureContentTypeForMode();
                                      },
                                      rawCtrl: _rawCtrl,
                                      jsonCtrl: _jsonCtrl,
                                      form: const [],
                                      multipart: const [],
                                      allowedModes: const ['raw', 'json'],
                                    ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _applyAndContinue(
    InterceptEditorStore ed,
    InterceptItem it,
  ) async {
    final isReq = it.direction == 'request';
    final hdrs = _headersDirty ? _toHeaderMap(_headers) : null;

    if (_mode == 'json') {
      final text = _jsonCtrl.text.trim();
      if (text.isNotEmpty) {
        try {
          jsonDecode(text);
        } catch (e) {
          sl<NotificationsService>().error('Invalid JSON', e.toString());
          return;
        }
      }
      _ensureContentTypeForMode();
    }

    final bodyStr = _mode == 'json' ? _jsonCtrl.text : _rawCtrl.text;
    final bodyB64 = (_isBinary || _isTruncated || !_bodyDirty)
        ? null
        : base64Encode(utf8.encode(bodyStr));
    final queue = context.read<InterceptQueueStore>();

    setState(() => _submitting = true);
    try {
      if (isReq) {
        await ed.continueRequest(
          method: _methodCtrl.text.trim().isEmpty
              ? null
              : _methodCtrl.text.trim(),
          url: _urlCtrl.text.trim().isEmpty ? null : _urlCtrl.text.trim(),
          // Пустая map означает "очистить заголовки", null означает "не менять".
          headers: hdrs,
          bodyBase64: bodyB64,
        );
      } else {
        final status = int.tryParse(_statusCtrl.text.trim());
        await ed.continueResponse(
          status: status,
          // Пустая map означает "очистить заголовки", null означает "не менять".
          headers: hdrs,
          bodyBase64: bodyB64,
        );
      }
      try {
        await queue.refresh();
      } catch (_) {}
      sl<NotificationsService>().info('Applied', 'Changes sent to backend');
    } catch (e) {
      sl<NotificationsService>().error('Failed', e.toString());
      try {
        await queue.refresh();
      } catch (_) {}
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  Future<void> _cancelItem(InterceptEditorStore ed) async {
    final queue = context.read<InterceptQueueStore>();
    setState(() => _submitting = true);
    try {
      await ed.cancel();
      try {
        await queue.refresh();
      } catch (_) {}
      sl<NotificationsService>().info('Canceled', 'Intercept canceled');
    } catch (e) {
      sl<NotificationsService>().error('Failed', e.toString());
      try {
        await queue.refresh();
      } catch (_) {}
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  Future<void> _dropRequest(InterceptEditorStore ed) async {
    final queue = context.read<InterceptQueueStore>();
    setState(() => _submitting = true);
    try {
      await ed.continueRequest(drop: true);
      try {
        await queue.refresh();
      } catch (_) {}
      sl<NotificationsService>().warn('Dropped', 'Request dropped');
    } catch (e) {
      sl<NotificationsService>().error('Failed', e.toString());
      try {
        await queue.refresh();
      } catch (_) {}
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _addOrUpdateHeader(String key, String valueTemplate) {
    final idx = _headers.indexWhere(
      (e) => e.key.toLowerCase() == key.toLowerCase(),
    );
    setState(() {
      if (idx >= 0) {
        _headers[idx] = _headers[idx].copyWith(value: valueTemplate);
      } else {
        _headers = [..._headers, KvPair(key, valueTemplate)];
      }
      _headersDirty = true;
      _updateContentInfo();
    });
  }

  void _ensureContentTypeForMode() {
    final hasCT = _headers.any((h) => h.key.toLowerCase() == 'content-type');
    if (_mode == 'json' && !hasCT) {
      _addOrUpdateHeader('Content-Type', 'application/json; charset=utf-8');
    }
  }

  String? _safeB64Decode(String? b64) {
    try {
      if (b64 == null || b64.isEmpty) return '';
      final bytes = base64Decode(b64);
      // В snapshot от бэка уже лежит декодированное тело (decodeForIntercept),
      // повторная декомпрессия по Content-Encoding здесь ломает данные.
      return utf8.decode(bytes);
    } catch (_) {
      return null;
    }
  }

  bool _looksLikeJson(String s) {
    final t = s.trim();
    return (t.startsWith('{') && t.endsWith('}')) ||
        (t.startsWith('[') && t.endsWith(']'));
  }

  List<KvPair> _fromHeaderMap(Map<String, List<String>> m) {
    final out = <KvPair>[];
    m.forEach((k, vals) {
      if (vals.isEmpty) {
        out.add(KvPair(k, ''));
      } else {
        for (final v in vals) {
          out.add(KvPair(k, v));
        }
      }
    });
    return out;
  }

  Map<String, List<String>> _toHeaderMap(List<KvPair> items) {
    final out = <String, List<String>>{};
    for (final i in items) {
      final key = i.key.trim();
      if (key.isEmpty) continue;
      out.putIfAbsent(key, () => <String>[]).add(i.value);
    }
    return out;
  }

  String? _firstHeader(List<KvPair> items, String headerName) {
    final name = headerName.toLowerCase();
    for (final i in items) {
      if (i.key.toLowerCase() == name) return i.value;
    }
    return null;
  }

  void _reloadBodyFromSnapshot(InterceptItem? it) {
    if (it == null) return;
    final b64 = it.direction == 'request'
        ? it.req?.bodyBase64
        : it.res?.bodyBase64;
    final bodyStr = _safeB64Decode(b64) ?? '';
    setState(() {
      _suppressDirty = true;
      _jsonCtrl.text = _looksLikeJson(bodyStr) ? bodyStr : '';
      _rawCtrl.text = !_looksLikeJson(bodyStr) ? bodyStr : '';
      _mode = _looksLikeJson(bodyStr) ? 'json' : 'raw';
      _bodyDirty = false;
      _suppressDirty = false;
    });
  }

  void _updateContentInfo() {
    _contentType = _firstHeader(_headers, 'Content-Type');
  }
}
