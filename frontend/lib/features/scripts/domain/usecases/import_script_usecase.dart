import '../entities/script.dart';
import '../repositories/scripts_repository.dart';

/// Use case for importing a script from ZIP file
class ImportScriptUseCase {
  final ScriptsRepository _repository;

  ImportScriptUseCase(this._repository);

  /// Import script from ZIP file bytes
  Future<Script> call(List<int> zipBytes) async {
    if (zipBytes.isEmpty) {
      throw ArgumentError('ZIP file data cannot be empty');
    }

    return await _repository.importFromZip(zipBytes);
  }
}
