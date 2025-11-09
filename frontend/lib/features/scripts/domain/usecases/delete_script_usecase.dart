import '../repositories/scripts_repository.dart';

/// Use case for deleting a script
class DeleteScriptUseCase {
  final ScriptsRepository _repository;

  DeleteScriptUseCase(this._repository);

  Future<void> call(String id) async {
    if (id.trim().isEmpty) {
      throw ArgumentError('Script ID cannot be empty');
    }

    return await _repository.delete(id);
  }
}
