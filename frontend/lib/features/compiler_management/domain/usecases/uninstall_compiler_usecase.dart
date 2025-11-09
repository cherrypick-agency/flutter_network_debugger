import '../repositories/compiler_repository.dart';

class UninstallCompilerUseCase {
  final CompilerRepository repository;

  UninstallCompilerUseCase(this.repository);

  Future<void> call(String language) {
    return repository.uninstallCompiler(language);
  }
}
