import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../../../theme/context_ext.dart';
import '../../../application/stores/breakpoints_store.dart';
import '../../../application/stores/intercept_editor_store.dart';
import '../../../application/stores/intercept_queue_store.dart';
import 'editor_panel.dart';
import 'queue_panel.dart';
import 'rules_panel.dart';

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

    // Загружаем конфиг/правила и поднимаем подписку на события очереди.
    Future.microtask(() async {
      try {
        await sl<BreakpointsStore>().load();
        await sl<InterceptQueueStore>().init();

        final q = sl<InterceptQueueStore>();
        if (mounted && q.items.isEmpty) {
          _tabs.index = 2;
        }
      } catch (e) {
        sl<NotificationsService>().error('Breakpoints', e.toString());
      }
    });
  }

  @override
  void dispose() {
    // Очередь живёт как singleton в DI, поэтому отписываемся вручную при закрытии.
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
                              style: context.appText.title,
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
                        children: [
                          AnimatedBuilder(
                            animation: _tabs,
                            builder: (context, child) {
                              return TickerMode(
                                enabled: _tabs.index == 0,
                                child: child!,
                              );
                            },
                            child: const QueuePanel(),
                          ),
                          const EditorPanel(),
                          const RulesPanel(),
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
