// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sessions_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$SessionsStore on _SessionsStore, Store {
  late final _$itemsAtom = Atom(name: '_SessionsStore.items', context: context);

  @override
  ObservableList<Session> get items {
    _$itemsAtom.reportRead();
    return super.items;
  }

  @override
  set items(ObservableList<Session> value) {
    _$itemsAtom.reportWrite(value, super.items, () {
      super.items = value;
    });
  }

  late final _$loadingAtom = Atom(
    name: '_SessionsStore.loading',
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

  late final _$loadAsyncAction = AsyncAction(
    '_SessionsStore.load',
    context: context,
  );

  @override
  Future<void> load({
    String? q,
    String? target,
    List<String>? tags,
    Set<String>? types,
    Set<String>? statusGroups,
  }) {
    return _$loadAsyncAction.run(
      () => super.load(
        q: q,
        target: target,
        tags: tags,
        types: types,
        statusGroups: statusGroups,
      ),
    );
  }

  late final _$_SessionsStoreActionController = ActionController(
    name: '_SessionsStore',
    context: context,
  );

  @override
  void clear() {
    final _$actionInfo = _$_SessionsStoreActionController.startAction(
      name: '_SessionsStore.clear',
    );
    try {
      return super.clear();
    } finally {
      _$_SessionsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void applyInit(List<Session> data) {
    final _$actionInfo = _$_SessionsStoreActionController.startAction(
      name: '_SessionsStore.applyInit',
    );
    try {
      return super.applyInit(data);
    } finally {
      _$_SessionsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void upsert(Session s) {
    final _$actionInfo = _$_SessionsStoreActionController.startAction(
      name: '_SessionsStore.upsert',
    );
    try {
      return super.upsert(s);
    } finally {
      _$_SessionsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void removeById(String id) {
    final _$actionInfo = _$_SessionsStoreActionController.startAction(
      name: '_SessionsStore.removeById',
    );
    try {
      return super.removeById(id);
    } finally {
      _$_SessionsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
items: ${items},
loading: ${loading}
    ''';
  }
}
