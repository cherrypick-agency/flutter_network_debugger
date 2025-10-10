import 'dart:async';
import '../../../../services/prefs.dart';
import '../stores/home_ui_store.dart';
import '../stores/sessions_store.dart';
import '../../../../core/di/di.dart';

class RecentWindowService {
  Timer? _timer;

  Future<void> initFromPrefs() async {
    final enabled = await PrefsService().loadRecentWindowEnabled();
    final minutes = await PrefsService().loadRecentWindowMinutes();
    final ui = sl<HomeUiStore>();
    ui.setRecentWindowEnabled(enabled);
    ui.setRecentWindowMinutes(minutes);
    _apply(enabled: enabled, minutes: minutes);
  }

  void apply({required bool enabled, required int minutes}) {
    PrefsService().saveRecentWindow(enabled: enabled, minutes: minutes);
    final ui = sl<HomeUiStore>();
    ui.setRecentWindowEnabled(enabled);
    ui.setRecentWindowMinutes(minutes);
    _apply(enabled: enabled, minutes: minutes);
    // Подтянем список сессий под новое окно
    try {
      sl<SessionsStore>().load();
    } catch (_) {}
  }

  void _apply({required bool enabled, required int minutes}) {
    _timer?.cancel();
    final ui = sl<HomeUiStore>();
    if (!enabled) {
      ui.setSince(null);
      return;
    }
    // первое применение сразу
    ui.setSince(DateTime.now().subtract(Duration(minutes: minutes)));
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      ui.setSince(DateTime.now().subtract(Duration(minutes: minutes)));
    });
  }

  void dispose() {
    _timer?.cancel();
    _timer = null;
  }
}
