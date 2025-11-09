// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'scripts_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$ScriptsStore on _ScriptsStore, Store {
  Computed<List<Script>>? _$enabledScriptsComputed;

  @override
  List<Script> get enabledScripts =>
      (_$enabledScriptsComputed ??= Computed<List<Script>>(
        () => super.enabledScripts,
        name: '_ScriptsStore.enabledScripts',
      )).value;
  Computed<List<Script>>? _$disabledScriptsComputed;

  @override
  List<Script> get disabledScripts =>
      (_$disabledScriptsComputed ??= Computed<List<Script>>(
        () => super.disabledScripts,
        name: '_ScriptsStore.disabledScripts',
      )).value;
  Computed<List<Script>>? _$filteredScriptsComputed;

  @override
  List<Script> get filteredScripts =>
      (_$filteredScriptsComputed ??= Computed<List<Script>>(
        () => super.filteredScripts,
        name: '_ScriptsStore.filteredScripts',
      )).value;

  late final _$scriptsAtom = Atom(
    name: '_ScriptsStore.scripts',
    context: context,
  );

  @override
  ObservableList<Script> get scripts {
    _$scriptsAtom.reportRead();
    return super.scripts;
  }

  @override
  set scripts(ObservableList<Script> value) {
    _$scriptsAtom.reportWrite(value, super.scripts, () {
      super.scripts = value;
    });
  }

  late final _$selectedScriptAtom = Atom(
    name: '_ScriptsStore.selectedScript',
    context: context,
  );

  @override
  Script? get selectedScript {
    _$selectedScriptAtom.reportRead();
    return super.selectedScript;
  }

  @override
  set selectedScript(Script? value) {
    _$selectedScriptAtom.reportWrite(value, super.selectedScript, () {
      super.selectedScript = value;
    });
  }

  late final _$isLoadingAtom = Atom(
    name: '_ScriptsStore.isLoading',
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

  late final _$isImportingAtom = Atom(
    name: '_ScriptsStore.isImporting',
    context: context,
  );

  @override
  bool get isImporting {
    _$isImportingAtom.reportRead();
    return super.isImporting;
  }

  @override
  set isImporting(bool value) {
    _$isImportingAtom.reportWrite(value, super.isImporting, () {
      super.isImporting = value;
    });
  }

  late final _$isExportingAtom = Atom(
    name: '_ScriptsStore.isExporting',
    context: context,
  );

  @override
  bool get isExporting {
    _$isExportingAtom.reportRead();
    return super.isExporting;
  }

  @override
  set isExporting(bool value) {
    _$isExportingAtom.reportWrite(value, super.isExporting, () {
      super.isExporting = value;
    });
  }

  late final _$exportingScriptIdAtom = Atom(
    name: '_ScriptsStore.exportingScriptId',
    context: context,
  );

  @override
  String? get exportingScriptId {
    _$exportingScriptIdAtom.reportRead();
    return super.exportingScriptId;
  }

  @override
  set exportingScriptId(String? value) {
    _$exportingScriptIdAtom.reportWrite(value, super.exportingScriptId, () {
      super.exportingScriptId = value;
    });
  }

  late final _$errorMessageAtom = Atom(
    name: '_ScriptsStore.errorMessage',
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

  late final _$searchQueryAtom = Atom(
    name: '_ScriptsStore.searchQuery',
    context: context,
  );

  @override
  String get searchQuery {
    _$searchQueryAtom.reportRead();
    return super.searchQuery;
  }

  @override
  set searchQuery(String value) {
    _$searchQueryAtom.reportWrite(value, super.searchQuery, () {
      super.searchQuery = value;
    });
  }

  late final _$runtimeFilterAtom = Atom(
    name: '_ScriptsStore.runtimeFilter',
    context: context,
  );

  @override
  ScriptRuntime? get runtimeFilter {
    _$runtimeFilterAtom.reportRead();
    return super.runtimeFilter;
  }

  @override
  set runtimeFilter(ScriptRuntime? value) {
    _$runtimeFilterAtom.reportWrite(value, super.runtimeFilter, () {
      super.runtimeFilter = value;
    });
  }

  late final _$triggerTypeFilterAtom = Atom(
    name: '_ScriptsStore.triggerTypeFilter',
    context: context,
  );

  @override
  TriggerType? get triggerTypeFilter {
    _$triggerTypeFilterAtom.reportRead();
    return super.triggerTypeFilter;
  }

  @override
  set triggerTypeFilter(TriggerType? value) {
    _$triggerTypeFilterAtom.reportWrite(value, super.triggerTypeFilter, () {
      super.triggerTypeFilter = value;
    });
  }

  late final _$enabledFilterAtom = Atom(
    name: '_ScriptsStore.enabledFilter',
    context: context,
  );

  @override
  bool? get enabledFilter {
    _$enabledFilterAtom.reportRead();
    return super.enabledFilter;
  }

  @override
  set enabledFilter(bool? value) {
    _$enabledFilterAtom.reportWrite(value, super.enabledFilter, () {
      super.enabledFilter = value;
    });
  }

  late final _$isCompilingAtom = Atom(
    name: '_ScriptsStore.isCompiling',
    context: context,
  );

  @override
  bool get isCompiling {
    _$isCompilingAtom.reportRead();
    return super.isCompiling;
  }

  @override
  set isCompiling(bool value) {
    _$isCompilingAtom.reportWrite(value, super.isCompiling, () {
      super.isCompiling = value;
    });
  }

  late final _$compilationErrorAtom = Atom(
    name: '_ScriptsStore.compilationError',
    context: context,
  );

  @override
  String? get compilationError {
    _$compilationErrorAtom.reportRead();
    return super.compilationError;
  }

  @override
  set compilationError(String? value) {
    _$compilationErrorAtom.reportWrite(value, super.compilationError, () {
      super.compilationError = value;
    });
  }

  late final _$loadScriptsAsyncAction = AsyncAction(
    '_ScriptsStore.loadScripts',
    context: context,
  );

  @override
  Future<void> loadScripts() {
    return _$loadScriptsAsyncAction.run(() => super.loadScripts());
  }

  late final _$createScriptAsyncAction = AsyncAction(
    '_ScriptsStore.createScript',
    context: context,
  );

  @override
  Future<Script> createScript(Script script) {
    return _$createScriptAsyncAction.run(() => super.createScript(script));
  }

  late final _$updateScriptAsyncAction = AsyncAction(
    '_ScriptsStore.updateScript',
    context: context,
  );

  @override
  Future<Script> updateScript(String id, Script script) {
    return _$updateScriptAsyncAction.run(() => super.updateScript(id, script));
  }

  late final _$deleteScriptAsyncAction = AsyncAction(
    '_ScriptsStore.deleteScript',
    context: context,
  );

  @override
  Future<void> deleteScript(String id) {
    return _$deleteScriptAsyncAction.run(() => super.deleteScript(id));
  }

  late final _$toggleScriptAsyncAction = AsyncAction(
    '_ScriptsStore.toggleScript',
    context: context,
  );

  @override
  Future<void> toggleScript(String id, bool enabled) {
    return _$toggleScriptAsyncAction.run(() => super.toggleScript(id, enabled));
  }

  late final _$compileScriptAsyncAction = AsyncAction(
    '_ScriptsStore.compileScript',
    context: context,
  );

  @override
  Future<Script> compileScript(String id, {bool optimize = true}) {
    return _$compileScriptAsyncAction.run(
      () => super.compileScript(id, optimize: optimize),
    );
  }

  late final _$importFromZipAsyncAction = AsyncAction(
    '_ScriptsStore.importFromZip',
    context: context,
  );

  @override
  Future<Script> importFromZip(List<int> zipBytes) {
    return _$importFromZipAsyncAction.run(() => super.importFromZip(zipBytes));
  }

  late final _$_ScriptsStoreActionController = ActionController(
    name: '_ScriptsStore',
    context: context,
  );

  @override
  void selectScript(Script script) {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.selectScript',
    );
    try {
      return super.selectScript(script);
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearSelection() {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.clearSelection',
    );
    try {
      return super.clearSelection();
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setSearchQuery(String query) {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.setSearchQuery',
    );
    try {
      return super.setSearchQuery(query);
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setRuntimeFilter(ScriptRuntime? runtime) {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.setRuntimeFilter',
    );
    try {
      return super.setRuntimeFilter(runtime);
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setTriggerTypeFilter(TriggerType? triggerType) {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.setTriggerTypeFilter',
    );
    try {
      return super.setTriggerTypeFilter(triggerType);
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setEnabledFilter(bool? enabled) {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.setEnabledFilter',
    );
    try {
      return super.setEnabledFilter(enabled);
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearFilters() {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.clearFilters',
    );
    try {
      return super.clearFilters();
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearError() {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.clearError',
    );
    try {
      return super.clearError();
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearCompilationError() {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.clearCompilationError',
    );
    try {
      return super.clearCompilationError();
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String getExportZipUrl(String id) {
    final _$actionInfo = _$_ScriptsStoreActionController.startAction(
      name: '_ScriptsStore.getExportZipUrl',
    );
    try {
      return super.getExportZipUrl(id);
    } finally {
      _$_ScriptsStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
scripts: ${scripts},
selectedScript: ${selectedScript},
isLoading: ${isLoading},
isImporting: ${isImporting},
isExporting: ${isExporting},
exportingScriptId: ${exportingScriptId},
errorMessage: ${errorMessage},
searchQuery: ${searchQuery},
runtimeFilter: ${runtimeFilter},
triggerTypeFilter: ${triggerTypeFilter},
enabledFilter: ${enabledFilter},
isCompiling: ${isCompiling},
compilationError: ${compilationError},
enabledScripts: ${enabledScripts},
disabledScripts: ${disabledScripts},
filteredScripts: ${filteredScripts}
    ''';
  }
}
