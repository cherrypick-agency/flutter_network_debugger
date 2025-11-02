//
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../domain/models.dart';
import '../../data/compose_repository.dart';
import '../../application/compose_store.dart';
import '../widgets/library_animated_tree.dart';
import '../widgets/folder_picker_dialog.dart';
import '../widgets/compose_http_details.dart';
import '../widgets/key_value_editor.dart';
import '../widgets/kv.dart';
import '../widgets/body_editor.dart';
import '../widgets/auth_editor.dart';
import '../widgets/history_list.dart';
import '../../../../../../core/di/di.dart';
import '../../../inspector/application/stores/home_ui_store.dart';
import '../../../../core/utils/debouncer.dart';
import 'package:app_http_client/application/app_http_exception.dart';
import '../../../../services/prefs.dart';
import 'dart:convert';

class ComposePage extends StatefulWidget {
  const ComposePage({super.key});
  @override
  State<ComposePage> createState() => _ComposePageState();
}

class _ComposePageState extends State<ComposePage> {
  String _method = 'GET';
  final TextEditingController _urlCtrl = TextEditingController();
  final List<KvPair> _headers = [];
  final List<KvPair> _query = [];
  String _bodyMode = 'raw';
  final TextEditingController _rawCtrl = TextEditingController();
  final TextEditingController _jsonCtrl = TextEditingController(
    text: '{\n  \n}',
  );
  final List<ComposeFormFieldDTO> _form = [];
  final List<ComposePartDTO> _multipart = [];
  String _authType = 'none';
  final TextEditingController _basicUser = TextEditingController();
  final TextEditingController _basicPass = TextEditingController();
  final TextEditingController _bearer = TextEditingController();
  final TextEditingController _apiKey = TextEditingController();
  final TextEditingController _apiKeyHeader = TextEditingController(
    text: 'X-Api-Key',
  );

  Map<String, dynamic>? _lastResponse;
  ComposeTemplateDTO? _lastSentTpl;
  bool _sending = false;
  bool _dirty = false;
  final TextEditingController _libSearchCtrl = TextEditingController();
  final Debouncer _autosave = Debouncer(const Duration(milliseconds: 900));
  int? _maxUploadMB;
  String? _currentTplId;
  String _currentTplName = 'Untitled';

  @override
  void dispose() {
    _urlCtrl.dispose();
    _rawCtrl.dispose();
    _jsonCtrl.dispose();
    _basicUser.dispose();
    _basicPass.dispose();
    _bearer.dispose();
    _apiKey.dispose();
    _apiKeyHeader.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    _loadComposeConfig();
    _loadDraftIfAny();
    // загрузим библиотеку один раз при входе, без FutureBuilder в дереве
    // ignore: discarded_futures
    WidgetsBinding.instance.addPostFrameCallback((_) {
      // ignore: discarded_futures
      sl<ComposeStore>().loadLibrary();
    });
  }

  void _markDirty() {
    if (!_dirty) {
      setState(() {
        _dirty = true;
      });
    }
  }

  void _scheduleAutosave() {
    if (!_dirty) return;
    _autosave.run(() {
      // локальный черновик
      // ignore: discarded_futures
      _saveDraft();
    });
  }

  Future<void> _saveDraft() async {
    try {
      final draft = <String, dynamic>{
        'tplId': _currentTplId,
        'tplName': _currentTplName,
        'method': _method,
        'url': _urlCtrl.text,
        'bodyMode': _bodyMode,
        'raw': _rawCtrl.text,
        'json': _jsonCtrl.text,
        'headers':
            _headers.map((e) => {'key': e.key, 'value': e.value}).toList(),
        'query': _query.map((e) => {'key': e.key, 'value': e.value}).toList(),
        'form': _form.map((e) => e.toJson()).toList(),
        'multipart': _multipart.map((e) => e.toJson()).toList(),
        'auth': {
          'type': _authType,
          'username': _basicUser.text,
          'password': _basicPass.text,
          'bearerToken': _bearer.text,
          'apiKey': _apiKey.text,
          'apiKeyHeader': _apiKeyHeader.text,
        },
      };
      if (_currentTplId != null) {
        await PrefsService().saveComposeDraftFor(_currentTplId!, draft);
      } else {
        await PrefsService().saveComposeDraft(draft);
      }
    } catch (_) {}
  }

