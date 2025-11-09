// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'performance_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$PerformanceStore on _PerformanceStore, Store {
  late final _$overviewAtom = Atom(
    name: '_PerformanceStore.overview',
    context: context,
  );

  @override
  PerformanceOverview? get overview {
    _$overviewAtom.reportRead();
    return super.overview;
  }

  @override
  set overview(PerformanceOverview? value) {
    _$overviewAtom.reportWrite(value, super.overview, () {
      super.overview = value;
    });
  }

  late final _$fromAtom = Atom(
    name: '_PerformanceStore.from',
    context: context,
  );

  @override
  DateTime get from {
    _$fromAtom.reportRead();
    return super.from;
  }

  @override
  set from(DateTime value) {
    _$fromAtom.reportWrite(value, super.from, () {
      super.from = value;
    });
  }

  late final _$toAtom = Atom(name: '_PerformanceStore.to', context: context);

  @override
  DateTime get to {
    _$toAtom.reportRead();
    return super.to;
  }

  @override
  set to(DateTime value) {
    _$toAtom.reportWrite(value, super.to, () {
      super.to = value;
    });
  }

  late final _$loadingAtom = Atom(
    name: '_PerformanceStore.loading',
    context: context,
  );

  @override
  bool get loading {
    _$loadingAtom.reportRead();
    return super.loading;
  }

  @override
  set loading(bool value) {
    _$loadingAtom.reportWrite(value, super.loading, () {
      super.loading = value;
    });
  }

  late final _$errorAtom = Atom(
    name: '_PerformanceStore.error',
    context: context,
  );

  @override
  String? get error {
    _$errorAtom.reportRead();
    return super.error;
  }

  @override
  set error(String? value) {
    _$errorAtom.reportWrite(value, super.error, () {
      super.error = value;
    });
  }

  late final _$loadMetricsAsyncAction = AsyncAction(
    '_PerformanceStore.loadMetrics',
    context: context,
  );

  @override
  Future<void> loadMetrics() {
    return _$loadMetricsAsyncAction.run(() => super.loadMetrics());
  }

  late final _$_PerformanceStoreActionController = ActionController(
    name: '_PerformanceStore',
    context: context,
  );

  @override
  void setTimeRange(DateTime newFrom, DateTime newTo) {
    final _$actionInfo = _$_PerformanceStoreActionController.startAction(
      name: '_PerformanceStore.setTimeRange',
    );
    try {
      return super.setTimeRange(newFrom, newTo);
    } finally {
      _$_PerformanceStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setLastHour() {
    final _$actionInfo = _$_PerformanceStoreActionController.startAction(
      name: '_PerformanceStore.setLastHour',
    );
    try {
      return super.setLastHour();
    } finally {
      _$_PerformanceStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setLastDay() {
    final _$actionInfo = _$_PerformanceStoreActionController.startAction(
      name: '_PerformanceStore.setLastDay',
    );
    try {
      return super.setLastDay();
    } finally {
      _$_PerformanceStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setLastWeek() {
    final _$actionInfo = _$_PerformanceStoreActionController.startAction(
      name: '_PerformanceStore.setLastWeek',
    );
    try {
      return super.setLastWeek();
    } finally {
      _$_PerformanceStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
overview: ${overview},
from: ${from},
to: ${to},
loading: ${loading},
error: ${error}
    ''';
  }
}
