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

/// Dependency Injection для multi_file_code_editor
/// Адаптирует EditorScaffold для работы с ScriptEditorStore
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

  /// Инициализация всех зависимостей
  Future<void> init() async {
    // Создаём event bus
    eventBus = MobxEventBus();

    // Создаём адаптеры репозиториев, которые работают с ScriptEditorStore
    final baseFileRepository = MobxFileRepository(
      scriptStore: _scriptStore,
      eventBus: eventBus,
    );
    // Оборачиваем репозиторий, чтобы сохранять несохранённые изменения между переключениями
    _unsavedOverlay = UnsavedOverlayFileRepository(
      base: baseFileRepository,
      eventBus: eventBus,
    );
    fileRepository = _unsavedOverlay!;
    folderRepository = MobxFolderRepository();

    // Создаём контроллеры
    fileTreeController = FileTreeController(
      folderRepository: folderRepository,
      fileRepository: fileRepository,
      eventBus: eventBus,
    );

    editorController = EditorController(
      fileRepository: fileRepository,
      eventBus: eventBus,
    );

    // Инициализируем plugin system
    await _initializePluginSystem();

    // Загружаем дерево файлов
    await fileTreeController.load();
  }

  /// Инициализация plugin system
  Future<void> _initializePluginSystem() async {
    // Создаём service implementations
    multiEditorService = MultiEditorService();
    fileNavigationService = EditorFileNavigationService(editorController);
    pluginUIRegistry = PluginUIRegistry();
    errorTracker = ErrorTracker(maxErrors: 100);

    // Создаём plugin context
    pluginContext = AppPluginContext(
      eventBus: eventBus,
      fileRepository: fileRepository,
      folderRepository: folderRepository,
      validationService: ValidationServiceImpl(),
      languageDetector: LanguageDetectorImpl(),
    );

    // Регистрируем services в context
    pluginContext.registerService<EditorService>(multiEditorService);
    pluginContext.registerService<FileNavigationService>(fileNavigationService);
    pluginContext.registerService<PluginUIService>(pluginUIRegistry);

    // Создаём plugin manager
    pluginManager = PluginManager(pluginContext, errorTracker: errorTracker);

    // Регистрируем и активируем plugins
    await _registerPlugins();
  }

  /// Регистрация всех plugins
  Future<void> _registerPlugins() async {
    // Регистрируем FileIconsPlugin из package
    await pluginManager.registerPlugin(FileIconsPlugin());

    // Регистрируем локальные plugins
    await pluginManager.registerPlugin(DartLanguagePlugin());
    await pluginManager.registerPlugin(FileStatsPlugin());
    await pluginManager.registerPlugin(RecentFilesPlugin());
  }

  /// Синхронизация файлов из ScriptEditorStore в FileTreeController
  Future<void> syncFilesFromStore() async {
    // Перезагружаем дерево файлов при изменении sourceFiles
    await fileTreeController.load();
  }

  /// Освобождение ресурсов
  void dispose() {
    // Освобождаем plugin system
    pluginManager.disposeAll();
    errorTracker.dispose();
    multiEditorService.dispose();
    pluginUIRegistry.dispose();

    // Освобождаем контроллеры
    fileTreeController.dispose();
    editorController.dispose();
    _unsavedOverlay?.dispose();
  }

  /// Принудительно сохраняет буфер редактора в ScriptEditorStore.
  /// Используем перед отправкой на сервер, чтобы не потерять последние правки.
  Future<void> flushUnsavedChanges() async {
    await _unsavedOverlay?.flushAll();
  }

  /// Перезагружает текущий выбранный файл в редакторе, чтобы сбросить dirty‑флаг.
  Future<void> reloadCurrentFileIfAny(String? selectedFile) async {
    if (selectedFile == null) return;
    await editorController.loadFile(selectedFile);
  }
}
