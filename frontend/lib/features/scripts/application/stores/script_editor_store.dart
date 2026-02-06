import 'package:mobx/mobx.dart';
import '../../domain/entities/script.dart';
import '../../domain/entities/match_rules.dart';
import '../../domain/entities/script_config.dart';
import '../../domain/entities/script_test_result.dart';
import '../../domain/usecases/test_script_usecase.dart';
import '../../domain/usecases/validate_syntax_usecase.dart';
import '../../domain/usecases/upload_project_usecase.dart';
import '../../domain/usecases/list_project_files_usecase.dart';
import '../../domain/usecases/download_project_usecase.dart';

part 'script_editor_store.g.dart';

/// Code creation mode for Extism runtime
enum CodeCreationMode {
  uploadWasm, // Upload pre-compiled WASM file
  writeSource, // Write source code (Rust/Go/etc) and compile
  importZip, // Import multi-file project from ZIP
}

/// Store for Script Editor Dialog
/// Manages form state, validation, and testing
class ScriptEditorStore = _ScriptEditorStore with _$ScriptEditorStore;

abstract class _ScriptEditorStore with Store {
  final TestScriptUseCase? _testUseCase;
  final ValidateSyntaxUseCase? _validateSyntaxUseCase;
  final UploadProjectUseCase? _uploadProjectUseCase;
  final ListProjectFilesUseCase? _listProjectFilesUseCase;
  final DownloadProjectUseCase? _downloadProjectUseCase;

  _ScriptEditorStore({
    TestScriptUseCase? testUseCase,
    ValidateSyntaxUseCase? validateSyntaxUseCase,
    UploadProjectUseCase? uploadProjectUseCase,
    ListProjectFilesUseCase? listProjectFilesUseCase,
    DownloadProjectUseCase? downloadProjectUseCase,
  }) : _testUseCase = testUseCase,
       _validateSyntaxUseCase = validateSyntaxUseCase,
       _uploadProjectUseCase = uploadProjectUseCase,
       _listProjectFilesUseCase = listProjectFilesUseCase,
       _downloadProjectUseCase = downloadProjectUseCase;

  // Form fields
  @observable
  String? editingScriptId; // null for new script, ID for editing

  @observable
  String name = '';

  @observable
  String description = '';

  @observable
  ScriptRuntime runtime = ScriptRuntime.extism;

  @observable
  String code = ''; // base64 encoded or source code

  @observable
  String language = 'rust';

  @observable
  CodeCreationMode codeCreationMode = CodeCreationMode.writeSource;

  // Multi-file support (for writeSource and importZip modes)
  @observable
  ObservableMap<String, String> sourceFiles = ObservableMap<String, String>();

  @observable
  String? selectedFile; // Currently selected file in multi-file editor

  @observable
  TriggerType triggerType = TriggerType.request;

  @observable
  int priority = 10;

  @observable
  bool enabled = true;

  // Match rules
  @observable
  bool useCustomMatchRules = false;

  @observable
  ObservableList<String> selectedMethods = ObservableList<String>();

  @observable
  String pathPattern = '';

  @observable
  String hostPattern = '';

  @observable
  PatternType patternType = PatternType.wildcard;

  // Config
  @observable
  int? timeoutMs;

  @observable
  int? memoryLimitMB;

  // Compilation state (preserved during edit)
  @observable
  CompilationStatus? compilationStatus;

  @observable
  String? compilationError;

  @observable
  DateTime? lastCompiledAt;

  // Validation
  @observable
  String? nameError;

  @observable
  String? codeError;

  @observable
  String? sourceFilesError;

  @observable
  String? errorMessage; // General error message

  // Test state
  @observable
  bool isTesting = false;

  @observable
  ScriptTestResult? testResult;

  @observable
  String? testError;

  // Syntax validation state
  @observable
  bool isValidating = false;

  @observable
  bool isValidSyntax = false;

  @observable
  String? syntaxValidationError;

  // Upload project state
  @observable
  bool isUploadingProject = false;

  // Tab state
  @observable
  int currentTab = 0; // 0: Code, 1: Settings, 2: Match Rules, 3: Test

  // Computed
  @computed
  bool get isValid =>
      nameError == null &&
      codeError == null &&
      sourceFilesError == null &&
      name.isNotEmpty &&
      (code.isNotEmpty || sourceFiles.isNotEmpty);

