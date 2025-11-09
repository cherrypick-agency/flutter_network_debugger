import '../entities/download_progress.dart';
import '../repositories/compiler_repository.dart';

class WatchProgressUseCase {
  final CompilerRepository repository;

  WatchProgressUseCase(this.repository);

  Stream<DownloadProgress> call(String language) {
    return repository.watchProgress(language);
  }
}
