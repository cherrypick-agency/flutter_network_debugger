import 'dart:async';
import 'package:flutter/foundation.dart';
import '../../domain/entities/intercept_item.dart';
import '../../domain/repositories/breakpoints_repository.dart';
import '../../../inspector/application/services/monitor_service.dart';
import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/utils/debouncer.dart';
import '../../domain/entities/decisions.dart';

class InterceptQueueStore extends ChangeNotifier {
  InterceptQueueStore(this._repo);
  final BreakpointsRepository _repo;

  List<InterceptItem> _items = const [];
  InterceptItem? _selected;
  MonitorListener? _listener;
  int _refreshStamp = 0;
  final Debouncer _debounced = Debouncer(const Duration(milliseconds: 200));
  bool _loading = false;
  String? _lastError;
  Timer? _retryTimer;
  int _retryAttempt = 0;

  List<InterceptItem> get items => _items;
  InterceptItem? get selected => _selected;
  bool get loading => _loading;
  String? get lastError => _lastError;

  Future<void> init() async {
    await refresh();
    // если уже подписаны — повторно не подписываемся
    if (_listener != null) {
      return;
    }
    final monitor = sl<MonitorService>();
    _listener = (Map<String, dynamic> ev) {
      try {
        final t = (ev['type'] ?? '').toString();
        if (t.startsWith('intercept_')) {
          _debounced.run(() {
            refresh().catchError((_) {});
          });
        }
      } catch (_) {}
    };
    monitor.addListener(_listener!);
  }

  void detach() {
    _debounced.dispose();
    _retryTimer?.cancel();
    _retryTimer = null;
    if (_listener != null) {
      try {
        sl<MonitorService>().removeListener(_listener!);
      } catch (_) {}
      _listener = null;
    }
  }

  Future<void> refresh() async {
    final stamp = ++_refreshStamp;
    _retryTimer?.cancel();
    _retryTimer = null;
    if (!_loading) {
      _loading = true;
      notifyListeners();
    }
    List<InterceptItem> nextItems;
    try {
      nextItems = await _repo.listPending(limit: 200);
    } catch (e) {
      if (stamp != _refreshStamp) return;
      _lastError = e.toString();
      _loading = false;
      // На ошибке оставляем старый список/выбор — так UX стабильнее при сетевых глюках.
      notifyListeners();
      _scheduleRetry();
      return;
    }
    if (stamp != _refreshStamp) return;

    _items = nextItems;

    // Обновляем выбранный элемент из списка pending (getItem здесь не нужен —
    // listPending уже отдаёт полный снапшот).
    if (_selected != null) {
      final selectedId = _selected!.id;
      final match = _items.where((e) => e.id == selectedId);
      _selected = match.isEmpty ? null : match.first;
    }

    if (_selected == null && _items.isNotEmpty) {
      _selected = _items.first;
    }

    _retryAttempt = 0;
    _lastError = null;
    _loading = false;
    notifyListeners();
  }

  void _scheduleRetry() {
    if (_listener == null) return; // если диалог уже закрыли — не ретраим
    _retryTimer?.cancel();
    final attempt = _retryAttempt.clamp(0, 6);
    final delayMs = (500 * (1 << attempt)).clamp(500, 10000);
    _retryAttempt = attempt + 1;
    _retryTimer = Timer(Duration(milliseconds: delayMs), () {
      refresh().catchError((_) {});
    });
  }

  void select(String id) {
    for (final it in _items) {
      if (it.id == id) {
        _selected = it;
        notifyListeners();
        return;
      }
    }
    // Если элемента уже нет в списке (например, пришло обновление очереди),
    // просто оставляем текущее значение.
    notifyListeners();
  }

  Future<void> quickContinue(String id) async {
    try {
      final it = _items.firstWhere((e) => e.id == id);
      if (it.direction == 'request') {
        await _repo.continueRequest(
          id,
          const RequestDecision(action: 'continue'),
        );
      } else {
        await _repo.continueResponse(
          id,
          const ResponseDecision(action: 'continue'),
        );
      }
    } catch (e) {
      sl<NotificationsService>().error('Continue', e.toString());
    }
    await refresh();
  }

  Future<void> quickCancel(String id) async {
    try {
      await _repo.cancel(id);
    } catch (e) {
      sl<NotificationsService>().error('Cancel', e.toString());
    }
    await refresh();
  }

  @override
  void dispose() {
    detach();
    super.dispose();
  }
}
