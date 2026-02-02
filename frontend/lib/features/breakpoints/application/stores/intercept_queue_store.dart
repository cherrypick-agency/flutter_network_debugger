import 'dart:async';
import 'package:flutter/foundation.dart';
import '../../domain/entities/intercept_item.dart';
import '../../domain/repositories/breakpoints_repository.dart';
import '../../../inspector/application/services/monitor_service.dart';
import '../../../../core/di/di.dart';
import '../../domain/entities/decisions.dart';

class InterceptQueueStore extends ChangeNotifier {
  InterceptQueueStore(this._repo);
  final BreakpointsRepository _repo;

  List<InterceptItem> _items = const [];
  InterceptItem? _selected;
  MonitorListener? _listener;
  int _refreshStamp = 0;

  List<InterceptItem> get items => _items;
  InterceptItem? get selected => _selected;

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
          refresh().catchError((_) {});
        }
      } catch (_) {}
    };
    monitor.addListener(_listener!);
  }

  void detach() {
    if (_listener != null) {
      try {
        sl<MonitorService>().removeListener(_listener!);
      } catch (_) {}
      _listener = null;
    }
  }

  Future<void> refresh() async {
    final stamp = ++_refreshStamp;
    List<InterceptItem> nextItems;
    try {
      nextItems = await _repo.listPending(limit: 200);
    } catch (_) {
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

    notifyListeners();
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
    } catch (_) {}
    await refresh();
  }

  Future<void> quickCancel(String id) async {
    try {
      await _repo.cancel(id);
    } catch (_) {}
    await refresh();
  }

  @override
  void dispose() {
    detach();
    super.dispose();
  }
}
