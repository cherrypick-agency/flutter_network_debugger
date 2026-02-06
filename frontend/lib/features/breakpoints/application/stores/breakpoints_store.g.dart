// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'breakpoints_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$BreakpointsStore on _BreakpointsStore, Store {
  late final _$configAtom = Atom(
    name: '_BreakpointsStore.config',
    context: context,
  );

  @override
  InterceptConfig? get config {
    _$configAtom.reportRead();
    return super.config;
  }

  @override
  set config(InterceptConfig? value) {
    _$configAtom.reportWrite(value, super.config, () {
      super.config = value;
    });
  }

  late final _$rulesAtom = Atom(
    name: '_BreakpointsStore.rules',
    context: context,
  );

  @override
  ObservableList<InterceptRule> get rules {
    _$rulesAtom.reportRead();
    return super.rules;
  }

  @override
  set rules(ObservableList<InterceptRule> value) {
    _$rulesAtom.reportWrite(value, super.rules, () {
      super.rules = value;
    });
  }

  late final _$loadingAtom = Atom(
    name: '_BreakpointsStore.loading',
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
    name: '_BreakpointsStore.lastError',
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

  late final _$loadAsyncAction = AsyncAction(
    '_BreakpointsStore.load',
    context: context,
  );

  @override
  Future<void> load() {
    return _$loadAsyncAction.run(() => super.load());
  }

  late final _$saveConfigAsyncAction = AsyncAction(
    '_BreakpointsStore.saveConfig',
    context: context,
  );

  @override
  Future<void> saveConfig(InterceptConfig cfg) {
    return _$saveConfigAsyncAction.run(() => super.saveConfig(cfg));
  }

  late final _$replaceRulesAsyncAction = AsyncAction(
    '_BreakpointsStore.replaceRules',
    context: context,
  );

  @override
  Future<void> replaceRules(List<InterceptRule> newRules) {
    return _$replaceRulesAsyncAction.run(() => super.replaceRules(newRules));
  }

  @override
  String toString() {
    return '''
config: ${config},
rules: ${rules},
loading: ${loading},
lastError: ${lastError}
    ''';
  }
}
