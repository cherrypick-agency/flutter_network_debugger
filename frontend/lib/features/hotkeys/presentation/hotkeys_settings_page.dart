import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../core/hotkeys/hotkeys_service.dart';
import '../../../core/di/di.dart';

class HotkeysSettingsPage extends StatefulWidget {
  const HotkeysSettingsPage({super.key});
  @override
  State<HotkeysSettingsPage> createState() => _HotkeysSettingsPageState();
}

class _HotkeysSettingsPageState extends State<HotkeysSettingsPage> {
  @override
  Widget build(BuildContext context) {
    final service = sl<HotkeysService>();
    return Scaffold(
      appBar: AppBar(title: const Text('Hotkeys')),
      body: StreamBuilder<void>(
        stream: service.changes,
        builder: (context, _) {
          final items = service.getAll();
          return ListView.separated(
            padding: const EdgeInsets.all(12),
            itemCount: items.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (_, i) {
              final b = items[i];
              return ListTile(
                dense: true,
                title: Text(b.label),
                subtitle: Text(_formatActivator(b.currentShortcut), style: Theme.of(context).textTheme.bodySmall),
                trailing: Wrap(spacing: 8, children: [
                  OutlinedButton(onPressed: () async { await _editBinding(context, b); }, child: const Text('Edit')),
                  TextButton(onPressed: () async { await sl<HotkeysService>().resetToDefault(b.id); }, child: const Text('Reset')),
                ]),
              );
            },
          );
        },
      ),
    );
  }

  String _formatActivator(ShortcutActivator a) {
    if (a is CharacterActivator) {
      final mods = <String>[];
      if (a.control) mods.add('Ctrl');
      if (a.alt) mods.add('Alt');
      // CharacterActivator не имеет флага shift
      if (a.meta) mods.add('Meta');
      mods.add(a.character.toUpperCase());
      return mods.join(' + ');
    }
    if (a is SingleActivator) {
      final mods = <String>[];
      if (a.control) mods.add('Ctrl');
      if (a.alt) mods.add('Alt');
      if (a.shift) mods.add('Shift');
      if (a.meta) mods.add('Meta');
      final name = a.trigger.keyLabel.isNotEmpty ? a.trigger.keyLabel.toUpperCase() : (a.trigger.debugName ?? a.trigger.keyId.toString());
      mods.add(name);
      return mods.join(' + ');
    }
    return a.toString();
  }

  Future<void> _editBinding(BuildContext context, HotkeyBinding b) async {
    ShortcutActivator? selected;
    await showDialog(context: context, builder: (_) {
      return _HotkeyCaptureDialog(
        title: b.label,
        onCaptured: (act){ selected = act; Navigator.of(context).pop(); },
      );
    });
    if (selected != null) {
      await sl<HotkeysService>().setBinding(b.id, selected!);
    }
  }
}

class _HotkeyCaptureDialog extends StatefulWidget {
  const _HotkeyCaptureDialog({required this.title, required this.onCaptured});
  final String title;
  final void Function(ShortcutActivator) onCaptured;
  @override
  State<_HotkeyCaptureDialog> createState() => _HotkeyCaptureDialogState();
}

class _HotkeyCaptureDialogState extends State<_HotkeyCaptureDialog> {
  String _hint = 'Press desired keys...';
  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget.title),
      content: Focus(
        autofocus: true,
        onKey: (node, evt) {
          if (evt is! RawKeyDownEvent) return KeyEventResult.handled;
          final key = evt.logicalKey;
          // Skip pure modifier presses
          if (key == LogicalKeyboardKey.shift || key == LogicalKeyboardKey.meta || key == LogicalKeyboardKey.control || key == LogicalKeyboardKey.alt) {
            return KeyEventResult.handled;
          }
          final act = SingleActivator(
            key,
            control: evt.isControlPressed,
            alt: evt.isAltPressed,
            shift: evt.isShiftPressed,
            meta: evt.isMetaPressed,
          );
          setState(() { _hint = 'Captured: ${key.keyLabel.isNotEmpty ? key.keyLabel : (key.debugName ?? key.keyId)}'; });
          widget.onCaptured(act);
          return KeyEventResult.handled;
        },
        child: SizedBox(width: 320, child: Text(_hint)),
      ),
      actions: [TextButton(onPressed: ()=> Navigator.of(context).pop(), child: const Text('Cancel'))],
    );
  }
}


