// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'updates_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$UpdatesStore on _UpdatesStore, Store {
  Computed<List<UpdateInfo>>? _$filteredReleasesComputed;

  @override
  List<UpdateInfo> get filteredReleases =>
      (_$filteredReleasesComputed ??= Computed<List<UpdateInfo>>(
            () => super.filteredReleases,
            name: '_UpdatesStore.filteredReleases',
          ))
          .value;

  late final _$releasesAtom = Atom(
    name: '_UpdatesStore.releases',
    context: context,
  );

  @override
  ObservableList<UpdateInfo> get releases {
    _$releasesAtom.reportRead();
    return super.releases;
  }

  @override
  set releases(ObservableList<UpdateInfo> value) {
    _$releasesAtom.reportWrite(value, super.releases, () {
      super.releases = value;
    });
  }

  late final _$isLoadingAtom = Atom(
    name: '_UpdatesStore.isLoading',
    context: context,
  );

  @override
  bool get isLoading {
    _$isLoadingAtom.reportRead();
    return super.isLoading;
  }

  @override
  set isLoading(bool value) {
    _$isLoadingAtom.reportWrite(value, super.isLoading, () {
      super.isLoading = value;
    });
  }

  late final _$errorMessageAtom = Atom(
    name: '_UpdatesStore.errorMessage',
    context: context,
  );

  @override
  String? get errorMessage {
    _$errorMessageAtom.reportRead();
    return super.errorMessage;
  }

  @override
  set errorMessage(String? value) {
    _$errorMessageAtom.reportWrite(value, super.errorMessage, () {
      super.errorMessage = value;
    });
  }

  late final _$currentPageAtom = Atom(
    name: '_UpdatesStore.currentPage',
    context: context,
  );

  @override
  int get currentPage {
    _$currentPageAtom.reportRead();
    return super.currentPage;
  }

  @override
  set currentPage(int value) {
    _$currentPageAtom.reportWrite(value, super.currentPage, () {
      super.currentPage = value;
    });
  }

  late final _$hasMoreReleasesAtom = Atom(
    name: '_UpdatesStore.hasMoreReleases',
    context: context,
  );

  @override
  bool get hasMoreReleases {
    _$hasMoreReleasesAtom.reportRead();
    return super.hasMoreReleases;
  }

  @override
  set hasMoreReleases(bool value) {
    _$hasMoreReleasesAtom.reportWrite(value, super.hasMoreReleases, () {
      super.hasMoreReleases = value;
    });
  }

  late final _$filterAtom = Atom(
    name: '_UpdatesStore.filter',
    context: context,
  );

  @override
  ReleaseFilter get filter {
    _$filterAtom.reportRead();
    return super.filter;
  }

  @override
  set filter(ReleaseFilter value) {
    _$filterAtom.reportWrite(value, super.filter, () {
      super.filter = value;
    });
  }

  late final _$availableUpdateAtom = Atom(
    name: '_UpdatesStore.availableUpdate',
    context: context,
  );

  @override
  UpdateInfo? get availableUpdate {
    _$availableUpdateAtom.reportRead();
    return super.availableUpdate;
  }

  @override
  set availableUpdate(UpdateInfo? value) {
    _$availableUpdateAtom.reportWrite(value, super.availableUpdate, () {
      super.availableUpdate = value;
    });
  }

  late final _$isCheckingForUpdatesAtom = Atom(
    name: '_UpdatesStore.isCheckingForUpdates',
    context: context,
  );

  @override
  bool get isCheckingForUpdates {
    _$isCheckingForUpdatesAtom.reportRead();
    return super.isCheckingForUpdates;
  }

  @override
  set isCheckingForUpdates(bool value) {
    _$isCheckingForUpdatesAtom.reportWrite(
      value,
      super.isCheckingForUpdates,
      () {
        super.isCheckingForUpdates = value;
      },
    );
  }

  late final _$loadReleasesAsyncAction = AsyncAction(
    '_UpdatesStore.loadReleases',
    context: context,
  );

  @override
  Future<void> loadReleases({bool loadMore = false}) {
    return _$loadReleasesAsyncAction.run(
      () => super.loadReleases(loadMore: loadMore),
    );
  }

  late final _$checkForUpdatesAsyncAction = AsyncAction(
    '_UpdatesStore.checkForUpdates',
    context: context,
  );

  @override
  Future<void> checkForUpdates() {
    return _$checkForUpdatesAsyncAction.run(() => super.checkForUpdates());
  }

  late final _$_UpdatesStoreActionController = ActionController(
    name: '_UpdatesStore',
    context: context,
  );

  @override
  void setFilter(ReleaseFilter newFilter) {
    final _$actionInfo = _$_UpdatesStoreActionController.startAction(
      name: '_UpdatesStore.setFilter',
    );
    try {
      return super.setFilter(newFilter);
    } finally {
      _$_UpdatesStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  void reset() {
    final _$actionInfo = _$_UpdatesStoreActionController.startAction(
      name: '_UpdatesStore.reset',
    );
    try {
      return super.reset();
    } finally {
      _$_UpdatesStoreActionController.endAction(_$actionInfo);
    }
  }

  @override
  String toString() {
    return '''
releases: ${releases},
isLoading: ${isLoading},
errorMessage: ${errorMessage},
currentPage: ${currentPage},
hasMoreReleases: ${hasMoreReleases},
filter: ${filter},
availableUpdate: ${availableUpdate},
isCheckingForUpdates: ${isCheckingForUpdates},
filteredReleases: ${filteredReleases}
    ''';
  }
}
