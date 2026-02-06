// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'intercept_queue_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$InterceptQueueStore on _InterceptQueueStore, Store {
  late final _$itemsAtom = Atom(
    name: '_InterceptQueueStore.items',
    context: context,
  );

  @override
  ObservableList<InterceptItem> get items {
    _$itemsAtom.reportRead();
    return super.items;
  }

  @override
  set items(ObservableList<InterceptItem> value) {
    _$itemsAtom.reportWrite(value, super.items, () {
      super.items = value;
    });
  }

  late final _$selectedAtom = Atom(
    name: '_InterceptQueueStore.selected',
    context: context,
  );

  @override
  InterceptItem? get selected {
    _$selectedAtom.reportRead();
    return super.selected;
  }

  @override
  set selected(InterceptItem? value) {
    _$selectedAtom.reportWrite(value, super.selected, () {
      super.selected = value;
    });
  }

  late final _$loadingAtom = Atom(
    name: '_InterceptQueueStore.loading',
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

  late final _$lastErrorAtom = Atom(
    name: '_InterceptQueueStore.lastError',
    context: context,
  );

  @override
  String? get lastError {
    _$lastErrorAtom.reportRead();
    return super.lastError;
  }

  @override
  set lastError(String? value) {
    _$lastErrorAtom.reportWrite(value, super.lastError, () {
      super.lastError = value;
    });
  }

  late final _$refreshAsyncAction = AsyncAction(
    '_InterceptQueueStore.refresh',
    context: context,
  );

  @override
  Future<void> refresh() {
    return _$refreshAsyncAction.run(() => super.refresh());
  }

  late final _$quickContinueAsyncAction = AsyncAction(
    '_InterceptQueueStore.quickContinue',
    context: context,
  );

  @override
  Future<void> quickContinue(String id) {
    return _$quickContinueAsyncAction.run(() => super.quickContinue(id));
  }

  late final _$quickCancelAsyncAction = AsyncAction(
    '_InterceptQueueStore.quickCancel',
    context: context,
  );

  @override
  Future<void> quickCancel(String id) {
    return _$quickCancelAsyncAction.run(() => super.quickCancel(id));
  }

  late final _$_InterceptQueueStoreActionController = ActionController(
    name: '_InterceptQueueStore',
    context: context,
  );

  @override
  void select(String id) {
    final _$actionInfo = _$_InterceptQueueStoreActionController.startAction(
      name: '_InterceptQueueStore.select',
    );
    try {
      return super.select(id);
    } finally {
      _$_InterceptQueueStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
items: ${items},
selected: ${selected},
loading: ${loading},
lastError: ${lastError}
    ''';
  }
}
