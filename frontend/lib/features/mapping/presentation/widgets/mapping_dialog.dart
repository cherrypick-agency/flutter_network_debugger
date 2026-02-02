import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../../domain/mapping_rule.dart';
import '../../data/mapping_api.dart';
import '../../data/mapping_repository_impl.dart';
import '../../application/stores/mapping_store.dart';
import '../../application/stores/mapping_config_store.dart';
import '../../domain/mapping_config.dart';
import '../../../../core/di/di.dart';
import 'package:app_http_client/application/app_http_client.dart' as app_http;
import '../../../../core/notifications/notifications_service.dart';
import 'mapping_rule_editor.dart';
import '../../../../theme/context_ext.dart';

class MappingDialog extends StatefulWidget {
  const MappingDialog({super.key});

  @override
  State<MappingDialog> createState() => _MappingDialogState();
}

class _MappingDialogState extends State<MappingDialog>
    with SingleTickerProviderStateMixin {
  late final TabController _tabs;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final http = sl<app_http.AppHttpClient>();
    final repo = MappingRepositoryImpl(MappingApi(http));
    final mq = MediaQuery.sizeOf(context);
    final w = (mq.width - 48).clamp(320.0, 900.0).toDouble();
    final h = (mq.height - 48).clamp(320.0, 600.0).toDouble();
    return Center(
      child: Material(
        color: Theme.of(context).colorScheme.surface,
        elevation: 8,
        borderRadius: BorderRadius.circular(12),
        child: SizedBox(
          width: w,
          height: h,
          child: MultiProvider(
            providers: [
              ChangeNotifierProvider(create: (_) => MappingStore(repo)..load()),
              ChangeNotifierProvider(
                create: (_) => MappingConfigStore(repo)..load(),
              ),
            ],
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text('Mapping', style: context.appText.title),
                      ),
                      IconButton(
                        onPressed: () => Navigator.of(context).pop(),
                        icon: const Icon(Icons.close),
                      ),
                    ],
                  ),
                ),
                const Divider(height: 1),
                TabBar(
                  controller: _tabs,
                  tabs: const [
                    Tab(text: 'Rules'),
                    Tab(text: 'Config'),
                    Tab(text: 'Activity'),
                  ],
                ),
                const Divider(height: 1),
                Expanded(
                  child: TabBarView(
                    controller: _tabs,
                    children: const [
                      _RulesPanel(),
                      _ConfigPanel(),
                      _ActivityPanel(),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _RulesPanel extends StatelessWidget {
  const _RulesPanel();
  @override
  Widget build(BuildContext context) {
    final store = context.watch<MappingStore>();
    final rows = store.rules;
    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              ElevatedButton.icon(
                onPressed: () async {
                  try {
                    await store.load();
                  } catch (e) {
                    sl<NotificationsService>().error('Failed', e.toString());
                  }
                },
                icon: const Icon(Icons.refresh),
                label: const Text('Refresh'),
              ),
              const SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: () async {
                  await showDialog<bool>(
                    context: context,
                    useRootNavigator: false,
                    builder: (_) => const Dialog(child: MappingRuleEditor()),
                  );
                },
                icon: const Icon(Icons.add),
                label: const Text('Add Local/Remote'),
              ),
              const Spacer(),
            ],
          ),
          const SizedBox(height: 8),
          Expanded(
            child: rows.isEmpty
                ? Align(
                    alignment: Alignment.bottomCenter,
                    child: Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.rule,
                            size: 36,
                            color: context.appColors.textSecondary,
                          ),
                          const SizedBox(height: 8),
                          Text('No rules yet', style: context.appText.subtitle),
                          const SizedBox(height: 4),
                          Text(
                            'Mapping rules let you rewrite or stub HTTP requests.\n'
                            'Add a Local rule to serve a file/blob (optionally set status/content-type),\n'
                            'or a Remote rule to proxy to a URL template. Patterns support glob/regex.',
                            style: context.appText.body,
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    ),
                  )
                : ReorderableListView.builder(
                    itemCount: rows.length,
                    onReorder: (oldIndex, newIndex) async {
                      try {
                        var ni = newIndex;
                        if (ni > oldIndex) ni -= 1;
                        final list = rows.toList();
                        final item = list.removeAt(oldIndex);
                        list.insert(ni, item);
                        final ids = list.map((e) => e.id).toList();
                        await context.read<MappingStore>().reorder(ids);
                      } catch (e) {
                        sl<NotificationsService>().error(
                          'Reorder failed',
                          e.toString(),
                        );
                      }
                    },
                    itemBuilder: (_, i) {
                      final r = rows[i];
                      return Column(
                        key: ValueKey(r.id),
                        children: [_RuleTile(r), const Divider(height: 1)],
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class _RuleTile extends StatelessWidget {
  const _RuleTile(this.r);
  final MappingRule r;
  @override
  Widget build(BuildContext context) {
    final text =
        '${r.kind.toUpperCase()}  ${r.patternType}:${r.hostPattern}${r.pathPattern}';
    return ListTile(
      dense: true,
      title: Text(text),
      subtitle: Text(
        r.kind == 'local'
            ? (r.filePath ?? r.blobPath ?? '')
            : (r.targetURLTemplate ?? ''),
      ),
      leading: const Icon(Icons.drag_indicator),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Switch(
            value: r.enabled,
            onChanged: (v) async {
              try {
                final store = context.read<MappingStore>();
                await store.upsert(
                  MappingRule(
                    id: r.id,
                    enabled: v,
                    priority: r.priority,
                    kind: r.kind,
                    stopProcessing: r.stopProcessing,
                    methods: r.methods,
                    hostPattern: r.hostPattern,
                    pathPattern: r.pathPattern,
                    patternType: r.patternType,
                    filePath: r.filePath,
                    blobPath: r.blobPath,
                    statusOverride: r.statusOverride,
                    contentTypeOverride: r.contentTypeOverride,
                    targetURLTemplate: r.targetURLTemplate,
                    preserveHost: r.preserveHost,
                  ),
                );
              } catch (e) {
                sl<NotificationsService>().error('Update failed', e.toString());
              }
            },
          ),
          IconButton(
            tooltip: 'Edit',
            onPressed: () async {
              await showDialog<bool>(
                context: context,
                useRootNavigator: false,
                builder: (_) => Dialog(child: MappingRuleEditor(initial: r)),
              );
            },
            icon: const Icon(Icons.edit),
          ),
          IconButton(
            tooltip: 'Delete',
            onPressed: () async {
              final ok = await showDialog<bool>(
                context: context,
                builder: (ctx) => AlertDialog(
                  title: const Text('Delete rule?'),
                  content: Text('This will remove "${r.id}"'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.of(ctx).pop(false),
                      child: const Text('Cancel'),
                    ),
                    ElevatedButton(
                      onPressed: () => Navigator.of(ctx).pop(true),
                      child: const Text('Delete'),
                    ),
                  ],
                ),
              );
              if (ok != true) return;
              try {
                await context.read<MappingStore>().delete(r.id);
              } catch (e) {
                sl<NotificationsService>().error('Delete failed', e.toString());
              }
            },
            icon: const Icon(Icons.delete_outline),
          ),
        ],
      ),
    );
  }
}

class _ConfigPanel extends StatefulWidget {
  const _ConfigPanel();
  @override
  State<_ConfigPanel> createState() => _ConfigPanelState();
}

class _ConfigPanelState extends State<_ConfigPanel> {
  late final TextEditingController _uploadCtrl;
  late final FocusNode _uploadFocus;
  late final MappingConfigStore _store;

  @override
  void initState() {
    super.initState();
    _uploadFocus = FocusNode();
    _uploadCtrl = TextEditingController(text: '20');
    _store = context.read<MappingConfigStore>();
    _store.addListener(_syncFromStore);
    _syncFromStore();
  }

  void _syncFromStore() {
    final cfg = _store.config;
    if (cfg == null) return;
    if (_uploadFocus.hasFocus) return;
    final desired = cfg.uploadMaxMB.toString();
    if (_uploadCtrl.text != desired) {
      _uploadCtrl.text = desired;
    }
  }

  @override
  void dispose() {
    _store.removeListener(_syncFromStore);
    _uploadFocus.dispose();
    _uploadCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final store = context.watch<MappingConfigStore>();
    final cfg = store.config;
    final canEdit = !store.loading && cfg != null;
    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              SizedBox(
                width: 220,
                child: SwitchListTile(
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Enabled'),
                  value: cfg?.enabled ?? true,
                  onChanged: !canEdit
                      ? null
                      : (v) async {
                          try {
                            final c = MappingConfig(
                              enabled: v,
                              uploadMaxMB: int.tryParse(_uploadCtrl.text) ?? 20,
                            );
                            await store.save(c);
                          } catch (e) {
                            sl<NotificationsService>().error(
                              'Save failed',
                              e.toString(),
                            );
                          }
                        },
                ),
              ),
              const Spacer(),
              IconButton(
                tooltip: 'Refresh',
                onPressed: () async {
                  try {
                    await store.load();
                    final newCfg = store.config;
                    _uploadCtrl.text = (newCfg?.uploadMaxMB ?? 20).toString();
                  } catch (e) {
                    sl<NotificationsService>().error('Failed', e.toString());
                  }
                },
                icon: const Icon(Icons.refresh),
              ),
            ],
          ),
          const SizedBox(height: 8),
          SizedBox(
            width: 220,
            child: TextField(
              controller: _uploadCtrl,
              focusNode: _uploadFocus,
              keyboardType: TextInputType.number,
              textInputAction: TextInputAction.done,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              onSubmitted: !canEdit
                  ? null
                  : (_) async {
                      try {
                        final val = int.tryParse(_uploadCtrl.text.trim()) ?? 20;
                        final c = MappingConfig(
                          enabled: cfg.enabled,
                          uploadMaxMB: val,
                        );
                        await store.save(c);
                      } catch (e) {
                        sl<NotificationsService>().error(
                          'Save failed',
                          e.toString(),
                        );
                      }
                    },
              decoration: const InputDecoration(labelText: 'Upload max (MB)'),
            ),
          ),
          const SizedBox(height: 12),
          ElevatedButton.icon(
            onPressed: !canEdit
                ? null
                : () async {
                    try {
                      final val = int.tryParse(_uploadCtrl.text.trim()) ?? 20;
                      final c = MappingConfig(
                        enabled: cfg.enabled,
                        uploadMaxMB: val,
                      );
                      await store.save(c);
                    } catch (e) {
                      sl<NotificationsService>().error(
                        'Save failed',
                        e.toString(),
                      );
                    }
                  },
            icon: const Icon(Icons.save),
            label: const Text('Save'),
          ),
        ],
      ),
    );
  }
}

class _ActivityPanel extends StatelessWidget {
  const _ActivityPanel();
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.history, size: 40, color: context.appColors.textSecondary),
          const SizedBox(height: 8),
          Text('No activity yet', style: context.appText.subtitle),
          const SizedBox(height: 4),
          Text(
            'Recent mapping hits and changes will appear here.',
            style: context.appText.body,
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
