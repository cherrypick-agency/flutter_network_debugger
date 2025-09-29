// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'session_details_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$SessionDetailsStore on _SessionDetailsStore, Store {
  late final _$sessionIdAtom = Atom(
    name: '_SessionDetailsStore.sessionId',
    context: context,
  );

  @override
  String? get sessionId {
    _$sessionIdAtom.reportRead();
    return super.sessionId;
  }

  @override
  set sessionId(String? value) {
    _$sessionIdAtom.reportWrite(value, super.sessionId, () {
      super.sessionId = value;
    });
  }

  late final _$framesAtom = Atom(
    name: '_SessionDetailsStore.frames',
    context: context,
  );

  @override
  ObservableList<Frame> get frames {
    _$framesAtom.reportRead();
    return super.frames;
  }

  @override
  set frames(ObservableList<Frame> value) {
    _$framesAtom.reportWrite(value, super.frames, () {
      super.frames = value;
    });
  }

  late final _$eventsAtom = Atom(
    name: '_SessionDetailsStore.events',
    context: context,
  );

  @override
  ObservableList<EventEntity> get events {
    _$eventsAtom.reportRead();
    return super.events;
  }

  @override
  set events(ObservableList<EventEntity> value) {
    _$eventsAtom.reportWrite(value, super.events, () {
      super.events = value;
    });
  }

  late final _$loadingAtom = Atom(
    name: '_SessionDetailsStore.loading',
    context: context,
  );

  @override
  bool get loading {
    _$loadingAtom.reportRead();
    return super.loading;
  }

  @override
  set loading(bool value) {
    _$loadingAtom.reportWrite(value, super.loading, () {
      super.loading = value;
    });
  }

  late final _$openAsyncAction = AsyncAction(
    '_SessionDetailsStore.open',
    context: context,
  );

  @override
  Future<void> open(String id) {
    return _$openAsyncAction.run(() => super.open(id));
  }

  late final _$loadMoreFramesAsyncAction = AsyncAction(
    '_SessionDetailsStore.loadMoreFrames',
    context: context,
  );

  @override
  Future<void> loadMoreFrames() {
    return _$loadMoreFramesAsyncAction.run(() => super.loadMoreFrames());
  }

  late final _$loadMoreEventsAsyncAction = AsyncAction(
    '_SessionDetailsStore.loadMoreEvents',
    context: context,
  );

  @override
  Future<void> loadMoreEvents() {
    return _$loadMoreEventsAsyncAction.run(() => super.loadMoreEvents());
  }

  @override
  String toString() {
    return '''
sessionId: ${sessionId},
frames: ${frames},
events: ${events},
loading: ${loading}
    ''';
  }
}
