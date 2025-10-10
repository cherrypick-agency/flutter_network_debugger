import 'package:flutter/material.dart';
import 'dart:async';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../../../theme/context_ext.dart';
import '../../application/stores/sessions_store.dart';
import 'waterfall_timeline.dart' show WaterfallTimeline;
import '../../application/stores/home_ui_store.dart';
import '../../../../core/di/di.dart';

class WaterfallTimelineFullscreenPage extends StatefulWidget {
  const WaterfallTimelineFullscreenPage({super.key, this.initialRange});
  final DateTimeRange? initialRange;
  @override
  State<WaterfallTimelineFullscreenPage> createState() =>
      _WaterfallTimelineFullscreenPageState();
}

class _WaterfallTimelineFullscreenPageState
    extends State<WaterfallTimelineFullscreenPage> {
  Timer? _poll;

  @override
  void initState() {
    super.initState();
    // Подтягиваем актуальные статусы завершения сессий, пока открыт фулскрин
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final store = context.read<SessionsStore>();
      store.load();
      _poll = Timer.periodic(const Duration(seconds: 2), (_) {
        if (!store.loading) {
          store.load();
        }
      });
    });
  }

  @override
  void dispose() {
    _poll?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Requests Timeline'),
        actions: [
          IconButton(
            tooltip: 'Close',
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).maybePop(),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('All sessions', style: context.appText.title),
            const SizedBox(height: 8),
            Expanded(
              child: Observer(
                builder: (_) {
                  final ui = sl<HomeUiStore>();
                  final since = ui.since.value;
                  final store = context.read<SessionsStore>();
                  final raw = store.items.toList();
                  var list = raw;
                  if (since != null) {
                    list = list
                        .where((s) {
                          final end = s.closedAt ?? DateTime.now();
                          return !end.isBefore(since);
                        })
                        .toList(growable: false);
                  }
                  if (widget.initialRange != null) {
                    list = list
                        .where((s) {
                          final st = s.startedAt;
                          if (st == null) return false;
                          final en = s.closedAt ?? DateTime.now();
                          final sel = widget.initialRange!;
                          return en.isAfter(sel.start) && st.isBefore(sel.end);
                        })
                        .toList(growable: false);
                  }
                  final fixedWindow =
                      ui.recentWindowEnabled.value
                          ? Duration(minutes: ui.recentWindowMinutes.value)
                          : null;
                  return WaterfallTimeline(
                    sessions: list,
                    autoCompressLanes: true,
                    expandToParent: true,
                    fixedWindow: fixedWindow,
                    initialRange: widget.initialRange,
                    onIntervalSelected: (range) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(
                            'Selected: ${range.start} — ${range.end}',
                          ),
                        ),
                      );
                    },
                    onSessionSelected: (s) {
                      Navigator.of(context).maybePop(s);
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}
