import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';

class MockProjectRepository implements ProjectRepository {
  final Map<String, Project> _projects = {};
  final Map<String, StreamController<Either<DomainFailure, Project>>>
  _watchers = {};
  String? _currentProjectId;

  int _idCounter = 0;

  String _generateId() => 'project_${++_idCounter}';

  @override
  Future<Either<DomainFailure, Project>> create({
    required String name,
    String? description,
    Map<String, dynamic>? settings,
    Map<String, dynamic>? metadata,
  }) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final id = _generateId();
    final now = DateTime.now();

    final project = Project(
      id: id,
      name: name,
      description: description,
      rootFolderId: 'root',
      createdAt: now,
      updatedAt: now,
      settings: settings ?? {},
      metadata: metadata ?? {},
    );

    _projects[id] = project;

    // Set as current if it's the first project
    _currentProjectId ??= id;

    _notifyWatchers(id, project);

    return Right(project);
  }

  @override
  Future<Either<DomainFailure, Project>> load(String id) async {
    await Future.delayed(const Duration(milliseconds: 50));

    final project = _projects[id];
    if (project == null) {
      return Left(
        DomainFailure.notFound(
          entityType: 'Project',
          entityId: id,
          message: 'Project with id "$id" not found',
        ),
      );
    }

    return Right(project);
  }

  @override
  Future<Either<DomainFailure, void>> save(Project project) async {
    await Future.delayed(const Duration(milliseconds: 50));

    if (!_projects.containsKey(project.id)) {
      return Left(
        DomainFailure.notFound(
          entityType: 'Project',
          entityId: project.id,
          message: 'Project with id "${project.id}" not found',
        ),
      );
    }

    final updated = project.copyWith(updatedAt: DateTime.now());
    _projects[project.id] = updated;
    _notifyWatchers(project.id, updated);

    return const Right(null);
  }

  @override
  Future<Either<DomainFailure, void>> delete(String id) async {
    await Future.delayed(const Duration(milliseconds: 50));

    if (!_projects.containsKey(id)) {
      return Left(
        DomainFailure.notFound(
          entityType: 'Project',
          entityId: id,
          message: 'Project with id "$id" not found',
        ),
      );
    }

    _projects.remove(id);

    // Clear current if deleting current project
    if (_currentProjectId == id) {
      _currentProjectId = _projects.keys.isNotEmpty
          ? _projects.keys.first
          : null;
    }

    _closeWatcher(id);

    return const Right(null);
  }

  @override
  Stream<Either<DomainFailure, Project>> watch(String id) {
    final controller =
        _watchers[id] ??
        StreamController<Either<DomainFailure, Project>>.broadcast();
    _watchers[id] = controller;

    final project = _projects[id];
    if (project != null) {
      Future.microtask(() => controller.add(Right(project)));
    }

    return controller.stream;
  }

  @override
  Future<Either<DomainFailure, List<Project>>> listAll() async {
    await Future.delayed(const Duration(milliseconds: 50));
    return Right(_projects.values.toList());
  }

  @override
  Future<Either<DomainFailure, Project>> getCurrent() async {
    await Future.delayed(const Duration(milliseconds: 50));

    if (_currentProjectId == null ||
        !_projects.containsKey(_currentProjectId)) {
      return Left(
        DomainFailure.notFound(
          entityType: 'Project',
          entityId: _currentProjectId ?? '',
          message: 'No current project set',
        ),
      );
    }

    return Right(_projects[_currentProjectId]!);
  }

  void _notifyWatchers(String id, Project project) {
    final watcher = _watchers[id];
    watcher?.add(Right(project));
  }

  void _closeWatcher(String id) {
    final watcher = _watchers[id];
    if (watcher != null) {
      watcher.close();
      _watchers.remove(id);
    }
  }

  void dispose() {
    for (final watcher in _watchers.values) {
      watcher.close();
    }
    _watchers.clear();
    _projects.clear();
  }
}
