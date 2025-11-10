import 'package:multi_editor_core/multi_editor_core.dart';
import 'package:multi_editor_plugins/multi_editor_plugins.dart';

/// Application-specific implementation of [PluginContext]
///
/// Provides plugins with access to:
/// - Core repositories (FileRepository, FolderRepository)
/// - Services (ValidationService, LanguageDetector, EditorService, etc.)
/// - Event bus for communication
/// - Configuration storage
///
/// Follows SOLID principles by implementing the PluginContext interface
/// and delegating to dependency-injected components.
class AppPluginContext implements PluginContext {
  final EventBus _eventBus;
  final FileRepository _fileRepository;
  final FolderRepository _folderRepository;
  final ValidationService _validationService;
  final LanguageDetector _languageDetector;

  /// Registry of services available to plugins
  /// Key: Service type, Value: Service instance
  final Map<Type, Object> _services = {};

  /// Plugin-specific configuration storage
  /// Key: Plugin ID, Value: Configuration map
  final Map<String, Map<String, dynamic>> _configurations = {};

  AppPluginContext({
    required EventBus eventBus,
    required FileRepository fileRepository,
    required FolderRepository folderRepository,
    required ValidationService validationService,
    required LanguageDetector languageDetector,
  }) : _eventBus = eventBus,
       _fileRepository = fileRepository,
       _folderRepository = folderRepository,
       _validationService = validationService,
       _languageDetector = languageDetector;

  /// Command bus is not implemented in this application
  ///
  /// Throws [UnimplementedError] if accessed.
  /// Can be implemented later if command pattern is needed.
  @override
  CommandBus get commands => throw UnimplementedError(
    'CommandBus is not implemented. Use EventBus for communication.',
  );

  /// Returns the event bus for plugin communication
  @override
  EventBus get events => _eventBus;

  /// Hook registry is not implemented in this application
  ///
  /// Throws [UnimplementedError] if accessed.
  /// Plugins can listen to events via EventBus instead.
  @override
  HookRegistry get hooks => throw UnimplementedError(
    'HookRegistry is not implemented. Use EventBus for event handling.',
  );

  /// Returns the file repository for file operations
  @override
  FileRepository get fileRepository => _fileRepository;

  /// Returns the folder repository for folder operations
  @override
  FolderRepository get folderRepository => _folderRepository;

  /// Project repository is not implemented in this application
  ///
  /// Throws [UnimplementedError] if accessed.
  /// The current implementation works with a single script editor instance.
  @override
  ProjectRepository get projectRepository => throw UnimplementedError(
    'ProjectRepository is not implemented. This editor works with individual files.',
  );

  /// Returns the validation service for name/path validation
  @override
  ValidationService get validationService => _validationService;

  /// Returns the language detector for file language detection
  @override
  LanguageDetector get languageDetector => _languageDetector;

  /// Gets plugin-specific configuration
  ///
  /// Returns an empty map if no configuration exists for the given key.
  ///
  /// [key] - Usually the plugin ID
  @override
  Map<String, dynamic> getConfiguration(String key) {
    return _configurations[key] ?? {};
  }

  /// Sets plugin-specific configuration
  ///
  /// [key] - Usually the plugin ID
  /// [value] - Configuration data to store
  @override
  void setConfiguration(String key, Map<String, dynamic> value) {
    _configurations[key] = value;
  }

  /// Gets a service by type from the service registry
  ///
  /// Returns null if the service is not registered.
  ///
  /// Example:
  /// ```dart
  /// final editorService = context.getService<EditorService>();
  /// ```
  @override
  T? getService<T extends Object>() {
    return _services[T] as T?;
  }

  /// Registers a service in the service registry
  ///
  /// Services are registered by their type and can be retrieved by plugins.
  ///
  /// Example:
  /// ```dart
  /// context.registerService<EditorService>(myEditorService);
  /// ```
  @override
  void registerService<T extends Object>(T service) {
    _services[T] = service;
  }
}
