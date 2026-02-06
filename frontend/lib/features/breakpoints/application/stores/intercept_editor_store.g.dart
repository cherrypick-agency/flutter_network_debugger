// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'intercept_editor_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$InterceptEditorStore on _InterceptEditorStore, Store {
  late final _$itemAtom = Atom(
    name: '_InterceptEditorStore.item',
    context: context,
  );

  @override
  InterceptItem? get item {
    _$itemAtom.reportRead();
    return super.item;
  }

  @override
  set item(InterceptItem? value) {
    _$itemAtom.reportWrite(value, super.item, () {
      super.item = value;
    });
  }

  late final _$_InterceptEditorStoreActionController = ActionController(
    name: '_InterceptEditorStore',
    context: context,
  );

  @override
  void setItem(InterceptItem? it) {
    final _$actionInfo = _$_InterceptEditorStoreActionController.startAction(
      name: '_InterceptEditorStore.setItem',
    );
    try {
      return super.setItem(it);
    } finally {
      _$_InterceptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
item: ${item}
    ''';
  }
}
