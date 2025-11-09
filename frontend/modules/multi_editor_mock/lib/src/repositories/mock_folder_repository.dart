import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';

class MockFolderRepository implements FolderRepository {
  final Map<String, Folder> _folders = {};
  int _idCounter = 0;

  String _generateId() => 'folder_${++_idCounter}';

  MockFolderRepository() {
    final rootFolder = Folder(
      id: 'root',
      name: 'root',
      parentId: null,
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
    );
    _folders['root'] = rootFolder;
  }

  @override
  Future<Either<DomainFailure, Folder>> create({
    required String name,
    String? parentId,
    Map<String, dynamic>? metadata,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final id = _generateId();
    final now = DateTime.now();

    final folder = Folder(
      id: id,
      name: name,
      parentId: parentId,
      createdAt: now,
      updatedAt: now,
      metadata: metadata ?? {},
    );

    _folders[id] = folder;
    return Right(folder);
  }

  @override
  Future<Either<DomainFailure, Folder>> load(String id) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final folder = _folders[id];
    if (folder == null) {
      return Left(DomainFailure.notFound(entityType: 'Folder', entityId: id));
    }

    return Right(folder);
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    await Future.delayed(const Duration(milliseconds: 50));

    if (!_folders.containsKey(id)) {
      return Left(DomainFailure.notFound(entityType: 'Folder', entityId: id));
    }

    _folders.remove(id);
    return const Right(null);
  }

  @override
  Future<Either<DomainFailure, Folder>> move({
    required String folderId,
    String? targetParentId,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final folder = _folders[folderId];
    if (folder == null) {
      return Left(
        DomainFailure.notFound(entityType: 'Folder', entityId: folderId),
      );
    }

    final updated = folder.move(targetParentId);
    _folders[folderId] = updated;
    return Right(updated);
  }

  @override
  Future<Either<DomainFailure, Folder>> rename({
    required String folderId,
    required String newName,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final folder = _folders[folderId];
    if (folder == null) {
      return Left(
        DomainFailure.notFound(entityType: 'Folder', entityId: folderId),
      );
    }

    final updated = folder.rename(newName);
    _folders[folderId] = updated;
    return Right(updated);
  }

  @override
  Stream<Either<DomainFailure, Folder>> watch(String id) async* {
    final folder = _folders[id];
    if (folder != null) {
      yield Right(folder);
    }
  }

  @override
  Future<Either<DomainFailure, List<Folder>>> listInFolder(
    String? parentId,
  ) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final folders = _folders.values
        .where((f) => f.parentId == parentId)
        .toList();
    return Right(folders);
  }

  @override
  Future<Either<DomainFailure, List<Folder>>> listAll() async {
    await Future.delayed(const Duration(milliseconds: 50));
    return Right(_folders.values.toList());
  }

  @override
  Future<Either<DomainFailure, Folder>> getRoot() async {
    await Future.delayed(const Duration(milliseconds: 50));

    final root = _folders['root'];
    if (root == null) {
      return const Left(
        DomainFailure.notFound(
          entityType: 'Folder',
          entityId: 'root',
          message: 'Root folder not found',
        ),
      );
    }

    return Right(root);
  }

  void clear() {
    _folders.clear();
    _idCounter = 0;
    final rootFolder = Folder(
      id: 'root',
      name: 'root',
      parentId: null,
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
    );
    _folders['root'] = rootFolder;
  }

  void dispose() {
    clear();
  }
}
