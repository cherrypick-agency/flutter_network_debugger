import '../repositories/compiler_repository.dart';

class InstallCompilerUseCase {
  final CompilerRepository repository;

  InstallCompilerUseCase(this.repository);

  Future<void> call(String language) {
    return repository.installCompiler(language);
  }
}
