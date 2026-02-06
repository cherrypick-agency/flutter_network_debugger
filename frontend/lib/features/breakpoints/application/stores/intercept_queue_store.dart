import 'dart:async';

import 'package:mobx/mobx.dart';

import '../../domain/entities/intercept_item.dart';
import '../../domain/repositories/breakpoints_repository.dart';
import '../../../inspector/application/services/monitor_service.dart';
import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/utils/debouncer.dart';
import '../../domain/entities/decisions.dart';

part 'intercept_queue_store.g.dart';

class InterceptQueueStore = _InterceptQueueStore with _$InterceptQueueStore;

abstract class _InterceptQueueStore with Store {
  final BreakpointsRepository _repo;

  _InterceptQueueStore(this._repo);

  @observable
  ObservableList<InterceptItem> items = ObservableList<InterceptItem>();

  @observable
  InterceptItem? selected;

  @observable
  bool loading = false;

  @observable
  String? lastError;

  MonitorListener? _listener;
  int _refreshStamp = 0;
  final Debouncer _debounced = Debouncer(const Duration(milliseconds: 200));
  Timer? _retryTimer;
  int _retryAttempt = 0;

  Future<void> init() async {
    await refresh();
    if (_listener != null) return;
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

  @action
  Future<void> refresh() async {
    final stamp = ++_refreshStamp;
    _retryTimer?.cancel();
    _retryTimer = null;
    if (!loading) {
      loading = true;
    }
    List<InterceptItem> nextItems;
    try {
      nextItems = await _repo.listPending(limit: 200);
    } catch (e) {
      if (stamp != _refreshStamp) return;
      lastError = e.toString();
      loading = false;
      _scheduleRetry();
      return;
    }
    if (stamp != _refreshStamp) return;

    items = ObservableList.of(nextItems);

    if (selected != null) {
      final selectedId = selected!.id;
      final match = items.where((e) => e.id == selectedId);
      selected = match.isEmpty ? null : match.first;
    }

    if (selected == null && items.isNotEmpty) {
      selected = items.first;
    }

    _retryAttempt = 0;
    lastError = null;
    loading = false;
  }

  void _scheduleRetry() {
    if (_listener == null) return;
    _retryTimer?.cancel();
    final attempt = _retryAttempt.clamp(0, 6);
    final delayMs = (500 * (1 << attempt)).clamp(500, 10000);
    _retryAttempt = attempt + 1;
    _retryTimer = Timer(Duration(milliseconds: delayMs), () {
      refresh().catchError((_) {});
    });
  }

  @action
  void select(String id) {
    for (final it in items) {
      if (it.id == id) {
        selected = it;
        return;
      }
    }
  }

  @action
  Future<void> quickContinue(String id) async {
    try {
      final it = items.firstWhere((e) => e.id == id);
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

  @action
  Future<void> quickCancel(String id) async {
    try {
      await _repo.cancel(id);
    } catch (e) {
      sl<NotificationsService>().error('Cancel', e.toString());
    }
    await refresh();
  }
}