  Future<void> _loadDraftIfAny() async {
    try {
      final raw = await PrefsService().loadComposeDraft();
      if (raw == null || raw.isEmpty) return;
      setState(() {
        _currentTplId = (raw['tplId'] ?? _currentTplId)?.toString();
        _currentTplName =
            (raw['tplName'] ?? _currentTplName)?.toString() ?? 'Untitled';
        _method = (raw['method'] ?? _method).toString();
        _urlCtrl.text = (raw['url'] ?? '').toString();
        _bodyMode = (raw['bodyMode'] ?? _bodyMode).toString();
        _rawCtrl.text = (raw['raw'] ?? '').toString();
        _jsonCtrl.text = (raw['json'] ?? '{\n  \n}').toString();
        _headers
          ..clear()
          ..addAll(
            ((raw['headers'] as List?) ?? const []).map((e) {
              final m = (e as Map).cast<String, dynamic>();
              return KvPair(
                (m['key'] ?? '').toString(),
                (m['value'] ?? '').toString(),
              );
            }),
          );
        _query
          ..clear()
          ..addAll(
            ((raw['query'] as List?) ?? const []).map((e) {
              final m = (e as Map).cast<String, dynamic>();
              return KvPair(
                (m['key'] ?? '').toString(),
                (m['value'] ?? '').toString(),
              );
            }),
          );
        _form
          ..clear()
          ..addAll(
            ((raw['form'] as List?) ?? const []).map(
              (e) => ComposeFormFieldDTO.fromJson(
                (e as Map).cast<String, dynamic>(),
              ),
            ),
          );
        _multipart
          ..clear()
          ..addAll(
            ((raw['multipart'] as List?) ?? const []).map(
              (e) =>
                  ComposePartDTO.fromJson((e as Map).cast<String, dynamic>()),
            ),
          );
        final auth = (raw['auth'] as Map?)?.cast<String, dynamic>();
        if (auth != null) {
          _authType = (auth['type'] ?? 'none').toString();
          _basicUser.text = (auth['username'] ?? '').toString();
          _basicPass.text = (auth['password'] ?? '').toString();
          _bearer.text = (auth['bearerToken'] ?? '').toString();
          _apiKey.text = (auth['apiKey'] ?? '').toString();
          _apiKeyHeader.text = (auth['apiKeyHeader'] ?? 'X-Api-Key').toString();
        }
      });
    } catch (_) {}
  }

