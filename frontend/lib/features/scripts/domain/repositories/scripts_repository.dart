import '../entities/script.dart';
import '../entities/script_test_result.dart';

/// Repository interface for Scripts domain
/// Following Repository Pattern and Dependency Inversion Principle
abstract class ScriptsRepository {
  /// Get all scripts
  Future<List<Script>> getAll();

  /// Get script by ID
  Future<Script> getById(String id);

  /// Create new script
  Future<Script> create(Script script);

  /// Update existing script
  Future<Script> update(String id, Script script);

  /// Delete script
  Future<void> delete(String id);

  /// Toggle script enabled/disabled status
  Future<void> toggle(String id, bool enabled);

  /// Test script with sample request
  /// Note: Backend may not have this endpoint yet, might need to add it
  Future<ScriptTestResult> test(Script script, TestRequest testRequest);

  /// Compile script source code to WASM
  Future<Script> compile(String id, {bool optimize = true});

  /// Export script as ZIP file (returns download URL)
  String getExportZipUrl(String id);

  /// Import script from ZIP file bytes
  Future<Script> importFromZip(List<int> zipBytes);
}
