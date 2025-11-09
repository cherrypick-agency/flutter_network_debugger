import '../entities/script.dart';
import '../repositories/scripts_repository.dart';

/// Use case for compiling script source code to WASM
class CompileScriptUseCase {
  final ScriptsRepository _repository;

  CompileScriptUseCase(this._repository);

  Future<Script> call(String id, {bool optimize = true}) async {
    if (id.trim().isEmpty) {
      throw ArgumentError('Script ID cannot be empty');
    }

    return await _repository.compile(id, optimize: optimize);
  }
}
