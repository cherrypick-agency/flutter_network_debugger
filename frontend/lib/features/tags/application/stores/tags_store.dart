import 'package:flutter/foundation.dart';
import 'package:mobx/mobx.dart';

import '../../../../core/network/error_utils.dart';
import '../../domain/entities/predefined_tag.dart';
import '../../domain/entities/session_annotation.dart';
import '../../domain/entities/session_tag.dart';
import '../../domain/repositories/tags_repository.dart';

part 'tags_store.g.dart';

class TagsStore = _TagsStore with _$TagsStore;

abstract class _TagsStore with Store {
  _TagsStore(this._repository);
  final TagsRepository _repository;

  @observable
  ObservableList<PredefinedTag> predefinedTags = ObservableList.of([]);

  @observable
  ObservableMap<String, List<SessionTag>> sessionTagsCache =
      ObservableMap<String, List<SessionTag>>.of({});

  @observable
  ObservableMap<String, List<SessionAnnotation>> sessionAnnotationsCache =
      ObservableMap<String, List<SessionAnnotation>>.of({});

  @observable
  bool loading = false;

  @observable
  String? error;

  @action
  Future<void> loadPredefinedTags() async {
    loading = true;
    error = null;
    try {
      final tags = await _repository.listPredefinedTags();
      predefinedTags = ObservableList.of(tags);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error loading predefined tags: $e\n$st');
      }
    } finally {
      loading = false;
    }
  }

  @action
  Future<void> createPredefinedTag({
    required String name,
    required String color,
    required String category,
    required int displayOrder,
  }) async {
    try {
      await _repository.createPredefinedTag(
        name: name,
        color: color,
        category: category,
        displayOrder: displayOrder,
      );
      await loadPredefinedTags();
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error creating predefined tag: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<void> deletePredefinedTag(String id) async {
    try {
      await _repository.deletePredefinedTag(id);
      predefinedTags.removeWhere((t) => t.id == id);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error deleting predefined tag: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<List<SessionTag>> loadSessionTags(String sessionId) async {
    try {
      final tags = await _repository.getSessionTags(sessionId);
      sessionTagsCache[sessionId] = tags;
      return tags;
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error loading session tags: $e\n$st');
      }
      return [];
    }
  }

  @action
  Future<void> addSessionTag(String sessionId, String tagName) async {
    try {
      await _repository.addSessionTag(sessionId, tagName);
      await loadSessionTags(sessionId);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error adding session tag: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<void> removeSessionTag(String sessionId, String tagName) async {
    try {
      await _repository.removeSessionTag(sessionId, tagName);
      await loadSessionTags(sessionId);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error removing session tag: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<void> bulkAddTags(
    List<String> sessionIds,
    List<String> tagNames,
  ) async {
    try {
      await _repository.bulkAddTags(sessionIds, tagNames);
      for (final sessionId in sessionIds) {
        await loadSessionTags(sessionId);
      }
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error bulk adding tags: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<void> bulkRemoveTags(
    List<String> sessionIds,
    List<String> tagNames,
  ) async {
    try {
      await _repository.bulkRemoveTags(sessionIds, tagNames);
      for (final sessionId in sessionIds) {
        await loadSessionTags(sessionId);
      }
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error bulk removing tags: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<List<SessionAnnotation>> loadSessionAnnotations(
    String sessionId,
  ) async {
    try {
      final annotations = await _repository.getSessionAnnotations(sessionId);
      sessionAnnotationsCache[sessionId] = annotations;
      return annotations;
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error loading session annotations: $e\n$st');
      }
      return [];
    }
  }

  @action
  Future<void> upsertSessionAnnotation(
    String sessionId,
    String key,
    String value,
  ) async {
    try {
      await _repository.upsertSessionAnnotation(sessionId, key, value);
      await loadSessionAnnotations(sessionId);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error upserting session annotation: $e\n$st');
      }
      rethrow;
    }
  }

  @action
  Future<void> deleteSessionAnnotation(String sessionId, String key) async {
    try {
      await _repository.deleteSessionAnnotation(sessionId, key);
      await loadSessionAnnotations(sessionId);
    } catch (e, st) {
      error = resolveErrorMessage(e, st).description;
      if (kDebugMode) {
        debugPrint('Error deleting session annotation: $e\n$st');
      }
      rethrow;
    }
  }

  List<SessionTag> getSessionTagsFromCache(String sessionId) {
    return sessionTagsCache[sessionId] ?? [];
  }

  List<SessionAnnotation> getSessionAnnotationsFromCache(String sessionId) {
    return sessionAnnotationsCache[sessionId] ?? [];
  }
}
