import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

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

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 3, vsync: this);

    // Load config/rules and subscribe to queue events.
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
    // The queue lives as a singleton in DI, so we unsubscribe manually when closing.
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
          child: MultiProvider(
            providers: [
              ChangeNotifierProvider.value(value: sl<BreakpointsStore>()),
              ChangeNotifierProvider.value(value: sl<InterceptQueueStore>()),
              ChangeNotifierProvider.value(value: sl<InterceptEditorStore>()),
            ],
            child: _SelectedInterceptItemBinder(
              key: const ValueKey('selectedInterceptItemBinder'),
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
    );
  }
}

class _SelectedInterceptItemBinder extends StatefulWidget {
  const _SelectedInterceptItemBinder({required this.child, super.key});

  final Widget child;

  @override
  State<_SelectedInterceptItemBinder> createState() =>
      _SelectedInterceptItemBinderState();
}

class _SelectedInterceptItemBinderState
    extends State<_SelectedInterceptItemBinder> {
  bool _boundOnce = false;
  String? _lastId;
  InterceptItem? _lastRef;

  @override
  Widget build(BuildContext context) {
    return Selector<InterceptQueueStore, InterceptItem?>(
      selector: (_, s) => s.selected,
      builder: (context, selected, child) {
        final id = selected?.id;
        final changed =
            !_boundOnce || id != _lastId || !identical(selected, _lastRef);

        if (changed) {
          _boundOnce = true;
          _lastId = id;
          _lastRef = selected;
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            context.read<InterceptEditorStore>().setItem(selected);
          });
        }

        return child!;
      },
      child: widget.child,
    );
  }
}
