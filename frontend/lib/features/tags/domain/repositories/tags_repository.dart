import '../entities/predefined_tag.dart';
import '../entities/session_annotation.dart';
import '../entities/session_tag.dart';

abstract class TagsRepository {
  // Predefined Tags
  Future<List<PredefinedTag>> listPredefinedTags();
  Future<void> createPredefinedTag({
    required String name,
    required String color,
    required String category,
    required int displayOrder,
  });
  Future<void> deletePredefinedTag(String id);

  // Session Tags
  Future<List<SessionTag>> getSessionTags(String sessionId);
  Future<void> addSessionTag(String sessionId, String tagName);
  Future<void> removeSessionTag(String sessionId, String tagName);
  Future<void> bulkAddTags(List<String> sessionIds, List<String> tagNames);
  Future<void> bulkRemoveTags(List<String> sessionIds, List<String> tagNames);

  // Session Annotations
  Future<List<SessionAnnotation>> getSessionAnnotations(String sessionId);
  Future<void> upsertSessionAnnotation(
    String sessionId,
    String key,
    String value,
  );
  Future<void> deleteSessionAnnotation(String sessionId, String key);
}
