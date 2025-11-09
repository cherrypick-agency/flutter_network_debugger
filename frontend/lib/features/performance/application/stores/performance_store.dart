import 'package:flutter/foundation.dart';
import 'package:mobx/mobx.dart';

import '../../../../core/network/error_utils.dart';
import '../../domain/entities/performance_overview.dart';
import '../../domain/repositories/performance_repository.dart';

part 'performance_store.g.dart';

class PerformanceStore = _PerformanceStore with _$PerformanceStore;

abstract class _PerformanceStore with Store {
  _PerformanceStore(this._repository);
  final PerformanceRepository _repository;

  @observable
  PerformanceOverview? overview;

  @observable
  DateTime from = DateTime.now().subtract(const Duration(hours: 1));

  @observable
  DateTime to = DateTime.now();

  @observable
  bool loading = false;

  @observable
  String? error;

  @action
  Future<void> loadMetrics() async {
    loading = true;
    error = null;
    try {
      overview = await _repository.getOverview(from, to);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error loading performance metrics: $e\n$st');
      }
    } finally {
      loading = false;
    }
  }

  @action
  void setTimeRange(DateTime newFrom, DateTime newTo) {
    from = newFrom;
    to = newTo;
    loadMetrics();
  }

  @action
  void setLastHour() {
    setTimeRange(
      DateTime.now().subtract(const Duration(hours: 1)),
      DateTime.now(),
    );
  }

  @action
  void setLastDay() {
    setTimeRange(
      DateTime.now().subtract(const Duration(days: 1)),
      DateTime.now(),
    );
  }

  @action
  void setLastWeek() {
    setTimeRange(
      DateTime.now().subtract(const Duration(days: 7)),
      DateTime.now(),
    );
  }
}
