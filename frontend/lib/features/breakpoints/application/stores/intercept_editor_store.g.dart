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

  late final _$submittingAtom = Atom(
    name: '_InterceptEditorStore.submitting',
    context: context,
  );

  @override
  bool get submitting {
    _$submittingAtom.reportRead();
    return super.submitting;
  }

  @override
  set submitting(bool value) {
    _$submittingAtom.reportWrite(value, super.submitting, () {
      super.submitting = value;
    });
  }

  late final _$lastErrorAtom = Atom(
    name: '_InterceptEditorStore.lastError',
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

  late final _$continueRequestAsyncAction = AsyncAction(
    '_InterceptEditorStore.continueRequest',
    context: context,
  );

  @override
  Future<void> continueRequest({
    String? method,
    String? url,
    Map<String, List<String>>? headers,
    String? bodyBase64,
    bool drop = false,
  }) {
    return _$continueRequestAsyncAction.run(
      () => super.continueRequest(
        method: method,
        url: url,
        headers: headers,
        bodyBase64: bodyBase64,
        drop: drop,
      ),
    );
  }

  late final _$continueResponseAsyncAction = AsyncAction(
    '_InterceptEditorStore.continueResponse',
    context: context,
  );

  @override
  Future<void> continueResponse({
    int? status,
    Map<String, List<String>>? headers,
    String? bodyBase64,
  }) {
    return _$continueResponseAsyncAction.run(
      () => super.continueResponse(
        status: status,
        headers: headers,
        bodyBase64: bodyBase64,
      ),
    );
  }

  late final _$cancelAsyncAction = AsyncAction(
    '_InterceptEditorStore.cancel',
    context: context,
  );

  @override
  Future<void> cancel() {
    return _$cancelAsyncAction.run(() => super.cancel());
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
item: ${item},
submitting: ${submitting},
lastError: ${lastError}
    ''';
  }
}
