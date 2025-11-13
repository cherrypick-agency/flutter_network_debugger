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

    // For multi-file projects (writeSource/importZip), validate dependencies
    // For single-file (uploadWasm), validate code
    final hasSourceFiles =
        script.dependencies != null && script.dependencies!.isNotEmpty;
    final hasCode = script.code.trim().isNotEmpty;

    if (!hasCode && !hasSourceFiles) {
      throw ArgumentError('Script code or source files are required');
    }

    return await _repository.update(id, script);
  }
}
