// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'compiler_list_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$CompilerListStore on _CompilerListStoreBase, Store {
  Computed<List<CompilerInfo>>? _$installedCompilersComputed;

  @override
  List<CompilerInfo> get installedCompilers =>
      (_$installedCompilersComputed ??= Computed<List<CompilerInfo>>(
        () => super.installedCompilers,
        name: '_CompilerListStoreBase.installedCompilers',
      )).value;
  Computed<List<CompilerInfo>>? _$availableCompilersComputed;

  @override
  List<CompilerInfo> get availableCompilers =>
      (_$availableCompilersComputed ??= Computed<List<CompilerInfo>>(
        () => super.availableCompilers,
        name: '_CompilerListStoreBase.availableCompilers',
      )).value;
  Computed<int>? _$totalCacheSizeComputed;

  @override
  int get totalCacheSize => (_$totalCacheSizeComputed ??= Computed<int>(
    () => super.totalCacheSize,
    name: '_CompilerListStoreBase.totalCacheSize',
  )).value;

  late final _$compilersAtom = Atom(
    name: '_CompilerListStoreBase.compilers',
    context: context,
  );

  @override
  ObservableList<CompilerInfo> get compilers {
    _$compilersAtom.reportRead();
    return super.compilers;
  }

  @override
  set compilers(ObservableList<CompilerInfo> value) {
    _$compilersAtom.reportWrite(value, super.compilers, () {
      super.compilers = value;
    });
  }

  late final _$isLoadingAtom = Atom(
    name: '_CompilerListStoreBase.isLoading',
    context: context,
  );

  @override
  bool get isLoading {
    _$isLoadingAtom.reportRead();
    return super.isLoading;
  }

  @override
  set isLoading(bool value) {
    _$isLoadingAtom.reportWrite(value, super.isLoading, () {
      super.isLoading = value;
    });
  }

  late final _$errorAtom = Atom(
    name: '_CompilerListStoreBase.error',
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

  late final _$loadCompilersAsyncAction = AsyncAction(
    '_CompilerListStoreBase.loadCompilers',
    context: context,
  );

  @override
  Future<void> loadCompilers() {
    return _$loadCompilersAsyncAction.run(() => super.loadCompilers());
  }

  late final _$installAsyncAction = AsyncAction(
    '_CompilerListStoreBase.install',
    context: context,
  );

  @override
  Future<void> install(String language) {
    return _$installAsyncAction.run(() => super.install(language));
  }

  late final _$uninstallAsyncAction = AsyncAction(
    '_CompilerListStoreBase.uninstall',
    context: context,
  );

  @override
  Future<void> uninstall(String language) {
    return _$uninstallAsyncAction.run(() => super.uninstall(language));
  }

  late final _$_CompilerListStoreBaseActionController = ActionController(
    name: '_CompilerListStoreBase',
    context: context,
  );

  @override
  void clearError() {
    final _$actionInfo = _$_CompilerListStoreBaseActionController.startAction(
      name: '_CompilerListStoreBase.clearError',
    );
    try {
      return super.clearError();
    } finally {
      _$_CompilerListStoreBaseActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
compilers: ${compilers},
isLoading: ${isLoading},
error: ${error},
installedCompilers: ${installedCompilers},
availableCompilers: ${availableCompilers},
totalCacheSize: ${totalCacheSize}
    ''';
  }
}
