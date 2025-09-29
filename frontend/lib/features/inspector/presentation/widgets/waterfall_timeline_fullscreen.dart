import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../../../theme/context_ext.dart';
import '../../application/stores/sessions_store.dart';
import '../../domain/entities/session.dart';
import 'waterfall_timeline.dart';

class WaterfallTimelineFullscreenPage extends StatelessWidget {
  const WaterfallTimelineFullscreenPage({super.key, this.initialRange});
  final DateTimeRange? initialRange;
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
              child: Consumer<SessionsStore>(builder: (context, store, _) {
                final sessions = store.items.toList();
                return WaterfallTimeline(
                  sessions: sessions,
                  initialRange: initialRange,
                  onIntervalSelected: (range) {
                    // optionally communicate back via Navigator.pop or state mgmt
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Selected: ${range.start} — ${range.end}')),
                    );
                  },
                  onSessionSelected: (s) {
                    Navigator.of(context).maybePop(s);
                  },
                );
              }),
            ),
          ],
        ),
      ),
    );
  }
}
