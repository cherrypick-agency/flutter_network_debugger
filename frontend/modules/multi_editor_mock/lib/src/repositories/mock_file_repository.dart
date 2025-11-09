import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';

class MockFileRepository implements FileRepository {
  final Map<String, FileDocument> _files = {};
  final Map<String, StreamController<Either<DomainFailure, FileDocument>>>
  _watchers = {};

  int _idCounter = 0;

  String _generateId() => 'file_${++_idCounter}';

  @override
  Future<Either<DomainFailure, FileDocument>> create({
    required String folderId,
    required String name,
    String? initialContent,
    String? language,
    Map<String, dynamic>? metadata,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final id = _generateId();
    final now = DateTime.now();

    final file = FileDocument(
      id: id,
      name: name,
      folderId: folderId,
      content: initialContent ?? '',
      language: language ?? 'plaintext',
      createdAt: now,
      updatedAt: now,
      metadata: metadata,
    );

    _files[id] = file;
    _notifyWatchers(id, file);

    return Right(file);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> load(String id) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final file = _files[id];
    if (file == null) {
      return Left(
        DomainFailure.notFound(entityType: 'FileDocument', entityId: id),
      );
    }

    return Right(file);
  }

  @override
  Future<Either<DomainFailure, void>> save(FileDocument file) async {
    await Future.delayed(const Duration(milliseconds: 50));

    if (!_files.containsKey(file.id)) {
      return Left(
        DomainFailure.notFound(entityType: 'FileDocument', entityId: file.id),
      );
    }

    _files[file.id] = file;
    _notifyWatchers(file.id, file);

    return const Right(null);
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    await Future.delayed(const Duration(milliseconds: 50));

    if (!_files.containsKey(id)) {
      return Left(
        DomainFailure.notFound(entityType: 'FileDocument', entityId: id),
      );
    }

    _files.remove(id);
    _closeWatcher(id);

    return const Right(null);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> move({
    required String fileId,
    required String targetFolderId,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final file = _files[fileId];
    if (file == null) {
      return Left(
        DomainFailure.notFound(entityType: 'FileDocument', entityId: fileId),
      );
    }

    final updated = file.move(targetFolderId);
    _files[fileId] = updated;
    _notifyWatchers(fileId, updated);

    return Right(updated);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> rename({
    required String fileId,
    required String newName,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final file = _files[fileId];
    if (file == null) {
      return Left(
        DomainFailure.notFound(entityType: 'FileDocument', entityId: fileId),
      );
    }

    final updated = file.rename(newName);
    _files[fileId] = updated;
    _notifyWatchers(fileId, updated);

    return Right(updated);
  }

  @override
  Future<Either<DomainFailure, FileDocument>> duplicate({
    required String fileId,
    String? newName,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final file = _files[fileId];
    if (file == null) {
      return Left(
        DomainFailure.notFound(entityType: 'FileDocument', entityId: fileId),
      );
    }

    final duplicateName = newName ?? 'Copy of ${file.name}';
    return create(
      folderId: file.folderId,
      name: duplicateName,
      initialContent: file.content,
      language: file.language,
      metadata: file.metadata,
    );
  }

  @override
  Stream<Either<DomainFailure, FileDocument>> watch(String id) {
    if (!_watchers.containsKey(id)) {
      _watchers[id] =
          StreamController<Either<DomainFailure, FileDocument>>.broadcast();
    }

    final file = _files[id];
    if (file != null) {
      Future.microtask(() {
        _watchers[id]?.add(Right(file));
      });
    }

    return _watchers[id]!.stream;
  }

  @override
  Future<Either<DomainFailure, List<FileDocument>>> listInFolder(
    String folderId,
  ) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final files = _files.values.where((f) => f.folderId == folderId).toList();
    return Right(files);
  }

  @override
  Future<Either<DomainFailure, List<FileDocument>>> search({
    String? query,
    String? language,
    String? folderId,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    var results = _files.values;

    if (folderId != null) {
      results = results.where((f) => f.folderId == folderId);
    }

    if (language != null) {
      results = results.where((f) => f.language == language);
    }

    if (query != null && query.isNotEmpty) {
      final lowerQuery = query.toLowerCase();
      results = results.where(
        (f) =>
            f.name.toLowerCase().contains(lowerQuery) ||
            f.content.toLowerCase().contains(lowerQuery),
      );
    }

    return Right(results.toList());
  }

  void _notifyWatchers(String id, FileDocument file) {
    _watchers[id]?.add(Right(file));
  }

  void _closeWatcher(String id) {
    _watchers[id]?.close();
    _watchers.remove(id);
  }

  void clear() {
    _files.clear();
    for (final watcher in _watchers.values) {
      watcher.close();
    }
    _watchers.clear();
    _idCounter = 0;
  }

  void dispose() {
    clear();
  }
}
