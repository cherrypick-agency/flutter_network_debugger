import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';

/// Wrapper repository that stores unsaved changes in memory
/// and mixes them in when loading a file. This allows switching between
/// files without losing typed content before explicit save.
class UnsavedOverlayFileRepository implements FileRepository {
  final FileRepository _base;
  final EventBus _eventBus;

  final Map<String, String> _unsavedByFileId = {};
  StreamSubscription<EditorEvent>? _subContentChanged;
  StreamSubscription<EditorEvent>? _subFileSaved;
  StreamSubscription<EditorEvent>? _subFileDeleted;

  UnsavedOverlayFileRepository({
    required FileRepository base,
    required EventBus eventBus,
  }) : _base = base,
       _eventBus = eventBus {
    // Cache unsaved content from editor events
    _subContentChanged = _eventBus.on<FileContentChanged>().listen((e) {
      _unsavedByFileId[e.fileId] = e.content;
    });
    // Clear cache after successful save or delete
    _subFileSaved = _eventBus.on<FileSaved>().listen((e) {
      _unsavedByFileId.remove(e.file.id);
    });
    _subFileDeleted = _eventBus.on<FileDeleted>().listen((e) {
      _unsavedByFileId.remove(e.fileId);
    });
  }

  void dispose() {
    _subContentChanged?.cancel();
    _subFileSaved?.cancel();
    _subFileDeleted?.cancel();
  }

  /// Forcefully saves all unsaved changes to base repository.
  /// Needed for cases when user presses Cmd+S before auto-save triggers.
  Future<void> flushAll() async {
    // Make a copy of keys since collection will change during save
    final pending = List<String>.from(_unsavedByFileId.keys);
    for (final fileId in pending) {
      final loaded = await load(fileId);
      await loaded.fold((_) async {}, (file) async {
        await save(file);
      });
    }
  }

  @override
  Future<Either<DomainFailure, FileDocument>> create({
    required String folderId,
    required String name,
    String? initialContent,
    String? language,
    Map<String, dynamic>? metadata,
  }) {
    return _base.create(
      folderId: folderId,
      name: name,
      initialContent: initialContent,
      language: language,
      metadata: metadata,
    );
  }

  @override
  Future<Either<DomainFailure, FileDocument>> load(String id) async {
    final res = await _base.load(id);
    return res.fold((l) => Left(l), (file) {
      final cached = _unsavedByFileId[id];
      if (cached != null) {
        // Return file with unsaved content
        return Right(file.updateContent(cached));
      }
      return Right(file);
    });
  }

  @override
  Future<Either<DomainFailure, void>> save(FileDocument file) async {
    final res = await _base.save(file);
    return res.fold((l) => Left(l), (_) {
      _unsavedByFileId.remove(file.id);
      return const Right(null);
    });
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    final res = await _base.delete(id);
    return res.fold((l) => Left(l), (_) {
      _unsavedByFileId.remove(id);
      return const Right(null);
    });
  }

  @override
  Future<Either<DomainFailure, FileDocument>> move({
    required String fileId,
    required String targetFolderId,
  }) {
    return _base.move(fileId: fileId, targetFolderId: targetFolderId);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> rename({
    required String fileId,
    required String newName,
  }) async {
    final res = await _base.rename(fileId: fileId, newName: newName);
    return res.fold((l) => Left(l), (file) {
      final cached = _unsavedByFileId.remove(fileId);
      if (cached != null) {
        _unsavedByFileId[file.id] = cached;
      }
      return Right(file);
    });
  }

  @override
  Future<Either<DomainFailure, FileDocument>> duplicate({
    required String fileId,
    String? newName,
  }) {
    return _base.duplicate(fileId: fileId, newName: newName);
  }

  @override
  Stream<Either<DomainFailure, FileDocument>> watch(String id) {
    return _base.watch(id);
  }

  @override
  Future<Either<DomainFailure, List<FileDocument>>> listInFolder(
    String folderId,
  ) {
    return _base.listInFolder(folderId);
  }

  @override
  Future<Either<DomainFailure, List<FileDocument>>> search({
    String? query,
    String? language,
    String? folderId,
  }) {
    return _base.search(query: query, language: language, folderId: folderId);
  }
}
