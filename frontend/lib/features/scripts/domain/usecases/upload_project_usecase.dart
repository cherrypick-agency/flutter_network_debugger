import '../repositories/scripts_repository.dart';

class UploadProjectUseCase {
  final ScriptsRepository _repository;
  UploadProjectUseCase(this._repository);

  Future<Map<String, dynamic>> call(String id, List<int> zipBytes) =>
      _repository.uploadProject(id, zipBytes);
}
