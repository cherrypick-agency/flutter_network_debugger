// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'export_import_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$ExportImportStore on _ExportImportStore, Store {
  Computed<bool>? _$canExportComputed;

  @override
  bool get canExport => (_$canExportComputed ??= Computed<bool>(
    () => super.canExport,
    name: '_ExportImportStore.canExport',
  )).value;
  Computed<bool>? _$canImportComputed;

  @override
  bool get canImport => (_$canImportComputed ??= Computed<bool>(
    () => super.canImport,
    name: '_ExportImportStore.canImport',
  )).value;
  Computed<String>? _$exportScopeDescriptionComputed;

  @override
  String get exportScopeDescription =>
      (_$exportScopeDescriptionComputed ??= Computed<String>(
        () => super.exportScopeDescription,
        name: '_ExportImportStore.exportScopeDescription',
      )).value;
  Computed<String>? _$visibleScopeDescriptionComputed;

  @override
  String get visibleScopeDescription =>
      (_$visibleScopeDescriptionComputed ??= Computed<String>(
        () => super.visibleScopeDescription,
        name: '_ExportImportStore.visibleScopeDescription',
      )).value;
  Computed<String>? _$allScopeDescriptionComputed;

  @override
  String get allScopeDescription =>
      (_$allScopeDescriptionComputed ??= Computed<String>(
        () => super.allScopeDescription,
        name: '_ExportImportStore.allScopeDescription',
      )).value;
  Computed<String?>? _$estimatedFileSizeComputed;

  @override
  String? get estimatedFileSize =>
      (_$estimatedFileSizeComputed ??= Computed<String?>(
        () => super.estimatedFileSize,
        name: '_ExportImportStore.estimatedFileSize',
      )).value;

  late final _$stateAtom = Atom(
    name: '_ExportImportStore.state',
    context: context,
  );

  @override
  ExportImportState get state {
    _$stateAtom.reportRead();
    return super.state;
  }

  @override
  set state(ExportImportState value) {
    _$stateAtom.reportWrite(value, super.state, () {
      super.state = value;
    });
  }

  late final _$selectedModeAtom = Atom(
    name: '_ExportImportStore.selectedMode',
    context: context,
  );

  @override
  ExportImportMode? get selectedMode {
    _$selectedModeAtom.reportRead();
    return super.selectedMode;
  }

  @override
  set selectedMode(ExportImportMode? value) {
    _$selectedModeAtom.reportWrite(value, super.selectedMode, () {
      super.selectedMode = value;
    });
  }

  late final _$exportScopeAtom = Atom(
    name: '_ExportImportStore.exportScope',
    context: context,
  );

  @override
  ExportScope get exportScope {
    _$exportScopeAtom.reportRead();
    return super.exportScope;
  }

  @override
  set exportScope(ExportScope value) {
    _$exportScopeAtom.reportWrite(value, super.exportScope, () {
      super.exportScope = value;
    });
  }

  late final _$includeBodiesAtom = Atom(
    name: '_ExportImportStore.includeBodies',
    context: context,
  );

  @override
  bool get includeBodies {
    _$includeBodiesAtom.reportRead();
    return super.includeBodies;
  }

  @override
  set includeBodies(bool value) {
    _$includeBodiesAtom.reportWrite(value, super.includeBodies, () {
      super.includeBodies = value;
    });
  }

  late final _$includeSensitiveAtom = Atom(
    name: '_ExportImportStore.includeSensitive',
    context: context,
  );

  @override
  bool get includeSensitive {
    _$includeSensitiveAtom.reportRead();
    return super.includeSensitive;
  }

  @override
  set includeSensitive(bool value) {
    _$includeSensitiveAtom.reportWrite(value, super.includeSensitive, () {
      super.includeSensitive = value;
    });
  }

  late final _$minifyAtom = Atom(
    name: '_ExportImportStore.minify',
    context: context,
  );

  @override
  bool get minify {
    _$minifyAtom.reportRead();
    return super.minify;
  }

  @override
  set minify(bool value) {
    _$minifyAtom.reportWrite(value, super.minify, () {
      super.minify = value;
    });
  }

  late final _$importModeAtom = Atom(
    name: '_ExportImportStore.importMode',
    context: context,
  );

  @override
  ImportMode get importMode {
    _$importModeAtom.reportRead();
    return super.importMode;
  }

  @override
  set importMode(ImportMode value) {
    _$importModeAtom.reportWrite(value, super.importMode, () {
      super.importMode = value;
    });
  }

  late final _$progressAtom = Atom(
    name: '_ExportImportStore.progress',
    context: context,
  );

  @override
  double get progress {
    _$progressAtom.reportRead();
    return super.progress;
  }

  @override
  set progress(double value) {
    _$progressAtom.reportWrite(value, super.progress, () {
      super.progress = value;
    });
  }

  late final _$progressMessageAtom = Atom(
    name: '_ExportImportStore.progressMessage',
    context: context,
  );

  @override
  String? get progressMessage {
    _$progressMessageAtom.reportRead();
    return super.progressMessage;
  }

  @override
  set progressMessage(String? value) {
    _$progressMessageAtom.reportWrite(value, super.progressMessage, () {
      super.progressMessage = value;
    });
  }

  late final _$errorMessageAtom = Atom(
    name: '_ExportImportStore.errorMessage',
    context: context,
  );

  @override
  String? get errorMessage {
    _$errorMessageAtom.reportRead();
    return super.errorMessage;
  }

  @override
  set errorMessage(String? value) {
    _$errorMessageAtom.reportWrite(value, super.errorMessage, () {
      super.errorMessage = value;
    });
  }

  late final _$successMessageAtom = Atom(
    name: '_ExportImportStore.successMessage',
    context: context,
  );

  @override
  String? get successMessage {
    _$successMessageAtom.reportRead();
    return super.successMessage;
  }

  @override
  set successMessage(String? value) {
    _$successMessageAtom.reportWrite(value, super.successMessage, () {
      super.successMessage = value;
    });
  }

  late final _$exportedCountAtom = Atom(
    name: '_ExportImportStore.exportedCount',
    context: context,
  );

  @override
  int? get exportedCount {
    _$exportedCountAtom.reportRead();
    return super.exportedCount;
  }

  @override
  set exportedCount(int? value) {
    _$exportedCountAtom.reportWrite(value, super.exportedCount, () {
      super.exportedCount = value;
    });
  }

  late final _$importedCountAtom = Atom(
    name: '_ExportImportStore.importedCount',
    context: context,
  );

  @override
  int? get importedCount {
    _$importedCountAtom.reportRead();
    return super.importedCount;
  }

  @override
  set importedCount(int? value) {
    _$importedCountAtom.reportWrite(value, super.importedCount, () {
      super.importedCount = value;
    });
  }

  late final _$failedCountAtom = Atom(
    name: '_ExportImportStore.failedCount',
    context: context,
  );

  @override
  int? get failedCount {
    _$failedCountAtom.reportRead();
    return super.failedCount;
  }

  @override
  set failedCount(int? value) {
    _$failedCountAtom.reportWrite(value, super.failedCount, () {
      super.failedCount = value;
    });
  }

  late final _$importFileDataAtom = Atom(
    name: '_ExportImportStore.importFileData',
    context: context,
  );

  @override
  Uint8List? get importFileData {
    _$importFileDataAtom.reportRead();
    return super.importFileData;
  }

  @override
  set importFileData(Uint8List? value) {
    _$importFileDataAtom.reportWrite(value, super.importFileData, () {
      super.importFileData = value;
    });
  }

  late final _$importFileNameAtom = Atom(
    name: '_ExportImportStore.importFileName',
    context: context,
  );

  @override
  String? get importFileName {
    _$importFileNameAtom.reportRead();
    return super.importFileName;
  }

  @override
  set importFileName(String? value) {
    _$importFileNameAtom.reportWrite(value, super.importFileName, () {
      super.importFileName = value;
    });
  }

  late final _$importFileSizeAtom = Atom(
    name: '_ExportImportStore.importFileSize',
    context: context,
  );

  @override
  int? get importFileSize {
    _$importFileSizeAtom.reportRead();
    return super.importFileSize;
  }

  @override
  set importFileSize(int? value) {
    _$importFileSizeAtom.reportWrite(value, super.importFileSize, () {
      super.importFileSize = value;
    });
  }

  late final _$visibleSessionsCountAtom = Atom(
    name: '_ExportImportStore.visibleSessionsCount',
    context: context,
  );

  @override
  int? get visibleSessionsCount {
    _$visibleSessionsCountAtom.reportRead();
    return super.visibleSessionsCount;
  }

  @override
  set visibleSessionsCount(int? value) {
    _$visibleSessionsCountAtom.reportWrite(
      value,
      super.visibleSessionsCount,
      () {
        super.visibleSessionsCount = value;
      },
    );
  }

  late final _$totalSessionsCountAtom = Atom(
    name: '_ExportImportStore.totalSessionsCount',
    context: context,
  );

  @override
  int? get totalSessionsCount {
    _$totalSessionsCountAtom.reportRead();
    return super.totalSessionsCount;
  }

  @override
  set totalSessionsCount(int? value) {
    _$totalSessionsCountAtom.reportWrite(value, super.totalSessionsCount, () {
      super.totalSessionsCount = value;
    });
  }

  late final _$harEntriesCountAtom = Atom(
    name: '_ExportImportStore.harEntriesCount',
    context: context,
  );

  @override
  int? get harEntriesCount {
    _$harEntriesCountAtom.reportRead();
    return super.harEntriesCount;
  }

  @override
  set harEntriesCount(int? value) {
    _$harEntriesCountAtom.reportWrite(value, super.harEntriesCount, () {
      super.harEntriesCount = value;
    });
  }

  late final _$performExportAsyncAction = AsyncAction(
    '_ExportImportStore.performExport',
    context: context,
  );

  @override
  Future<void> performExport(List<String>? sessionIds) {
    return _$performExportAsyncAction.run(
      () => super.performExport(sessionIds),
    );
  }

  late final _$performImportAsyncAction = AsyncAction(
    '_ExportImportStore.performImport',
    context: context,
  );

  @override
  Future<void> performImport() {
    return _$performImportAsyncAction.run(() => super.performImport());
  }

  late final _$_ExportImportStoreActionController = ActionController(
    name: '_ExportImportStore',
    context: context,
  );

  @override
  void selectMode(ExportImportMode mode) {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.selectMode',
    );
    try {
      return super.selectMode(mode);
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void backToModeSelection() {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.backToModeSelection',
    );
    try {
      return super.backToModeSelection();
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setExportScope(ExportScope scope) {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.setExportScope',
    );
    try {
      return super.setExportScope(scope);
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void toggleIncludeBodies() {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.toggleIncludeBodies',
    );
    try {
      return super.toggleIncludeBodies();
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void toggleIncludeSensitive() {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.toggleIncludeSensitive',
    );
    try {
      return super.toggleIncludeSensitive();
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void toggleMinify() {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.toggleMinify',
    );
    try {
      return super.toggleMinify();
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setImportMode(ImportMode mode) {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.setImportMode',
    );
    try {
      return super.setImportMode(mode);
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setVisibleSessionsCount(int count) {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.setVisibleSessionsCount',
    );
    try {
      return super.setVisibleSessionsCount(count);
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setTotalSessionsCount(int count) {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.setTotalSessionsCount',
    );
    try {
      return super.setTotalSessionsCount(count);
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setImportFile(Uint8List data, String fileName, int fileSize) {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.setImportFile',
    );
    try {
      return super.setImportFile(data, fileName, fileSize);
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearImportFile() {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.clearImportFile',
    );
    try {
      return super.clearImportFile();
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void reset() {
    final _$actionInfo = _$_ExportImportStoreActionController.startAction(
      name: '_ExportImportStore.reset',
    );
    try {
      return super.reset();
    } finally {
      _$_ExportImportStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
state: ${state},
selectedMode: ${selectedMode},
exportScope: ${exportScope},
includeBodies: ${includeBodies},
includeSensitive: ${includeSensitive},
minify: ${minify},
importMode: ${importMode},
progress: ${progress},
progressMessage: ${progressMessage},
errorMessage: ${errorMessage},
successMessage: ${successMessage},
exportedCount: ${exportedCount},
importedCount: ${importedCount},
failedCount: ${failedCount},
importFileData: ${importFileData},
importFileName: ${importFileName},
importFileSize: ${importFileSize},
visibleSessionsCount: ${visibleSessionsCount},
totalSessionsCount: ${totalSessionsCount},
harEntriesCount: ${harEntriesCount},
canExport: ${canExport},
canImport: ${canImport},
exportScopeDescription: ${exportScopeDescription},
visibleScopeDescription: ${visibleScopeDescription},
allScopeDescription: ${allScopeDescription},
estimatedFileSize: ${estimatedFileSize}
    ''';
  }
}
