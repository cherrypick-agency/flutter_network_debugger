import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:archive/archive.dart';
import '../../application/stores/script_editor_store.dart';
import '../../application/stores/scripts_store.dart';
import '../../domain/entities/script.dart';
import 'package:file_picker/file_picker.dart';
import 'script_settings_form.dart';
import 'match_rules_form.dart';
import 'script_test_tab.dart';
import 'code_tab_content.dart';
import 'multi_file_editor_section.dart';
import '../../infrastructure/editor_di.dart';
import '../../data/services/scripts_api_service.dart';
import '../../../../core/di/di.dart';
import '../../../compiler_management/presentation/stores/compiler_list_store.dart';
import '../../domain/utils/zip_validator.dart';

// Intent classes for keyboard shortcuts
class SaveScriptIntent extends Intent {
  const SaveScriptIntent();
}

class TestScriptIntent extends Intent {
  const TestScriptIntent();
}

class CompileScriptIntent extends Intent {
  const CompileScriptIntent();
}

class CloseEditorIntent extends Intent {
  const CloseEditorIntent();
}

/// Full-screen script editor dialog with tabs
class ScriptEditorDialog extends StatefulWidget {
  final ScriptsStore scriptsStore;
  final Script? editingScript; // null for create, Script for edit

  const ScriptEditorDialog({
    super.key,
    required this.scriptsStore,
    this.editingScript,
  });

  @override
  State<ScriptEditorDialog> createState() => _ScriptEditorDialogState();

  static Future<void> show(
    BuildContext context,
    ScriptsStore scriptsStore, {
    Script? editingScript,
  }) async {
    return showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => ScriptEditorDialog(
        scriptsStore: scriptsStore,
        editingScript: editingScript,
      ),
    );
  }
}

class _ScriptEditorDialogState extends State<ScriptEditorDialog> {
  late final ScriptEditorStore _editorStore;
  late final CompilerListStore _compilerStore;
  EditorDI? _editorDI; // Multi-file editor DI (lazy init)
  Future<void>? _initEditorFuture; // fix FutureBuilder jitter
  bool _isSaving = false;
  bool _isDownloadingProject = false;
  bool _loadingCompilers = true;
  Map<String, bool>?
  _canCompileByLanguage; // compilation availability by language
  final FocusNode _hotkeysFocus = FocusNode(
    debugLabel: 'script_editor_hotkeys',
  );
  DateTime? _saveGuardUntil;
  DateTime? _compileGuardUntil;

  @override
  void initState() {
    super.initState();
    _editorStore = sl<ScriptEditorStore>();
    _compilerStore = sl<CompilerListStore>();

    // Load compilers asynchronously
    _loadCompilersAsync();

    if (widget.editingScript != null) {
      _editorStore.initForEdit(widget.editingScript!);
    } else {
      _editorStore.initForNewScript();
    }
    // Start lazy editor initialization once
    _initEditorFuture = _initEditorDI();
  }

  @override
  void dispose() {
    _editorDI?.dispose();
    _hotkeysFocus.dispose();
    super.dispose();
  }

  /// Load compilers list from backend
  Future<void> _loadCompilersAsync() async {
    // Load from two sources:
    // 1) cache statuses (manager page) via CompilerListStore
    // 2) compilation availability (from system or cache) via /_api/v1/scripts/compilers
    final futures = <Future<void>>[];
    if (_compilerStore.compilers.isEmpty) {
      futures.add(_compilerStore.loadCompilers());
    }
    futures.add(_loadCompilersAvailability());
    await Future.wait(futures);
    if (mounted) {
      setState(() => _loadingCompilers = false);
    }
  }

  Future<void> _loadCompilersAvailability() async {
    try {
      final api = sl<ScriptsApiService>();
      final map = await api.getCompilersAvailability();
      // normalize keys to lowercase for convenience
      _canCompileByLanguage = map.map((k, v) => MapEntry(k.toLowerCase(), v));
    } catch (_) {
      // silently ignore — don't block UI, just remain without hint
      _canCompileByLanguage = null;
    }
  }

