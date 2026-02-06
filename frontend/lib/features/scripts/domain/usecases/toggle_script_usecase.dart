import '../entities/script.dart';
import '../repositories/scripts_repository.dart';

/// Use case for toggling script enabled status
class ToggleScriptUseCase {
  final ScriptsRepository _repository;

  ToggleScriptUseCase(this._repository);

  Future<Script> call(String id, bool enabled) async {
    if (id.trim().isEmpty) {
      throw ArgumentError('Script ID cannot be empty');
    }

    return await _repository.toggle(id, enabled);
  }
}
