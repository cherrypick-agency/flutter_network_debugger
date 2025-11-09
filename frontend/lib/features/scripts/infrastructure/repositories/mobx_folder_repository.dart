import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';

/// Простой адаптер FolderRepository
/// У нас только одна root папка, так как все файлы хранятся плоским списком
class MobxFolderRepository implements FolderRepository {
  // Хранилище папок (только root)
  final Map<String, Folder> _folders = {
    'root': Folder(
      id: 'root',
      name: 'Project Files',
      parentId: null,
      createdAt: DateTime.now(),
      updatedAt: DateTime.now(),
    ),
  };

  @override
  Future<Either<DomainFailure, Folder>> create({
    required String name,
    String? parentId,
    Map<String, dynamic>? metadata,
  }) async {
    try {
      final id = '${parentId ?? 'root'}_$name';

      if (_folders.containsKey(id)) {
        return Left(
          DomainFailure.alreadyExists(
            entityType: 'Folder',
            entityId: id,
            message: 'Folder already exists: $name',
          ),
        );
      }

      final folder = Folder(
        id: id,
        name: name,
        parentId: parentId,
        createdAt: DateTime.now(),
        updatedAt: DateTime.now(),
        metadata: metadata ?? {},
      );

      _folders[id] = folder;

      return Right(folder);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to create folder: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, Folder>> load(String id) async {
    try {
      if (!_folders.containsKey(id)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'Folder',
            entityId: id,
            message: 'Folder not found: $id',
          ),
        );
      }

      return Right(_folders[id]!);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to load folder: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    try {
      if (id == 'root') {
        return Left(
          DomainFailure.permissionDenied(
            operation: 'delete',
            resource: 'root folder',
            message: 'Cannot delete root folder',
          ),
        );
      }

      if (!_folders.containsKey(id)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'Folder',
            entityId: id,
            message: 'Folder not found: $id',
          ),
        );
      }

      _folders.remove(id);
      return const Right(null);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to delete folder: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, Folder>> move({
    required String folderId,
    String? targetParentId,
  }) async {
    try {
      if (!_folders.containsKey(folderId)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'Folder',
            entityId: folderId,
            message: 'Folder not found: $folderId',
          ),
        );
      }

      final folder = _folders[folderId]!;
      final moved = folder.move(targetParentId);
      _folders[folderId] = moved;

      return Right(moved);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to move folder: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, Folder>> rename({
    required String folderId,
    required String newName,
  }) async {
    try {
      if (!_folders.containsKey(folderId)) {
        return Left(
          DomainFailure.notFound(
            entityType: 'Folder',
            entityId: folderId,
            message: 'Folder not found: $folderId',
          ),
        );
      }

      final folder = _folders[folderId]!;
      final renamed = folder.rename(newName);
      _folders[folderId] = renamed;

      return Right(renamed);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to rename folder: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, List<Folder>>> listInFolder(
    String? parentId,
  ) async {
    try {
      // Фильтруем папки по parentId
      final allFolders = _folders.values.toList();
      final filtered = <Folder>[];

      for (final folder in allFolders) {
        // folder.parentId может быть null, сравниваем правильно
        final folderParentId = folder.parentId;
        if ((parentId == null && folderParentId == null) ||
            (parentId != null && folderParentId == parentId)) {
          filtered.add(folder);
        }
      }

      return Right(filtered);
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to list folders: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, List<Folder>>> listAll() async {
    try {
      return Right(_folders.values.toList());
    } catch (e) {
      return Left(
        DomainFailure.unexpected(message: 'Failed to list folders: $e'),
      );
    }
  }

  @override
  Future<Either<DomainFailure, Folder>> getRoot() async {
    return load('root');
  }

  @override
  Stream<Either<DomainFailure, Folder>> watch(String id) {
    // Простая реализация без реального watching
    return Stream.empty();
  }
}
