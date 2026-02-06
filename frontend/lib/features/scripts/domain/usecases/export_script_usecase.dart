import '../repositories/scripts_repository.dart';

/// Use case for exporting a script as ZIP file
class ExportScriptUseCase {
  final ScriptsRepository _repository;

  ExportScriptUseCase(this._repository);

  Future<List<int>> call(String id) {
    if (id.trim().isEmpty) {
      throw ArgumentError('Script ID cannot be empty');
    }

    return _repository.downloadExportZip(id);
  }
}
