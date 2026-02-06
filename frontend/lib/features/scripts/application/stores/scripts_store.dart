import 'package:mobx/mobx.dart';
import '../../domain/entities/script.dart';
import '../../domain/usecases/get_scripts_usecase.dart';
import '../../domain/usecases/create_script_usecase.dart';
import '../../domain/usecases/update_script_usecase.dart';
import '../../domain/usecases/delete_script_usecase.dart';
import '../../domain/usecases/toggle_script_usecase.dart';
import '../../domain/usecases/compile_script_usecase.dart';
import '../../domain/usecases/export_script_usecase.dart';
import '../../domain/usecases/import_script_usecase.dart';
import '../../domain/usecases/download_project_usecase.dart';

part 'scripts_store.g.dart';

/// Main store for Scripts feature
/// Manages list of scripts and CRUD operations
class ScriptsStore = _ScriptsStore with _$ScriptsStore;

abstract class _ScriptsStore with Store {
  final GetScriptsUseCase _getScriptsUseCase;
  final CreateScriptUseCase _createUseCase;
  final UpdateScriptUseCase _updateUseCase;
  final DeleteScriptUseCase _deleteUseCase;
  final ToggleScriptUseCase _toggleUseCase;
  final CompileScriptUseCase _compileUseCase;
  final ExportScriptUseCase _exportUseCase;
  final ImportScriptUseCase _importUseCase;
  final DownloadProjectUseCase? _downloadProjectUseCase;

  _ScriptsStore({
    required GetScriptsUseCase getScriptsUseCase,
    required CreateScriptUseCase createUseCase,
    required UpdateScriptUseCase updateUseCase,
    required DeleteScriptUseCase deleteUseCase,
    required ToggleScriptUseCase toggleUseCase,
    required CompileScriptUseCase compileUseCase,
    required ExportScriptUseCase exportUseCase,
    required ImportScriptUseCase importUseCase,
    DownloadProjectUseCase? downloadProjectUseCase,
  }) : _getScriptsUseCase = getScriptsUseCase,
       _createUseCase = createUseCase,
       _updateUseCase = updateUseCase,
       _deleteUseCase = deleteUseCase,
       _toggleUseCase = toggleUseCase,
       _compileUseCase = compileUseCase,
       _exportUseCase = exportUseCase,
       _importUseCase = importUseCase,
       _downloadProjectUseCase = downloadProjectUseCase;

  // Observable state
  @observable
  ObservableList<Script> scripts = ObservableList<Script>();

  @observable
  Script? selectedScript;

  @observable
  bool isLoading = false;

  @observable
  bool isImporting = false;

  @observable
  bool isExporting = false;

  @observable
  String? exportingScriptId; // Track which script is being exported

  @observable
  String? downloadingProjectScriptId; // Track which script project is being downloaded

  @observable
  String? errorMessage;

  @observable
  String searchQuery = '';

  @observable
  ScriptRuntime? runtimeFilter;

  @observable
  TriggerType? triggerTypeFilter;

  @observable
  bool? enabledFilter;

  @observable
  bool isCompiling = false;

  @observable
  String? compilationError;

  // Computed values
  @computed
  List<Script> get enabledScripts => scripts.where((s) => s.enabled).toList();

  @computed
  List<Script> get disabledScripts => scripts.where((s) => !s.enabled).toList();

  @computed
  List<Script> get filteredScripts {
    var filtered = scripts.toList();

    // Search filter
    if (searchQuery.isNotEmpty) {
      final query = searchQuery.toLowerCase();
      filtered = filtered
          .where(
            (s) =>
                s.name.toLowerCase().contains(query) ||
                (s.description?.toLowerCase().contains(query) ?? false),
          )
          .toList();
    }

    // Runtime filter
    if (runtimeFilter != null) {
      filtered = filtered.where((s) => s.runtime == runtimeFilter).toList();
    }

    // Trigger type filter
    if (triggerTypeFilter != null) {
      filtered = filtered
          .where((s) => s.triggerType == triggerTypeFilter)
          .toList();
    }

    // Enabled filter
    if (enabledFilter != null) {
      filtered = filtered.where((s) => s.enabled == enabledFilter).toList();
    }

    // Sort by priority (higher first), then by name
    filtered.sort((a, b) {
      final priorityCompare = b.priority.compareTo(a.priority);
      if (priorityCompare != 0) return priorityCompare;
      return a.name.compareTo(b.name);
    });

    return filtered;
  }

  // Actions
  @action
  Future<void> loadScripts() async {
    try {
      isLoading = true;
      errorMessage = null;

      final loadedScripts = await _getScriptsUseCase();
      scripts.clear();
      scripts.addAll(loadedScripts);
    } catch (e) {
      errorMessage = 'Failed to load scripts: ${e.toString()}';
    } finally {
      isLoading = false;
    }
  }

