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
import 'code_mode_selector.dart';
import 'multi_file_editor_widget.dart';
import 'wasm_upload_zone.dart';
import '../../infrastructure/editor_di.dart';
import '../../../../core/di/di.dart';
import '../../../compiler_management/presentation/stores/compiler_list_store.dart';

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
  bool _isSaving = false;
  bool _loadingCompilers = true;

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
  }

  @override
  void dispose() {
    _editorDI?.dispose();
    super.dispose();
  }

  /// Load compilers list from backend
  Future<void> _loadCompilersAsync() async {
    if (_compilerStore.compilers.isEmpty) {
      await _compilerStore.loadCompilers();
    }
    if (mounted) {
      setState(() => _loadingCompilers = false);
    }
  }

  /// Initialize EditorDI lazily when needed for writeSource mode
  Future<void> _initEditorDI() async {
    if (_editorDI != null) return;

    _editorDI = EditorDI(scriptStore: _editorStore);
    await _editorDI!.init();

    if (mounted) setState(() {});
  }

  /// Check if compiler for the current language is installed
  /// Returns true optimistically while loading to avoid false negatives
  bool _isCompilerInstalled(String language) {
    if (_loadingCompilers) return true; // Optimistic check while loading
    return _compilerStore.installedCompilers.any(
      (c) => c.language.toLowerCase() == language.toLowerCase(),
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
              if (!_isSaving && _editorStore.isValid) {
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
              if (_editorStore.runtime == ScriptRuntime.extism &&
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
                          !isCompiling && !_isSaving && _hasCompilableContent();

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
                          label: Text(isCompiling ? 'Compiling...' : 'Compile'),
                        ),
                      );
                    },
                  ),
                  const SizedBox(width: 8),
                  Observer(
                    builder: (_) => Tooltip(
                      message: 'Save (Ctrl+S)',
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
                      _buildCodeTab(),
                      _buildSettingsTab(),
                      _buildMatchRulesTab(),
                      _buildTestTab(),
                    ],
                  );
                },
              ), // body Observer
            ), // Scaffold
          ), // DefaultTabController
        ), // Dialog.fullscreen
      ), // Actions
    ); // Shortcuts
  }

  Widget _buildCodeTab() {
    return Observer(
      builder: (_) {
        return Column(
          children: [
            // Compilation status banner for Extism runtime
            if (_editorStore.runtime == ScriptRuntime.extism &&
                _editorStore.isEditing)
              _buildCompilationStatusBanner(),
            // Mode selector for Extism runtime
            if (_editorStore.runtime == ScriptRuntime.extism)
              CodeModeSelector(
                selectedMode: _editorStore.codeCreationMode,
                onModeChanged: _editorStore.setCodeCreationMode,
              ),
            // Toolbar for Extism runtime (not in uploadWasm mode)
            if (_editorStore.runtime == ScriptRuntime.extism &&
                _editorStore.codeCreationMode != CodeCreationMode.uploadWasm)
              _buildExtismToolbar(),
            // Content based on mode
            Expanded(
              child:
                  _editorStore.runtime == ScriptRuntime.extism &&
                      _editorStore.codeCreationMode ==
                          CodeCreationMode.uploadWasm
                  ? _buildUploadWasmMode()
                  : _buildMultiFileEditor(),
            ),
          ],
        );
      },
    );
  }

  Widget _buildExtismToolbar() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Theme.of(
          context,
        ).colorScheme.surfaceContainerHighest.withOpacity(0.2),
        border: Border(
          bottom: BorderSide(color: Theme.of(context).dividerColor, width: 1),
        ),
      ),
      child: Row(
        children: [
          // Import ZIP button (only in importZip mode)
          if (_editorStore.codeCreationMode == CodeCreationMode.importZip)
            OutlinedButton.icon(
              onPressed: _importProjectFromZip,
              icon: const Icon(Icons.folder_zip, size: 18),
              label: const Text('Import ZIP'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildMultiFileEditor() {
    return FutureBuilder<void>(
      future: _initEditorDI(),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting ||
            _editorDI == null) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: Text('Failed to initialize editor: ${snapshot.error}'),
          );
        }

        return MultiFileEditorWidget(editorDI: _editorDI!);
      },
    );
  }

  Widget _buildUploadWasmMode() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: WasmUploadZone(
        onWasmUploaded: (base64) {
          _editorStore.setCode(base64);
        },
        onClear: () {
          _editorStore.setCode('');
        },
      ),
    );
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

        // Validate file size (max 10MB)
        const maxSizeBytes = 10 * 1024 * 1024; // 10MB
        if (bytes.length > maxSizeBytes) {
          throw Exception('File is too large. Maximum size is 10MB.');
        }

        // Validate ZIP magic bytes (0x50 0x4B = "PK")
        if (bytes.length < 4 || bytes[0] != 0x50 || bytes[1] != 0x4B) {
          throw Exception(
            'Invalid ZIP file. Please select a valid ZIP archive.',
          );
        }

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

  Widget _buildSettingsTab() {
    return ScriptSettingsForm(store: _editorStore);
  }

  Widget _buildMatchRulesTab() {
    return MatchRulesForm(store: _editorStore);
  }

  Widget _buildTestTab() {
    return ScriptTestTab(store: _editorStore);
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

    setState(() => _isSaving = true);

    try {
      final script = _editorStore.buildScript();

      if (_editorStore.isEditing) {
        await widget.scriptsStore.updateScript(
          _editorStore.editingScriptId!,
          script,
        );
      } else {
        await widget.scriptsStore.createScript(script);
      }

      if (mounted) {
        Navigator.of(context).pop();
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

  Widget _buildCompilationStatusBanner() {
    return Observer(
      builder: (_) {
        // Get script from store to check compilation status
        final script = widget.scriptsStore.scripts.cast<Script?>().firstWhere(
          (s) => s?.id == _editorStore.editingScriptId,
          orElse: () => null,
        );

        if (script == null) return const SizedBox.shrink();

        // Show compilation error if any
        if (script.compilationError != null &&
            script.compilationError!.isNotEmpty) {
          return Container(
            margin: const EdgeInsets.all(16),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.red.shade50,
              border: Border.all(color: Colors.red.shade200),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(Icons.error_outline, color: Colors.red.shade700),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Compilation Failed',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.red.shade700,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        script.compilationError!,
                        style: TextStyle(
                          color: Colors.red.shade900,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        }

        // Show compilation success status
        if (script.compilationStatus == 'success' &&
            script.lastCompiledAt != null) {
          return Container(
            margin: const EdgeInsets.all(16),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.green.shade50,
              border: Border.all(color: Colors.green.shade200),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(Icons.check_circle_outline, color: Colors.green.shade700),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Compiled Successfully',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.green.shade700,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Last compiled: ${_formatTimestamp(script.lastCompiledAt!)}',
                        style: TextStyle(
                          color: Colors.green.shade900,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        }

        // Show pending status if sourceCode exists but not compiled yet
        if (script.sourceCode != null && script.sourceCode!.isNotEmpty) {
          return Container(
            margin: const EdgeInsets.all(16),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.orange.shade50,
              border: Border.all(color: Colors.orange.shade200),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.warning_amber_outlined,
                  color: Colors.orange.shade700,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    '${_editorStore.codeCreationMode == CodeCreationMode.importZip ? "Project files" : "Source code"} detected - Click "Compile" to generate WASM',
                    style: TextStyle(
                      color: Colors.orange.shade900,
                      fontSize: 13,
                    ),
                  ),
                ),
              ],
            ),
          );
        }

        return const SizedBox.shrink();
      },
    );
  }

  String _formatTimestamp(DateTime timestamp) {
    final now = DateTime.now();
    final difference = now.difference(timestamp);

    if (difference.inMinutes < 1) {
      return 'just now';
    } else if (difference.inHours < 1) {
      return '${difference.inMinutes} minute${difference.inMinutes == 1 ? '' : 's'} ago';
    } else if (difference.inDays < 1) {
      return '${difference.inHours} hour${difference.inHours == 1 ? '' : 's'} ago';
    } else {
      return '${timestamp.day}/${timestamp.month}/${timestamp.year} ${timestamp.hour}:${timestamp.minute.toString().padLeft(2, '0')}';
    }
  }

  Future<void> _compileScript() async {
    // Guard: prevent double-click and race conditions
    if (_isSaving || widget.scriptsStore.isCompiling) {
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
