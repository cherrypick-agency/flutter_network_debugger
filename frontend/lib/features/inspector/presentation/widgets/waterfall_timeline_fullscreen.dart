import 'package:flutter/material.dart';
import 'dart:async';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../../../theme/context_ext.dart';
import '../../application/stores/sessions_store.dart';
import 'waterfall_timeline.dart';
import '../../../../services/prefs.dart';

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
              child: FutureBuilder<DateTime?>(
                future: PrefsService().loadSince(),
                builder: (context, snap) {
                  final since = snap.data;
                  return Observer(
                    builder: (_) {
                      final store = context.read<SessionsStore>();
                      final raw = store.items.toList();
                      var list = raw;
                      if (since != null) {
                        list = list
                            .where((s) {
                              final st = s.startedAt;
                              return st == null || !st.isBefore(since);
                            })
                            .toList(growable: false);
                      }
                      if (widget.initialRange != null) {
                        list = list
                            .where((s) {
                              final st = s.startedAt;
                              return st == null ||
                                  !st.isBefore(widget.initialRange!.start);
                            })
                            .toList(growable: false);
                      }
                      return WaterfallTimeline(
                        sessions: list,
                        autoCompressLanes: true,
                        expandToParent: true,
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
