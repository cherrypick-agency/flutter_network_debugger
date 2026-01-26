import 'package:flutter/services.dart';
import 'package:flutter/widgets.dart';

/// Следит за lifecycle и сбрасывает состояние клавиатуры при смене активности.
///
/// На desktop иногда теряются KeyUp события при потере фокуса/сворачивании окна,
/// из‑за чего следующий KeyDown может вызвать assert в `HardwareKeyboard`.
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
    // HardwareKeyboard API отличается между версиями Flutter — вызываем мягко.
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
