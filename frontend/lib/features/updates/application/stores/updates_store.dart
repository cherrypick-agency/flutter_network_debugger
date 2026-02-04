import 'package:mobx/mobx.dart';
import 'package:logging/logging.dart';
import '../services/updates_service.dart';
import '../../domain/entities/update_info.dart';

part 'updates_store.g.dart';

/// Release filter type
enum ReleaseFilter {
  all, // All releases
  stable, // Only stable (not pre-release)
  prerelease, // Only pre-release
}

class UpdatesStore = _UpdatesStore with _$UpdatesStore;

abstract class _UpdatesStore with Store {
  final UpdatesService _updatesService;
  final _log = Logger('UpdatesStore');

  _UpdatesStore(this._updatesService);

  /// List of loaded releases
  @observable
  ObservableList<UpdateInfo> releases = ObservableList();

  /// Loading state
  @observable
  bool isLoading = false;

  /// Error message
  @observable
  String? errorMessage;

  /// Current pagination page
  @observable
  int currentPage = 1;

  /// Whether there are more releases to load
  @observable
  bool hasMoreReleases = true;

  /// Current filter
  @observable
  ReleaseFilter filter = ReleaseFilter.all;

  /// Information about available update (if any)
  @observable
  UpdateInfo? availableUpdate;

  /// Update check in progress
  @observable
  bool isCheckingForUpdates = false;

  /// Filtered list of releases based on selected filter
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

  /// Number of releases per page
  static const int _perPage = 15;

  /// Loads releases (first page or Load More)
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
        includePrerelease: true, // Always load all, filter locally
      );

      if (loadMore) {
        // Add to existing
        releases.addAll(newReleases);
        currentPage = page;
      } else {
        // Replace list (initial load or refresh)
        releases.clear();
        releases.addAll(newReleases);
        currentPage = 1;
      }

      // If received fewer than perPage, no more releases available
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

  /// Checks for available updates
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

  /// Sets release filter
  @action
  void setFilter(ReleaseFilter newFilter) {
    _log.info('Setting filter: $newFilter');
    filter = newFilter;
  }

  /// Reset state (for reload)
  @action
  void reset() {
    releases.clear();
    currentPage = 1;
    hasMoreReleases = true;
    errorMessage = null;
    availableUpdate = null;
  }
}
