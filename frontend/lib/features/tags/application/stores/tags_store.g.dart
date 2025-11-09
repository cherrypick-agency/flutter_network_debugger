// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tags_store.dart';

// **************************************************************************
// StoreGenerator
// **************************************************************************

// ignore_for_file: non_constant_identifier_names, unnecessary_brace_in_string_interps, unnecessary_lambdas, prefer_expression_function_bodies, lines_longer_than_80_chars, avoid_as, avoid_annotating_with_dynamic, no_leading_underscores_for_local_identifiers

mixin _$TagsStore on _TagsStore, Store {
  late final _$predefinedTagsAtom = Atom(
    name: '_TagsStore.predefinedTags',
    context: context,
  );

  @override
  ObservableList<PredefinedTag> get predefinedTags {
    _$predefinedTagsAtom.reportRead();
    return super.predefinedTags;
  }

  @override
  set predefinedTags(ObservableList<PredefinedTag> value) {
    _$predefinedTagsAtom.reportWrite(value, super.predefinedTags, () {
      super.predefinedTags = value;
    });
  }

  late final _$sessionTagsCacheAtom = Atom(
    name: '_TagsStore.sessionTagsCache',
    context: context,
  );

  @override
  ObservableMap<String, List<SessionTag>> get sessionTagsCache {
    _$sessionTagsCacheAtom.reportRead();
    return super.sessionTagsCache;
  }

  @override
  set sessionTagsCache(ObservableMap<String, List<SessionTag>> value) {
    _$sessionTagsCacheAtom.reportWrite(value, super.sessionTagsCache, () {
      super.sessionTagsCache = value;
    });
  }

  late final _$sessionAnnotationsCacheAtom = Atom(
    name: '_TagsStore.sessionAnnotationsCache',
    context: context,
  );

  @override
  ObservableMap<String, List<SessionAnnotation>> get sessionAnnotationsCache {
    _$sessionAnnotationsCacheAtom.reportRead();
    return super.sessionAnnotationsCache;
  }

  @override
  set sessionAnnotationsCache(
    ObservableMap<String, List<SessionAnnotation>> value,
  ) {
    _$sessionAnnotationsCacheAtom.reportWrite(
      value,
      super.sessionAnnotationsCache,
      () {
        super.sessionAnnotationsCache = value;
      },
    );
  }

  late final _$loadingAtom = Atom(name: '_TagsStore.loading', context: context);

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

  late final _$errorAtom = Atom(name: '_TagsStore.error', context: context);

  @override
  String? get error {
    _$errorAtom.reportRead();
    return super.error;
  }

  @override
  set error(String? value) {
    _$errorAtom.reportWrite(value, super.error, () {
      super.error = value;
    });
  }

  late final _$loadPredefinedTagsAsyncAction = AsyncAction(
    '_TagsStore.loadPredefinedTags',
    context: context,
  );

  @override
  Future<void> loadPredefinedTags() {
    return _$loadPredefinedTagsAsyncAction.run(
      () => super.loadPredefinedTags(),
    );
  }

  late final _$createPredefinedTagAsyncAction = AsyncAction(
    '_TagsStore.createPredefinedTag',
    context: context,
  );

  @override
  Future<void> createPredefinedTag({
    required String name,
    required String color,
    required String category,
    required int displayOrder,
  }) {
    return _$createPredefinedTagAsyncAction.run(
      () => super.createPredefinedTag(
        name: name,
        color: color,
        category: category,
        displayOrder: displayOrder,
      ),
    );
  }

  late final _$deletePredefinedTagAsyncAction = AsyncAction(
    '_TagsStore.deletePredefinedTag',
    context: context,
  );

  @override
  Future<void> deletePredefinedTag(String id) {
    return _$deletePredefinedTagAsyncAction.run(
      () => super.deletePredefinedTag(id),
    );
  }

  late final _$loadSessionTagsAsyncAction = AsyncAction(
    '_TagsStore.loadSessionTags',
    context: context,
  );

  @override
  Future<List<SessionTag>> loadSessionTags(String sessionId) {
    return _$loadSessionTagsAsyncAction.run(
      () => super.loadSessionTags(sessionId),
    );
  }

  late final _$addSessionTagAsyncAction = AsyncAction(
    '_TagsStore.addSessionTag',
    context: context,
  );

  @override
  Future<void> addSessionTag(String sessionId, String tagName) {
    return _$addSessionTagAsyncAction.run(
      () => super.addSessionTag(sessionId, tagName),
    );
  }

  late final _$removeSessionTagAsyncAction = AsyncAction(
    '_TagsStore.removeSessionTag',
    context: context,
  );

  @override
  Future<void> removeSessionTag(String sessionId, String tagName) {
    return _$removeSessionTagAsyncAction.run(
      () => super.removeSessionTag(sessionId, tagName),
    );
  }

  late final _$bulkAddTagsAsyncAction = AsyncAction(
    '_TagsStore.bulkAddTags',
    context: context,
  );

  @override
  Future<void> bulkAddTags(List<String> sessionIds, List<String> tagNames) {
    return _$bulkAddTagsAsyncAction.run(
      () => super.bulkAddTags(sessionIds, tagNames),
    );
  }

  late final _$bulkRemoveTagsAsyncAction = AsyncAction(
    '_TagsStore.bulkRemoveTags',
    context: context,
  );

  @override
  Future<void> bulkRemoveTags(List<String> sessionIds, List<String> tagNames) {
    return _$bulkRemoveTagsAsyncAction.run(
      () => super.bulkRemoveTags(sessionIds, tagNames),
    );
  }

  late final _$loadSessionAnnotationsAsyncAction = AsyncAction(
    '_TagsStore.loadSessionAnnotations',
    context: context,
  );

  @override
  Future<List<SessionAnnotation>> loadSessionAnnotations(String sessionId) {
    return _$loadSessionAnnotationsAsyncAction.run(
      () => super.loadSessionAnnotations(sessionId),
    );
  }

  late final _$upsertSessionAnnotationAsyncAction = AsyncAction(
    '_TagsStore.upsertSessionAnnotation',
    context: context,
  );

  @override
  Future<void> upsertSessionAnnotation(
    String sessionId,
    String key,
    String value,
  ) {
    return _$upsertSessionAnnotationAsyncAction.run(
      () => super.upsertSessionAnnotation(sessionId, key, value),
    );
  }

  late final _$deleteSessionAnnotationAsyncAction = AsyncAction(
    '_TagsStore.deleteSessionAnnotation',
    context: context,
  );

  @override
  Future<void> deleteSessionAnnotation(String sessionId, String key) {
    return _$deleteSessionAnnotationAsyncAction.run(
      () => super.deleteSessionAnnotation(sessionId, key),
    );
  }

  @override
  String toString() {
    return '''
predefinedTags: ${predefinedTags},
sessionTagsCache: ${sessionTagsCache},
sessionAnnotationsCache: ${sessionAnnotationsCache},
loading: ${loading},
error: ${error}
    ''';
  }
}
