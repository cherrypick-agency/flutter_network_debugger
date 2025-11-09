import '../entities/compiler_info.dart';
import '../entities/download_progress.dart';

/// PORT - Repository interface (Clean Architecture)
/// This is implemented by the Data layer
abstract class CompilerRepository {
  /// Get list of all available compilers
  Future<List<CompilerInfo>> getCompilers();

  /// Get info about specific compiler
  Future<CompilerInfo> getCompiler(String language);

  /// Install compiler
  Future<void> installCompiler(String language);

  /// Uninstall compiler
  Future<void> uninstallCompiler(String language);

  /// Watch installation progress (SSE stream)
  Stream<DownloadProgress> watchProgress(String language);

  /// Get cache size
  Future<int> getCacheSize();

  /// Clear all cache
  Future<void> clearCache();
}
