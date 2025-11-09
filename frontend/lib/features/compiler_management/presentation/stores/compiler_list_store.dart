import 'package:mobx/mobx.dart';
import '../../domain/entities/compiler_info.dart';
import '../../domain/usecases/get_compilers_usecase.dart';
import '../../domain/usecases/install_compiler_usecase.dart';
import '../../domain/usecases/uninstall_compiler_usecase.dart';

part 'compiler_list_store.g.dart';

class CompilerListStore = _CompilerListStoreBase with _$CompilerListStore;

abstract class _CompilerListStoreBase with Store {
  final GetCompilersUseCase getCompilers;
  final InstallCompilerUseCase installCompiler;
  final UninstallCompilerUseCase uninstallCompiler;

  _CompilerListStoreBase({
    required this.getCompilers,
    required this.installCompiler,
    required this.uninstallCompiler,
  });

  @observable
  ObservableList<CompilerInfo> compilers = ObservableList<CompilerInfo>();

  @observable
  bool isLoading = false;

  @observable
  String? error;

  @computed
  List<CompilerInfo> get installedCompilers =>
      compilers.where((c) => c.isInstalled).toList();

  @computed
  List<CompilerInfo> get availableCompilers =>
      compilers.where((c) => c.isNotInstalled).toList();

  @computed
  int get totalCacheSize =>
      installedCompilers.fold(0, (sum, c) => sum + c.size);

  @action
  Future<void> loadCompilers() async {
    isLoading = true;
    error = null;

    try {
      final result = await getCompilers();
      compilers = ObservableList.of(result);
    } catch (e) {
      error = e.toString();
    } finally {
      isLoading = false;
    }
  }

  @action
  Future<void> install(String language) async {
    try {
      await installCompiler(language);
      await loadCompilers();
    } catch (e) {
      error = 'Failed to install $language: $e';
    }
  }

  @action
  Future<void> uninstall(String language) async {
    try {
      await uninstallCompiler(language);
      await loadCompilers();
    } catch (e) {
      error = 'Failed to uninstall $language: $e';
    }
  }

  @action
  void clearError() {
    error = null;
  }
}
