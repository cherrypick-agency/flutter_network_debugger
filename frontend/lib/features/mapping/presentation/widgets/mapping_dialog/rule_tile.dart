import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../application/stores/mapping_store.dart';
import '../../../domain/mapping_rule.dart';
import '../mapping_rule_editor.dart';

class MappingRuleTile extends StatelessWidget {
  const MappingRuleTile({required this.rule, super.key});

  final MappingRule rule;

  @override
  Widget build(BuildContext context) {
    final r = rule;
    final title =
        '${r.kind.toUpperCase()}  ${r.patternType}: '
        '${r.hostPattern}${r.pathPattern}';

    return ListTile(
      dense: true,
      title: Text(title),
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
                await context.read<MappingStore>().upsert(
                  _copyWithEnabled(r, v),
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

  MappingRule _copyWithEnabled(MappingRule r, bool enabled) {
    return MappingRule(
      id: r.id,
      enabled: enabled,
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
    );
  }
}
