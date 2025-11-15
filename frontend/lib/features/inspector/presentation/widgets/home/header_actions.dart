import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';

import '../../../../filters/application/stores/sessions_filters_store.dart';
import '../../../application/stores/sessions_store.dart';
import '../../../application/stores/home_ui_store.dart';
import 'capture_settings_dialog.dart';

// Top action bar: recording, filters, theme, hotkeys, settings
class HeaderActions extends StatelessWidget {
  const HeaderActions({
    super.key,
    required this.showFilters,
    required this.onToggleFilters,
    required this.onToggleTheme,
    required this.onOpenHotkeys,
    required this.onOpenSettings,
    this.onOpenUpdates,
    this.onOpenIntegrations,
    required this.isRecording,
    required this.onToggleRecording,
    required this.themeMode,
    this.timelineVisible = true,
    required this.onToggleTimeline,
    this.onOpenCompose,
    this.onOpenScripts,
    this.onOpenBreakpoints,
    this.onOpenMapping,
    this.onOpenExportImport,
    // this.onOpenPerformance,
  });

  final bool showFilters;
  final VoidCallback onToggleFilters;
  final VoidCallback? onToggleTheme;
  final VoidCallback onOpenHotkeys;
  final VoidCallback onOpenSettings;
  final VoidCallback? onOpenUpdates;
  final VoidCallback? onOpenIntegrations;
  final bool isRecording;
  final VoidCallback onToggleRecording;
  final ThemeMode themeMode;
  final bool timelineVisible;
  final VoidCallback onToggleTimeline;
  final VoidCallback? onOpenCompose;
  final VoidCallback? onOpenScripts;
  final VoidCallback? onOpenBreakpoints;
  final VoidCallback? onOpenMapping;
  final VoidCallback? onOpenExportImport;
  // final VoidCallback? onOpenPerformance;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        MenuAnchor(
          builder: (context, controller, child) {
            return IconButton(
              onPressed: onToggleRecording,
              onLongPress: () {
                controller.isOpen ? controller.close() : controller.open();
              },
              tooltip: isRecording
                  ? 'Stop recording (long press for settings)'
                  : 'Start recording (long press for settings)',
              icon: Icon(
                isRecording ? Icons.stop_circle : Icons.radio_button_checked,
                color: isRecording ? Colors.red : Colors.grey,
              ),
            );
          },
          menuChildren: [
            MenuItemButton(
              leadingIcon: const Icon(Icons.tune),
              child: const Text('Open settings'),
              onPressed: () async {
                final applied = await showDialog<bool>(
                  context: context,
                  builder: (_) => CaptureSettingsDialog(
                    initialRecording: isRecording,
                    initialScope: context
                        .read<HomeUiStore>()
                        .captureScope
                        .value,
                    initialIncludePaused: context
                        .read<HomeUiStore>()
                        .includePaused
                        .value,
                  ),
                );
                if (applied == true) {
                  try {
                    await context.read<SessionsStore>().load();
                  } catch (_) {}
                }
              },
            ),
          ],
        ),
        Observer(
          builder: (_) {
            final hasActive = context.read<SessionsFiltersStore>().hasActive;
            final showFiltersActive = showFilters;
            final color = showFiltersActive
                ? Theme.of(context).colorScheme.primary
                : Theme.of(context).colorScheme.onSurfaceVariant;
            return Stack(
              clipBehavior: Clip.none,
              children: [
                IconButton(
                  onPressed: onToggleFilters,
                  tooltip: 'Filters',
                  icon: Icon(Icons.filter_list, color: color),
                ),
                if (hasActive)
                  Positioned(
                    right: 10,
                    top: 10,
                    child: Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.primary,
                        shape: BoxShape.circle,
                      ),
                    ),
                  ),
              ],
            );
          },
        ),
        // Toggle timeline visibility
        IconButton(
          onPressed: onToggleTimeline,
          tooltip: 'Timeline',
          icon: Icon(
            Icons.view_timeline,
            color: timelineVisible
                ? Theme.of(context).colorScheme.primary
                : Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        IconButton(
          onPressed: onToggleTheme,
          tooltip: 'Theme',
          icon: Icon(
            themeMode == ThemeMode.dark
                ? Icons.dark_mode
                : themeMode == ThemeMode.system
                ? Icons.brightness_auto
                : Icons.light_mode,
          ),
        ),
        IconButton(
          onPressed: onOpenHotkeys,
          tooltip: 'Hotkeys',
          icon: const Icon(Icons.keyboard),
        ),
        IconButton(
          onPressed: onOpenSettings,
          tooltip: 'Settings',
          icon: const Icon(Icons.settings),
        ),
        if (onOpenUpdates != null)
          IconButton(
            onPressed: onOpenUpdates,
            tooltip: 'Updates',
            icon: const Icon(Icons.system_update_alt),
          ),
        if (onOpenCompose != null)
          IconButton(
            onPressed: onOpenCompose,
            tooltip: 'Compose',
            icon: const Icon(Icons.playlist_add),
          ),
        if (onOpenScripts != null)
          IconButton(
            onPressed: onOpenScripts,
            tooltip: 'Scripts',
            icon: const Icon(Icons.code),
          ),
        if (onOpenBreakpoints != null)
          IconButton(
            onPressed: onOpenBreakpoints,
            tooltip: 'Breakpoints',
            icon: const Icon(Icons.bug_report),
          ),
        if (onOpenMapping != null)
          IconButton(
            onPressed: onOpenMapping,
            tooltip: 'Mapping',
            icon: const Icon(Icons.link),
          ),
        // if (onOpenPerformance != null)
        //   IconButton(
        //     onPressed: onOpenPerformance,
        //     tooltip: 'Performance Insights',
        //     icon: const Icon(Icons.analytics),
        //   ),
        // Export/Import HAR
        if (onOpenExportImport != null)
          IconButton(
            onPressed: onOpenExportImport,
            tooltip: 'Export / Import HAR',
            icon: const Icon(Icons.import_export),
          ),
        if (onOpenIntegrations != null)
          IconButton(
            onPressed: onOpenIntegrations,
            tooltip: 'Integrations',
            icon: const Icon(Icons.shield),
          ),
      ],
    );
  }
}
