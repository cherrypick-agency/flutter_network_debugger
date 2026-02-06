import '../repositories/scripts_repository.dart';

class ListProjectFilesUseCase {
  final ScriptsRepository _repository;
  ListProjectFilesUseCase(this._repository);

  Future<Map<String, dynamic>> call(String id) =>
      _repository.listProjectFiles(id);
}
