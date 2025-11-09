// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'installation_progress_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$InstallationProgressStore on _InstallationProgressStoreBase, Store {
  late final _$progressMapAtom = Atom(
    name: '_InstallationProgressStoreBase.progressMap',
    context: context,
  );

  @override
  ObservableMap<String, DownloadProgress> get progressMap {
    _$progressMapAtom.reportRead();
    return super.progressMap;
  }

  @override
  set progressMap(ObservableMap<String, DownloadProgress> value) {
    _$progressMapAtom.reportWrite(value, super.progressMap, () {
      super.progressMap = value;
    });
  }

  late final _$_InstallationProgressStoreBaseActionController =
      ActionController(
        name: '_InstallationProgressStoreBase',
        context: context,
      );

  @override
  void startWatching(String language) {
    final _$actionInfo = _$_InstallationProgressStoreBaseActionController
        .startAction(name: '_InstallationProgressStoreBase.startWatching');
    try {
      return super.startWatching(language);
    } finally {
      _$_InstallationProgressStoreBaseActionController.endAction(_$actionInfo);
    }
  }

  @override
  void stopWatching(String language) {
    final _$actionInfo = _$_InstallationProgressStoreBaseActionController
        .startAction(name: '_InstallationProgressStoreBase.stopWatching');
    try {
      return super.stopWatching(language);
    } finally {
      _$_InstallationProgressStoreBaseActionController.endAction(_$actionInfo);
    }
  }

  @override
  void stopAll() {
    final _$actionInfo = _$_InstallationProgressStoreBaseActionController
        .startAction(name: '_InstallationProgressStoreBase.stopAll');
    try {
      return super.stopAll();
    } finally {
      _$_InstallationProgressStoreBaseActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
progressMap: ${progressMap}
    ''';
  }
}
