import '../entities/compiler_info.dart';
import '../repositories/compiler_repository.dart';

class GetCompilersUseCase {
  final CompilerRepository repository;

  GetCompilersUseCase(this.repository);

  Future<List<CompilerInfo>> call() {
    return repository.getCompilers();
  }
}
