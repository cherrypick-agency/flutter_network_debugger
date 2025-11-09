import 'package:multi_editor_core/multi_editor_core.dart';
import 'package:multi_editor_ui/multi_editor_ui.dart';
import '../application/stores/script_editor_store.dart';
import 'repositories/mobx_file_repository.dart';
import 'repositories/mobx_folder_repository.dart';
import 'events/mobx_event_bus.dart';

/// Dependency Injection для multi_file_code_editor
/// Адаптирует EditorScaffold для работы с ScriptEditorStore
class EditorDI {
  late final FileRepository fileRepository;
  late final FolderRepository folderRepository;
  late final EventBus eventBus;

  late final FileTreeController fileTreeController;
  late final EditorController editorController;

  final ScriptEditorStore _scriptStore;

  EditorDI({required ScriptEditorStore scriptStore})
    : _scriptStore = scriptStore;

  /// Инициализация всех зависимостей
  Future<void> init() async {
    // Создаём event bus
    eventBus = MobxEventBus();

    // Создаём адаптеры репозиториев, которые работают с ScriptEditorStore
    fileRepository = MobxFileRepository(
      scriptStore: _scriptStore,
      eventBus: eventBus,
    );
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

    // Загружаем дерево файлов
    await fileTreeController.load();
  }

  /// Синхронизация файлов из ScriptEditorStore в FileTreeController
  Future<void> syncFilesFromStore() async {
    // Перезагружаем дерево файлов при изменении sourceFiles
    await fileTreeController.load();
  }

  /// Освобождение ресурсов
  void dispose() {
    fileTreeController.dispose();
    editorController.dispose();
  }
}
