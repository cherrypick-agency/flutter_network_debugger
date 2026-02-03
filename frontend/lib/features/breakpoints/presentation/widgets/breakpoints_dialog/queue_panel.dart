import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:provider/provider.dart';

import '../../../../../core/widgets/inline_status_banner.dart';
import '../../../../../theme/context_ext.dart';
import '../../../application/stores/intercept_queue_store.dart';

class QueuePanel extends StatefulWidget {
  const QueuePanel({super.key});

  @override
  State<QueuePanel> createState() => _QueuePanelState();
}

class _QueuePanelState extends State<QueuePanel>
    with SingleTickerProviderStateMixin {
  late final Ticker _ticker;
  Duration _lastRedraw = Duration.zero;

  @override
  void initState() {
    super.initState();
    _ticker = createTicker((elapsed) {
      if (!mounted) return;
      if (elapsed - _lastRedraw >= const Duration(milliseconds: 250)) {
        _lastRedraw = elapsed;
        setState(() {});
      }
    })..start();
  }

  @override
  void dispose() {
    _ticker.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final store = context.watch<InterceptQueueStore>();
    final items = store.items;
    final selectedId = store.selected?.id;
    final loading = store.loading;
    final err = store.lastError;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 12, 12, 8),
          child: Text(
            'Pending intercepted requests and responses waiting for action.',
            style: context.appText.body.copyWith(
              color: context.appColors.textSecondary,
            ),
          ),
        ),
        InlineStatusBanner(
          loading: loading && items.isNotEmpty,
          loadingText: 'Refreshing queue…',
          errorText: err,
          onRetry: () => store.refresh(),
        ),
        Expanded(
          child: loading && items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : items.isEmpty
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

                    final titleLeft = it.direction == 'request'
                        ? (it.req?.method ?? '')
                        : (it.res?.status ?? '');

                    final subtitle = it.direction == 'request'
                        ? (it.req?.url ?? '')
                        : (it.res?.contentType ?? '');

                    return ListTile(
                      selected: selectedId == it.id,
                      title: Text('$titleLeft • ${it.direction}'),
                      subtitle: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(subtitle),
                          const SizedBox(height: 4),
                          LinearProgressIndicator(value: pct),
                        ],
                      ),
                      onTap: () {
                        store.select(it.id);
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
