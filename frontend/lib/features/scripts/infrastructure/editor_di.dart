import 'package:multi_editor_core/multi_editor_core.dart';
import 'package:multi_editor_ui/multi_editor_ui.dart';
import 'package:multi_editor_plugins/multi_editor_plugins.dart';
import 'package:multi_editor_plugin_file_icons/multi_editor_plugin_file_icons.dart';
import '../application/stores/script_editor_store.dart';
import 'repositories/mobx_file_repository.dart';
import 'repositories/mobx_folder_repository.dart';
import 'repositories/unsaved_overlay_file_repository.dart';
import 'events/mobx_event_bus.dart';
import 'multi_editor/plugin/app_plugin_context.dart';
import 'multi_editor/services/multi_editor_service.dart';
import 'multi_editor/services/file_navigation_service.dart';
import 'multi_editor/services/validation_service_impl.dart';
import 'multi_editor/services/language_detector_impl.dart';
import 'multi_editor/services/plugin_ui_registry.dart';
import 'package:multi_editor_plugin_dart/multi_editor_plugin_dart.dart';
import 'package:multi_editor_plugin_file_stats/multi_editor_plugin_file_stats.dart';
import 'package:multi_editor_plugin_recent_files/multi_editor_plugin_recent_files.dart';

/// Dependency Injection for multi_file_code_editor
/// Adapts EditorScaffold to work with ScriptEditorStore
class EditorDI {
  // Core components
  late final FileRepository fileRepository;
  late final FolderRepository folderRepository;
  late final EventBus eventBus;
  UnsavedOverlayFileRepository? _unsavedOverlay;

  late final FileTreeController fileTreeController;
  late final EditorController editorController;

  // Plugin system components
  late final AppPluginContext pluginContext;
  late final PluginManager pluginManager;
  late final ErrorTracker errorTracker;
  late final MultiEditorService multiEditorService;
  late final EditorFileNavigationService fileNavigationService;
  late final PluginUIRegistry pluginUIRegistry;

  final ScriptEditorStore _scriptStore;

  EditorDI({required ScriptEditorStore scriptStore})
    : _scriptStore = scriptStore;

  /// Initialize all dependencies
  Future<void> init() async {
    // Create event bus
    eventBus = MobxEventBus();

    // Create repository adapters that work with ScriptEditorStore
    final baseFileRepository = MobxFileRepository(
      scriptStore: _scriptStore,
      eventBus: eventBus,
    );
    // Wrap repository to preserve unsaved changes between file switches
    _unsavedOverlay = UnsavedOverlayFileRepository(
      base: baseFileRepository,
      eventBus: eventBus,
    );
    fileRepository = _unsavedOverlay!;
    folderRepository = MobxFolderRepository();

    // Create controllers
    fileTreeController = FileTreeController(
      folderRepository: folderRepository,
      fileRepository: fileRepository,
      eventBus: eventBus,
    );

    editorController = EditorController(
      fileRepository: fileRepository,
      eventBus: eventBus,
    );

    // Initialize plugin system
    await _initializePluginSystem();

    // Load file tree
    await fileTreeController.load();
  }

  /// Initialize plugin system
  Future<void> _initializePluginSystem() async {
    // Create service implementations
    multiEditorService = MultiEditorService();
    fileNavigationService = EditorFileNavigationService(editorController);
    pluginUIRegistry = PluginUIRegistry();
    errorTracker = ErrorTracker(maxErrors: 100);

    // Create plugin context
    pluginContext = AppPluginContext(
      eventBus: eventBus,
      fileRepository: fileRepository,
      folderRepository: folderRepository,
      validationService: ValidationServiceImpl(),
      languageDetector: LanguageDetectorImpl(),
    );

    // Register services in context
    pluginContext.registerService<EditorService>(multiEditorService);
    pluginContext.registerService<FileNavigationService>(fileNavigationService);
    pluginContext.registerService<PluginUIService>(pluginUIRegistry);

    // Create plugin manager
    pluginManager = PluginManager(pluginContext, errorTracker: errorTracker);

    // Register and activate plugins
    await _registerPlugins();
  }

  /// Register all plugins
  Future<void> _registerPlugins() async {
    // Register FileIconsPlugin from package
    await pluginManager.registerPlugin(FileIconsPlugin());

    // Register local plugins
    await pluginManager.registerPlugin(DartLanguagePlugin());
    await pluginManager.registerPlugin(FileStatsPlugin());
    await pluginManager.registerPlugin(RecentFilesPlugin());
  }

  /// Synchronize files from ScriptEditorStore to FileTreeController
  Future<void> syncFilesFromStore() async {
    // Reload file tree when sourceFiles change
    await fileTreeController.load();
  }

  /// Release resources
  void dispose() {
    // Release plugin system
    pluginManager.disposeAll();
    errorTracker.dispose();
    multiEditorService.dispose();
    pluginUIRegistry.dispose();

    // Release controllers
    fileTreeController.dispose();
    editorController.dispose();
    _unsavedOverlay?.dispose();
  }

  /// Forcefully saves editor buffer to ScriptEditorStore.
  /// Used before sending to server to not lose latest edits.
  Future<void> flushUnsavedChanges() async {
    await _unsavedOverlay?.flushAll();
  }

  /// Reloads currently selected file in editor to reset dirty flag.
  Future<void> reloadCurrentFileIfAny(String? selectedFile) async {
    if (selectedFile == null) return;
    await editorController.loadFile(selectedFile);
  }
}
