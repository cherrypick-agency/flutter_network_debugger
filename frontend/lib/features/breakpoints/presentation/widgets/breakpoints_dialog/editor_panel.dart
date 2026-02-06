import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/hotkeys/hotkeys_service.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../application/stores/breakpoints_store.dart';
import '../../../application/stores/intercept_editor_store.dart';
import '../../../application/stores/intercept_queue_store.dart';
import '../../../domain/entities/intercept_item.dart';
import '../../../../compose/presentation/widgets/kv.dart';
import 'editor_panel_body_section.dart';
import 'editor_panel_headers_editor.dart';
import 'editor_panel_toolbar.dart';

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
  InterceptItem? _lastItemRef;
  bool _suppressDirty = false;
  bool _headersDirty = false;
  bool _bodyDirty = false;
  bool _methodDirty = false;
  bool _urlDirty = false;
  bool _statusDirty = false;

  @override
  void initState() {
    super.initState();
    _rawCtrl.addListener(_markBodyDirty);
    _jsonCtrl.addListener(_markBodyDirty);
    _methodCtrl.addListener(_markMethodDirty);
    _urlCtrl.addListener(_markUrlDirty);
    _statusCtrl.addListener(_markStatusDirty);
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

  void _markMethodDirty() {
    if (_suppressDirty) return;
    _methodDirty = true;
  }

  void _markUrlDirty() {
    if (_suppressDirty) return;
    _urlDirty = true;
  }

  void _markStatusDirty() {
    if (_suppressDirty) return;
    _statusDirty = true;
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
      _methodDirty = false;
      _urlDirty = false;
      _statusDirty = false;
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
    _methodDirty = false;
    _urlDirty = false;
    _statusDirty = false;
    _suppressDirty = false;
  }

  @override
  Widget build(BuildContext context) {
    final ed = sl<InterceptEditorStore>();

    return Observer(builder: (_) {
      final it = ed.item;
      if (it == null) {
        return const Center(child: Text('Queue is empty'));
      }

      final canAutoSyncFromItem =
          !_submitting &&
          !_headersDirty &&
          !_bodyDirty &&
          !_methodDirty &&
          !_urlDirty &&
          !_statusDirty;
      if (_lastItemId != it.id) {
        _populateFromItem(it);
        _lastItemId = it.id;
        _lastItemRef = it;
      } else if (canAutoSyncFromItem && !identical(it, _lastItemRef)) {
        // Item may be updated (e.g., server refresh) without changing id.
        _populateFromItem(it);
        _lastItemRef = it;
      }

      final isReq = it.direction == 'request';
      final hk = sl<HotkeysService>();
      final bindings = hk.buildHandlers({
        'breakpoints.applyContinue': () => _applyAndContinue(ed, it),
        'breakpoints.applyContinue.ctrl': () => _applyAndContinue(ed, it),
        'breakpoints.cancel': () => _cancelItem(ed),
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
                EditorPanelToolbar(
                  isRequest: isReq,
                  methodController: _methodCtrl,
                  urlController: _urlCtrl,
                  statusController: _statusCtrl,
                  responseContentType: it.res?.contentType,
                  submitting: _submitting,
                  onAddAuth: () => _addOrUpdateHeader('Authorization', 'Bearer '),
                  onAddJson: _ensureContentTypeForMode,
                  onDrop: () => _dropRequest(ed),
                  onCancel: () => _cancelItem(ed),
                  onContinue: () => _applyAndContinue(ed, it),
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: Row(
                    children: [
                      Expanded(
                        child: EditorPanelHeadersEditor(
                          headers: _headers,
                          onChanged: (v) {
                            setState(() {
                              _headers = v;
                              _headersDirty = true;
                              _updateContentInfo();
                            });
                          },
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: EditorPanelBodySection(
                          contentType: _contentType,
                          reencode:
                              sl<BreakpointsStore>().config?.reencode == true,
                          isBinary: _isBinary,
                          isTruncated: _isTruncated,
                          mode: _mode,
                          onModeChanged: (m) {
                            setState(() => _mode = m);
                            _ensureContentTypeForMode();
                          },
                          rawController: _rawCtrl,
                          jsonController: _jsonCtrl,
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
    });
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
    final queue = sl<InterceptQueueStore>();

    setState(() => _submitting = true);
    var success = false;
    try {
      if (isReq) {
        final method = _methodCtrl.text.trim();
        final url = _urlCtrl.text.trim();
        await ed.continueRequest(
          method: _methodDirty ? (method.isEmpty ? null : method) : null,
          url: _urlDirty ? (url.isEmpty ? null : url) : null,
          // Empty map means "clear headers", null means "do not change".
          headers: hdrs,
          bodyBase64: bodyB64,
        );
      } else {
        int? status;
        if (_statusDirty) {
          status = int.tryParse(_statusCtrl.text.trim());
          if (status == null || status < 100 || status > 599) {
            sl<NotificationsService>().error(
              'Invalid status',
              'Must be a number in range 100..599',
            );
            return;
          }
        }
        await ed.continueResponse(
          status: status,
          // Empty map means "clear headers", null means "do not change".
          headers: hdrs,
          bodyBase64: bodyB64,
        );
      }
      success = true;
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
      if (success) {
        _headersDirty = false;
        _bodyDirty = false;
        _methodDirty = false;
        _urlDirty = false;
        _statusDirty = false;
      }
      if (mounted) setState(() => _submitting = false);
    }
  }

  Future<void> _cancelItem(InterceptEditorStore ed) async {
    if (_submitting) return;
    final queue = sl<InterceptQueueStore>();
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
    if (_submitting) return;
    final queue = sl<InterceptQueueStore>();
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
      // The snapshot from the backend already contains the decoded body (decodeForIntercept),
      // re-decompressing by Content-Encoding here corrupts the data.
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

  void _updateContentInfo() {
    _contentType = _firstHeader(_headers, 'Content-Type');
  }
}