  @computed
  bool get isEditing => editingScriptId != null;

  @computed
  String get title => isEditing ? 'Edit Script' : 'Create Script';

  @computed
  MatchRules? get matchRules {
    if (!useCustomMatchRules) return null;

    return MatchRules(
      methods: selectedMethods.toList(),
      pathPattern: pathPattern.isEmpty ? null : pathPattern,
      hostPattern: hostPattern.isEmpty ? null : hostPattern,
      patternType: patternType,
    );
  }

  @computed
  ScriptConfig? get config {
    if (timeoutMs == null && memoryLimitMB == null) return null;

    return ScriptConfig(timeoutMs: timeoutMs, memoryLimitMB: memoryLimitMB);
  }

  // Actions
  @action
  void initForNewScript() {
    reset();
    editingScriptId = null;
    // Create starter file and auto-select it
    final starterFileName = _getStarterFileName();
    final starterCode = _getStarterCode();
    addSourceFile(starterFileName, starterCode);
  }

  @action
  void initForEdit(Script script) {
    reset();
    editingScriptId = script.id;
    name = script.name;
    description = script.description ?? '';
    runtime = script.runtime;
    code = script.code;
    language = script.language;
    triggerType = script.triggerType;
    priority = script.priority;
    enabled = script.enabled;

    // Restore source files from dependencies for multi-file projects
    if (script.dependencies != null && script.dependencies!.isNotEmpty) {
      sourceFiles.clear();
      sourceFiles.addAll(script.dependencies!);
      selectedFile = sourceFiles.keys.firstOrNull;
      // If we have dependencies, set appropriate mode
      codeCreationMode = CodeCreationMode.writeSource;
    }

    // Match rules
    if (script.matchRules != null && !script.matchRules!.isEmpty) {
      useCustomMatchRules = true;
      selectedMethods.clear();
      selectedMethods.addAll(script.matchRules!.methods);
      pathPattern = script.matchRules!.pathPattern ?? '';
      hostPattern = script.matchRules!.hostPattern ?? '';
      patternType = script.matchRules!.patternType;
    }

    // Config
    if (script.config != null) {
      timeoutMs = script.config!.timeoutMs;
      memoryLimitMB = script.config!.memoryLimitMB;
    }

    // Compilation state
    compilationStatus = script.compilationStatus;
    compilationError = script.compilationError;
    lastCompiledAt = script.lastCompiledAt;
  }

  @action
  void reset() {
    editingScriptId = null;
    name = '';
    description = '';
    runtime = ScriptRuntime.extism;
    code = '';
    language = 'rust';
    codeCreationMode = CodeCreationMode.writeSource;
    sourceFiles.clear();
    selectedFile = null;
    triggerType = TriggerType.request;
    priority = 10;
    enabled = true;
    useCustomMatchRules = false;
    selectedMethods.clear();
    pathPattern = '';
    hostPattern = '';
    patternType = PatternType.wildcard;
    timeoutMs = null;
    memoryLimitMB = null;
    compilationStatus = null;
    compilationError = null;
    lastCompiledAt = null;
    nameError = null;
    codeError = null;
    sourceFilesError = null;
    currentTab = 0;
    testResult = null;
    testError = null;
    isValidating = false;
    isValidSyntax = false;
    syntaxValidationError = null;
    isUploadingProject = false;
    errorMessage = null;
  }

  @action
  void setName(String value) {
    name = value;
    validateName();
  }

  @action
  void setDescription(String value) {
    description = value;
  }

  @action
  void setRuntime(ScriptRuntime value) {
    runtime = value;

    // Auto-adjust language
    if (value == ScriptRuntime.dart) {
      language = 'dart';
      if (sourceFiles.isNotEmpty) {
        sourceFiles.clear();
        selectedFile = null;
      }
    } else if (value == ScriptRuntime.extism && language == 'dart') {
      language = 'rust';
    }
  }

  @action
  void setCode(String value) {
    code = value;
    clearValidation();
    validateCode();
  }

  @action
  void setLanguage(String value) {
    language = value;
  }

