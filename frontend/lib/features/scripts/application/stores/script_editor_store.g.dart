// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'script_editor_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$ScriptEditorStore on _ScriptEditorStore, Store {
  Computed<bool>? _$isValidComputed;

  @override
  bool get isValid => (_$isValidComputed ??= Computed<bool>(
    () => super.isValid,
    name: '_ScriptEditorStore.isValid',
  )).value;
  Computed<bool>? _$isEditingComputed;

  @override
  bool get isEditing => (_$isEditingComputed ??= Computed<bool>(
    () => super.isEditing,
    name: '_ScriptEditorStore.isEditing',
  )).value;
  Computed<String>? _$titleComputed;

  @override
  String get title => (_$titleComputed ??= Computed<String>(
    () => super.title,
    name: '_ScriptEditorStore.title',
  )).value;
  Computed<MatchRules?>? _$matchRulesComputed;

  @override
  MatchRules? get matchRules => (_$matchRulesComputed ??= Computed<MatchRules?>(
    () => super.matchRules,
    name: '_ScriptEditorStore.matchRules',
  )).value;
  Computed<ScriptConfig?>? _$configComputed;

  @override
  ScriptConfig? get config => (_$configComputed ??= Computed<ScriptConfig?>(
    () => super.config,
    name: '_ScriptEditorStore.config',
  )).value;

  late final _$editingScriptIdAtom = Atom(
    name: '_ScriptEditorStore.editingScriptId',
    context: context,
  );

  @override
  String? get editingScriptId {
    _$editingScriptIdAtom.reportRead();
    return super.editingScriptId;
  }

  @override
  set editingScriptId(String? value) {
    _$editingScriptIdAtom.reportWrite(value, super.editingScriptId, () {
      super.editingScriptId = value;
    });
  }

  late final _$nameAtom = Atom(
    name: '_ScriptEditorStore.name',
    context: context,
  );

  @override
  String get name {
    _$nameAtom.reportRead();
    return super.name;
  }

  @override
  set name(String value) {
    _$nameAtom.reportWrite(value, super.name, () {
      super.name = value;
    });
  }

  late final _$descriptionAtom = Atom(
    name: '_ScriptEditorStore.description',
    context: context,
  );

  @override
  String get description {
    _$descriptionAtom.reportRead();
    return super.description;
  }

  @override
  set description(String value) {
    _$descriptionAtom.reportWrite(value, super.description, () {
      super.description = value;
    });
  }

  late final _$runtimeAtom = Atom(
    name: '_ScriptEditorStore.runtime',
    context: context,
  );

  @override
  ScriptRuntime get runtime {
    _$runtimeAtom.reportRead();
    return super.runtime;
  }

  @override
  set runtime(ScriptRuntime value) {
    _$runtimeAtom.reportWrite(value, super.runtime, () {
      super.runtime = value;
    });
  }

  late final _$codeAtom = Atom(
    name: '_ScriptEditorStore.code',
    context: context,
  );

  @override
  String get code {
    _$codeAtom.reportRead();
    return super.code;
  }

  @override
  set code(String value) {
    _$codeAtom.reportWrite(value, super.code, () {
      super.code = value;
    });
  }

  late final _$languageAtom = Atom(
    name: '_ScriptEditorStore.language',
    context: context,
  );

  @override
  String get language {
    _$languageAtom.reportRead();
    return super.language;
  }

  @override
  set language(String value) {
    _$languageAtom.reportWrite(value, super.language, () {
      super.language = value;
    });
  }

  late final _$codeCreationModeAtom = Atom(
    name: '_ScriptEditorStore.codeCreationMode',
    context: context,
  );

  @override
  CodeCreationMode get codeCreationMode {
    _$codeCreationModeAtom.reportRead();
    return super.codeCreationMode;
  }

  @override
  set codeCreationMode(CodeCreationMode value) {
    _$codeCreationModeAtom.reportWrite(value, super.codeCreationMode, () {
      super.codeCreationMode = value;
    });
  }

  late final _$sourceFilesAtom = Atom(
    name: '_ScriptEditorStore.sourceFiles',
    context: context,
  );

  @override
  ObservableMap<String, String> get sourceFiles {
    _$sourceFilesAtom.reportRead();
    return super.sourceFiles;
  }

  @override
  set sourceFiles(ObservableMap<String, String> value) {
    _$sourceFilesAtom.reportWrite(value, super.sourceFiles, () {
      super.sourceFiles = value;
    });
  }

  late final _$selectedFileAtom = Atom(
    name: '_ScriptEditorStore.selectedFile',
    context: context,
  );

  @override
  String? get selectedFile {
    _$selectedFileAtom.reportRead();
    return super.selectedFile;
  }

  @override
  set selectedFile(String? value) {
    _$selectedFileAtom.reportWrite(value, super.selectedFile, () {
      super.selectedFile = value;
    });
  }

  late final _$triggerTypeAtom = Atom(
    name: '_ScriptEditorStore.triggerType',
    context: context,
  );

  @override
  TriggerType get triggerType {
    _$triggerTypeAtom.reportRead();
    return super.triggerType;
  }

  @override
  set triggerType(TriggerType value) {
    _$triggerTypeAtom.reportWrite(value, super.triggerType, () {
      super.triggerType = value;
    });
  }

  late final _$priorityAtom = Atom(
    name: '_ScriptEditorStore.priority',
    context: context,
  );

  @override
  int get priority {
    _$priorityAtom.reportRead();
    return super.priority;
  }

  @override
  set priority(int value) {
    _$priorityAtom.reportWrite(value, super.priority, () {
      super.priority = value;
    });
  }

  late final _$enabledAtom = Atom(
    name: '_ScriptEditorStore.enabled',
    context: context,
  );

  @override
  bool get enabled {
    _$enabledAtom.reportRead();
    return super.enabled;
  }

  @override
  set enabled(bool value) {
    _$enabledAtom.reportWrite(value, super.enabled, () {
      super.enabled = value;
    });
  }

  late final _$useCustomMatchRulesAtom = Atom(
    name: '_ScriptEditorStore.useCustomMatchRules',
    context: context,
  );

  @override
  bool get useCustomMatchRules {
    _$useCustomMatchRulesAtom.reportRead();
    return super.useCustomMatchRules;
  }

  @override
  set useCustomMatchRules(bool value) {
    _$useCustomMatchRulesAtom.reportWrite(value, super.useCustomMatchRules, () {
      super.useCustomMatchRules = value;
    });
  }

  late final _$selectedMethodsAtom = Atom(
    name: '_ScriptEditorStore.selectedMethods',
    context: context,
  );

  @override
  ObservableList<String> get selectedMethods {
    _$selectedMethodsAtom.reportRead();
    return super.selectedMethods;
  }

  @override
  set selectedMethods(ObservableList<String> value) {
    _$selectedMethodsAtom.reportWrite(value, super.selectedMethods, () {
      super.selectedMethods = value;
    });
  }

  late final _$pathPatternAtom = Atom(
    name: '_ScriptEditorStore.pathPattern',
    context: context,
  );

  @override
  String get pathPattern {
    _$pathPatternAtom.reportRead();
    return super.pathPattern;
  }

  @override
  set pathPattern(String value) {
    _$pathPatternAtom.reportWrite(value, super.pathPattern, () {
      super.pathPattern = value;
    });
  }

  late final _$hostPatternAtom = Atom(
    name: '_ScriptEditorStore.hostPattern',
    context: context,
  );

  @override
  String get hostPattern {
    _$hostPatternAtom.reportRead();
    return super.hostPattern;
  }

  @override
  set hostPattern(String value) {
    _$hostPatternAtom.reportWrite(value, super.hostPattern, () {
      super.hostPattern = value;
    });
  }

  late final _$patternTypeAtom = Atom(
    name: '_ScriptEditorStore.patternType',
    context: context,
  );

  @override
  PatternType get patternType {
    _$patternTypeAtom.reportRead();
    return super.patternType;
  }

  @override
  set patternType(PatternType value) {
    _$patternTypeAtom.reportWrite(value, super.patternType, () {
      super.patternType = value;
    });
  }

  late final _$timeoutMsAtom = Atom(
    name: '_ScriptEditorStore.timeoutMs',
    context: context,
  );

  @override
  int? get timeoutMs {
    _$timeoutMsAtom.reportRead();
    return super.timeoutMs;
  }

  @override
  set timeoutMs(int? value) {
    _$timeoutMsAtom.reportWrite(value, super.timeoutMs, () {
      super.timeoutMs = value;
    });
  }

  late final _$memoryLimitMBAtom = Atom(
    name: '_ScriptEditorStore.memoryLimitMB',
    context: context,
  );

  @override
  int? get memoryLimitMB {
    _$memoryLimitMBAtom.reportRead();
    return super.memoryLimitMB;
  }

  @override
  set memoryLimitMB(int? value) {
    _$memoryLimitMBAtom.reportWrite(value, super.memoryLimitMB, () {
      super.memoryLimitMB = value;
    });
  }

  late final _$compilationStatusAtom = Atom(
    name: '_ScriptEditorStore.compilationStatus',
    context: context,
  );

  @override
  CompilationStatus? get compilationStatus {
    _$compilationStatusAtom.reportRead();
    return super.compilationStatus;
  }

  @override
  set compilationStatus(CompilationStatus? value) {
    _$compilationStatusAtom.reportWrite(value, super.compilationStatus, () {
      super.compilationStatus = value;
    });
  }

  late final _$compilationErrorAtom = Atom(
    name: '_ScriptEditorStore.compilationError',
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

  late final _$lastCompiledAtAtom = Atom(
    name: '_ScriptEditorStore.lastCompiledAt',
    context: context,
  );

  @override
  DateTime? get lastCompiledAt {
    _$lastCompiledAtAtom.reportRead();
    return super.lastCompiledAt;
  }

  @override
  set lastCompiledAt(DateTime? value) {
    _$lastCompiledAtAtom.reportWrite(value, super.lastCompiledAt, () {
      super.lastCompiledAt = value;
    });
  }

  late final _$nameErrorAtom = Atom(
    name: '_ScriptEditorStore.nameError',
    context: context,
  );

  @override
  String? get nameError {
    _$nameErrorAtom.reportRead();
    return super.nameError;
  }

  @override
  set nameError(String? value) {
    _$nameErrorAtom.reportWrite(value, super.nameError, () {
      super.nameError = value;
    });
  }

  late final _$codeErrorAtom = Atom(
    name: '_ScriptEditorStore.codeError',
    context: context,
  );

  @override
  String? get codeError {
    _$codeErrorAtom.reportRead();
    return super.codeError;
  }

  @override
  set codeError(String? value) {
    _$codeErrorAtom.reportWrite(value, super.codeError, () {
      super.codeError = value;
    });
  }

  late final _$sourceFilesErrorAtom = Atom(
    name: '_ScriptEditorStore.sourceFilesError',
    context: context,
  );

  @override
  String? get sourceFilesError {
    _$sourceFilesErrorAtom.reportRead();
    return super.sourceFilesError;
  }

  @override
  set sourceFilesError(String? value) {
    _$sourceFilesErrorAtom.reportWrite(value, super.sourceFilesError, () {
      super.sourceFilesError = value;
    });
  }

  late final _$errorMessageAtom = Atom(
    name: '_ScriptEditorStore.errorMessage',
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

  late final _$isTestingAtom = Atom(
    name: '_ScriptEditorStore.isTesting',
    context: context,
  );

  @override
  bool get isTesting {
    _$isTestingAtom.reportRead();
    return super.isTesting;
  }

  @override
  set isTesting(bool value) {
    _$isTestingAtom.reportWrite(value, super.isTesting, () {
      super.isTesting = value;
    });
  }

  late final _$testResultAtom = Atom(
    name: '_ScriptEditorStore.testResult',
    context: context,
  );

  @override
  ScriptTestResult? get testResult {
    _$testResultAtom.reportRead();
    return super.testResult;
  }

  @override
  set testResult(ScriptTestResult? value) {
    _$testResultAtom.reportWrite(value, super.testResult, () {
      super.testResult = value;
    });
  }

  late final _$testErrorAtom = Atom(
    name: '_ScriptEditorStore.testError',
    context: context,
  );

  @override
  String? get testError {
    _$testErrorAtom.reportRead();
    return super.testError;
  }

  @override
  set testError(String? value) {
    _$testErrorAtom.reportWrite(value, super.testError, () {
      super.testError = value;
    });
  }

  late final _$isValidatingAtom = Atom(
    name: '_ScriptEditorStore.isValidating',
    context: context,
  );

  @override
  bool get isValidating {
    _$isValidatingAtom.reportRead();
    return super.isValidating;
  }

  @override
  set isValidating(bool value) {
    _$isValidatingAtom.reportWrite(value, super.isValidating, () {
      super.isValidating = value;
    });
  }

  late final _$isValidSyntaxAtom = Atom(
    name: '_ScriptEditorStore.isValidSyntax',
    context: context,
  );

  @override
  bool get isValidSyntax {
    _$isValidSyntaxAtom.reportRead();
    return super.isValidSyntax;
  }

  @override
  set isValidSyntax(bool value) {
    _$isValidSyntaxAtom.reportWrite(value, super.isValidSyntax, () {
      super.isValidSyntax = value;
    });
  }

  late final _$syntaxValidationErrorAtom = Atom(
    name: '_ScriptEditorStore.syntaxValidationError',
    context: context,
  );

  @override
  String? get syntaxValidationError {
    _$syntaxValidationErrorAtom.reportRead();
    return super.syntaxValidationError;
  }

  @override
  set syntaxValidationError(String? value) {
    _$syntaxValidationErrorAtom.reportWrite(
      value,
      super.syntaxValidationError,
      () {
        super.syntaxValidationError = value;
      },
    );
  }

  late final _$isUploadingProjectAtom = Atom(
    name: '_ScriptEditorStore.isUploadingProject',
    context: context,
  );

  @override
  bool get isUploadingProject {
    _$isUploadingProjectAtom.reportRead();
    return super.isUploadingProject;
  }

  @override
  set isUploadingProject(bool value) {
    _$isUploadingProjectAtom.reportWrite(value, super.isUploadingProject, () {
      super.isUploadingProject = value;
    });
  }

  late final _$currentTabAtom = Atom(
    name: '_ScriptEditorStore.currentTab',
    context: context,
  );

  @override
  int get currentTab {
    _$currentTabAtom.reportRead();
    return super.currentTab;
  }

  @override
  set currentTab(int value) {
    _$currentTabAtom.reportWrite(value, super.currentTab, () {
      super.currentTab = value;
    });
  }

  late final _$testScriptAsyncAction = AsyncAction(
    '_ScriptEditorStore.testScript',
    context: context,
  );

  @override
  Future<void> testScript(TestRequest testRequest) {
    return _$testScriptAsyncAction.run(() => super.testScript(testRequest));
  }

  late final _$validateSyntaxAsyncAction = AsyncAction(
    '_ScriptEditorStore.validateSyntax',
    context: context,
  );

  @override
  Future<void> validateSyntax() {
    return _$validateSyntaxAsyncAction.run(() => super.validateSyntax());
  }

  late final _$uploadProjectZipAsyncAction = AsyncAction(
    '_ScriptEditorStore.uploadProjectZip',
    context: context,
  );

  @override
  Future<Map<String, dynamic>?> uploadProjectZip(List<int> zipBytes) {
    return _$uploadProjectZipAsyncAction.run(
      () => super.uploadProjectZip(zipBytes),
    );
  }

  late final _$loadProjectFilesAsyncAction = AsyncAction(
    '_ScriptEditorStore.loadProjectFiles',
    context: context,
  );

  @override
  Future<List<Map<String, dynamic>>> loadProjectFiles() {
    return _$loadProjectFilesAsyncAction.run(() => super.loadProjectFiles());
  }

  late final _$_ScriptEditorStoreActionController = ActionController(
    name: '_ScriptEditorStore',
    context: context,
  );

  @override
  void initForNewScript() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.initForNewScript',
    );
    try {
      return super.initForNewScript();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void initForEdit(Script script) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.initForEdit',
    );
    try {
      return super.initForEdit(script);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void reset() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.reset',
    );
    try {
      return super.reset();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setName(String value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setName',
    );
    try {
      return super.setName(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setDescription(String value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setDescription',
    );
    try {
      return super.setDescription(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setRuntime(ScriptRuntime value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setRuntime',
    );
    try {
      return super.setRuntime(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setCode(String value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setCode',
    );
    try {
      return super.setCode(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setLanguage(String value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setLanguage',
    );
    try {
      return super.setLanguage(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setCodeCreationMode(CodeCreationMode mode) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setCodeCreationMode',
    );
    try {
      return super.setCodeCreationMode(mode);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void addSourceFile(String filename, String content) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.addSourceFile',
    );
    try {
      return super.addSourceFile(filename, content);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void updateSourceFile(String filename, String content) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.updateSourceFile',
    );
    try {
      return super.updateSourceFile(filename, content);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void removeSourceFile(String filename) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.removeSourceFile',
    );
    try {
      return super.removeSourceFile(filename);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void selectFile(String filename) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.selectFile',
    );
    try {
      return super.selectFile(filename);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearSourceFiles() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.clearSourceFiles',
    );
    try {
      return super.clearSourceFiles();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setTriggerType(TriggerType value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setTriggerType',
    );
    try {
      return super.setTriggerType(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setPriority(int value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setPriority',
    );
    try {
      return super.setPriority(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setEnabled(bool value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setEnabled',
    );
    try {
      return super.setEnabled(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setUseCustomMatchRules(bool value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setUseCustomMatchRules',
    );
    try {
      return super.setUseCustomMatchRules(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void toggleMethod(String method) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.toggleMethod',
    );
    try {
      return super.toggleMethod(method);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setPathPattern(String value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setPathPattern',
    );
    try {
      return super.setPathPattern(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setHostPattern(String value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setHostPattern',
    );
    try {
      return super.setHostPattern(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setPatternType(PatternType value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setPatternType',
    );
    try {
      return super.setPatternType(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setTimeoutMs(int? value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setTimeoutMs',
    );
    try {
      return super.setTimeoutMs(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setMemoryLimitMB(int? value) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setMemoryLimitMB',
    );
    try {
      return super.setMemoryLimitMB(value);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void setCurrentTab(int tab) {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.setCurrentTab',
    );
    try {
      return super.setCurrentTab(tab);
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void validateName() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.validateName',
    );
    try {
      return super.validateName();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void validateCode() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.validateCode',
    );
    try {
      return super.validateCode();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void validateSourceFiles() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.validateSourceFiles',
    );
    try {
      return super.validateSourceFiles();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  bool validate() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.validate',
    );
    try {
      return super.validate();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearError() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.clearError',
    );
    try {
      return super.clearError();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void clearValidation() {
    final _$actionInfo = _$_ScriptEditorStoreActionController.startAction(
      name: '_ScriptEditorStore.clearValidation',
    );
    try {
      return super.clearValidation();
    } finally {
      _$_ScriptEditorStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
editingScriptId: ${editingScriptId},
name: ${name},
description: ${description},
runtime: ${runtime},
code: ${code},
language: ${language},
codeCreationMode: ${codeCreationMode},
sourceFiles: ${sourceFiles},
selectedFile: ${selectedFile},
triggerType: ${triggerType},
priority: ${priority},
enabled: ${enabled},
useCustomMatchRules: ${useCustomMatchRules},
selectedMethods: ${selectedMethods},
pathPattern: ${pathPattern},
hostPattern: ${hostPattern},
patternType: ${patternType},
timeoutMs: ${timeoutMs},
memoryLimitMB: ${memoryLimitMB},
compilationStatus: ${compilationStatus},
compilationError: ${compilationError},
lastCompiledAt: ${lastCompiledAt},
nameError: ${nameError},
codeError: ${codeError},
sourceFilesError: ${sourceFilesError},
errorMessage: ${errorMessage},
isTesting: ${isTesting},
testResult: ${testResult},
testError: ${testError},
isValidating: ${isValidating},
isValidSyntax: ${isValidSyntax},
syntaxValidationError: ${syntaxValidationError},
isUploadingProject: ${isUploadingProject},
currentTab: ${currentTab},
isValid: ${isValid},
isEditing: ${isEditing},
title: ${title},
matchRules: ${matchRules},
config: ${config}
    ''';
  }
}
