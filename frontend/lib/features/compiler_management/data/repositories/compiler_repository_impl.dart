import '../../domain/entities/compiler_info.dart';
import '../../domain/entities/download_progress.dart';
import '../../domain/repositories/compiler_repository.dart';
import '../datasources/compiler_api.dart';

class CompilerRepositoryImpl implements CompilerRepository {
  final CompilerApi api;

  CompilerRepositoryImpl(this.api);

  @override
  Future<List<CompilerInfo>> getCompilers() async {
    final models = await api.getCompilers();
    return models.map((model) => model.toEntity()).toList();
  }

  @override
  Future<CompilerInfo> getCompiler(String language) async {
    final model = await api.getCompiler(language);
    return model.toEntity();
  }

  @override
  Future<void> installCompiler(String language) {
    return api.installCompiler(language);
  }

  @override
  Future<void> uninstallCompiler(String language) {
    return api.uninstallCompiler(language);
  }

  @override
  Stream<DownloadProgress> watchProgress(String language) {
    return api.watchProgress(language).map((model) => model.toEntity());
  }

  @override
  Future<int> getCacheSize() {
    return api.getCacheSize();
  }

  @override
  Future<void> clearCache() {
    return api.clearCache();
  }
}
