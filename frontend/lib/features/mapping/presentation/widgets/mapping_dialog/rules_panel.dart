import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../../../core/widgets/inline_status_banner.dart';
import '../../../../../theme/context_ext.dart';
import '../../../application/stores/mapping_store.dart';
import '../mapping_rule_editor.dart';
import 'rule_tile.dart';

class MappingRulesPanel extends StatelessWidget {
  const MappingRulesPanel({super.key});

  @override
  Widget build(BuildContext context) {
    final store = context.watch<MappingStore>();
    final rows = store.rules;
    final loading = store.loading;
    final err = store.lastError;

    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InlineStatusBanner(
            loading: loading && rows.isNotEmpty,
            loadingText: 'Refreshing rules…',
            errorText: err,
            onRetry: () => store.load(),
          ),
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
            child: loading
                ? const Center(child: CircularProgressIndicator())
                : rows.isEmpty
                ? const _EmptyRulesState()
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
                        children: [
                          MappingRuleTile(rule: r),
                          const Divider(height: 1),
                        ],
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class _EmptyRulesState extends StatelessWidget {
  const _EmptyRulesState();

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: Alignment.bottomCenter,
      child: Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.rule, size: 36, color: context.appColors.textSecondary),
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
    );
  }
}
