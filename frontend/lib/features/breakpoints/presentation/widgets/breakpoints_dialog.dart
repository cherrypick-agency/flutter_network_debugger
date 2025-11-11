import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'dart:convert';
// services not needed here; hotkeys are wired via HotkeysService
import '../../../../core/hotkeys/hotkeys_service.dart';
import 'package:archive/archive.dart';
import '../../../breakpoints/application/stores/breakpoints_store.dart';
import '../../../breakpoints/application/stores/intercept_queue_store.dart';
import '../../../breakpoints/application/stores/intercept_editor_store.dart';
import '../../../../core/di/di.dart';
import '../../../compose/presentation/widgets/key_value_editor.dart';
import '../../../compose/presentation/widgets/kv.dart';
import '../../../compose/presentation/widgets/body_editor.dart';
import '../../domain/entities/intercept_item.dart';
import '../../domain/entities/intercept_rule.dart';
import '../../domain/entities/intercept_config.dart';
import '../../../../core/notifications/notifications_service.dart';

class BreakpointsDialog extends StatefulWidget {
  const BreakpointsDialog({super.key});
  @override
  State<BreakpointsDialog> createState() => _BreakpointsDialogState();
}

class _BreakpointsDialogState extends State<BreakpointsDialog>
    with SingleTickerProviderStateMixin {
  late final TabController _tabs;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);
    // инициализация данных
    Future.microtask(() async {
      await sl<BreakpointsStore>().load();
      await sl<InterceptQueueStore>().init();
      // если очередь пуста — открываем Rules по умолчанию
      final q = sl<InterceptQueueStore>();
      if (mounted && q.items.isEmpty) {
        _tabs.index = 2;
      }
    });
  }

  @override
  void dispose() {
    // снимаем подписку очереди на монитор при закрытии диалога
    try {
      sl<InterceptQueueStore>().detach();
    } catch (_) {}
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 1200, maxHeight: 800),
        child: ScaffoldMessenger(
          child: Scaffold(
            backgroundColor: Colors.transparent,
            body: Material(
              elevation: 12,
              color: Theme.of(context).colorScheme.surface,
              borderRadius: BorderRadius.circular(12),
              child: MultiProvider(
                providers: [
                  ChangeNotifierProvider.value(value: sl<BreakpointsStore>()),
                  ChangeNotifierProvider.value(
                    value: sl<InterceptQueueStore>(),
                  ),
                  ChangeNotifierProvider.value(
                    value: sl<InterceptEditorStore>(),
                  ),
                ],
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Header
                    Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 12,
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              'Breakpoints',
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                          ),
                          IconButton(
                            onPressed: () => Navigator.of(context).pop(),
                            icon: Icon(Icons.close, color: cs.onSurfaceVariant),
                          ),
                        ],
                      ),
                    ),
                    const Divider(height: 1),
                    // Tabs
                    TabBar(
                      controller: _tabs,
                      tabs: const [
                        Tab(text: 'Queue'),
                        Tab(text: 'Editor'),
                        Tab(text: 'Rules'),
                      ],
                    ),
                    const Divider(height: 1),
                    Expanded(
                      child: TabBarView(
                        controller: _tabs,
                        children: const [
                          _QueuePanel(),
                          _EditorPanel(),
                          _RulesPanel(),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _QueuePanel extends StatelessWidget {
  const _QueuePanel();
  @override
  Widget build(BuildContext context) {
    final store = context.watch<InterceptQueueStore>();
    final items = store.items;
    final cs = Theme.of(context).colorScheme;
    final descriptionStyle = Theme.of(
      context,
    ).textTheme.bodySmall?.copyWith(color: cs.onSurfaceVariant);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
          child: Text(
            'Pending intercepted requests and responses waiting for action.',
            style: descriptionStyle,
          ),
        ),
        Expanded(
          child: items.isEmpty
              ? const Center(child: Text('Queue is empty'))
              : ListView.separated(
                  padding: const EdgeInsets.all(12),
                  itemCount: items.length,
                  separatorBuilder: (_, __) => const Divider(height: 1),
                  itemBuilder: (_, i) {
                    final it = items[i];
                    final ttlMs =
                        it.deadline.millisecondsSinceEpoch -
                        DateTime.now().toUtc().millisecondsSinceEpoch;
                    final totalMs = it.deadline
                        .difference(it.createdAt)
                        .inMilliseconds;
                    final pct = totalMs > 0
                        ? (1 - (ttlMs.clamp(0, totalMs) / totalMs))
                        : 0.0;
                    return ListTile(
                      title: Text(
                        '${it.direction == 'request' ? it.req?.method ?? '' : it.res?.status ?? ''} • ${it.direction}',
                      ),
                      subtitle: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            it.direction == 'request'
                                ? (it.req?.url ?? '')
                                : (it.res?.contentType ?? ''),
                          ),
                          const SizedBox(height: 4),
                          LinearProgressIndicator(value: pct),
                        ],
                      ),
                      onTap: () {
                        store.select(it.id);
                        context.read<InterceptEditorStore>().setItem(it);
                      },
                      trailing: Wrap(
                        spacing: 4,
                        children: [
                          IconButton(
                            tooltip: 'Continue',
                            onPressed: () => store.quickContinue(it.id),
                            icon: const Icon(Icons.play_arrow),
                          ),
                          IconButton(
                            tooltip: 'Cancel',
                            onPressed: () => store.quickCancel(it.id),
                            icon: const Icon(Icons.close),
                          ),
                        ],
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}

class _EditorPanel extends StatelessWidget {
  const _EditorPanel();
  @override
  Widget build(BuildContext context) {
    return const _EditorPanelStateful();
  }
}

class _EditorPanelStateful extends StatefulWidget {
  const _EditorPanelStateful();
  @override
  State<_EditorPanelStateful> createState() => _EditorPanelStatefulState();
}

class _EditorPanelStatefulState extends State<_EditorPanelStateful> {
  final TextEditingController _methodCtrl = TextEditingController();
  final TextEditingController _urlCtrl = TextEditingController();
  final TextEditingController _statusCtrl = TextEditingController();
  final TextEditingController _rawCtrl = TextEditingController();
  final TextEditingController _jsonCtrl = TextEditingController();
  List<KvPair> _headers = const [];
  String _mode = 'raw';
  bool _isBinary = false;
  bool _isTruncated = false;
  String? _contentEncoding;
  bool _viewDecompressed = true;
  bool _submitting = false;
  String? _contentType;
  String? _lastItemId;

  void _populateFromItem(InterceptItem it) {
    final isReq = it.direction == 'request';
    if (isReq) {
      _methodCtrl.text = it.req?.method ?? '';
      _urlCtrl.text = it.req?.url ?? '';
      _headers = _fromHeaderMap(it.req?.headers ?? const {});
      _contentEncoding = _firstHeader(_headers, 'Content-Encoding');
      _contentType = _firstHeader(_headers, 'Content-Type');
      final bodyStr = _safeB64Decode(it.req?.bodyBase64);
      _jsonCtrl.text = _looksLikeJson(bodyStr) ? _prettyJson(bodyStr) : '';
      _rawCtrl.text = !_looksLikeJson(bodyStr) ? bodyStr : '';
      _mode = _looksLikeJson(bodyStr) ? 'json' : 'raw';
      _isBinary =
          (it.req?.bodyBase64 != null &&
          it.req!.bodyBase64!.isNotEmpty &&
          bodyStr.isEmpty);
      _isTruncated = it.req?.bodyTruncated == true;
    } else {
      _statusCtrl.text = (it.res?.status ?? 0).toString();
      _headers = _fromHeaderMap(it.res?.headers ?? const {});
      _contentEncoding = _firstHeader(_headers, 'Content-Encoding');
      _contentType = _firstHeader(_headers, 'Content-Type');
      final bodyStr = _safeB64Decode(it.res?.bodyBase64);
      _jsonCtrl.text = _looksLikeJson(bodyStr) ? _prettyJson(bodyStr) : '';
      _rawCtrl.text = !_looksLikeJson(bodyStr) ? bodyStr : '';
      _mode = _looksLikeJson(bodyStr) ? 'json' : 'raw';
      _isBinary =
          (it.res?.bodyBase64 != null &&
          it.res!.bodyBase64!.isNotEmpty &&
          bodyStr.isEmpty);
      _isTruncated = it.res?.bodyTruncated == true;
    }
  }

  @override
  Widget build(BuildContext context) {
    final ed = context.watch<InterceptEditorStore>();
    final it = ed.item;
    if (it == null) {
      return const Center(child: Text('Queue is empty'));
    }
    if (_lastItemId != it.id) {
      _populateFromItem(it);
      _lastItemId = it.id;
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
                  if (isReq) ...[
                    SizedBox(
                      width: 100,
                      child: TextField(
                        controller: _methodCtrl,
                        decoration: const InputDecoration(labelText: 'Method'),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: TextField(
                        controller: _urlCtrl,
                        decoration: const InputDecoration(labelText: 'URL'),
                      ),
                    ),
                  ] else ...[
                    SizedBox(
                      width: 120,
                      child: TextField(
                        controller: _statusCtrl,
                        decoration: const InputDecoration(labelText: 'Status'),
                        keyboardType: TextInputType.number,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        it.res?.contentType ?? '',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                  const SizedBox(width: 12),
                  TextButton(
                    onPressed: () =>
                        _addOrUpdateHeader('Authorization', 'Bearer '),
                    child: const Text('+Auth'),
                  ),
                  const SizedBox(width: 4),
                  TextButton(
                    onPressed: () => _ensureContentTypeForMode(),
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
                    // Headers
                    Expanded(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          border: Border.all(
                            color: Theme.of(context).colorScheme.outline,
                          ),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: KeyValueEditor(
                          items: _headers,
                          onChanged: (v) {
                            setState(() {
                              _headers = v;
                              _updateContentInfo();
                            });
                          },
                          labelKey: 'Header',
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    // Body
                    Expanded(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          border: Border.all(
                            color: Theme.of(context).colorScheme.outline,
                          ),
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
                                  if (_contentEncoding != null &&
                                      _contentEncoding!.isNotEmpty)
                                    Chip(
                                      label: Text(
                                        _contentEncoding!.toLowerCase(),
                                      ),
                                    ),
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
                            if (_contentEncoding != null &&
                                _contentEncoding!.isNotEmpty &&
                                !_isBinary)
                              Padding(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                ),
                                child: Row(
                                  children: [
                                    Switch(
                                      value: _viewDecompressed,
                                      onChanged: (v) {
                                        setState(() => _viewDecompressed = v);
                                        _reloadBodyFromSnapshot(ed.item);
                                      },
                                    ),
                                    const SizedBox(width: 8),
                                    const Text(
                                      'Показывать декомпрессированное',
                                    ),
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
                                        Chip(
                                          label: const Text(
                                            'Binary body — editing disabled',
                                          ),
                                        ),
                                      if (_isTruncated)
                                        Chip(
                                          label: const Text(
                                            'Truncated preview',
                                          ),
                                        ),
                                    ],
                                  ),
                                ),
                              ),
                            Expanded(
                              child: _isBinary
                                  ? const Center(
                                      child: Text(
                                        'Body is binary, not displayed',
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
    final hdrs = _toHeaderMap(_headers);
    if (_mode == 'json') {
      // JSON валидация перед отправкой
      try {
        jsonDecode(_jsonCtrl.text);
      } catch (e) {
        sl<NotificationsService>().error('Invalid JSON', e.toString());
        return;
      }
      _ensureContentTypeForMode();
    }
    final bodyStr = _mode == 'json' ? _jsonCtrl.text : _rawCtrl.text;
    final bodyB64 = bodyStr.isEmpty ? null : base64Encode(utf8.encode(bodyStr));
    final queue = context.read<InterceptQueueStore>();
    setState(() => _submitting = true);
    try {
      if (isReq) {
        await ed.continueRequest(
          method: _methodCtrl.text.trim().isEmpty
              ? null
              : _methodCtrl.text.trim(),
          url: _urlCtrl.text.trim().isEmpty ? null : _urlCtrl.text.trim(),
          headers: hdrs.isEmpty ? null : hdrs,
          bodyBase64: bodyB64,
        );
      } else {
        final status = int.tryParse(_statusCtrl.text.trim());
        await ed.continueResponse(
          status: status,
          headers: hdrs.isEmpty ? null : hdrs,
          bodyBase64: bodyB64,
        );
      }
      // refresh queue after successful apply
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
    });
  }

  void _ensureContentTypeForMode() {
    final hasCT = _headers.any((h) => h.key.toLowerCase() == 'content-type');
    if (_mode == 'json' && !hasCT) {
      _addOrUpdateHeader('Content-Type', 'application/json; charset=utf-8');
    }
  }

  String _safeB64Decode(String? b64) {
    try {
      if (b64 == null || b64.isEmpty) return '';
      final bytes = base64Decode(b64);
      List<int> out = bytes;
      if (_viewDecompressed && _contentEncoding != null) {
        final enc = _contentEncoding!.toLowerCase();
        if (enc.contains('gzip')) {
          out = GZipDecoder().decodeBytes(bytes, verify: false);
        } else if (enc.contains('deflate')) {
          out = ZLibDecoder().decodeBytes(bytes, verify: false);
        }
      }
      return utf8.decode(out);
    } catch (_) {
      return '';
    }
  }

  bool _looksLikeJson(String s) {
    final t = s.trim();
    return (t.startsWith('{') && t.endsWith('}')) ||
        (t.startsWith('[') && t.endsWith(']'));
  }

  String _prettyJson(String s) {
    try {
      final obj = jsonDecode(s);
      const encoder = JsonEncoder.withIndent('  ');
      return encoder.convert(obj);
    } catch (_) {
      return s;
    }
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
      if (i.key.trim().isEmpty) continue;
      out.putIfAbsent(i.key, () => <String>[]).add(i.value);
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
    final bodyStr = _safeB64Decode(b64);
    setState(() {
      _jsonCtrl.text = _looksLikeJson(bodyStr) ? _prettyJson(bodyStr) : '';
      _rawCtrl.text = !_looksLikeJson(bodyStr) ? bodyStr : '';
      _mode = _looksLikeJson(bodyStr) ? 'json' : 'raw';
    });
  }

  void _updateContentInfo() {
    _contentEncoding = _firstHeader(_headers, 'Content-Encoding');
    _contentType = _firstHeader(_headers, 'Content-Type');
  }
}

class _RulesPanel extends StatelessWidget {
  const _RulesPanel();
  @override
  Widget build(BuildContext context) {
    return const _RulesEditor();
  }
}

class _RulesEditor extends StatefulWidget {
  const _RulesEditor();
  @override
  State<_RulesEditor> createState() => _RulesEditorState();
}

class _RulesEditorState extends State<_RulesEditor> {
  InterceptConfig? _cfg;
  late List<InterceptRule> _rules;
  bool _savingCfg = false;
  bool _savingRules = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final bp = context.read<BreakpointsStore>();
    _cfg ??= bp.config;
    _rules = bp.rules
        .map(
          (e) => InterceptRule(
            id: e.id,
            enabled: e.enabled,
            priority: e.priority,
            action: e.action,
            once: e.once,
            stopProcessing: e.stopProcessing,
            when: e.when,
          ),
        )
        .toList(growable: true);
  }

  @override
  Widget build(BuildContext context) {
    final bp = context.watch<BreakpointsStore>();
    final cfg = _cfg ?? bp.config;
    // если правила ещё не подхвачены (например, пришли после async load) — синхронизируем
    if (_rules.isEmpty && bp.rules.isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        setState(() {
          _rules = bp.rules
              .map(
                (e) => InterceptRule(
                  id: e.id,
                  enabled: e.enabled,
                  priority: e.priority,
                  action: e.action,
                  once: e.once,
                  stopProcessing: e.stopProcessing,
                  when: e.when,
                ),
              )
              .toList(growable: true);
          _cfg ??= bp.config;
        });
      });
    }
    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Config', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          if (cfg != null) _buildConfig(cfg),
          const SizedBox(height: 16),
          Row(
            children: [
              ElevatedButton(
                onPressed: _savingCfg || cfg == null
                    ? null
                    : () async {
                        setState(() => _savingCfg = true);
                        try {
                          await bp.saveConfig(cfg);
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Config saved')),
                          );
                        } catch (e) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text('Failed to save config: $e'),
                            ),
                          );
                        } finally {
                          setState(() => _savingCfg = false);
                        }
                      },
                child: Text(_savingCfg ? 'Saving…' : 'Save Config'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Text('Rules', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _rules.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (ctx, i) => _buildRuleItem(i),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              OutlinedButton.icon(
                onPressed: _addRule,
                icon: const Icon(Icons.add),
                label: const Text('Add Rule'),
              ),
              const SizedBox(width: 8),
              ElevatedButton(
                onPressed: _savingRules
                    ? null
                    : () async {
                        setState(() => _savingRules = true);
                        try {
                          await bp.replaceRules(_rules);
                          await bp.load();
                          setState(() {
                            _rules = bp.rules
                                .map(
                                  (e) => InterceptRule(
                                    id: e.id,
                                    enabled: e.enabled,
                                    priority: e.priority,
                                    action: e.action,
                                    once: e.once,
                                    stopProcessing: e.stopProcessing,
                                    when: e.when,
                                  ),
                                )
                                .toList(growable: true);
                          });
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Rules saved')),
                          );
                        } catch (e) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(content: Text('Failed to save rules: $e')),
                          );
                        } finally {
                          setState(() => _savingRules = false);
                        }
                      },
                child: Text(_savingRules ? 'Saving…' : 'Save Rules'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildConfig(InterceptConfig cfg) {
    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: [
        FilterChip(
          selected: cfg.enabled,
          label: const Text('Enabled'),
          onSelected: (v) => setState(() => _cfg = _copyCfg(cfg, enabled: v)),
        ),
        FilterChip(
          selected: cfg.requests,
          label: const Text('Requests'),
          onSelected: (v) => setState(() => _cfg = _copyCfg(cfg, requests: v)),
        ),
        FilterChip(
          selected: cfg.responses,
          label: const Text('Responses'),
          onSelected: (v) => setState(() => _cfg = _copyCfg(cfg, responses: v)),
        ),
        SizedBox(
          width: 160,
          child: TextFormField(
            initialValue: cfg.timeoutMs.toString(),
            decoration: const InputDecoration(labelText: 'Timeout (ms)'),
            keyboardType: TextInputType.number,
            onChanged: (v) => setState(
              () => _cfg = _copyCfg(
                cfg,
                timeoutMs: int.tryParse(v) ?? cfg.timeoutMs,
              ),
            ),
          ),
        ),
        SizedBox(
          width: 160,
          child: TextFormField(
            initialValue: cfg.queueMax.toString(),
            decoration: const InputDecoration(labelText: 'Queue max'),
            keyboardType: TextInputType.number,
            onChanged: (v) => setState(
              () => _cfg = _copyCfg(
                cfg,
                queueMax: int.tryParse(v) ?? cfg.queueMax,
              ),
            ),
          ),
        ),
        SizedBox(
          width: 180,
          child: TextFormField(
            initialValue: cfg.bodyMaxBytes.toString(),
            decoration: const InputDecoration(labelText: 'Body max bytes'),
            keyboardType: TextInputType.number,
            onChanged: (v) => setState(
              () => _cfg = _copyCfg(
                cfg,
                bodyMaxBytes: int.tryParse(v) ?? cfg.bodyMaxBytes,
              ),
            ),
          ),
        ),
        FilterChip(
          selected: cfg.reencode,
          label: const Text('Re-encode bodies'),
          onSelected: (v) => setState(() => _cfg = _copyCfg(cfg, reencode: v)),
        ),
        SizedBox(
          width: 220,
          child: DropdownButtonFormField<String>(
            value: cfg.overflow,
            decoration: const InputDecoration(labelText: 'Overflow policy'),
            onChanged: (v) => setState(
              () => _cfg = _copyCfg(cfg, overflow: v ?? cfg.overflow),
            ),
            items: const [
              'auto-continue-oldest',
              'drop-new',
            ].map((e) => DropdownMenuItem(value: e, child: Text(e))).toList(),
          ),
        ),
      ],
    );
  }

  Widget _buildRuleItem(int i) {
    final r = _rules[i];
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Checkbox(
            value: r.enabled,
            onChanged: (v) => setState(
              () => _rules[i] = _copyRule(r, enabled: v ?? r.enabled),
            ),
          ),
          SizedBox(
            width: 90,
            child: TextFormField(
              initialValue: r.priority.toString(),
              decoration: const InputDecoration(labelText: 'Priority'),
              keyboardType: TextInputType.number,
              onChanged: (v) => setState(
                () => _rules[i] = _copyRule(
                  r,
                  priority: int.tryParse(v) ?? r.priority,
                ),
              ),
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 140,
            child: DropdownButtonFormField<String>(
              value: r.action,
              decoration: const InputDecoration(labelText: 'Action'),
              onChanged: (v) => setState(
                () => _rules[i] = _copyRule(r, action: v ?? r.action),
              ),
              items: const [
                'request',
                'response',
                'both',
              ].map((e) => DropdownMenuItem(value: e, child: Text(e))).toList(),
            ),
          ),
          const SizedBox(width: 8),
          FilterChip(
            selected: r.once,
            label: const Text('Once'),
            onSelected: (v) =>
                setState(() => _rules[i] = _copyRule(r, once: v)),
          ),
          const SizedBox(width: 8),
          FilterChip(
            selected: r.stopProcessing,
            label: const Text('Stop'),
            onSelected: (v) =>
                setState(() => _rules[i] = _copyRule(r, stopProcessing: v)),
          ),
          const Spacer(),
          IconButton(
            onPressed: () => setState(() => _rules.removeAt(i)),
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Delete',
          ),
        ],
      ),
    );
  }

  void _addRule() {
    setState(() {
      _rules.add(
        InterceptRule(
          id: '',
          enabled: true,
          priority: (_rules.isEmpty
              ? 0
              : (_rules.map((e) => e.priority).reduce((a, b) => a > b ? a : b) +
                    1)),
          action: 'both',
          once: false,
          stopProcessing: true,
          when: const InterceptWhen(),
        ),
      );
    });
  }

  InterceptConfig _copyCfg(
    InterceptConfig c, {
    bool? enabled,
    bool? requests,
    bool? responses,
    int? timeoutMs,
    int? queueMax,
    int? bodyMaxBytes,
    bool? reencode,
    String? overflow,
  }) {
    return InterceptConfig(
      enabled: enabled ?? c.enabled,
      requests: requests ?? c.requests,
      responses: responses ?? c.responses,
      timeoutMs: timeoutMs ?? c.timeoutMs,
      queueMax: queueMax ?? c.queueMax,
      bodyMaxBytes: bodyMaxBytes ?? c.bodyMaxBytes,
      reencode: reencode ?? c.reencode,
      overflow: overflow ?? c.overflow,
    );
  }

  InterceptRule _copyRule(
    InterceptRule r, {
    bool? enabled,
    int? priority,
    String? action,
    bool? once,
    bool? stopProcessing,
  }) {
    return InterceptRule(
      id: r.id,
      enabled: enabled ?? r.enabled,
      priority: priority ?? r.priority,
      action: action ?? r.action,
      once: once ?? r.once,
      stopProcessing: stopProcessing ?? r.stopProcessing,
      when: r.when,
    );
  }
}
