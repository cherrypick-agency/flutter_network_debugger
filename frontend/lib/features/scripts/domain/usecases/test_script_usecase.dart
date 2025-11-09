import '../entities/script.dart';
import '../entities/script_test_result.dart';
import '../repositories/scripts_repository.dart';

/// Use case for testing a script before deploying
class TestScriptUseCase {
  final ScriptsRepository _repository;

  TestScriptUseCase(this._repository);

  Future<ScriptTestResult> call(Script script, TestRequest testRequest) async {
    // Validate script
    if (script.code.trim().isEmpty) {
      throw ArgumentError('Script code cannot be empty');
    }

    // Validate test request
    if (testRequest.url.trim().isEmpty) {
      throw ArgumentError('Test request URL cannot be empty');
    }

    return await _repository.test(script, testRequest);
  }
}
