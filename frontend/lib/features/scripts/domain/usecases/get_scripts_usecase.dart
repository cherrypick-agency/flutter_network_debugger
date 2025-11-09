import '../entities/script.dart';
import '../repositories/scripts_repository.dart';

/// Use case for retrieving all scripts
class GetScriptsUseCase {
  final ScriptsRepository _repository;

  GetScriptsUseCase(this._repository);

  /// Retrieve all scripts from the repository
  Future<List<Script>> call() async {
    return await _repository.getAll();
  }
}
