import '../repositories/scripts_repository.dart';

class DownloadProjectUseCase {
  final ScriptsRepository _repository;
  DownloadProjectUseCase(this._repository);

  String call(String id) => _repository.getDownloadProjectUrl(id);
}