  @action
  void setCodeCreationMode(CodeCreationMode mode) {
    // Clear old data when switching modes
    if (mode != codeCreationMode) {
      if (mode == CodeCreationMode.uploadWasm) {
        // Switching TO uploadWasm: clear sourceFiles
        sourceFiles.clear();
        selectedFile = null;
        sourceFilesError = null;
      } else if (mode == CodeCreationMode.importZip) {
        // Switching TO importZip: clear sourceFiles to show upload zone
        sourceFiles.clear();
        selectedFile = null;
        sourceFilesError = null;
      } else if (codeCreationMode == CodeCreationMode.uploadWasm) {
        // Switching FROM uploadWasm: clear code
        code = '';
        codeError = null;
      }

      // Auto-create starter file when switching TO writeSource with empty sourceFiles
      if (mode == CodeCreationMode.writeSource && sourceFiles.isEmpty) {
        final starterFileName = _getStarterFileName();
        final starterCode = _getStarterCode();
        addSourceFile(starterFileName, starterCode);
      }
    }

    codeCreationMode = mode;
  }

  // Allowed file extensions for source files
  static const _allowedExtensions = {
    'rs',
    'go',
    'ts',
    'js',
    'c',
    'cpp',
    'cc',
    'cxx',
    'h',
    'hpp',
    'toml',
    'json',
    'yaml',
    'yml',
    'md',
    'txt',
    'mod',
    'sum',
    'py',
    'java',
    'kt',
  };

  @action
  void addSourceFile(String filename, String content) {
    // Validate file extension
    final extension = filename.split('.').last.toLowerCase();
    if (!_allowedExtensions.contains(extension)) {
      throw ArgumentError(
        'File type ".$extension" is not supported. Allowed types: ${_allowedExtensions.join(', ')}',
      );
    }

    sourceFiles[filename] = content;
    if (selectedFile == null) {
      selectedFile = filename;
    }
    clearValidation();
  }

  @action
  void updateSourceFile(String filename, String content) {
    sourceFiles[filename] = content;
    clearValidation();
  }

  @action
  void removeSourceFile(String filename) {
    sourceFiles.remove(filename);
    if (selectedFile == filename) {
      selectedFile = sourceFiles.keys.isNotEmpty
          ? sourceFiles.keys.first
          : null;
    }
    clearValidation();
  }

  @action
  void selectFile(String filename) {
    if (sourceFiles.containsKey(filename)) {
      selectedFile = filename;
    }
  }

  @action
  void clearSourceFiles() {
    sourceFiles.clear();
    selectedFile = null;
    clearValidation();
  }

  @action
  void setTriggerType(TriggerType value) {
    triggerType = value;
  }

  @action
  void setPriority(int value) {
    priority = value.clamp(0, 100);
  }

  @action
  void setEnabled(bool value) {
    enabled = value;
  }

  @action
  void setUseCustomMatchRules(bool value) {
    useCustomMatchRules = value;
  }

  @action
  void toggleMethod(String method) {
    if (selectedMethods.contains(method)) {
      selectedMethods.remove(method);
    } else {
      selectedMethods.add(method);
    }
  }

  @action
  void setPathPattern(String value) {
    pathPattern = value;
  }

  @action
  void setHostPattern(String value) {
    hostPattern = value;
  }

  @action
  void setPatternType(PatternType value) {
    patternType = value;
  }

  @action
  void setTimeoutMs(int? value) {
    timeoutMs = value;
  }

  @action
  void setMemoryLimitMB(int? value) {
    memoryLimitMB = value;
  }

  @action
  void setCurrentTab(int tab) {
    currentTab = tab;
  }

  @action
  void validateName() {
    if (name.trim().isEmpty) {
      nameError = 'Name is required';
    } else if (name.trim().length < 3) {
      nameError = 'Name must be at least 3 characters';
    } else {
      nameError = null;
    }
  }

  @action
  void validateCode() {
    if (code.trim().isEmpty) {
      codeError = 'Code is required';
    } else {
      codeError = null;
    }
  }

  @action
  void validateSourceFiles() {
    // Only validate sourceFiles for writeSource and importZip modes
    if (runtime == ScriptRuntime.extism &&
        (codeCreationMode == CodeCreationMode.writeSource ||
            codeCreationMode == CodeCreationMode.importZip)) {
      if (sourceFiles.isEmpty) {
        sourceFilesError = 'Add at least one source file to continue';
      } else {
        sourceFilesError = null;
      }
    } else {
      sourceFilesError = null;
    }
  }

