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
    test('loadPredefinedTags успешно загружает теги', () async {
      // Подготовка
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

      // Действие
      await store.loadPredefinedTags();

      // Проверка
      expect(store.predefinedTags.length, 2);
      expect(store.predefinedTags[0].name, 'Bug');
      expect(store.predefinedTags[1].name, 'Feature');
      expect(store.loading, false);
      expect(store.error, null);
    });

    test('loadPredefinedTags обрабатывает ошибки', () async {
      // Подготовка
      mockRepo.shouldThrowError = true;

      // Действие
      await store.loadPredefinedTags();

      // Проверка
      expect(store.predefinedTags.isEmpty, true);
      expect(store.loading, false);
      expect(store.error, isNotNull);
    });

    test('createPredefinedTag создает новый тег', () async {
      // Действие
      await store.createPredefinedTag(
        name: 'Critical',
        color: '#ff0000',
        category: 'priority',
        displayOrder: 1,
      );

      // Проверка
      expect(store.predefinedTags.length, 1);
      expect(store.predefinedTags[0].name, 'Critical');
    });

    test('deletePredefinedTag удаляет тег', () async {
      // Подготовка
      await store.createPredefinedTag(
        name: 'Bug',
        color: '#ff0000',
        category: 'issue',
        displayOrder: 1,
      );
      final tagId = store.predefinedTags[0].id;

      // Действие
      await store.deletePredefinedTag(tagId);

      // Проверка
      expect(store.predefinedTags.isEmpty, true);
    });
  });

  group('Session Tags', () {
    test('loadSessionTags загружает теги сессии', () async {
      // Подготовка
      const sessionId = 'session-1';
      await mockRepo.addSessionTag(sessionId, 'important');
      await mockRepo.addSessionTag(sessionId, 'bug');

      // Действие
      final tags = await store.loadSessionTags(sessionId);

      // Проверка
      expect(tags.length, 2);
      expect(tags[0].tagName, 'important');
      expect(tags[1].tagName, 'bug');
      expect(store.sessionTagsCache[sessionId]?.length, 2);
    });

    test('loadSessionTags обрабатывает ошибки', () async {
      // Подготовка
      mockRepo.shouldThrowError = true;

      // Действие
      final tags = await store.loadSessionTags('session-1');

      // Проверка
      expect(tags.isEmpty, true);
      expect(store.error, isNotNull);
    });

    test('addSessionTag добавляет тег к сессии', () async {
      // Подготовка
      const sessionId = 'session-1';

      // Действие
      await store.addSessionTag(sessionId, 'urgent');

      // Проверка
      final tags = store.sessionTagsCache[sessionId];
      expect(tags?.length, 1);
      expect(tags?.first.tagName, 'urgent');
    });

    test('removeSessionTag удаляет тег из сессии', () async {
      // Подготовка
      const sessionId = 'session-1';
      await store.addSessionTag(sessionId, 'urgent');

      // Действие
      await store.removeSessionTag(sessionId, 'urgent');

      // Проверка
      final tags = store.sessionTagsCache[sessionId];
      expect(tags?.isEmpty, true);
    });

    test('bulkAddTags добавляет теги к нескольким сессиям', () async {
      // Подготовка
      final sessionIds = ['s1', 's2', 's3'];
      final tagNames = ['bug', 'critical'];

      // Действие
      await store.bulkAddTags(sessionIds, tagNames);

      // Проверка
      for (final sessionId in sessionIds) {
        final tags = store.sessionTagsCache[sessionId];
        expect(tags?.length, 2);
      }
    });

    test('bulkRemoveTags удаляет теги из нескольких сессий', () async {
      // Подготовка
      final sessionIds = ['s1', 's2'];
      final tagNames = ['bug', 'critical'];
      await store.bulkAddTags(sessionIds, tagNames);

      // Действие
      await store.bulkRemoveTags(sessionIds, tagNames);

      // Проверка
      for (final sessionId in sessionIds) {
        final tags = store.sessionTagsCache[sessionId];
        expect(tags?.isEmpty, true);
      }
    });

    test('getSessionTagsFromCache возвращает теги из кеша', () {
      // Подготовка
      const sessionId = 'session-1';
      store.sessionTagsCache[sessionId] = [
        SessionTag(
          id: '1',
          sessionId: sessionId,
          tagName: 'cached-tag',
          createdAt: DateTime.now(),
        ),
      ];

      // Действие
      final tags = store.getSessionTagsFromCache(sessionId);

      // Проверка
      expect(tags.length, 1);
      expect(tags.first.tagName, 'cached-tag');
    });

    test('getSessionTagsFromCache возвращает пустой список если нет кеша', () {
      // Действие
      final tags = store.getSessionTagsFromCache('non-existent');

      // Проверка
      expect(tags.isEmpty, true);
    });
  });

  group('Session Annotations', () {
    test('loadSessionAnnotations загружает аннотации', () async {
      // Подготовка
      const sessionId = 'session-1';
      await mockRepo.upsertSessionAnnotation(sessionId, 'env', 'production');
      await mockRepo.upsertSessionAnnotation(sessionId, 'region', 'us-east-1');

      // Действие
      final annotations = await store.loadSessionAnnotations(sessionId);

      // Проверка
      expect(annotations.length, 2);
      expect(store.sessionAnnotationsCache[sessionId]?.length, 2);
    });

    test('upsertSessionAnnotation создает новую аннотацию', () async {
      // Подготовка
      const sessionId = 'session-1';

      // Действие
      await store.upsertSessionAnnotation(sessionId, 'env', 'staging');

      // Проверка
      final annotations = store.sessionAnnotationsCache[sessionId];
      expect(annotations?.length, 1);
      expect(annotations?.first.key, 'env');
      expect(annotations?.first.value, 'staging');
    });

    test('upsertSessionAnnotation обновляет существующую аннотацию', () async {
      // Подготовка
      const sessionId = 'session-1';
      await store.upsertSessionAnnotation(sessionId, 'env', 'production');

      // Действие - обновление
      await store.upsertSessionAnnotation(sessionId, 'env', 'staging');

      // Проверка
      final annotations = store.sessionAnnotationsCache[sessionId];
      expect(annotations?.length, 1);
      expect(annotations?.first.value, 'staging');
    });

    test('deleteSessionAnnotation удаляет аннотацию', () async {
      // Подготовка
      const sessionId = 'session-1';
      await store.upsertSessionAnnotation(sessionId, 'env', 'production');

      // Действие
      await store.deleteSessionAnnotation(sessionId, 'env');

      // Проверка
      final annotations = store.sessionAnnotationsCache[sessionId];
      expect(annotations?.isEmpty, true);
    });

    test('getSessionAnnotationsFromCache возвращает аннотации из кеша', () {
      // Подготовка
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

      // Действие
      final annotations = store.getSessionAnnotationsFromCache(sessionId);

      // Проверка
      expect(annotations.length, 1);
      expect(annotations.first.key, 'env');
    });

    test(
      'getSessionAnnotationsFromCache возвращает пустой список если нет кеша',
      () {
        // Действие
        final annotations = store.getSessionAnnotationsFromCache(
          'non-existent',
        );

        // Проверка
        expect(annotations.isEmpty, true);
      },
    );
  });

  group('Error handling', () {
    test('обработка ошибок при создании предопределенного тега', () async {
      // Подготовка
      mockRepo.shouldThrowError = true;

      // Действие & Проверка
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

    test('обработка ошибок при удалении предопределенного тега', () async {
      // Подготовка
      mockRepo.shouldThrowError = true;

      // Действие & Проверка
      try {
        await store.deletePredefinedTag('tag-1');
        fail('Expected exception');
      } catch (e) {
        expect(e, isException);
        expect(store.error, isNotNull);
      }
    });

    test('обработка ошибок при добавлении тега к сессии', () async {
      // Подготовка
      mockRepo.shouldThrowError = true;

      // Действие & Проверка
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