  @action
  Future<Script> createScript(Script script) async {
    try {
      isLoading = true;
      errorMessage = null;

      final createdScript = await _createUseCase(script);
      scripts.add(createdScript);
      return createdScript;
    } catch (e) {
      errorMessage = 'Failed to create script: ${e.toString()}';
      rethrow;
    } finally {
      isLoading = false;
    }
  }

  @action
  Future<Script> updateScript(String id, Script script) async {
    try {
      isLoading = true;
      errorMessage = null;

      final updatedScript = await _updateUseCase(id, script);

      final index = scripts.indexWhere((s) => s.id == id);
      if (index != -1) {
        scripts[index] = updatedScript;
      }

      if (selectedScript?.id == id) {
        selectedScript = updatedScript;
      }

      return updatedScript;
    } catch (e) {
      errorMessage = 'Failed to update script: ${e.toString()}';
      rethrow;
    } finally {
      isLoading = false;
    }
  }

  @action
  Future<void> deleteScript(String id) async {
    final removedIndex = scripts.indexWhere((s) => s.id == id);
    if (removedIndex == -1) return;
    final removedScript = scripts[removedIndex];

    // Optimistic: remove immediately
    scripts.removeAt(removedIndex);
    if (selectedScript?.id == id) {
      selectedScript = null;
    }

    try {
      errorMessage = null;
      await _deleteUseCase(id);
    } catch (e) {
      // Rollback on error
      scripts.insert(removedIndex, removedScript);
      errorMessage = 'Failed to delete script: ${e.toString()}';
      rethrow;
    }
  }

  @action
  Future<void> toggleScript(String id, bool enabled) async {
    try {
      final responseData = await _toggleUseCase(id, enabled);

      // Try to parse full script from response, fallback to optimistic update
      Script? updatedScript;
      if (responseData.containsKey('name')) {
        try {
          updatedScript = Script.fromJson(responseData);
        } catch (_) {
          // Fallback to optimistic update
        }
      }

      final index = scripts.indexWhere((s) => s.id == id);
      if (index != -1) {
        scripts[index] = updatedScript ?? scripts[index].copyWith(enabled: enabled);
      }

      if (selectedScript?.id == id) {
        selectedScript = updatedScript ?? selectedScript!.copyWith(enabled: enabled);
      }
    } catch (e) {
      errorMessage = 'Failed to toggle script: ${e.toString()}';
      rethrow;
    }
  }

  @action
  void selectScript(Script script) {
    selectedScript = script;
  }

  @action
  void clearSelection() {
    selectedScript = null;
  }

  @action
  void setSearchQuery(String query) {
    searchQuery = query;
  }

  @action
  void setRuntimeFilter(ScriptRuntime? runtime) {
    runtimeFilter = runtime;
  }

  @action
  void setTriggerTypeFilter(TriggerType? triggerType) {
    triggerTypeFilter = triggerType;
  }

  @action
  void setEnabledFilter(bool? enabled) {
    enabledFilter = enabled;
  }

  @action
  void clearFilters() {
    searchQuery = '';
    runtimeFilter = null;
    triggerTypeFilter = null;
    enabledFilter = null;
  }

  @action
  void clearError() {
    errorMessage = null;
  }

  @action
  Future<Script> compileScript(String id, {bool optimize = true}) async {
    try {
      isCompiling = true;
      compilationError = null;

      final compiledScript = await _compileUseCase(id, optimize: optimize);

      // Update script in list
      final index = scripts.indexWhere((s) => s.id == id);
      if (index != -1) {
        scripts[index] = compiledScript;
      }

      // Update selected script if it's the one being compiled
      if (selectedScript?.id == id) {
        selectedScript = compiledScript;
      }

      return compiledScript;
    } catch (e) {
      compilationError = 'Failed to compile script: ${e.toString()}';
      rethrow;
    } finally {
      isCompiling = false;
    }
  }

  @action
  void clearCompilationError() {
    compilationError = null;
  }

  /// Get export ZIP URL for downloading
  @action
  String getExportZipUrl(String id) {
    return _exportUseCase(id);
  }

  String? getDownloadProjectUrl(String id) {
    return _downloadProjectUseCase?.call(id);
  }

  /// Import script from ZIP file bytes
  @action
  Future<Script> importFromZip(List<int> zipBytes) async {
    try {
      isImporting = true;
      isLoading = true;
      errorMessage = null;

      final importedScript = await _importUseCase(zipBytes);
      scripts.add(importedScript);

      return importedScript;
    } catch (e) {
      errorMessage = 'Failed to import script: ${e.toString()}';
      rethrow;
    } finally {
      isImporting = false;
      isLoading = false;
    }
  }
}