  @action
  bool validate() {
    // Auto-generate name if empty
    if (name.trim().isEmpty) {
      final timestamp = DateTime.now().millisecondsSinceEpoch % 1000000;
      name = 'Untitled Script $timestamp';
    }

    validateName();

    // For writeSource/importZip modes, validate sourceFiles instead of code
    if (runtime == ScriptRuntime.extism &&
        (codeCreationMode == CodeCreationMode.writeSource ||
            codeCreationMode == CodeCreationMode.importZip)) {
      validateSourceFiles();
      codeError = null; // Don't require code field for multi-file mode
    } else {
      validateCode();
      sourceFilesError = null; // Don't require sourceFiles for other modes
    }

    return isValid;
  }

  @action
  Future<void> testScript(TestRequest testRequest) async {
    final testUseCase = _testUseCase;
    if (testUseCase == null) {
      testError = 'Test use case not configured';
      return;
    }

    try {
      isTesting = true;
      testError = null;
      testResult = null;

      final script = buildScript();
      testResult = await testUseCase(script, testRequest);
    } catch (e) {
      testError = 'Test failed: ${e.toString()}';
    } finally {
      isTesting = false;
    }
  }

  @action
  void clearError() {
    errorMessage = null;
  }

  @action
  void clearValidation() {
    isValidSyntax = false;
    syntaxValidationError = null;
  }

  @action
  Future<void> validateSyntax() async {
    final useCase = _validateSyntaxUseCase;
    if (useCase == null) return;
    try {
      isValidating = true;
      syntaxValidationError = null;
      isValidSyntax = false;

      final result = await useCase(
        sourceCode: code.isNotEmpty ? code : sourceFiles.values.join('\n'),
        language: language,
        dependencies: sourceFiles.isNotEmpty
            ? Map<String, String>.from(sourceFiles)
            : null,
      );

      isValidSyntax = result['valid'] == true;
      if (!isValidSyntax) {
        syntaxValidationError = result['error']?.toString();
      }
    } catch (e) {
      isValidSyntax = false;
      syntaxValidationError = e.toString();
    } finally {
      isValidating = false;
    }
  }

  @action
  Future<Map<String, dynamic>?> uploadProjectZip(List<int> zipBytes) async {
    final useCase = _uploadProjectUseCase;
    final scriptId = editingScriptId;
    if (useCase == null || scriptId == null) return null;
    try {
      isUploadingProject = true;
      final result = await useCase(scriptId, zipBytes);
      return result;
    } catch (e) {
      errorMessage = 'Failed to upload project: ${e.toString()}';
      return null;
    } finally {
      isUploadingProject = false;
    }
  }

  @action
  Future<List<Map<String, dynamic>>> loadProjectFiles() async {
    final useCase = _listProjectFilesUseCase;
    final scriptId = editingScriptId;
    if (useCase == null || scriptId == null) return [];
    try {
      final result = await useCase(scriptId);
      final files = result['files'];
      if (files is List) {
        return files.cast<Map<String, dynamic>>();
      }
      return [];
    } catch (e) {
      errorMessage = 'Failed to load project files: ${e.toString()}';
      return [];
    }
  }

  Future<List<int>> downloadProjectZip() async {
    final useCase = _downloadProjectUseCase;
    final scriptId = editingScriptId;
    if (useCase == null || scriptId == null) {
      throw Exception('Cannot download: script not saved yet');
    }
    return useCase(scriptId);
  }

  /// Build Script entity from form data
  Script buildScript() {
    // Using multi-file editor for all runtimes now
    // Store ALL files in dependencies map
    String? sourceCodeValue;
    Map<String, String>? dependenciesValue;

    if (sourceFiles.isNotEmpty) {
      dependenciesValue = Map<String, String>.from(sourceFiles);
      // sourceCode is null for multi-file projects
      // Backend will use Dependencies map for compilation
      sourceCodeValue = null;
    }

    return Script(
      id: editingScriptId ?? '',
      name: name,
      description: description.isEmpty ? null : description,
      runtime: runtime,
      code: code,
      language: language,
      triggerType: triggerType,
      priority: priority,
      enabled: enabled,
      matchRules: matchRules,
      config: config,
      createdAt: null,
      updatedAt: null,
      sourceCode: sourceCodeValue,
      dependencies: dependenciesValue,
      compilationStatus: compilationStatus,
      compilationError: compilationError,
      lastCompiledAt: lastCompiledAt,
      validationStatus: null,
      validationError: null,
    );
  }

