import 'dart:async';
import 'package:mobx/mobx.dart';
import '../../domain/entities/download_progress.dart';
import '../../domain/usecases/watch_progress_usecase.dart';

part 'installation_progress_store.g.dart';

class InstallationProgressStore = _InstallationProgressStoreBase
    with _$InstallationProgressStore;

abstract class _InstallationProgressStoreBase with Store {
  final WatchProgressUseCase watchProgress;

  _InstallationProgressStoreBase({required this.watchProgress});

  final Map<String, StreamSubscription> _subscriptions = {};

  @observable
  ObservableMap<String, DownloadProgress> progressMap =
      ObservableMap<String, DownloadProgress>();

  @action
  void startWatching(String language) {
    if (_subscriptions.containsKey(language)) {
      return;
    }

    final subscription = watchProgress(language).listen(
      (progress) {
        progressMap[language] = progress;

        if (progress.isComplete) {
          stopWatching(language);
        }
      },
      onError: (error) {
        stopWatching(language);
      },
      cancelOnError: true,
    );

    _subscriptions[language] = subscription;
  }

  @action
  void stopWatching(String language) {
    _subscriptions[language]?.cancel();
    _subscriptions.remove(language);
    progressMap.remove(language);
  }

  @action
  void stopAll() {
    for (final sub in _subscriptions.values) {
      sub.cancel();
    }
    _subscriptions.clear();
    progressMap.clear();
  }

  DownloadProgress? getProgress(String language) {
    return progressMap[language];
  }

  void dispose() {
    stopAll();
  }
}
