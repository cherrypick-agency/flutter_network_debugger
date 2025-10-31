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

  List<InterceptItem> get items => _items;
  InterceptItem? get selected => _selected;

  Future<void> init() async {
    await refresh();
    final monitor = sl<MonitorService>();
    _listener = (Map<String, dynamic> ev) {
      try {
        final t = (ev['type'] ?? '').toString();
        if (t.startsWith('intercept_')) {
          refresh();
        }
      } catch (_) {}
    };
    monitor.addListener(_listener!);
  }

  Future<void> refresh() async {
    _items = await _repo.listPending(limit: 200);
    if (_selected != null) {
      // обновим выбранный элемент, если он ещё есть
      final id = _selected!.id;
      try {
        _selected = await _repo.getItem(id);
      } catch (_) {}
    } else if (_items.isNotEmpty) {
      _selected = _items.first;
    }
    notifyListeners();
  }

  void select(String id) {
    _selected = _items.firstWhere((e) => e.id == id, orElse: () => _selected!);
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
    if (_listener != null) {
      try {
        sl<MonitorService>().removeListener(_listener!);
      } catch (_) {}
    }
    super.dispose();
  }
}
