import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';
import '../../application/stores/script_editor_store.dart';

/// Адаптер FileRepository для работы с ScriptEditorStore
/// Преобразует ObservableMap String->String в FileDocument entities
class MobxFileRepository implements FileRepository {
  final ScriptEditorStore scriptStore;
  final EventBus eventBus;

  // Контроллеры для watch streams
  final Map<String, StreamController<Either<DomainFailure, FileDocument>>>
  _watchControllers = {};

  // Mapping между sanitized ID и оригинальным filename
  final Map<String, String> _idToFilename = {};

  // Хранилище folderId для каждого файла (по sanitized id)
  // По умолчанию все файлы считаются в 'root'
  final Map<String, String> _fileIdToFolderId = {};

  MobxFileRepository({required this.scriptStore, required this.eventBus});

  /// Sanitizes filename to be used as ID (replaces dots with underscores)
  /// to avoid conflicts with animated_tree_view PATH_SEPARATOR
  String _sanitizeId(String filename) {
    return filename.replaceAll('.', '_');
  }

  /// Registers mapping between sanitized ID and original filename
  void _registerIdMapping(String filename) {
    final sanitizedId = _sanitizeId(filename);
    _idToFilename[sanitizedId] = filename;
  }

  /// Gets original filename from sanitized ID
  String? _getFilename(String sanitizedId) {
    return _idToFilename[sanitizedId];
  }

  @override
  Future<Either<DomainFailure, FileDocument>> create({
    required String folderId,
    required String name,
    String? initialContent,
    String? language,
    Map<String, dynamic>? metadata,
  }) async {
    try {
      // Определяем язык по расширению файла, если не указан
      final detectedLanguage = language ?? _detectLanguageFromFileName(name);

      // Добавляем файл в store
      scriptStore.addSourceFile(name, initialContent ?? '');

      // Регистрируем mapping между ID и filename
      _registerIdMapping(name);
      // Сохраняем привязку к папке
      _fileIdToFolderId[_sanitizeId(name)] = folderId;

      // Создаём FileDocument entity
      final file = FileDocument(
        id: _sanitizeId(name), // используем sanitized имя файла как ID
        name: name,
        content: initialContent ?? '',
        language: detectedLanguage,
        folderId: folderId,
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        metadata: metadata,
      );

      return Right(file);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to create file: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, FileDocument>> load(String id) async {
    try {
      // Получаем оригинальное имя файла из sanitized ID
      String? filename = _getFilename(id);

      // Если mapping не найден, пробуем альтернативные стратегии
      if (filename == null) {
        // Стратегия 1: Попробовать использовать id как есть (может быть уже original filename)
        if (scriptStore.sourceFiles.containsKey(id)) {
          filename = id;
          _registerIdMapping(id); // Регистрируем для будущего использования
        }
        // Стратегия 2: Попробовать reverse sanitize (main_rs → main.rs)
        else {
          final unsanitized = id.replaceAll('_', '.');
          if (scriptStore.sourceFiles.containsKey(unsanitized)) {
            filename = unsanitized;
            _registerIdMapping(
              unsanitized,
            ); // Регистрируем для будущего использования
          } else {
            return Left(
              DomainFailure.notFound(
                entityType: 'File',
                entityId: id,
                message:
                    'File mapping not found and file not found in sourceFiles: $id',
              ),
            );
          }
        }
      }

      // Ищем файл в sourceFiles по имени файла
      if (!scriptStore.sourceFiles.containsKey(filename)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: id,
            message: 'File not found: $filename',
          ),
        );
      }

      final content = scriptStore.sourceFiles[filename]!;
      final language = _detectLanguageFromFileName(filename);

      final file = FileDocument(
        id: id,
        name: filename,
        content: content,
        language: language,
        folderId: _fileIdToFolderId[id] ?? 'root',
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
      );

      return Right(file);
    } catch (e) {
      return Left(DomainFailure.unexpected(message: 'Failed to load file: $e'));
    }
  }

