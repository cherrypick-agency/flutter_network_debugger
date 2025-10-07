import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';

import '../../../../filters/application/stores/sessions_filters_store.dart';
import '../../../application/stores/sessions_store.dart';
import '../../../application/stores/home_ui_store.dart';
import 'capture_settings_dialog.dart';

// Верхняя панель действий: запись, фильтры, тема, хоткеи, настройки
class HeaderActions extends StatelessWidget {
  const HeaderActions({
    super.key,
    required this.showFilters,
    required this.onToggleFilters,
    required this.onToggleTheme,
    required this.onOpenHotkeys,
    required this.onOpenSettings,
    required this.isRecording,
    required this.onToggleRecording,
    required this.themeMode,
  });

  final bool showFilters;
  final VoidCallback onToggleFilters;
  final VoidCallback? onToggleTheme;
  final VoidCallback onOpenHotkeys;
  final VoidCallback onOpenSettings;
  final bool isRecording;
  final VoidCallback onToggleRecording;
  final ThemeMode themeMode;

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
              tooltip:
                  isRecording
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
                  builder:
                      (_) => CaptureSettingsDialog(
                        initialRecording: isRecording,
                        initialScope:
                            context.read<HomeUiStore>().captureScope.value,
                        initialIncludePaused:
                            context.read<HomeUiStore>().includePaused.value,
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
            return Stack(
              clipBehavior: Clip.none,
              children: [
                IconButton(
                  onPressed: onToggleFilters,
                  tooltip: 'Filters',
                  icon: const Icon(Icons.filter_list),
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
      ],
    );
  }
}
