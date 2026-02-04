import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/features/tags/application/stores/tags_store.dart';
import 'package:frontend/features/tags/domain/entities/predefined_tag.dart';
import 'package:frontend/features/tags/domain/entities/session_tag.dart';
import 'package:frontend/features/tags/domain/entities/session_annotation.dart';
import 'package:frontend/features/tags/domain/repositories/tags_repository.dart';

// Mock repository for testing
class MockTagsRepository implements TagsRepository {
  final List<PredefinedTag> _predefinedTags = [];
  final Map<String, List<SessionTag>> _sessionTags = {};
  final Map<String, List<SessionAnnotation>> _annotations = {};
  bool shouldThrowError = false;

  @override
  Future<List<PredefinedTag>> listPredefinedTags() async {
    if (shouldThrowError) throw Exception('Mock error');
    return List.from(_predefinedTags);
  }

  @override
  Future<void> createPredefinedTag({
    required String name,
    required String color,
    required String category,
    required int displayOrder,
  }) async {
    if (shouldThrowError) throw Exception('Mock error');
    _predefinedTags.add(
      PredefinedTag(
        id: 'tag-${_predefinedTags.length + 1}',
        name: name,
        color: color,
        category: category,
        displayOrder: displayOrder,
        createdAt: DateTime.now(),
      ),
    );
  }

  @override
  Future<void> deletePredefinedTag(String id) async {
    if (shouldThrowError) throw Exception('Mock error');
    _predefinedTags.removeWhere((t) => t.id == id);
  }

  @override
  Future<List<SessionTag>> getSessionTags(String sessionId) async {
    if (shouldThrowError) throw Exception('Mock error');
    return _sessionTags[sessionId] ?? [];
  }

  @override
  Future<void> addSessionTag(String sessionId, String tagName) async {
    if (shouldThrowError) throw Exception('Mock error');
    final tags = _sessionTags[sessionId] ?? [];
    tags.add(
      SessionTag(
        id: 'tag-${tags.length + 1}',
        sessionId: sessionId,
        tagName: tagName,
        createdAt: DateTime.now(),
      ),
    );
    _sessionTags[sessionId] = tags;
  }

  @override
  Future<void> removeSessionTag(String sessionId, String tagName) async {
    if (shouldThrowError) throw Exception('Mock error');
    final tags = _sessionTags[sessionId];
    if (tags != null) {
      tags.removeWhere((t) => t.tagName == tagName);
    }
  }

  @override
  Future<void> bulkAddTags(
    List<String> sessionIds,
    List<String> tagNames,
  ) async {
    if (shouldThrowError) throw Exception('Mock error');
    for (final sessionId in sessionIds) {
      for (final tagName in tagNames) {
        await addSessionTag(sessionId, tagName);
      }
    }
  }

  @override
  Future<void> bulkRemoveTags(
    List<String> sessionIds,
    List<String> tagNames,
  ) async {
    if (shouldThrowError) throw Exception('Mock error');
    for (final sessionId in sessionIds) {
      for (final tagName in tagNames) {
        await removeSessionTag(sessionId, tagName);
      }
    }
  }

  @override
  Future<List<SessionAnnotation>> getSessionAnnotations(
    String sessionId,
  ) async {
    if (shouldThrowError) throw Exception('Mock error');
    return _annotations[sessionId] ?? [];
  }

  @override
  Future<void> upsertSessionAnnotation(
    String sessionId,
    String key,
    String value,
  ) async {
    if (shouldThrowError) throw Exception('Mock error');
    final annotations = _annotations[sessionId] ?? [];
    final existingIndex = annotations.indexWhere((a) => a.key == key);

    if (existingIndex >= 0) {
      annotations[existingIndex] = SessionAnnotation(
        id: annotations[existingIndex].id,
        sessionId: sessionId,
        key: key,
        value: value,
        createdAt: annotations[existingIndex].createdAt,
        updatedAt: DateTime.now(),
      );
    } else {
      annotations.add(
        SessionAnnotation(
          id: 'ann-${annotations.length + 1}',
          sessionId: sessionId,
          key: key,
          value: value,
          createdAt: DateTime.now(),
          updatedAt: DateTime.now(),
        ),
      );
    }
    _annotations[sessionId] = annotations;
  }

  @override
  Future<void> deleteSessionAnnotation(String sessionId, String key) async {
    if (shouldThrowError) throw Exception('Mock error');
    final annotations = _annotations[sessionId];
    if (annotations != null) {
      annotations.removeWhere((a) => a.key == key);
    }
  }
}

