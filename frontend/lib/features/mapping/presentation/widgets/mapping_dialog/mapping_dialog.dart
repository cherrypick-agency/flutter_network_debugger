import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../../../theme/context_ext.dart';
import '../../../application/stores/mapping_config_store.dart';
import '../../../application/stores/mapping_store.dart';
import 'config_panel.dart';
import 'rules_panel.dart';

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
    _tabs = TabController(length: 2, vsync: this);

    Future.microtask(() async {
      try {
        await sl<MappingStore>().load();
        await sl<MappingConfigStore>().load();
      } catch (e) {
        sl<NotificationsService>().error('Mapping', e.toString());
      }
    });
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 900, maxHeight: 600),
        child: Material(
          color: cs.surface,
          elevation: 8,
          borderRadius: BorderRadius.circular(12),
          child: MultiProvider(
            providers: [
              ChangeNotifierProvider.value(value: sl<MappingStore>()),
              ChangeNotifierProvider.value(value: sl<MappingConfigStore>()),
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
                        icon: Icon(Icons.close, color: cs.onSurfaceVariant),
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
                  ],
                ),
                const Divider(height: 1),
                Expanded(
                  child: TabBarView(
                    controller: _tabs,
                    children: const [
                      MappingRulesPanel(),
                      MappingConfigPanel(),
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
