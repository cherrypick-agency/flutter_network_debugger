import '../entities/script.dart';
import '../repositories/scripts_repository.dart';

/// Use case for updating an existing script
class UpdateScriptUseCase {
  final ScriptsRepository _repository;

  UpdateScriptUseCase(this._repository);

  Future<Script> call(String id, Script script) async {
    // Business validation
    if (script.name.trim().isEmpty) {
      throw ArgumentError('Script name cannot be empty');
    }

    if (script.code.trim().isEmpty) {
      throw ArgumentError('Script code cannot be empty');
    }

    return await _repository.update(id, script);
  }
}