  Future<void> _loadDraftForTpl(String id) async {
    try {
      final raw = await PrefsService().loadComposeDraftFor(id);
      if (raw == null || raw.isEmpty) return;
      final restore = await showDialog<bool>(
        context: context,
        builder:
            (ctx) => AlertDialog(
              title: const Text('Восстановить черновик?'),
              content: const Text(
                'Для этого шаблона найден черновик. Восстановить изменения?',
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(false),
                  child: const Text('Ignore'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(ctx).pop(true),
                  child: const Text('Restore'),
                ),
              ],
            ),
      );
      if (restore != true) return;
      setState(() {
        _currentTplId = (raw['tplId'] ?? _currentTplId)?.toString();
        _currentTplName =
            (raw['tplName'] ?? _currentTplName)?.toString() ?? 'Untitled';
        _method = (raw['method'] ?? _method).toString();
        _urlCtrl.text = (raw['url'] ?? '').toString();
        _bodyMode = (raw['bodyMode'] ?? _bodyMode).toString();
        _rawCtrl.text = (raw['raw'] ?? '').toString();
        _jsonCtrl.text = (raw['json'] ?? '{\n  \n}').toString();
        _headers
          ..clear()
          ..addAll(
            ((raw['headers'] as List?) ?? const []).map((e) {
              final m = (e as Map).cast<String, dynamic>();
              return KvPair(
                (m['key'] ?? '').toString(),
                (m['value'] ?? '').toString(),
              );
            }),
          );
        _query
          ..clear()
          ..addAll(
            ((raw['query'] as List?) ?? const []).map((e) {
              final m = (e as Map).cast<String, dynamic>();
              return KvPair(
                (m['key'] ?? '').toString(),
                (m['value'] ?? '').toString(),
              );
            }),
          );
        _form
          ..clear()
          ..addAll(
            ((raw['form'] as List?) ?? const []).map(
              (e) => ComposeFormFieldDTO.fromJson(
                (e as Map).cast<String, dynamic>(),
              ),
            ),
          );
        _multipart
          ..clear()
          ..addAll(
            ((raw['multipart'] as List?) ?? const []).map(
              (e) =>
                  ComposePartDTO.fromJson((e as Map).cast<String, dynamic>()),
            ),
          );
        final auth = (raw['auth'] as Map?)?.cast<String, dynamic>();
        if (auth != null) {
          _authType = (auth['type'] ?? 'none').toString();
          _basicUser.text = (auth['username'] ?? '').toString();
          _basicPass.text = (auth['password'] ?? '').toString();
          _bearer.text = (auth['bearerToken'] ?? '').toString();
          _apiKey.text = (auth['apiKey'] ?? '').toString();
          _apiKeyHeader.text = (auth['apiKeyHeader'] ?? 'X-Api-Key').toString();
        }
        _dirty = true;
      });
    } catch (_) {}
  }

