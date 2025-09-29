// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'aggregate_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$AggregateStore on _AggregateStore, Store {
  late final _$groupsAtom = Atom(
    name: '_AggregateStore.groups',
    context: context,
  );

  @override
  ObservableList<Map<String, dynamic>> get groups {
    _$groupsAtom.reportRead();
    return super.groups;
  }

  @override
  set groups(ObservableList<Map<String, dynamic>> value) {
    _$groupsAtom.reportWrite(value, super.groups, () {
      super.groups = value;
    });
  }

  late final _$loadingAtom = Atom(
    name: '_AggregateStore.loading',
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
    '_AggregateStore.load',
    context: context,
  );

  @override
  Future<void> load({String groupBy = 'domain'}) {
    return _$loadAsyncAction.run(() => super.load(groupBy: groupBy));
  }

  @override
  String toString() {
    return '''
groups: ${groups},
loading: ${loading}
    ''';
  }
}