  /// Initialize EditorDI lazily when needed for writeSource mode
  Future<void> _initEditorDI() async {
    if (_editorDI != null) return;

    _editorDI = EditorDI(scriptStore: _editorStore);
    await _editorDI!.init();

    // Auto-open selected file if exists
    if (_editorStore.selectedFile != null) {
      await _editorDI!.editorController.loadFile(_editorStore.selectedFile!);
    }

    if (mounted) setState(() {});
  }

  /// Check if compiler for the current language is installed
  /// Returns true optimistically while loading to avoid false negatives
  bool _isCompilerInstalled(String language) {
    if (_loadingCompilers) return true; // Optimistic check while loading
    final lang = language.toLowerCase();
    // If backend says compilation is available (system or cache) — ok.
    final canCompile = _canCompileByLanguage?[lang];
    if (canCompile == true) return true;
    // Otherwise — fallback to cache (as before).
    return _compilerStore.installedCompilers.any(
      (c) => c.language.toLowerCase() == lang,
    );
  }

  /// Check if there is compilable content available
  /// Used by both Compile button visibility and enabled state
  bool _hasCompilableContent() {
    return _editorStore.isEditing ||
        ((_editorStore.codeCreationMode == CodeCreationMode.writeSource ||
                _editorStore.codeCreationMode == CodeCreationMode.importZip) &&
            _editorStore.sourceFiles.isNotEmpty);
  }

  bool _shouldProceedSave() {
    final now = DateTime.now();
    if (_saveGuardUntil != null && now.isBefore(_saveGuardUntil!)) {
      return false;
    }
    _saveGuardUntil = now.add(const Duration(milliseconds: 300));
    return true;
  }

  bool _shouldProceedCompile() {
    final now = DateTime.now();
    if (_compileGuardUntil != null && now.isBefore(_compileGuardUntil!)) {
      return false;
    }
    _compileGuardUntil = now.add(const Duration(milliseconds: 300));
    return true;
  }