void main() {
  late TagsStore store;
  late MockTagsRepository mockRepo;

  setUp(() {
    mockRepo = MockTagsRepository();
    store = TagsStore(mockRepo);
  });

  group('Predefined Tags', () {
    test('loadPredefinedTags successfully loads tags', () async {
      // Setup
      await mockRepo.createPredefinedTag(
        name: 'Bug',
        color: '#ff0000',
        category: 'issue',
        displayOrder: 1,
      );
      await mockRepo.createPredefinedTag(
        name: 'Feature',
        color: '#00ff00',
        category: 'type',
        displayOrder: 2,
      );

      // Action
      await store.loadPredefinedTags();

      // Verification
      expect(store.predefinedTags.length, 2);
      expect(store.predefinedTags[0].name, 'Bug');
      expect(store.predefinedTags[1].name, 'Feature');
      expect(store.loading, false);
      expect(store.error, null);
    });

    test('loadPredefinedTags handles errors', () async {
      // Setup
      mockRepo.shouldThrowError = true;

      // Action
      await store.loadPredefinedTags();

      // Verification
      expect(store.predefinedTags.isEmpty, true);
      expect(store.loading, false);
      expect(store.error, isNotNull);
    });

    test('createPredefinedTag creates new tag', () async {
      // Action
      await store.createPredefinedTag(
        name: 'Critical',
        color: '#ff0000',
        category: 'priority',
        displayOrder: 1,
      );

      // Verification
      expect(store.predefinedTags.length, 1);
      expect(store.predefinedTags[0].name, 'Critical');
    });

    test('deletePredefinedTag deletes tag', () async {
      // Setup
      await store.createPredefinedTag(
        name: 'Bug',
        color: '#ff0000',
        category: 'issue',
        displayOrder: 1,
      );
      final tagId = store.predefinedTags[0].id;

      // Action
      await store.deletePredefinedTag(tagId);

      // Verification
      expect(store.predefinedTags.isEmpty, true);
    });
  });

  group('Session Tags', () {
    test('loadSessionTags loads session tags', () async {
      // Setup
      const sessionId = 'session-1';
      await mockRepo.addSessionTag(sessionId, 'important');
      await mockRepo.addSessionTag(sessionId, 'bug');

      // Action
      final tags = await store.loadSessionTags(sessionId);

      // Verification
      expect(tags.length, 2);
      expect(tags[0].tagName, 'important');
      expect(tags[1].tagName, 'bug');
      expect(store.sessionTagsCache[sessionId]?.length, 2);
    });

    test('loadSessionTags handles errors', () async {
      // Setup
      mockRepo.shouldThrowError = true;

      // Action
      final tags = await store.loadSessionTags('session-1');

      // Verification
      expect(tags.isEmpty, true);
      expect(store.error, isNotNull);
    });

    test('addSessionTag adds tag to session', () async {
      // Setup
      const sessionId = 'session-1';

      // Action
      await store.addSessionTag(sessionId, 'urgent');

      // Verification
      final tags = store.sessionTagsCache[sessionId];
      expect(tags?.length, 1);
      expect(tags?.first.tagName, 'urgent');
    });

    test('removeSessionTag removes tag from session', () async {
      // Setup
      const sessionId = 'session-1';
      await store.addSessionTag(sessionId, 'urgent');

      // Action
      await store.removeSessionTag(sessionId, 'urgent');

      // Verification
      final tags = store.sessionTagsCache[sessionId];
      expect(tags?.isEmpty, true);
    });

    test('bulkAddTags adds tags to multiple sessions', () async {
      // Setup
      final sessionIds = ['s1', 's2', 's3'];
      final tagNames = ['bug', 'critical'];

      // Action
      await store.bulkAddTags(sessionIds, tagNames);

      // Verification
      for (final sessionId in sessionIds) {
        final tags = store.sessionTagsCache[sessionId];
        expect(tags?.length, 2);
      }
    });

    test('bulkRemoveTags removes tags from multiple sessions', () async {
      // Setup
      final sessionIds = ['s1', 's2'];
      final tagNames = ['bug', 'critical'];
      await store.bulkAddTags(sessionIds, tagNames);

      // Action
      await store.bulkRemoveTags(sessionIds, tagNames);

      // Verification
      for (final sessionId in sessionIds) {
        final tags = store.sessionTagsCache[sessionId];
        expect(tags?.isEmpty, true);
      }
    });

    test('getSessionTagsFromCache returns tags from cache', () {
      // Setup
      const sessionId = 'session-1';
      store.sessionTagsCache[sessionId] = [
        SessionTag(
          id: '1',
          sessionId: sessionId,
          tagName: 'cached-tag',
          createdAt: DateTime.now(),
        ),
      ];

      // Action
      final tags = store.getSessionTagsFromCache(sessionId);

      // Verification
      expect(tags.length, 1);
      expect(tags.first.tagName, 'cached-tag');
    });

    test('getSessionTagsFromCache returns empty list if no cache', () {
      // Action
      final tags = store.getSessionTagsFromCache('non-existent');

      // Verification
      expect(tags.isEmpty, true);
    });
  });

  group('Session Annotations', () {
    test('loadSessionAnnotations loads annotations', () async {
      // Setup
      const sessionId = 'session-1';
      await mockRepo.upsertSessionAnnotation(sessionId, 'env', 'production');
      await mockRepo.upsertSessionAnnotation(sessionId, 'region', 'us-east-1');

      // Action
      final annotations = await store.loadSessionAnnotations(sessionId);

      // Verification
      expect(annotations.length, 2);
      expect(store.sessionAnnotationsCache[sessionId]?.length, 2);
    });

    test('upsertSessionAnnotation creates new annotation', () async {
      // Setup
      const sessionId = 'session-1';

      // Action
      await store.upsertSessionAnnotation(sessionId, 'env', 'staging');

      // Verification
      final annotations = store.sessionAnnotationsCache[sessionId];
      expect(annotations?.length, 1);
      expect(annotations?.first.key, 'env');
      expect(annotations?.first.value, 'staging');
    });

    test('upsertSessionAnnotation updates existing annotation', () async {
      // Setup
      const sessionId = 'session-1';
      await store.upsertSessionAnnotation(sessionId, 'env', 'production');

      // Action - update
      await store.upsertSessionAnnotation(sessionId, 'env', 'staging');

      // Verification
      final annotations = store.sessionAnnotationsCache[sessionId];
      expect(annotations?.length, 1);
      expect(annotations?.first.value, 'staging');
    });

    test('deleteSessionAnnotation deletes annotation', () async {
      // Setup
      const sessionId = 'session-1';
      await store.upsertSessionAnnotation(sessionId, 'env', 'production');

      // Action
      await store.deleteSessionAnnotation(sessionId, 'env');

      // Verification
      final annotations = store.sessionAnnotationsCache[sessionId];
      expect(annotations?.isEmpty, true);
    });

    test('getSessionAnnotationsFromCache returns annotations from cache', () {
      // Setup
      const sessionId = 'session-1';
      store.sessionAnnotationsCache[sessionId] = [
        SessionAnnotation(
          id: '1',
          sessionId: sessionId,
          key: 'env',
          value: 'test',
          createdAt: DateTime.now(),
          updatedAt: DateTime.now(),
        ),
      ];

      // Action
      final annotations = store.getSessionAnnotationsFromCache(sessionId);

      // Verification
      expect(annotations.length, 1);
      expect(annotations.first.key, 'env');
    });

    test('getSessionAnnotationsFromCache returns empty list if no cache', () {
      // Action
      final annotations = store.getSessionAnnotationsFromCache('non-existent');

      // Verification
      expect(annotations.isEmpty, true);
    });
  });

  group('Error handling', () {
    test('error handling when creating predefined tag', () async {
      // Setup
      mockRepo.shouldThrowError = true;

      // Action & Verification
      try {
        await store.createPredefinedTag(
          name: 'Bug',
          color: '#ff0000',
          category: 'issue',
          displayOrder: 1,
        );
        fail('Expected exception');
      } catch (e) {
        expect(e, isException);
        expect(store.error, isNotNull);
      }
    });

    test('error handling when deleting predefined tag', () async {
      // Setup
      mockRepo.shouldThrowError = true;

      // Action & Verification
      try {
        await store.deletePredefinedTag('tag-1');
        fail('Expected exception');
      } catch (e) {
        expect(e, isException);
        expect(store.error, isNotNull);
      }
    });

    test('error handling when adding tag to session', () async {
      // Setup
      mockRepo.shouldThrowError = true;

      // Action & Verification
      try {
        await store.addSessionTag('session-1', 'tag');
        fail('Expected exception');
      } catch (e) {
        expect(e, isException);
        expect(store.error, isNotNull);
      }
    });
  });
}
