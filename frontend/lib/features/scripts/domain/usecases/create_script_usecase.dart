import '../entities/script.dart';
import '../repositories/scripts_repository.dart';

/// Use case for creating a new script
/// Following Single Responsibility Principle
class CreateScriptUseCase {
  final ScriptsRepository _repository;

  CreateScriptUseCase(this._repository);

  /// Execute the use case
  /// Validates script data before creating
  Future<Script> call(Script script) async {
    // Business validation
    _validate(script);

    // Create through repository
    return await _repository.create(script);
  }

  void _validate(Script script) {
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

    if (script.priority < 0 || script.priority > 100) {
      throw ArgumentError('Priority must be between 0 and 100');
    }
  }
}
