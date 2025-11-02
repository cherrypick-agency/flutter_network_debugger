import 'package:mobx/mobx.dart';
import 'package:logging/logging.dart';
import '../services/updates_service.dart';
import '../../domain/entities/update_info.dart';

part 'updates_store.g.dart';

/// Тип фильтра релизов
enum ReleaseFilter {
  all, // Все релизы
  stable, // Только stable (не pre-release)
  prerelease, // Только pre-release
}

class UpdatesStore = _UpdatesStore with _$UpdatesStore;

abstract class _UpdatesStore with Store {
  final UpdatesService _updatesService;
  final _log = Logger('UpdatesStore');

  _UpdatesStore(this._updatesService);

  /// Список загруженных релизов
  @observable
  ObservableList<UpdateInfo> releases = ObservableList();

  /// Состояние загрузки
  @observable
  bool isLoading = false;

  /// Сообщение об ошибке
  @observable
  String? errorMessage;

  /// Текущая страница пагинации
  @observable
  int currentPage = 1;

  /// Есть ли еще релизы для загрузки
  @observable
  bool hasMoreReleases = true;

  /// Текущий фильтр
  @observable
  ReleaseFilter filter = ReleaseFilter.all;

  /// Информация о доступном обновлении (если есть)
  @observable
  UpdateInfo? availableUpdate;

  /// Проверка обновления в процессе
  @observable
  bool isCheckingForUpdates = false;

  /// Отфильтрованный список релизов на основе выбранного фильтра
  @computed
  List<UpdateInfo> get filteredReleases {
    switch (filter) {
      case ReleaseFilter.all:
        return releases.toList();
      case ReleaseFilter.stable:
        return releases.where((r) => !r.isPrerelease).toList();
      case ReleaseFilter.prerelease:
        return releases.where((r) => r.isPrerelease).toList();
    }
  }

  /// Количество релизов на странице
  static const int _perPage = 15;

  /// Загружает релизы (первая страница или Load More)
  @action
  Future<void> loadReleases({bool loadMore = false}) async {
    if (isLoading) return;

    isLoading = true;
    errorMessage = null;

    try {
      final page = loadMore ? currentPage + 1 : 1;
      _log.info('Loading releases (page: $page, loadMore: $loadMore)');

      final newReleases = await _updatesService.getAllReleases(
        page: page,
        perPage: _perPage,
        includePrerelease: true, // Всегда загружаем все, фильтруем локально
      );

      if (loadMore) {
        // Добавляем к существующим
        releases.addAll(newReleases);
        currentPage = page;
      } else {
        // Заменяем список (первая загрузка или обновление)
        releases.clear();
        releases.addAll(newReleases);
        currentPage = 1;
      }

      // Если получили меньше чем perPage, значит больше нет релизов
      hasMoreReleases = newReleases.length >= _perPage;

      _log.info(
        'Loaded ${newReleases.length} releases (total: ${releases.length})',
      );
    } catch (e, stack) {
      _log.warning('Failed to load releases', e, stack);

      if (e.toString().contains('rate limit')) {
        errorMessage =
            'GitHub API rate limit exceeded. Please try again later.';
      } else if (e.toString().contains('Failed to fetch')) {
        errorMessage = 'Failed to load releases. Please check your connection.';
      } else {
        errorMessage = 'An error occurred while loading releases.';
      }
    } finally {
      isLoading = false;
    }
  }

  /// Проверяет наличие доступных обновлений
  @action
  Future<void> checkForUpdates() async {
    if (isCheckingForUpdates) return;

    isCheckingForUpdates = true;

    try {
      _log.info('Checking for updates...');
      final update = await _updatesService.checkForUpdates(forceCheck: true);
      availableUpdate = update;

      if (update != null) {
        _log.info('Update available: ${update.version}');
      } else {
        _log.info('No updates available');
      }
    } catch (e, stack) {
      _log.warning('Failed to check for updates', e, stack);
    } finally {
      isCheckingForUpdates = false;
    }
  }

  /// Устанавливает фильтр релизов
  @action
  void setFilter(ReleaseFilter newFilter) {
    _log.info('Setting filter: $newFilter');
    filter = newFilter;
  }

  /// Сброс состояния (для повторной загрузки)
  @action
  void reset() {
    releases.clear();
    currentPage = 1;
    hasMoreReleases = true;
    errorMessage = null;
    availableUpdate = null;
  }
}
