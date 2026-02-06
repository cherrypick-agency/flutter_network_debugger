import '../repositories/scripts_repository.dart';

class DownloadProjectUseCase {
  final ScriptsRepository _repository;
  DownloadProjectUseCase(this._repository);

  Future<List<int>> call(String id) {
    if (id.trim().isEmpty) {
      throw ArgumentError('Script ID cannot be empty');
    }
    return _repository.downloadProjectZip(id);
  }
}
