import '../repositories/scripts_repository.dart';

class ValidateSyntaxUseCase {
  final ScriptsRepository _repository;
  ValidateSyntaxUseCase(this._repository);

  Future<Map<String, dynamic>> call({
    required String sourceCode,
    required String language,
    Map<String, String>? dependencies,
  }) =>
      _repository.validateSyntax(
        sourceCode: sourceCode,
        language: language,
        dependencies: dependencies,
      );
}