  Future<void> _loadComposeConfig() async {
    try {
      final cfg = await sl<ComposeRepository>().config();
      setState(() {
        _maxUploadMB = (cfg['maxUploadMB'] as num?)?.toInt() ?? 10;
      });
    } catch (_) {
      // оставим дефолт
      setState(() {
        _maxUploadMB = 10;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return CallbackShortcuts(
      bindings: <ShortcutActivator, VoidCallback>{
        const SingleActivator(LogicalKeyboardKey.enter, control: true): _onSend,
        const SingleActivator(LogicalKeyboardKey.enter, meta: true): _onSend,
        const SingleActivator(LogicalKeyboardKey.keyS, control: true): _onSave,
        const SingleActivator(LogicalKeyboardKey.keyS, meta: true): _onSave,
        const SingleActivator(LogicalKeyboardKey.escape): () {
          FocusManager.instance.primaryFocus?.unfocus();
        },
      },
      child: Focus(
        autofocus: true,
        child: Theme(
          data: Theme.of(context).copyWith(
            visualDensity: const VisualDensity(horizontal: -2, vertical: -2),
            inputDecorationTheme: InputDecorationTheme(
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 10,
                vertical: 8,
              ),
              filled: true,
              fillColor: Theme.of(context).colorScheme.surfaceVariant,
              labelStyle: const TextStyle(fontSize: 12),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(6),
                borderSide: BorderSide(
                  color: Theme.of(context).colorScheme.outlineVariant,
                  width: 1,
                ),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(6),
                borderSide: BorderSide(
                  color: Theme.of(context).colorScheme.outlineVariant,
                  width: 1,
                ),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(6),
                borderSide: BorderSide(
                  color: Theme.of(context).colorScheme.primary,
                  width: 1.2,
                ),
              ),
              errorBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(6),
                borderSide: BorderSide(
                  color: Theme.of(context).colorScheme.error,
                  width: 1,
                ),
              ),
            ),
          ),
          child: Scaffold(
            appBar: AppBar(title: Text('Compose Request${_dirty ? ' *' : ''}')),
            body: Row(
              children: [
                // Left: library placeholder (v1 minimal)
                Container(
                  width: 280,
                  decoration: BoxDecoration(
                    border: Border(right: BorderSide(color: cs.outlineVariant)),
                  ),
                  child: Column(
                    children: [
                      Padding(
                        padding: const EdgeInsets.all(8),
                        child: TextField(
                          controller: _libSearchCtrl,
                          decoration: const InputDecoration(
                            isDense: true,
                            labelText: 'Search',
                            prefixIcon: Icon(Icons.search, size: 16),
                          ),
                          onChanged: (_) => setState(() {}),
                        ),
                      ),
                      Expanded(
                        child: LibraryAnimatedTree(
                          store: sl<ComposeStore>(),
                          onSelect: (tpl) async {
                            await _maybeApplyTemplate(tpl);
                          },
                          searchQuery: _libSearchCtrl.text,
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: SingleChildScrollView(
                    child: Column(
                      children: [
                        Padding(
                          padding: const EdgeInsets.all(12),
                          child: Row(
                            children: [
                              // Method
                              SizedBox(
                                width: 110,
                                child: DropdownButtonFormField<String>(
                                  value: _method,
                                  onChanged: (v) {
                                    setState(() => _method = v ?? 'GET');
                                    _markDirty();
                                    _scheduleAutosave();
                                  },
                                  isDense: true,
                                  isExpanded: true,
                                  iconSize: 16,
                                  style: Theme.of(context).textTheme.bodyMedium
                                      ?.copyWith(fontSize: 12),
                                  items:
                                      const [
                                            'GET',
                                            'POST',
                                            'PUT',
                                            'DELETE',
                                            'PATCH',
                                            'HEAD',
                                            'OPTIONS',
                                          ]
                                          .map(
                                            (m) => DropdownMenuItem(
                                              value: m,
                                              child: Text(m),
                                            ),
                                          )
                                          .toList(),
                                ),
                              ),
                              const SizedBox(width: 8),
                              // URL
                              Expanded(
                                child: TextFormField(
                                  controller: _urlCtrl,
                                  decoration: const InputDecoration(
                                    labelText: 'URL',
                                    hintText: 'https://api.example.com/v1',
                                  ),
                                  onChanged: (_) {
                                    setState(() {
                                      _dirty = true;
                                    });
                                    _scheduleAutosave();
                                  },
                                ),
                              ),
                              const SizedBox(width: 8),
                              Row(
                                children: [
                                  IconButton(
                                    tooltip: 'Copy cURL',
                                    onPressed: _copyCurl,
                                    icon: const Icon(Icons.copy_all),
                                  ),
                                ],
                              ),
                              const SizedBox(width: 8),
                              OutlinedButton.icon(
                                onPressed: _onSaveToCollection,
                                icon: const Icon(Icons.folder_open),
                                label: const Text('Save to...'),
                              ),
                              const SizedBox(width: 8),
                              FilledButton.icon(
                                onPressed: _sending ? null : _onSend,
                                icon: const Icon(Icons.send),
                                label:
                                    _sending
                                        ? Row(
                                          children: [
                                            const SizedBox(
                                              height: 16,
                                              width: 16,
                                              child: CircularProgressIndicator(
                                                strokeWidth: 2,
                                              ),
                                            ),
                                            const SizedBox(width: 8),
                                            const Text('Sending...'),
                                          ],
                                        )
                                        : const Text('Send'),
                              ),
                            ],
                          ),
                        ),
                        DefaultTabController(
                          length: 4,
                          child: Column(
                            children: [
                              const TabBar(
                                tabs: [
                                  Tab(text: 'Headers'),
                                  Tab(text: 'Query'),
                                  Tab(text: 'Body'),
                                  Tab(text: 'Auth'),
                                ],
                              ),
                              Builder(
                                builder: (ctx) {
                                  final ctrl = DefaultTabController.of(ctx);
                                  return AnimatedBuilder(
                                    animation: ctrl!,
                                    builder: (ctx, _) {
                                      final idx = ctrl.index;
                                      return IndexedStack(
                                        index: idx,
                                        children: [
                                          // Headers — делаем контент прокручиваемым, а
                                          // редактор ключ-значение растягиваем по контенту
                                          SingleChildScrollView(
                                            padding: EdgeInsets.zero,
                                            child: Column(
                                              children: [
                                                Padding(
                                                  padding:
                                                      const EdgeInsets.symmetric(
                                                        horizontal: 12,
                                                        vertical: 8,
                                                      ),
                                                  child: Row(
                                                    crossAxisAlignment:
                                                        CrossAxisAlignment
                                                            .start,
                                                    children: [
                                                      Padding(
                                                        padding:
                                                            const EdgeInsets.only(
                                                              top: 8,
                                                              right: 8,
                                                            ),
                                                        child: Text(
                                                          'Add:',
                                                          style:
                                                              Theme.of(context)
                                                                  .textTheme
                                                                  .bodySmall,
                                                        ),
                                                      ),
                                                      Expanded(
                                                        child: Wrap(
                                                          spacing: 8,
                                                          runSpacing: 8,
                                                          children: [
                                                            _quickHeaderButton(
                                                              'Content-Type',
                                                              'application/json',
                                                            ),
                                                            _quickHeaderButton(
                                                              'Authorization',
                                                              'Bearer ',
                                                            ),
                                                            _quickHeaderButton(
                                                              'Accept',
                                                              'application/json',
                                                            ),
                                                          ],
                                                        ),
                                                      ),
                                                    ],
                                                  ),
                                                ),
                                                KeyValueEditor(
                                                  key: const ValueKey(
                                                    'headers',
                                                  ),
                                                  items: _headers,
                                                  onChanged: (list) {
                                                    setState(() {
                                                      _dirty = true;
                                                    });
                                                    _scheduleAutosave();
                                                  },
                                                  labelKey: 'Header',
                                                ),
                                              ],
                                            ),
                                          ),
                                          SingleChildScrollView(
                                            padding: EdgeInsets.zero,
                                            child: KeyValueEditor(
                                              key: const ValueKey('query'),
                                              items: _query,
                                              onChanged: (list) {
                                                setState(() {
                                                  _dirty = true;
                                                });
                                                _scheduleAutosave();
                                              },
                                              labelKey: 'Param',
                                            ),
                                          ),
                                          BodyEditor(
                                            mode: _bodyMode,
                                            onModeChanged: (v) {
                                              _markDirty();
                                              setState(() {
                                                _bodyMode = v;
                                              });
                                              _ensureContentType();
                                              _scheduleAutosave();
                                            },
                                            rawCtrl: _rawCtrl,
                                            jsonCtrl: _jsonCtrl,
                                            form: _form,
                                            multipart: _multipart,
                                            maxUploadMB: _maxUploadMB,
                                            onFormChanged: (_) {
                                              _markDirty();
                                              _scheduleAutosave();
                                            },
                                            onMultipartChanged: (_) {
                                              _markDirty();
                                              _scheduleAutosave();
                                            },
                                          ),
                                          AuthEditor(
                                            authType: _authType,
                                            onAuthTypeChanged: (v) {
                                              setState(() => _authType = v);
                                              _markDirty();
                                              _scheduleAutosave();
                                            },
                                            basicUser: _basicUser,
                                            basicPass: _basicPass,
                                            bearer: _bearer,
                                            apiKeyHeader: _apiKeyHeader,
                                            apiKey: _apiKey,
                                          ),
                                        ],
                                      );
                                    },
                                  );
                                },
                              ),
                              const Divider(height: 1),
                              Builder(
                                builder: (_) {
                                  final hasData = _lastResponse != null;
                                  final details = ComposeHttpDetails(
                                    data: _lastResponse,
                                    template: _lastSentTpl,
                                    onOpenSession: (id) {
                                      try {
                                        // ignore: discarded_futures
                                        sl<HomeUiStore>().setSelectedSessionId(
                                          id,
                                        );
                                      } catch (_) {}
                                      Navigator.of(context).pop();
                                    },
                                  );
                                  // Если данных ещё нет — не фиксируем высоту (даём
                                  // занять столько, сколько нужно), но ограничим максимумом.
                                  // Если данные есть, панель деталей требует ограниченной
                                  // высоты, поэтому зададим 500.
                                  if (hasData) {
                                    return SizedBox(
                                      height: 500,
                                      child: details,
                                    );
                                  }
                                  return ConstrainedBox(
                                    constraints: const BoxConstraints(
                                      maxHeight: 500,
                                    ),
                                    child: details,
                                  );
                                },
                              ),
                              SizedBox(
                                height: 160,
                                child: FutureBuilder<Map<String, dynamic>>(
                                  future: sl<ComposeRepository>().history(
                                    limit: 20,
                                  ),
                                  builder: (ctx, snap) {
                                    final items =
                                        (snap.data?['items'] as List?)
                                            ?.cast<Map>()
                                            .map(
                                              (e) => e.cast<String, dynamic>(),
                                            )
                                            .toList() ??
                                        const <Map<String, dynamic>>[];
                                    return ComposeHistoryList(
                                      items: items,
                                      currentTemplateId: _currentTplId,
                                      onTapSessionId: (sid) {
                                        if (sid != null && sid.isNotEmpty) {
                                          setState(() {
                                            _lastResponse = {'sessionId': sid};
                                            _lastSentTpl = null;
                                          });
                                        }
                                      },
                                      onTapTemplate: (tid) {
                                        final t =
                                            sl<ComposeStore>()
                                                .requestsById[tid];
                                        if (t != null) {
                                          setState(() {
                                            _applyTemplate(t);
                                          });
                                        }
                                      },
                                    );
                                  },
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _onSend() async {
    setState(() => _sending = true);
    try {
      if (_bodyMode == 'json') {
        try {
          jsonDecode(_jsonCtrl.text);
        } catch (e) {
          setState(() => _sending = false);
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Невалидный JSON: ${e.toString()}')),
          );
          return;
        }
      }
      _ensureContentType();
      final repo = sl<ComposeRepository>();
      final tpl = ComposeTemplateDTO(
        id: 'tpl-${DateTime.now().microsecondsSinceEpoch}',
        name: 'Untitled',
        method: _method,
        url: _urlCtrl.text.trim(),
        headers:
            _headers
                .map((e) => ComposeHeaderDTO(key: e.key, value: e.value))
                .toList(),
        query:
            _query
                .map((e) => ComposeQueryDTO(key: e.key, value: e.value))
                .toList(),
        body: ComposeBodyDTO(
          mode: _bodyMode,
          raw: _bodyMode == 'raw' ? _rawCtrl.text : null,
          json: _bodyMode == 'json' ? _jsonCtrl.text : null,
          form: _bodyMode == 'form' ? _form : const [],
          multipart: _bodyMode == 'multipart' ? _multipart : const [],
        ),
        auth:
            _authType == 'none'
                ? null
                : ComposeAuthDTO(
                  type: _authType,
                  username: _basicUser.text,
                  password: _basicPass.text,
                  bearerToken: _bearer.text,
                  apiKey: _apiKey.text,
                  apiKeyHeader: _apiKeyHeader.text,
                ),
      );
      final res = await repo.send(tpl);
      setState(() {
        _lastSentTpl = tpl;
        _lastResponse = res;
      });
    } catch (e) {
      setState(() => _lastResponse = _friendlyError(e));
    } finally {
      setState(() => _sending = false);
    }
  }

  void _ensureContentType() {
    final hasCT = _headers.any((h) => h.key.toLowerCase() == 'content-type');
    if (hasCT) return;
    String? v;
    if (_bodyMode == 'json') v = 'application/json; charset=utf-8';
    if (_bodyMode == 'form') v = 'application/x-www-form-urlencoded';
    if (v != null) {
      _headers.add(KvPair('Content-Type', v));
    }
  }

  Future<void> _onSave() async {
    await _saveTemplate(showToast: true);
  }

  Future<void> _saveTemplate({required bool showToast}) async {
    try {
      final repo = sl<ComposeRepository>();
      _currentTplId ??= 'tpl-${DateTime.now().microsecondsSinceEpoch}';
      final tpl = ComposeTemplateDTO(
        id: _currentTplId!,
        name: _currentTplName,
        method: _method,
        url: _urlCtrl.text.trim(),
        headers:
            _headers
                .map((e) => ComposeHeaderDTO(key: e.key, value: e.value))
                .toList(),
        query:
            _query
                .map((e) => ComposeQueryDTO(key: e.key, value: e.value))
                .toList(),
        body: ComposeBodyDTO(
          mode: _bodyMode,
          raw: _bodyMode == 'raw' ? _rawCtrl.text : null,
          json: _bodyMode == 'json' ? _jsonCtrl.text : null,
          form: _bodyMode == 'form' ? _form : const [],
          multipart: _bodyMode == 'multipart' ? _multipart : const [],
        ),
        auth:
            _authType == 'none'
                ? null
                : ComposeAuthDTO(
                  type: _authType,
                  username: _basicUser.text,
                  password: _basicPass.text,
                  bearerToken: _bearer.text,
                  apiKey: _apiKey.text,
                  apiKeyHeader: _apiKeyHeader.text,
                ),
      );
      await repo.upsertRequest(tpl);
      _dirty = false;
      if (mounted && showToast) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('Сохранено в библиотеку')));
      }
      try {
        if (_currentTplId != null) {
          await PrefsService().clearComposeDraftFor(_currentTplId!);
        }
      } catch (_) {}
      await sl<ComposeStore>().loadLibrary();
      setState(() {});
    } catch (_) {}
  }

  Future<void> _onSaveToCollection() async {
    try {
      // 1) Сохраняем запрос без всплывашки
      await _saveTemplate(showToast: false);
      final store = sl<ComposeStore>();
      if (store.collections.isEmpty || _currentTplId == null) return;

      // 2) Выбор папки в древовидном виде
      final pick = await showDialog<FolderSelection>(
        context: context,
        builder: (_) => const FolderPickerDialog(),
      );
      if (pick == null) return;

      // 3) Перемещаем в выбранную папку
      await sl<ComposeRepository>().moveRequest(
        id: _currentTplId!,
        collectionId: pick.collectionId,
        folderId: pick.folderId,
        index: -1,
      );
      await store.loadLibrary();
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(const SnackBar(content: Text('Сохранено в коллекцию')));
      }
    } catch (_) {}
  }

  Widget _quickHeaderButton(String key, String valueTemplate) {
    // Hide button if header already exists
    final exists = _headers.any(
      (e) => e.key.toLowerCase() == key.toLowerCase(),
    );
    if (exists) {
      return const SizedBox.shrink();
    }

    return OutlinedButton(
      style: OutlinedButton.styleFrom(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        minimumSize: const Size(0, 32),
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      ),
      onPressed: () {
        setState(() {
          _headers.add(KvPair(key, valueTemplate));
          _dirty = true;
        });
      },
      child: Text(key, style: const TextStyle(fontSize: 12)),
    );
  }

  Future<void> _copyCurl() async {
    final buf = StringBuffer();
    buf.write('curl');
    buf.write(' -X ' + _method);
    for (final h in _headers) {
      if (h.key.trim().isEmpty) continue;
      buf.write(' -H ' + _shellEscape('${h.key}: ${h.value}'));
    }
    if (_bodyMode == 'raw' && _rawCtrl.text.isNotEmpty) {
      buf.write(' --data ' + _shellEscape(_rawCtrl.text));
    }
    if (_bodyMode == 'json' && _jsonCtrl.text.isNotEmpty) {
      buf.write(' --data ' + _shellEscape(_jsonCtrl.text));
    }
    buf.write(' ' + _shellEscape(_urlCtrl.text.trim()));
    await Clipboard.setData(ClipboardData(text: buf.toString()));
    if (mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('cURL скопирован')));
    }
  }

  String _shellEscape(String s) {
    final escaped = s.replaceAll("'", "'\\''");
    return "'" + escaped + "'";
  }

  Future<void> _maybeApplyTemplate(ComposeTemplateDTO tpl) async {
    if (_dirty) {
      final act = await showDialog<String>(
        context: context,
        builder:
            (ctx) => AlertDialog(
              title: const Text('Сохранить изменения?'),
              content: const Text('Есть несохранённые изменения. Что сделать?'),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop('cancel'),
                  child: const Text('Cancel'),
                ),
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop('discard'),
                  child: const Text('Discard'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(ctx).pop('save'),
                  child: const Text('Save'),
                ),
              ],
            ),
      );
      if (act == 'save') {
        await _onSave();
      } else if (act == 'cancel') {
        return;
      }
    }
    setState(() {
      _applyTemplate(tpl);
    });
    // после применения — предложить восстановить черновик для этого шаблона
    await _loadDraftForTpl(tpl.id);
  }

  void _applyTemplate(ComposeTemplateDTO tpl) {
    _method = tpl.method;
    _urlCtrl.text = tpl.url;
    _headers
      ..clear()
      ..addAll(tpl.headers.map((e) => KvPair(e.key, e.value)));
    _query
      ..clear()
      ..addAll(tpl.query.map((e) => KvPair(e.key, e.value)));
    _bodyMode = tpl.body.mode;
    _rawCtrl.text = tpl.body.raw ?? '';
    _jsonCtrl.text = tpl.body.json ?? '{\n  \n}';
    _form
      ..clear()
      ..addAll(tpl.body.form);
    _multipart
      ..clear()
      ..addAll(tpl.body.multipart);
    if (tpl.auth == null) {
      _authType = 'none';
      _basicUser.clear();
      _basicPass.clear();
      _bearer.clear();
      _apiKey.clear();
    } else {
      _authType = tpl.auth!.type;
      _basicUser.text = tpl.auth!.username ?? '';
      _basicPass.text = tpl.auth!.password ?? '';
      _bearer.text = tpl.auth!.bearerToken ?? '';
      _apiKey.text = tpl.auth!.apiKey ?? '';
      _apiKeyHeader.text = tpl.auth!.apiKeyHeader ?? 'X-Api-Key';
    }
    _dirty = false;
    _currentTplId = tpl.id;
    _currentTplName = tpl.name;
  }

  Map<String, dynamic> _friendlyError(Object e) {
    String msg = 'Не удалось выполнить запрос';
    try {
      if (e is AppHttpException) {
        switch (e.type) {
          case AppHttpErrorType.connectionTimeout:
          case AppHttpErrorType.sendTimeout:
          case AppHttpErrorType.receiveTimeout:
            msg = 'Таймаут запроса';
            break;
          case AppHttpErrorType.badCertificate:
            msg = 'Ошибка сертификата TLS';
            break;
          case AppHttpErrorType.connectionError:
            msg = 'Ошибка сети/соединения (DNS/сокет)';
            break;
          default:
            msg = e.message.isNotEmpty ? e.message : msg;
        }
        if (e is AppHttpServerException) {
          msg = e.messageFromServer.isNotEmpty ? e.messageFromServer : msg;
        }
      } else {
        msg = e.toString();
      }
    } catch (_) {}
    return {'error': true, 'message': msg};
  }

  List<ComposeFolderModel> _flattenFolders(ComposeFolderModel root) {
    final out = <ComposeFolderModel>[];
    void walk(ComposeFolderModel f) {
      out.add(f);
      for (final c in f.folders) {
        walk(c);
      }
    }

    walk(root);
    return out;
  }
}

// obsolete local viewer removed (moved to widgets/response_viewer.dart)
