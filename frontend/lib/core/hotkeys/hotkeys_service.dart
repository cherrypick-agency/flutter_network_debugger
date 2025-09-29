import 'dart:async';
import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart';
import 'package:shared_preferences/shared_preferences.dart';

class HotkeyBinding {
  HotkeyBinding({
    required this.id,
    required this.label,
    required this.defaultShortcut,
    ShortcutActivator? current,
  }) : currentShortcut = current ?? defaultShortcut;

  final String id;
  final String label;
  final ShortcutActivator defaultShortcut;
  ShortcutActivator currentShortcut;
}

class HotkeysService {
  static const String _prefsPrefix = 'hk_';

  final Map<String, HotkeyBinding> _bindings = {};
  final StreamController<void> _changes = StreamController<void>.broadcast();
  Stream<void> get changes => _changes.stream;

  Future<void> init() async {
    // defaults (mac-friendly where appropriate)
    // defaults
    _bindings.clear();
    _register(HotkeyBinding(
      id: 'sessions.refresh',
      label: 'Refresh sessions',
      defaultShortcut: SingleActivator(LogicalKeyboardKey.keyR, meta: true),
    ));
    _register(HotkeyBinding(
      id: 'sessions.refresh.ctrl',
      label: 'Refresh sessions (Ctrl for non-macOS)',
      defaultShortcut: const SingleActivator(LogicalKeyboardKey.keyR, control: true),
    ));
    _register(HotkeyBinding(
      id: 'sessions.delete',
      label: 'Delete selected session',
      defaultShortcut: const SingleActivator(LogicalKeyboardKey.delete),
    ));
    _register(HotkeyBinding(
      id: 'sessions.focusSearch',
      label: 'Focus sessions search',
      defaultShortcut: const CharacterActivator('/'),
    ));
    // JSON search
    _register(HotkeyBinding(
      id: 'jsonSearch.next',
      label: 'JSON search: Next match',
      defaultShortcut: const SingleActivator(LogicalKeyboardKey.enter),
    ));
    _register(HotkeyBinding(
      id: 'jsonSearch.prev',
      label: 'JSON search: Previous match',
      defaultShortcut: const SingleActivator(LogicalKeyboardKey.enter, shift: true),
    ));
    _register(HotkeyBinding(
      id: 'jsonSearch.close',
      label: 'JSON search: Close',
      defaultShortcut: const SingleActivator(LogicalKeyboardKey.escape),
    ));

    await _loadFromPrefs();
  }

  List<HotkeyBinding> getAll() => _bindings.values.toList(growable: false);

  HotkeyBinding? getById(String id) => _bindings[id];

  Future<void> setBinding(String id, ShortcutActivator activator) async {
    final b = _bindings[id];
    if (b == null) return;
    b.currentShortcut = activator;
    await _saveToPrefs(id, activator);
    _changes.add(null);
  }

  Future<void> resetToDefault(String id) async {
    final b = _bindings[id];
    if (b == null) return;
    b.currentShortcut = b.defaultShortcut;
    await _saveToPrefs(id, b.defaultShortcut);
    _changes.add(null);
  }

  Map<ShortcutActivator, VoidCallback> buildHandlers(Map<String, VoidCallback> byId) {
    final map = <ShortcutActivator, VoidCallback>{};
    for (final e in byId.entries) {
      final b = _bindings[e.key];
      if (b != null) {
        map[b.currentShortcut] = e.value;
      }
    }
    return map;
  }

  void _register(HotkeyBinding b) {
    _bindings[b.id] = b;
  }

  Future<void> _loadFromPrefs() async {
    final p = await SharedPreferences.getInstance();
    for (final id in _bindings.keys) {
      final s = p.getString('$_prefsPrefix$id');
      if (s == null || s.isEmpty) continue;
      final parsed = _parse(s);
      if (parsed != null) {
        _bindings[id]!.currentShortcut = parsed;
      }
    }
  }

  Future<void> _saveToPrefs(String id, ShortcutActivator activator) async {
    final p = await SharedPreferences.getInstance();
    await p.setString('$_prefsPrefix$id', _format(activator));
  }

  // --- serialization ---
  String _format(ShortcutActivator a) {
    if (a is CharacterActivator) {
      final mods = <String>[];
      if (a.control) mods.add('Ctrl');
      if (a.alt) mods.add('Alt');
      // CharacterActivator не имеет флага shift
      if (a.meta) mods.add('Meta');
      mods.add(a.character);
      return mods.join('+');
    }
    if (a is SingleActivator) {
      final mods = <String>[];
      if (a.control) mods.add('Ctrl');
      if (a.alt) mods.add('Alt');
      if (a.shift) mods.add('Shift');
      if (a.meta) mods.add('Meta');
      mods.add(a.trigger.keyLabel.isNotEmpty ? a.trigger.keyLabel : a.trigger.debugName ?? a.trigger.keyId.toString());
      return mods.join('+');
    }
    return a.toString();
  }

  ShortcutActivator? _parse(String s) {
    final parts = s.split('+');
    bool ctrl = false, alt = false, shift = false, meta = false;
    for (int i = 0; i < parts.length - 1; i++) {
      switch (parts[i].toLowerCase()) {
        case 'ctrl': ctrl = true; break;
        case 'alt': alt = true; break;
        case 'shift': shift = true; break;
        case 'meta': meta = true; break;
      }
    }
    final last = parts.isNotEmpty ? parts.last : '';
    // Character activator if single printable char
    if (last.length == 1) {
      return CharacterActivator(last, control: ctrl, alt: alt, meta: meta);
    }
    // Map common names
    final key = _keyFromName(last);
    if (key != null) {
      return SingleActivator(key, control: ctrl, alt: alt, shift: shift, meta: meta);
    }
    return null;
  }

  LogicalKeyboardKey? _keyFromName(String name) {
    final n = name.toLowerCase();
    switch (n) {
      case 'enter': return LogicalKeyboardKey.enter;
      case 'escape':
      case 'esc': return LogicalKeyboardKey.escape;
      case 'delete': return LogicalKeyboardKey.delete;
      case 'backspace': return LogicalKeyboardKey.backspace;
      case 'r': return LogicalKeyboardKey.keyR;
      case 'slash': return LogicalKeyboardKey.slash;
    }
    // Try map by keyLabel for letters
    if (n.length == 1) {
      final c = n.toUpperCase();
      final code = c.codeUnitAt(0);
      if (code >= 65 && code <= 90) {
        return LogicalKeyboardKey(code);
      }
    }
    return null;
  }
}