  @override
  Widget build(BuildContext context) {
    return Shortcuts(
      shortcuts: <LogicalKeySet, Intent>{
        // Ctrl+S / Cmd+S - Save
        LogicalKeySet(LogicalKeyboardKey.control, LogicalKeyboardKey.keyS):
            const SaveScriptIntent(),
        LogicalKeySet(LogicalKeyboardKey.meta, LogicalKeyboardKey.keyS):
            const SaveScriptIntent(),

        // Ctrl+T / Cmd+T - Test (switch to Test tab)
        LogicalKeySet(LogicalKeyboardKey.control, LogicalKeyboardKey.keyT):
            const TestScriptIntent(),
        LogicalKeySet(LogicalKeyboardKey.meta, LogicalKeyboardKey.keyT):
            const TestScriptIntent(),

        // Ctrl+Shift+C / Cmd+Shift+C - Compile
        LogicalKeySet(
          LogicalKeyboardKey.control,
          LogicalKeyboardKey.shift,
          LogicalKeyboardKey.keyC,
        ): const CompileScriptIntent(),
        LogicalKeySet(
          LogicalKeyboardKey.meta,
          LogicalKeyboardKey.shift,
          LogicalKeyboardKey.keyC,
        ): const CompileScriptIntent(),

        // Esc - Close dialog
        LogicalKeySet(LogicalKeyboardKey.escape): const CloseEditorIntent(),
      },
      child: Actions(
        actions: <Type, Action<Intent>>{
          SaveScriptIntent: CallbackAction<SaveScriptIntent>(
            onInvoke: (_) {
              if (_shouldProceedSave() && !_isSaving && _editorStore.isValid) {
                _save();
              }
              return null;
            },
          ),
          TestScriptIntent: CallbackAction<TestScriptIntent>(
            onInvoke: (_) {
              // Switch to Test tab (tab index 3)
              _editorStore.setCurrentTab(3);
              return null;
            },
          ),
          CompileScriptIntent: CallbackAction<CompileScriptIntent>(
            onInvoke: (_) {
              // Compile if Extism script and compilable content is present
              if (_shouldProceedCompile() &&
                  _editorStore.runtime == ScriptRuntime.extism &&
                  !widget.scriptsStore.isCompiling &&
                  !_isSaving &&
                  _hasCompilableContent()) {
                _compileScript();
              }
              return null;
            },
          ),
          CloseEditorIntent: CallbackAction<CloseEditorIntent>(
            onInvoke: (_) {
              _confirmClose(context);
              return null;
            },
          ),
        },
        child: RawKeyboardListener(
          focusNode: _hotkeysFocus,
          autofocus: true,
          onKey: (event) {
            // Fallback for macOS: intercept ⌘S and ⌘⇧C at platform level,
            // if they didn't reach Shortcuts due to editor platform view.
            if (event is RawKeyDownEvent) {
              final isMeta = event.isMetaPressed;
              final isCtrl = event.isControlPressed;
              final key = event.logicalKey;
              // Save: Cmd+S or Ctrl+S
              if ((isMeta || isCtrl) && key == LogicalKeyboardKey.keyS) {
                if (_shouldProceedSave() &&
                    !_isSaving &&
                    _editorStore.isValid) {
                  _save();
                }
                // prevent duplicate processing
                return;
              }
              // Compile: Cmd+Shift+C or Ctrl+Shift+C
              if ((isMeta || isCtrl) &&
                  event.isShiftPressed &&
                  key == LogicalKeyboardKey.keyC) {
                if (_shouldProceedCompile() &&
                    _editorStore.runtime == ScriptRuntime.extism &&
                    !widget.scriptsStore.isCompiling &&
                    !_isSaving &&
                    _hasCompilableContent()) {
                  _compileScript();
                }
                return;
              }
            }
          },
          child: Dialog.fullscreen(
            child: DefaultTabController(
              length: 4,
              child: Scaffold(
                appBar: AppBar(
                  title: Observer(builder: (_) => Text(_editorStore.title)),
                  leading: IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => _confirmClose(context),
                  ),
                  actions: [
                    // Compile button - show for Extism scripts in writeSource mode or when editing
                    Observer(
                      builder: (_) {
                        // Show compile button for Extism runtime when compilable content is present
                        final shouldShowCompile =
                            _editorStore.runtime == ScriptRuntime.extism &&
                            _hasCompilableContent();

                        if (!shouldShowCompile) {
                          return const SizedBox.shrink();
                        }

                        final isCompiling = widget.scriptsStore.isCompiling;
                        final canCompile =
                            !isCompiling &&
                            !_isSaving &&
                            _hasCompilableContent();

                        // Dynamic tooltip message based on state
                        final tooltipMessage = !canCompile
                            ? (_isSaving
                                  ? 'Saving script...'
                                  : isCompiling
                                  ? 'Compiling...'
                                  : 'No source files to compile')
                            : 'Compile (Ctrl+Shift+C)';

                        return Tooltip(
                          message: tooltipMessage,
                          child: ElevatedButton.icon(
                            onPressed: canCompile ? _compileScript : null,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: canCompile
                                  ? Colors.green
                                  : Colors.grey,
                              foregroundColor: Colors.white,
                            ),
                            icon: isCompiling
                                ? const SizedBox(
                                    width: 16,
                                    height: 16,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                      color: Colors.white,
                                    ),
                                  )
                                : const Icon(Icons.build),
                            label: Text(
                              isCompiling ? 'Compiling...' : 'Compile',
                            ),
                          ),
                        );
                      },
                    ),
                    const SizedBox(width: 8),
                    Observer(
                      builder: (_) => Tooltip(
                        message:
                            Theme.of(context).platform == TargetPlatform.macOS
                            ? 'Save (⌘S)'
                            : 'Save (Ctrl+S)',
                        child: TextButton.icon(
                          onPressed: _isSaving || !_editorStore.isValid
                              ? null
                              : _save,
                          icon: _isSaving
                              ? const SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Icon(Icons.save),
                          label: Text(_isSaving ? 'Saving...' : 'Save'),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                  ],
                  bottom: PreferredSize(
                    preferredSize: const Size.fromHeight(48),
                    child: TabBar(
                      isScrollable: true,
                      tabs: const [
                        Tab(text: 'Code', icon: Icon(Icons.code, size: 20)),
                        Tab(
                          text: 'Settings',
                          icon: Icon(Icons.settings, size: 20),
                        ),
                        Tab(
                          text: 'Match Rules',
                          icon: Icon(Icons.filter_alt, size: 20),
                        ),
                        Tab(text: 'Test', icon: Icon(Icons.science, size: 20)),
                      ],
                      onTap: _editorStore.setCurrentTab,
                    ),
                  ),
                ),
                body: Observer(
                  builder: (_) {
                    if (_editorStore.errorMessage != null) {
                      return Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(
                              Icons.error_outline,
                              size: 64,
                              color: Colors.red,
                            ),
                            const SizedBox(height: 16),
                            Text(
                              _editorStore.errorMessage!,
                              style: const TextStyle(color: Colors.red),
                            ),
                          ],
                        ),
                      );
                    }

                    return TabBarView(
                      physics: const NeverScrollableScrollPhysics(),
                      children: [
                        CodeTabContent(
                          editorStore: _editorStore,
                          scriptsStore: widget.scriptsStore,
                          isDownloadingProject: _isDownloadingProject,
                          onDownloadProject: _downloadProject,
                          buildMultiFileEditor: () => MultiFileEditorSection(
                            initEditorFuture: _initEditorFuture,
                            editorDI: _editorDI,
                          ),
                          onImportZip: _importProjectFromZip,
                        ),
                        ScriptSettingsForm(store: _editorStore),
                        MatchRulesForm(store: _editorStore),
                        ScriptTestTab(store: _editorStore),
                      ],
                    );
                  },
                ), // body Observer
              ), // Scaffold
            ), // DefaultTabController
          ), // Dialog.fullscreen
        ), // RawKeyboardListener
      ), // Actions
    ); // Shortcuts
  }

  Future<void> _importProjectFromZip() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['zip'],
        withData: true,
      );

      if (result != null && result.files.isNotEmpty) {
        final file = result.files.first;
        if (file.bytes == null) {
          throw Exception('Failed to read file data');
        }

        final bytes = file.bytes!;

        ZipValidator.validate(bytes);

        // Decode ZIP archive
        final archive = ZipDecoder().decodeBytes(bytes);

        // Allowed file extensions
        const allowedExtensions = {
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
        };

        int importedCount = 0;

        // Extract files
        for (final fileEntry in archive) {
          // Skip directories
          if (fileEntry.isFile) {
            final filename = fileEntry.name;
            final extension = filename.split('.').last.toLowerCase();

            // Only import allowed file types
            if (allowedExtensions.contains(extension)) {
              try {
                // Decode file content as UTF-8
                final content = utf8.decode(fileEntry.content as List<int>);
                _editorStore.addSourceFile(filename, content);
                importedCount++;
              } catch (e) {
                // Skip binary or non-UTF8 files
                continue;
              }
            }
          }
        }

        if (importedCount == 0) {
          throw Exception('No valid source files found in ZIP archive');
        }

        // Upload ZIP to backend if script is already saved
        if (_editorStore.isEditing) {
          await _editorStore.uploadProjectZip(bytes);
        }

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                'Imported $importedCount file${importedCount > 1 ? 's' : ''} from ZIP',
              ),
              backgroundColor: Colors.green,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Import failed: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _save() async {
    if (!_editorStore.validate()) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please fix validation errors'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    // Before saving, forcefully flush unsaved editor changes
    await _editorDI?.flushUnsavedChanges();

    setState(() => _isSaving = true);

    try {
      final script = _editorStore.buildScript();

      if (_editorStore.isEditing) {
        await widget.scriptsStore.updateScript(
          _editorStore.editingScriptId!,
          script,
        );
      } else {
        final createdScript = await widget.scriptsStore.createScript(script);
        // Stay in editor and switch to edit mode for the new script
        _editorStore.initForEdit(createdScript);
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              _editorStore.isEditing
                  ? 'Script updated successfully'
                  : 'Script created successfully',
            ),
            backgroundColor: Colors.green,
          ),
        );
        // Reset unsaved changes indicator by reloading active file
        await _editorDI?.reloadCurrentFileIfAny(_editorStore.selectedFile);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isSaving = false);
      }
    }
  }

  Future<void> _compileScript() async {
    // Guard: prevent double-click and race conditions
    if (_isSaving || widget.scriptsStore.isCompiling || _isDownloadingProject) {
      return;
    }

    // Check if compiler is installed
    if (!_isCompilerInstalled(_editorStore.language)) {
      final shouldNavigate = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Compiler not installed'),
          content: Text(
            'The compiler for ${_editorStore.language} is not installed. '
            'Please install it from the Compiler Management page to compile your script.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            FilledButton.icon(
              onPressed: () => Navigator.of(context).pop(true),
              icon: const Icon(Icons.extension),
              label: const Text('Go to Compilers'),
            ),
          ],
        ),
      );

      if (shouldNavigate == true && mounted) {
        Navigator.of(context).pushNamed('/compilers');
      }
      return;
    }

    // Always validate and save before compiling
    if (!_editorStore.validate()) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Please fix validation errors before compiling'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    // Additional validation for compilation: ensure source files exist for importZip mode
    if (_editorStore.codeCreationMode == CodeCreationMode.importZip &&
        _editorStore.sourceFiles.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Please import project files before compiling'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    // Save the script first (create new or update existing)
    setState(() => _isSaving = true);
    try {
      // Before saving, forcefully flush unsaved editor changes
      await _editorDI?.flushUnsavedChanges();

      final script = _editorStore.buildScript();

      if (_editorStore.isEditing) {
        // Update existing script
        await widget.scriptsStore.updateScript(
          _editorStore.editingScriptId!,
          script,
        );
      } else {
        // Create new script
        final createdScript = await widget.scriptsStore.createScript(script);
        _editorStore.initForEdit(createdScript);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to save script: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
      setState(() => _isSaving = false);
      return;
    }
    setState(() => _isSaving = false);

    try {
      final compiledScript = await widget.scriptsStore.compileScript(
        _editorStore.editingScriptId!,
        optimize: true,
      );

      // Update editor with compiled script
      _editorStore.initForEdit(compiledScript);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Script compiled successfully!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Compilation failed: ${e.toString()}'),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 5),
          ),
        );
      }
    }
  }

  Future<void> _downloadProject() async {
    if (_isDownloadingProject) return;

    setState(() => _isDownloadingProject = true);

    try {
      final bytes = await _editorStore.downloadProjectZip();

      final fileName =
          '${_editorStore.name.replaceAll(RegExp(r'[^\w\s-]'), '_')}_project.zip';

      final result = await FilePicker.platform.saveFile(
        dialogTitle: 'Save Project',
        fileName: fileName,
        bytes: Uint8List.fromList(bytes),
      );

      if (result != null && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Project downloaded successfully'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Download failed: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isDownloadingProject = false);
      }
    }
  }

  Future<void> _confirmClose(BuildContext context) async {
    // Check if there are unsaved changes
    if (_editorStore.code.isNotEmpty || _editorStore.name.isNotEmpty) {
      final confirm = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Discard changes?'),
          content: const Text(
            'You have unsaved changes. Are you sure you want to close?',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () => Navigator.of(context).pop(true),
              child: const Text('Discard'),
            ),
          ],
        ),
      );

      if (confirm == true && mounted) {
        Navigator.of(context).pop();
      }
    } else {
      Navigator.of(context).pop();
    }
  }
}
