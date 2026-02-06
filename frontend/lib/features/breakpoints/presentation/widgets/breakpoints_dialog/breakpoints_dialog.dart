import 'package:flutter/material.dart';
import 'package:mobx/mobx.dart';

import '../../../../../core/di/di.dart';
import '../../../../../core/notifications/notifications_service.dart';
import '../../../../../theme/context_ext.dart';
import '../../../application/stores/breakpoints_store.dart';
import '../../../application/stores/intercept_editor_store.dart';
import '../../../application/stores/intercept_queue_store.dart';
import '../../../domain/entities/intercept_item.dart';
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
  ReactionDisposer? _selectedBinder;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);

    _selectedBinder = reaction<InterceptItem?>(
      (_) => sl<InterceptQueueStore>().selected,
      (item) => sl<InterceptEditorStore>().setItem(item),
    );

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
    _selectedBinder?.call();
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
        child: Material(
          elevation: 12,
          color: Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(12),
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
    );
  }
}