  @override
  Future<Either<DomainFailure, void>> save(FileDocument file) async {
    try {
      // Получаем оригинальное имя файла из sanitized ID
      final filename = _getFilename(file.id);
      if (filename == null) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: file.id,
            message: 'File mapping not found: ${file.id}',
          ),
        );
      }

      // Проверяем что файл существует
      final loadResult = await load(file.id);

      return loadResult.fold((failure) => Left(failure), (existingFile) {
        // Обновляем содержимое файла используя оригинальное имя
        scriptStore.updateSourceFile(filename, file.content);
        return const Right(null);
      });
    } catch (e) {
      return Left(DomainFailure.unexpected(message: 'Failed to save file: $e'));
    }
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    try {
      // Получаем оригинальное имя файла
      final filename = _getFilename(id);
      if (filename == null) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: id,
            message: 'File mapping not found: $id',
          ),
        );
      }

      scriptStore.removeSourceFile(filename);
      _idToFilename.remove(id); // Удаляем mapping
      return const Right(null);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to delete file: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, FileDocument>> rename({
    required String fileId,
    required String newName,
  }) async {
    try {
      // Получаем старое имя файла
      final oldFilename = _getFilename(fileId);
      if (oldFilename == null) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: fileId,
            message: 'File mapping not found: $fileId',
          ),
        );
      }

      if (!scriptStore.sourceFiles.containsKey(oldFilename)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: fileId,
            message: 'File not found: $oldFilename',
          ),
        );
      }

      final content = scriptStore.sourceFiles[oldFilename]!;
      final currentFolderId = _fileIdToFolderId[fileId] ?? 'root';

      // Удаляем старый файл и mapping
      scriptStore.removeSourceFile(oldFilename);
      _idToFilename.remove(fileId);
      _fileIdToFolderId.remove(fileId);

      // Создаём новый файл с новым именем
      scriptStore.addSourceFile(newName, content);
      _registerIdMapping(newName);

      final language = _detectLanguageFromFileName(newName);
      final newId = _sanitizeId(newName);
      // Переносим привязку к папке на новый id
      _fileIdToFolderId[newId] = currentFolderId;

      final file = FileDocument(
        id: newId,
        name: newName,
        content: content,
        language: language,
        folderId: currentFolderId,
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
      );

      return Right(file);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to rename file: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, FileDocument>> move({
    required String fileId,
    required String targetFolderId,
  }) async {
    // Обновляем привязку файла к папке и возвращаем обновлённый документ
    _fileIdToFolderId[fileId] = targetFolderId;
    return load(fileId);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> duplicate({
    required String fileId,
    String? newName,
  }) async {
    try {
      // Получаем оригинальное имя файла
      final filename = _getFilename(fileId);
      if (filename == null) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: fileId,
            message: 'File mapping not found: $fileId',
          ),
        );
      }

      if (!scriptStore.sourceFiles.containsKey(filename)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: fileId,
            message: 'File not found: $filename',
          ),
        );
      }

      final content = scriptStore.sourceFiles[filename]!;
      final name = newName ?? '${filename}_copy';
      final newId = _sanitizeId(name);
      final folderId = _fileIdToFolderId[fileId] ?? 'root';

      final created = await create(
        folderId: folderId,
        name: name,
        initialContent: content,
      );
      // Подстрахуемся, что маппинг выставлен корректно
      _fileIdToFolderId[newId] = folderId;
      return created;
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to duplicate file: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, List<FileDocument>>> listInFolder(
    String folderId,
  ) async {
    return search(folderId: folderId);
  }

  @override
  Future<Either<DomainFailure, List<FileDocument>>> search({
    String? query,
    String? language,
    String? folderId,
  }) async {
    try {
      final files = scriptStore.sourceFiles.entries.map((entry) {
        final filename = entry.key;
        final lang = _detectLanguageFromFileName(filename);

        // Регистрируем mapping если его еще нет
        _registerIdMapping(filename);

        final id = _sanitizeId(filename);
        final assignedFolderId = _fileIdToFolderId[id] ?? 'root';

        return FileDocument(
          id: id,
          name: filename,
          content: entry.value,
          language: lang,
          folderId: assignedFolderId,
          createdAt: DateTime.now(),
          updatedAt: DateTime.now(),
        );
      }).toList();

      // Фильтруем по query если указан
      var result = files;

      if (query != null && query.isNotEmpty) {
        result = result
            .where(
              (file) => file.name.toLowerCase().contains(query.toLowerCase()),
            )
            .toList();
      }

      // Фильтруем по языку если указан
      if (language != null && language.isNotEmpty) {
        result = result.where((file) => file.language == language).toList();
      }

      // Фильтруем по папке если указана
      if (folderId != null) {
        result = result.where((file) => file.folderId == folderId).toList();
      }

      return Right(result);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to search files: $e'),
      );
    }
  }

  @override
  Stream<Either<DomainFailure, FileDocument>> watch(String id) {
    // Создаём stream controller если его ещё нет
    if (!_watchControllers.containsKey(id)) {
      _watchControllers[id] =
          StreamController<Either<DomainFailure, FileDocument>>.broadcast();
    }

    // Возвращаем stream
    return _watchControllers[id]!.stream;
  }

  /// Определяет язык программирования по расширению файла
  String _detectLanguageFromFileName(String fileName) {
    final extension = fileName.split('.').last.toLowerCase();

    switch (extension) {
      case 'dart':
        return 'dart';
      case 'js':
        return 'javascript';
      case 'ts':
        return 'typescript';
      case 'rs':
        return 'rust';
      case 'go':
        return 'go';
      case 'py':
        return 'python';
      case 'c':
        return 'c';
      case 'cpp':
      case 'cc':
      case 'cxx':
        return 'cpp';
      case 'json':
        return 'json';
      case 'md':
        return 'markdown';
      case 'toml':
        return 'toml';
      case 'yaml':
      case 'yml':
        return 'yaml';
      case 'xml':
        return 'xml';
      case 'html':
        return 'html';
      case 'css':
        return 'css';
      default:
        return 'text';
    }
  }

  /// Уведомляет watchers об изменении файла
  void notifyFileChanged(String fileId, FileDocument file) {
    if (_watchControllers.containsKey(fileId)) {
      _watchControllers[fileId]!.add(Right(file));
    }
  }

  /// Освобождает ресурсы
  void dispose() {
    for (final controller in _watchControllers.values) {
      controller.close();
    }
    _watchControllers.clear();
  }
}
