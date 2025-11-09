import '../domain/entities/predefined_tag.dart';
import '../domain/entities/session_annotation.dart';
import '../domain/entities/session_tag.dart';
import '../domain/repositories/tags_repository.dart';
import 'tags_api.dart';

class TagsRepositoryImpl implements TagsRepository {
  TagsRepositoryImpl(this._api);
  final TagsApi _api;

  @override
  Future<List<PredefinedTag>> listPredefinedTags() async {
    final list = await _api.listPredefinedTags();
    return list
        .whereType<Map<String, dynamic>>()
        .map(PredefinedTag.fromJson)
        .toList();
  }

  @override
  Future<void> createPredefinedTag({
    required String name,
    required String color,
    required String category,
    required int displayOrder,
  }) async {
    await _api.createPredefinedTag({
      'name': name,
      'color': color,
      'category': category,
      'displayOrder': displayOrder,
    });
  }

  @override
  Future<void> deletePredefinedTag(String id) async {
    await _api.deletePredefinedTag(id);
  }

  @override
  Future<List<SessionTag>> getSessionTags(String sessionId) async {
    final list = await _api.getSessionTags(sessionId);
    return list
        .whereType<Map<String, dynamic>>()
        .map(SessionTag.fromJson)
        .toList();
  }

  @override
  Future<void> addSessionTag(String sessionId, String tagName) async {
    await _api.addSessionTag(sessionId, tagName);
  }

  @override
  Future<void> removeSessionTag(String sessionId, String tagName) async {
    await _api.removeSessionTag(sessionId, tagName);
  }

  @override
  Future<void> bulkAddTags(
    List<String> sessionIds,
    List<String> tagNames,
  ) async {
    await _api.bulkTags({
      'operation': 'add',
      'sessionIds': sessionIds,
      'tagNames': tagNames,
    });
  }

  @override
  Future<void> bulkRemoveTags(
    List<String> sessionIds,
    List<String> tagNames,
  ) async {
    await _api.bulkTags({
      'operation': 'remove',
      'sessionIds': sessionIds,
      'tagNames': tagNames,
    });
  }

  @override
  Future<List<SessionAnnotation>> getSessionAnnotations(
    String sessionId,
  ) async {
    final list = await _api.getSessionAnnotations(sessionId);
    return list
        .whereType<Map<String, dynamic>>()
        .map(SessionAnnotation.fromJson)
        .toList();
  }

  @override
  Future<void> upsertSessionAnnotation(
    String sessionId,
    String key,
    String value,
  ) async {
    await _api.upsertSessionAnnotation(sessionId, key, value);
  }

  @override
  Future<void> deleteSessionAnnotation(String sessionId, String key) async {
    await _api.deleteSessionAnnotation(sessionId, key);
  }
}