  /// Get starter file name based on current language
  String _getStarterFileName() {
    switch (language) {
      case 'rust':
        return 'main.rs';
      case 'go':
        return 'main.go';
      case 'typescript':
        return 'index.ts';
      case 'javascript':
        return 'index.js';
      case 'python':
        return 'main.py';
      case 'c':
        return 'main.c';
      case 'cpp':
        return 'main.cpp';
      case 'java':
        return 'Main.java';
      default:
        return 'main.rs'; // Default to Rust
    }
  }

  /// Get starter code template based on current language
  String _getStarterCode() {
    switch (language) {
      case 'rust':
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

use extism_pdk::*;

#[plugin_fn]
pub fn process_request(input: String) -> FnResult<String> {
    // Process the HTTP request
    // input contains JSON with request details

    Ok(input)
}

#[plugin_fn]
pub fn process_response(input: String) -> FnResult<String> {
    // Process the HTTP response
    // input contains JSON with response details

    Ok(input)
}
''';

      case 'go':
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

package main

import (
\t"github.com/extism/go-pdk"
)

//export process_request
func processRequest() int32 {
\t// Process the HTTP request
\tinput := pdk.InputString()
\t
\tpdk.OutputString(input)
\treturn 0
}

//export process_response
func processResponse() int32 {
\t// Process the HTTP response
\tinput := pdk.InputString()
\t
\tpdk.OutputString(input)
\treturn 0
}

func main() {}
''';

      case 'typescript':
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

declare var Host: any;

export function process_request(): number {
    // Process the HTTP request
    const input = Host.inputString();

    Host.outputString(input);
    return 0;
}

export function process_response(): number {
    // Process the HTTP response
    const input = Host.inputString();

    Host.outputString(input);
    return 0;
}
''';

      case 'javascript':
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

function process_request() {
    // Process the HTTP request
    const input = Host.inputString();

    Host.outputString(input);
    return 0;
}

function process_response() {
    // Process the HTTP response
    const input = Host.inputString();

    Host.outputString(input);
    return 0;
}

module.exports = { process_request, process_response };
''';

      case 'python':
        return '''# Extism Plugin Example
#
# This plugin processes HTTP requests/responses
# Export functions that match your trigger type

from extism import host_fn

@host_fn
def process_request():
    # Process the HTTP request
    input_data = host_fn.input_string()

    host_fn.output_string(input_data)

@host_fn
def process_response():
    # Process the HTTP response
    input_data = host_fn.input_string()

    host_fn.output_string(input_data)
''';

      case 'c':
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

#include <extism-pdk.h>
#include <string.h>

int32_t process_request() {
    // Process the HTTP request
    uint64_t len = extism_input_length();
    uint8_t *input = extism_input_load_u8(len);

    extism_output_set(input, len);
    return 0;
}

int32_t process_response() {
    // Process the HTTP response
    uint64_t len = extism_input_length();
    uint8_t *input = extism_input_load_u8(len);

    extism_output_set(input, len);
    return 0;
}
''';

      case 'cpp':
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

#include <extism-pdk.hpp>
#include <string>

extern "C" {

int32_t process_request() {
    // Process the HTTP request
    std::string input = extism::input_string();

    extism::output_string(input);
    return 0;
}

int32_t process_response() {
    // Process the HTTP response
    std::string input = extism::input_string();

    extism::output_string(input);
    return 0;
}

}
''';

      default:
        // Fallback to Rust template for unknown languages
        return '''// Extism Plugin Example
//
// This plugin processes HTTP requests/responses
// Export functions that match your trigger type

use extism_pdk::*;

#[plugin_fn]
pub fn process_request(input: String) -> FnResult<String> {
    // Process the HTTP request
    // input contains JSON with request details

    Ok(input)
}

#[plugin_fn]
pub fn process_response(input: String) -> FnResult<String> {
    // Process the HTTP response
    // input contains JSON with response details

    Ok(input)
}
''';
    }
  }
}
