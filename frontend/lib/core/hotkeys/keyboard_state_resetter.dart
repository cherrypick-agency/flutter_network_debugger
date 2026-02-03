import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart';

/// Monitors lifecycle and resets keyboard state on activity change.
///
/// On desktop, KeyUp events are sometimes lost when focus is lost/window is minimized,
/// which can cause the next KeyDown to trigger an assert in `HardwareKeyboard`.
class KeyboardStateResetter extends StatefulWidget {
  const KeyboardStateResetter({required this.child, super.key});

  final Widget child;

  @override
  State<KeyboardStateResetter> createState() => _KeyboardStateResetterState();
}

class _KeyboardStateResetterState extends State<KeyboardStateResetter>
    with WidgetsBindingObserver {
  static const bool _isFlutterTest = bool.fromEnvironment('FLUTTER_TEST');

  @override
  void initState() {
    super.initState();
    if (_isFlutterTest) return;
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.resumed:
      case AppLifecycleState.inactive:
      case AppLifecycleState.paused:
      case AppLifecycleState.detached:
      case AppLifecycleState.hidden:
        _clearKeyboardState();
        break;
    }
  }

  void _clearKeyboardState() {
    // HardwareKeyboard API differs between Flutter versions — call softly.
    try {
      final dynamic hw = HardwareKeyboard.instance;
      hw.clearState();
    } catch (_) {}
  }

  @override
  void dispose() {
    if (_isFlutterTest) {
      super.dispose();
      return;
    }
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
