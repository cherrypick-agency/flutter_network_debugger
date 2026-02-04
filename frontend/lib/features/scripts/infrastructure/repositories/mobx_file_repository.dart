import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';
import '../../application/stores/script_editor_store.dart';

/// FileRepository adapter for working with ScriptEditorStore
/// Converts ObservableMap String->String to FileDocument entities
class MobxFileRepository implements FileRepository {
  final ScriptEditorStore scriptStore;
  final EventBus eventBus;

  // Controllers for watch streams
  final Map<String, StreamController<Either<DomainFailure, FileDocument>>>
  _watchControllers = {};

  // Mapping between sanitized ID and original filename
  final Map<String, String> _idToFilename = {};

  // Storage for folderId for each file (by sanitized id)
  // By default all files are considered in 'root'
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
      // Detect language from file extension if not specified
      final detectedLanguage = language ?? _detectLanguageFromFileName(name);

      // Add file to store
      scriptStore.addSourceFile(name, initialContent ?? '');

      // Register mapping between ID and filename
      _registerIdMapping(name);
      // Save folder binding
      _fileIdToFolderId[_sanitizeId(name)] = folderId;

      // Create FileDocument entity
      final file = FileDocument(
        id: _sanitizeId(name), // use sanitized filename as ID
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
      // Get original filename from sanitized ID
      String? filename = _getFilename(id);

      // If mapping not found, try alternative strategies
      if (filename == null) {
        // Strategy 1: Try using id as is (may already be original filename)
        if (scriptStore.sourceFiles.containsKey(id)) {
          filename = id;
          _registerIdMapping(id); // Register for future use
        }
        // Strategy 2: Try reverse sanitize (main_rs -> main.rs)
        else {
          final unsanitized = id.replaceAll('_', '.');
          if (scriptStore.sourceFiles.containsKey(unsanitized)) {
            filename = unsanitized;
            _registerIdMapping(unsanitized); // Register for future use
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

      // Find file in sourceFiles by filename
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
      // Get original filename from sanitized ID
      String? filename = _getFilename(file.id);
      // Fallback: some plugins/editor may return id as original filename
      // In this case accept id as filename directly.
      filename ??= scriptStore.sourceFiles.containsKey(file.id)
          ? file.id
          : null;
      // Another fallback: reverse sanitize (main_rs -> main.rs)
      filename ??= () {
        final unsanitized = file.id.replaceAll('_', '.');
        return scriptStore.sourceFiles.containsKey(unsanitized)
            ? unsanitized
            : null;
      }();
      if (filename == null) {
        return Left(
          DomainFailure.notFound(
            entityType: 'File',
            entityId: file.id,
            message: 'File mapping not found: ${file.id}',
          ),
        );
      }

      // Check that file exists
      final loadResult = await load(file.id);

      return loadResult.fold((failure) => Left(failure), (existingFile) {
        // Update file content using original name
        scriptStore.updateSourceFile(filename!, file.content);
        return const Right(null);
      });
    } catch (e) {
      return Left(DomainFailure.unexpected(message: 'Failed to save file: $e'));
    }
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    try {
      // Get original filename
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
      _idToFilename.remove(id); // Remove mapping
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
      // Get old filename
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

      // Remove old file and mapping
      scriptStore.removeSourceFile(oldFilename);
      _idToFilename.remove(fileId);
      _fileIdToFolderId.remove(fileId);

      // Create new file with new name
      scriptStore.addSourceFile(newName, content);
      _registerIdMapping(newName);

      final language = _detectLanguageFromFileName(newName);
      final newId = _sanitizeId(newName);
      // Transfer folder binding to new id
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
    // Update file folder binding and return updated document
    _fileIdToFolderId[fileId] = targetFolderId;
    return load(fileId);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> duplicate({
    required String fileId,
    String? newName,
  }) async {
    try {
      // Get original filename
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
      // Ensure mapping is set correctly
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

        // Register mapping if not exists yet
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

      // Filter by query if specified
      var result = files;

      if (query != null && query.isNotEmpty) {
        result = result
            .where(
              (file) => file.name.toLowerCase().contains(query.toLowerCase()),
            )
            .toList();
      }

      // Filter by language if specified
      if (language != null && language.isNotEmpty) {
        result = result.where((file) => file.language == language).toList();
      }

      // Filter by folder if specified
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
    // Create stream controller if not exists yet
    if (!_watchControllers.containsKey(id)) {
      _watchControllers[id] =
          StreamController<Either<DomainFailure, FileDocument>>.broadcast();
    }

    // Return stream
    return _watchControllers[id]!.stream;
  }

  /// Detects programming language from file extension
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

  /// Notifies watchers about file change
  void notifyFileChanged(String fileId, FileDocument file) {
    if (_watchControllers.containsKey(fileId)) {
      _watchControllers[fileId]!.add(Right(file));
    }
  }

  /// Releases resources
  void dispose() {
    for (final controller in _watchControllers.values) {
      controller.close();
    }
    _watchControllers.clear();
  }
}
