import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';

import '../../../application/stores/sessions_store.dart';
import '../waterfall_timeline.dart' show WaterfallTimeline;
import '../waterfall_timeline_fullscreen.dart';
import '../timeline_settings_button.dart';
import '../../../application/stores/home_ui_store.dart';
import '../../../../../../core/di/di.dart';
import '../../../application/services/recent_window_service.dart';

// Block with timeline and floating controls (fit/clear/settings/fullscreen)
class TimelineBlock extends StatelessWidget {
  const TimelineBlock({
    super.key,
    required this.since,
    required this.wfFitAll,
    required this.onFitAllChanged,
    required this.onSelectSession,
    required this.onClearAllSessions,
    required this.selectedRange,
    required this.onRangeChanged,
    required this.onRangeCleared,
    this.ignoredIds = const <String>{},
  });

  // If not null — hide sessions that started before this date
  final DateTime? since;
  final bool wfFitAll;
  final ValueChanged<bool> onFitAllChanged;
  final ValueChanged<String> onSelectSession;
  final Future<void> Function() onClearAllSessions;
  final DateTimeRange? selectedRange;
  final ValueChanged<DateTimeRange> onRangeChanged;
  final VoidCallback onRangeCleared;
  final Set<String> ignoredIds;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 8),
      child: Stack(
        children: [
          Observer(
            builder: (_) {
              final raw = context.watch<SessionsStore>().items.toList();
              final visibleRaw =
                  raw; // legacy ignore removed; captureId handles visibility
              final sessions =
                  since == null
                      ? visibleRaw
                      : visibleRaw
                          .where((s) {
                            final end = s.closedAt ?? DateTime.now();
                            // Show if end >= since
                            return !end.isBefore(since!);
                          })
                          .toList(growable: false);
              final ui = sl<HomeUiStore>();
              final fixedWindow =
                  ui.recentWindowEnabled.value
                      ? Duration(minutes: ui.recentWindowMinutes.value)
                      : null;
              return WaterfallTimeline(
                sessions: sessions,
                autoCompressLanes: true,
                fixedWindow: fixedWindow,
                fitAll: wfFitAll,
                onFitAllChanged: onFitAllChanged,
                onIntervalSelected: (range) => onRangeChanged(range),
                onIntervalCleared: onRangeCleared,
                onSessionSelected: (s) => onSelectSession(s.id),
                hoveredSessionIdExt: sl<HomeUiStore>().hoveredSessionId.value,
                selectedSessionIdExt: sl<HomeUiStore>().selectedSessionId.value,
                initialRange:
                    since == null
                        ? null
                        : DateTimeRange(start: since!, end: DateTime.now()),
              );
            },
          ),
          Positioned(
            right: 0,
            bottom: 0,
            child: Container(
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface.withOpacity(0.45),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: Theme.of(context).colorScheme.outlineVariant,
                  width: 1,
                ),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const SizedBox(width: 4),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 2),
                    child: InkWell(
                      borderRadius: BorderRadius.circular(6),
                      onTap: () => onFitAllChanged(!wfFitAll),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(6),
                          color:
                              wfFitAll
                                  ? Theme.of(
                                    context,
                                  ).colorScheme.primary.withOpacity(0.15)
                                  : Colors.transparent,
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              wfFitAll
                                  ? Icons.check_box
                                  : Icons.check_box_outline_blank,
                              size: 14,
                              color:
                                  wfFitAll
                                      ? Theme.of(context).colorScheme.primary
                                      : Theme.of(
                                        context,
                                      ).colorScheme.onSurfaceVariant,
                            ),
                            const SizedBox(width: 6),
                            const Text('fit', style: TextStyle(fontSize: 12)),
                          ],
                        ),
                      ),
                    ),
                  ),
                  // Mini-checkbox for window cropping (crop)
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 2),
                    child: InkWell(
                      borderRadius: BorderRadius.circular(6),
                      onTap: () {
                        final ui = sl<HomeUiStore>();
                        final enabled = ui.recentWindowEnabled.value;
                        sl<RecentWindowService>().apply(
                          enabled: !enabled,
                          minutes: ui.recentWindowMinutes.value,
                        );
                      },
                      child: Builder(
                        builder: (context) {
                          final crop =
                              sl<HomeUiStore>().recentWindowEnabled.value;
                          return Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 4,
                            ),
                            decoration: BoxDecoration(
                              borderRadius: BorderRadius.circular(6),
                              color:
                                  crop
                                      ? Theme.of(
                                        context,
                                      ).colorScheme.primary.withOpacity(0.15)
                                      : Colors.transparent,
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  crop
                                      ? Icons.check_box
                                      : Icons.check_box_outline_blank,
                                  size: 14,
                                  color:
                                      crop
                                          ? Theme.of(
                                            context,
                                          ).colorScheme.primary
                                          : Theme.of(
                                            context,
                                          ).colorScheme.onSurfaceVariant,
                                ),
                                const SizedBox(width: 6),
                                const Text(
                                  'crop',
                                  style: TextStyle(fontSize: 12),
                                ),
                              ],
                            ),
                          );
                        },
                      ),
                    ),
                  ),
                  IconButton(
                    tooltip: 'Clear all sessions',
                    iconSize: 16,
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    icon: Icon(
                      Icons.delete_forever,
                      size: 16,
                      color: Theme.of(context).colorScheme.error,
                    ),
                    onPressed: () async => onClearAllSessions(),
                  ),
                  SizedBox(
                    width: 28,
                    height: 28,
                    child: FittedBox(
                      child: TimelineSettingsButton(
                        getFit: () => wfFitAll,
                        setFit: (v) => onFitAllChanged(v),
                        getCrop:
                            () => sl<HomeUiStore>().recentWindowEnabled.value,
                        setCrop: (v) {
                          final ui = sl<HomeUiStore>();
                          ui.setRecentWindowEnabled(v);
                          // instantly recalculate since and update list
                          final minutes = ui.recentWindowMinutes.value;
                          sl<RecentWindowService>().apply(
                            enabled: v,
                            minutes: minutes,
                          );
                        },
                        getMinutes:
                            () => sl<HomeUiStore>().recentWindowMinutes.value,
                        setMinutes: (m) {
                          final ui = sl<HomeUiStore>();
                          ui.setRecentWindowMinutes(m);
                          sl<RecentWindowService>().apply(
                            enabled: ui.recentWindowEnabled.value,
                            minutes: m,
                          );
                        },
                      ),
                    ),
                  ),
                  IconButton(
                    tooltip: 'Open fullscreen',
                    iconSize: 16,
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(
                      minWidth: 28,
                      minHeight: 28,
                    ),
                    icon: const Icon(Icons.fullscreen, size: 16),
                    onPressed: () async {
                      final res = await Navigator.of(context).push<dynamic>(
                        MaterialPageRoute(
                          builder:
                              (_) => WaterfallTimelineFullscreenPage(
                                initialRange: selectedRange,
                              ),
                        ),
                      );
                      if (res is String) {
                        onSelectSession(res);
                      }
                    },
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
