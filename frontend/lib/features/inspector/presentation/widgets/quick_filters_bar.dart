import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';

import '../../../../theme/context_ext.dart';
import '../../../../core/di/di.dart';
import '../../application/stores/home_ui_store.dart';

// Compact quick filters panel below the timeline
class QuickFiltersBar extends StatelessWidget {
  const QuickFiltersBar({super.key});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final textStyle = Theme.of(context).textTheme.labelSmall;

    // Subtle background similar to timeline controls on the right
    final bg = cs.surface.withValues(alpha: 0.45);
    // Remove border — keep only soft background

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(10),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 6),
      child: Observer(
        builder: (_) {
          final ui = sl<HomeUiStore>();
          final allSelected =
              ui.quickTypes.isEmpty && ui.quickStatusGroups.isEmpty;

          Widget chip({
            required String label,
            required bool selected,
            required VoidCallback onTap,
          }) {
            return ChoiceChip(
              label: Text(label, style: textStyle),
              labelPadding: const EdgeInsets.symmetric(horizontal: 6),
              selected: selected,
              visualDensity: const VisualDensity(horizontal: -4, vertical: -4),
              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              selectedColor: cs.primary.withValues(alpha: 0.15),
              onSelected: (_) => onTap(),
            );
          }

          final typeMap = <String, String>{
            'http': 'HTTP',
            'https': 'HTTPS',
            'ws': 'WebSocket',
            'json': 'JSON',
            'form': 'Form',
            'xml': 'XML',
            'js': 'JS',
            'css': 'CSS',
            'graphql': 'GraphQL',
            'document': 'Document',
            'media': 'Media',
            'other': 'Other',
          };
          final statusList = <String>['1xx', '2xx', '3xx', '4xx', '5xx'];

          // Horizontal scroll, single row
          return SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                chip(
                  label: 'All',
                  selected: allSelected,
                  onTap: () async {
                    ui.clearQuickFilters();
                  },
                ),
                const SizedBox(width: 6),
                const _DividerDot(),
                const SizedBox(width: 6),
                // Types
                ...[
                  for (final e in typeMap.entries) ...[
                    chip(
                      label: e.value,
                      selected: ui.quickTypes.contains(e.key),
                      onTap: () async {
                        ui.toggleQuickType(e.key);
                      },
                    ),
                    const SizedBox(width: 6),
                  ],
                ],
                const _DividerDot(),
                const SizedBox(width: 6),
                // Statuses
                ...[
                  for (final g in statusList) ...[
                    chip(
                      label: g,
                      selected: ui.quickStatusGroups.contains(g),
                      onTap: () async {
                        ui.toggleQuickStatusGroup(g);
                      },
                    ),
                    const SizedBox(width: 6),
                  ],
                ],
              ],
            ),
          );
        },
      ),
    );
  }
}

class _DividerDot extends StatelessWidget {
  const _DividerDot();
  @override
  Widget build(BuildContext context) {
    return Container(
      width: 6,
      height: 6,
      margin: const EdgeInsets.symmetric(horizontal: 2),
      decoration: BoxDecoration(
        color: context.appColors.border.withValues(alpha: 0.7),
        shape: BoxShape.circle,
      ),
    );
  }
}
