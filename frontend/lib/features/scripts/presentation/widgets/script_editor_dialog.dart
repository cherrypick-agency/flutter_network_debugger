import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:archive/archive.dart';
import 'package:dio/dio.dart';
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
import '../../data/services/scripts_api_service.dart';
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
        ), // RawKeyboardListener
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
            // Validation status banner
            _buildValidationStatusBanner(),
            // Mode selector for Extism runtime
            if (_editorStore.runtime == ScriptRuntime.extism)
              CodeModeSelector(
                selectedMode: _editorStore.codeCreationMode,
                onModeChanged: _editorStore.setCodeCreationMode,
              ),
            // Toolbar for Extism runtime (NOT in uploadWasm mode)
            if (_editorStore.runtime == ScriptRuntime.extism &&
                _editorStore.codeCreationMode != CodeCreationMode.uploadWasm)
              _buildExtismToolbar(),
            // Content based on mode
            Expanded(child: _buildModeContent()),
          ],
        );
      },
    );
  }

  Widget _buildValidationStatusBanner() {
    return Observer(
      builder: (_) {
        if (_editorStore.isValidSyntax) {
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
                  child: Text(
                    'Syntax is valid',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Colors.green.shade700,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close, size: 18),
                  onPressed: _editorStore.clearValidation,
                  tooltip: 'Dismiss',
                ),
              ],
            ),
          );
        }

        if (_editorStore.syntaxValidationError != null) {
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
                        'Validation Error',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.red.shade700,
                        ),
                      ),
                      const SizedBox(height: 4),
                      SelectableText(
                        _editorStore.syntaxValidationError!,
                        style: TextStyle(
                          color: Colors.red.shade900,
                          fontSize: 13,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close, size: 18),
                  onPressed: _editorStore.clearValidation,
                  tooltip: 'Dismiss',
                ),
              ],
            ),
          );
        }

        return const SizedBox.shrink();
      },
    );
  }

  Widget _buildExtismToolbar() {
    return Observer(
      builder: (_) {
        final hasSourceFiles = _editorStore.sourceFiles.isNotEmpty;
        final isImportZipWithFiles =
            _editorStore.codeCreationMode == CodeCreationMode.importZip &&
            hasSourceFiles;
        final isEditing = _editorStore.isEditing;
        final isValidating = _editorStore.isValidating;

        // Build list of toolbar buttons
        final leftButtons = <Widget>[];

        // Validate button (if source files exist)
        if (hasSourceFiles) {
          leftButtons.add(
            OutlinedButton.icon(
              onPressed: isValidating ? null : _editorStore.validateSyntax,
              icon: isValidating
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.fact_check_outlined, size: 18),
              label: Text(isValidating ? 'Validating...' : 'Validate'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
              ),
            ),
          );
        }

        // Re-import ZIP button
        if (isImportZipWithFiles) {
          if (leftButtons.isNotEmpty) leftButtons.add(const SizedBox(width: 8));
          leftButtons.add(
            OutlinedButton.icon(
              onPressed: () {
                _editorStore.clearSourceFiles();
              },
              icon: const Icon(Icons.refresh, size: 18),
              label: const Text('Re-import ZIP'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
              ),
            ),
          );
        }

        // Download Project button (right-aligned, only for saved scripts)
        Widget? rightButton;
        if (isEditing) {
          rightButton = OutlinedButton.icon(
            onPressed: _isDownloadingProject ? null : _downloadProject,
            icon: _isDownloadingProject
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.download, size: 18),
            label: Text(
              _isDownloadingProject ? 'Downloading...' : 'Download Project',
            ),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 8,
              ),
            ),
          );
        }

        // If no buttons at all — don't render the toolbar
        if (leftButtons.isEmpty && rightButton == null) {
          return const SizedBox.shrink();
        }

        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: Theme.of(
              context,
            ).colorScheme.surfaceContainerHighest.withOpacity(0.2),
            border: Border(
              bottom: BorderSide(
                color: Theme.of(context).dividerColor,
                width: 1,
              ),
            ),
          ),
          child: Row(
            children: [
              ...leftButtons,
              const Spacer(),
              if (rightButton != null) rightButton,
            ],
          ),
        );
      },
    );
  }

  Widget _buildModeContent() {
    // Upload WASM mode
    if (_editorStore.runtime == ScriptRuntime.extism &&
        _editorStore.codeCreationMode == CodeCreationMode.uploadWasm) {
      return _buildUploadWasmMode();
    }

    // Import ZIP mode - empty files
    if (_editorStore.codeCreationMode == CodeCreationMode.importZip &&
        _editorStore.sourceFiles.isEmpty) {
      return _buildImportZipMode();
    }

    // Default: Multi-file editor
    return _buildMultiFileEditor();
  }

  Widget _buildMultiFileEditor() {
    return FutureBuilder<void>(
      // stable future is important to prevent spinner flicker on each rebuild
      future: _initEditorFuture,
      builder: (context, snapshot) {
        // If editor is not created yet — show loader
        if (_editorDI == null) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError && _editorDI == null) {
          return Center(
            child: Text('Failed to initialize editor: ${snapshot.error}'),
          );
        }

        return MultiFileEditorWidget(editorDI: _editorDI!);
      },
    );
  }

  Widget _buildUploadWasmMode() {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 600),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: WasmUploadZone(
            onWasmUploaded: (base64) {
              _editorStore.setCode(base64);
            },
            onClear: () {
              _editorStore.setCode('');
            },
          ),
        ),
      ),
    );
  }

  Widget _buildImportZipMode() {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 600),
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Card(
            elevation: 2,
            child: Padding(
              padding: const EdgeInsets.all(48),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // ZIP Icon
                  const Icon(Icons.folder_zip, size: 80, color: Colors.blue),
                  const SizedBox(height: 24),

                  // Title
                  Text(
                    'Import Project from ZIP',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 12),

                  // Description
                  Text(
                    'Upload a ZIP archive containing your source code files',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: Theme.of(
                        context,
                      ).colorScheme.onSurface.withOpacity(0.6),
                    ),
                  ),
                  const SizedBox(height: 32),

                  // Main button
                  ElevatedButton.icon(
                    onPressed: _importProjectFromZip,
                    icon: const Icon(Icons.upload_file),
                    label: const Text('Select ZIP File'),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 32,
                        vertical: 16,
                      ),
                    ),
                  ),

                  const SizedBox(height: 24),

                  // Divider with "or"
                  Row(
                    children: [
                      const Expanded(child: Divider()),
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        child: Text(
                          'or',
                          style: TextStyle(
                            color: Theme.of(
                              context,
                            ).colorScheme.onSurface.withOpacity(0.6),
                          ),
                        ),
                      ),
                      const Expanded(child: Divider()),
                    ],
                  ),

                  const SizedBox(height: 24),

                  // Drag-drop hint
                  Container(
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: Theme.of(context).colorScheme.outline,
                        width: 2,
                        style: BorderStyle.solid,
                      ),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Column(
                      children: [
                        Icon(
                          Icons.cloud_upload_outlined,
                          size: 48,
                          color: Theme.of(
                            context,
                          ).colorScheme.onSurface.withOpacity(0.4),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          'Drag and drop ZIP file here',
                          style: TextStyle(
                            color: Theme.of(
                              context,
                            ).colorScheme.onSurface.withOpacity(0.6),
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 32),

                  // Info section
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.primaryContainer,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(
                              Icons.info_outline,
                              size: 16,
                              color: Theme.of(context).colorScheme.primary,
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'Supported formats:',
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Theme.of(
                                  context,
                                ).colorScheme.onPrimaryContainer,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          '.rs .go .ts .js .c .cpp .h .toml .json .yaml .md .txt',
                          style: TextStyle(
                            fontSize: 12,
                            color: Theme.of(
                              context,
                            ).colorScheme.onPrimaryContainer.withOpacity(0.8),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Max file size: 10 MB',
                          style: TextStyle(
                            fontSize: 12,
                            color: Theme.of(
                              context,
                            ).colorScheme.onPrimaryContainer.withOpacity(0.8),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
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

    final url = _editorStore.getDownloadProjectUrl();
    if (url == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Cannot download: script not saved yet'),
            backgroundColor: Colors.orange,
          ),
        );
      }
      return;
    }

    setState(() => _isDownloadingProject = true);

    try {
      final dio = Dio();
      final response = await dio.get<List<int>>(
        url,
        options: Options(responseType: ResponseType.bytes),
      );

      if (response.data == null) {
        throw Exception('Failed to download file');
      }

      final fileName =
          '${_editorStore.name.replaceAll(RegExp(r'[^\w\s-]'), '_')}_project.zip';

      final result = await FilePicker.platform.saveFile(
        dialogTitle: 'Save Project',
        fileName: fileName,
        bytes: Uint8List.fromList(response.data!),
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
