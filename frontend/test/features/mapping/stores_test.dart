import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/features/mapping/application/stores/mapping_config_store.dart';
import 'package:frontend/features/mapping/application/stores/mapping_store.dart';
import 'package:frontend/features/mapping/domain/mapping_config.dart';
import 'package:frontend/features/mapping/domain/mapping_rule.dart';
import 'package:frontend/features/mapping/domain/repositories/mapping_repository.dart';

void main() {
  group('MappingStore', () {
    late _MockRepo repo;

    setUp(() {
      repo = _MockRepo();
    });

    test('load() fills rules list', () async {
      repo.listRulesFn = () async => [
            const MappingRule(id: 'r1', priority: 10),
            const MappingRule(id: 'r2', priority: 20),
          ];

      final store = MappingStore(repo);
      await store.load();

      expect(store.rules.length, 2);
      expect(store.loading, false);
      expect(store.lastError, isNull);
    });

    test('load() sets lastError on failure', () async {
      final store = MappingStore(_FailingRepo());

      try {
        await store.load();
        fail('Expected exception');
      } catch (_) {}

      expect(store.loading, false);
      expect(store.lastError, isNotNull);
      expect(store.lastError, contains('listRules failed'));
    });

    test('upsert() adds new rule', () async {
      repo.listRulesFn = () async => [
            const MappingRule(id: 'r1', priority: 10),
          ];
      repo.upsertFn = (rule) async =>
          rule.copyWith(id: 'r-new');

      final store = MappingStore(repo);
      await store.load();
      expect(store.rules.length, 1);

      await store.upsert(const MappingRule(id: '', priority: 50));
      expect(store.rules.length, 2);
      expect(store.rules.any((r) => r.id == 'r-new'), true);
    });

    test('upsert() updates existing rule', () async {
      repo.listRulesFn = () async => [
            const MappingRule(id: 'r1', priority: 10, hostPattern: 'old.com'),
          ];
      repo.upsertFn = (rule) async => rule;

      final store = MappingStore(repo);
      await store.load();
      expect(store.rules.length, 1);

      await store.upsert(
        const MappingRule(id: 'r1', priority: 10, hostPattern: 'new.com'),
      );

      expect(store.rules.length, 1);
      expect(store.rules.first.hostPattern, 'new.com');
    });

    test('upsert() sorts by priority', () async {
      repo.listRulesFn = () async => [
            const MappingRule(id: 'r1', priority: 100),
          ];
      repo.upsertFn = (rule) async => rule;

      final store = MappingStore(repo);
      await store.load();

      await store.upsert(
        const MappingRule(id: 'r2', priority: 5),
      );

      expect(store.rules.length, 2);
      expect(store.rules.first.id, 'r2');
      expect(store.rules.last.id, 'r1');
    });

    test('delete() removes rule', () async {
      repo.listRulesFn = () async => [
            const MappingRule(id: 'r1', priority: 10),
            const MappingRule(id: 'r2', priority: 20),
          ];
      repo.deleteFn = (id) async {};

      final store = MappingStore(repo);
      await store.load();
      expect(store.rules.length, 2);

      await store.delete('r1');
      expect(store.rules.length, 1);
      expect(store.rules.first.id, 'r2');
    });

    test('delete() with unknown id still calls repo', () async {
      repo.listRulesFn = () async => [
            const MappingRule(id: 'r1', priority: 10),
          ];
      var deleteCalled = false;
      repo.deleteFn = (id) async {
        deleteCalled = true;
      };

      final store = MappingStore(repo);
      await store.load();

      await store.delete('unknown-id');
      expect(deleteCalled, true);
      expect(store.rules.length, 1);
    });

    test('reorder() calls repo then reloads', () async {
      var reorderCalled = false;
      var listCallCount = 0;

      repo.reorderFn = (ids) async {
        reorderCalled = true;
      };
      repo.listRulesFn = () async {
        listCallCount++;
        return [
          const MappingRule(id: 'r1', priority: 10),
          const MappingRule(id: 'r2', priority: 20),
        ];
      };

      final store = MappingStore(repo);
      await store.load();
      expect(listCallCount, 1);

      await store.reorder(['r2', 'r1']);
      expect(reorderCalled, true);
      expect(listCallCount, 2);
    });
  });

  group('MappingConfigStore', () {
    late _MockRepo repo;

    setUp(() {
      repo = _MockRepo();
    });

    test('load() fills config', () async {
      repo.getConfigFn = () async =>
          const MappingConfig(enabled: true, uploadMaxMB: 50);

      final store = MappingConfigStore(repo);
      await store.load();

      expect(store.config, isNotNull);
      expect(store.config!.enabled, true);
      expect(store.config!.uploadMaxMB, 50);
      expect(store.loading, false);
      expect(store.lastError, isNull);
    });

    test('load() sets lastError on failure', () async {
      final store = MappingConfigStore(_FailingRepo());

      try {
        await store.load();
        fail('Expected exception');
      } catch (_) {}

      expect(store.loading, false);
      expect(store.lastError, isNotNull);
      expect(store.lastError, contains('getConfig failed'));
      expect(store.config, isNull);
    });

    test('save() updates local config', () async {
      repo.getConfigFn = () async =>
          const MappingConfig(enabled: true, uploadMaxMB: 20);
      repo.setConfigFn = (cfg) async => cfg;

      final store = MappingConfigStore(repo);
      await store.load();
      expect(store.config!.enabled, true);

      const updated = MappingConfig(enabled: false, uploadMaxMB: 100);
      await store.save(updated);

      expect(store.config!.enabled, false);
      expect(store.config!.uploadMaxMB, 100);
    });
  });

  group('MappingRule entity', () {
    test('fromJson parses correctly', () {
      final json = <String, dynamic>{
        'id': 'rule-1',
        'enabled': true,
        'priority': 42,
        'kind': 'local',
        'stopProcessing': false,
        'methods': ['GET', 'POST'],
        'hostPattern': '*.example.com',
        'pathPattern': '/api/*',
        'patternType': 'glob',
        'filePath': '/tmp/response.json',
        'blobPath': null,
        'statusOverride': 201,
        'contentTypeOverride': 'application/json',
        'targetURLTemplate': null,
        'preserveHost': true,
      };

      final rule = MappingRule.fromJson(json);

      expect(rule.id, 'rule-1');
      expect(rule.enabled, true);
      expect(rule.priority, 42);
      expect(rule.kind, 'local');
      expect(rule.stopProcessing, false);
      expect(rule.methods, ['GET', 'POST']);
      expect(rule.hostPattern, '*.example.com');
      expect(rule.pathPattern, '/api/*');
      expect(rule.patternType, 'glob');
      expect(rule.filePath, '/tmp/response.json');
      expect(rule.blobPath, isNull);
      expect(rule.statusOverride, 201);
      expect(rule.contentTypeOverride, 'application/json');
      expect(rule.targetURLTemplate, isNull);
      expect(rule.preserveHost, true);
    });

    test('toJson serializes remote rule correctly', () {
      const rule = MappingRule(
        id: 'rule-1',
        enabled: true,
        priority: 10,
        kind: 'remote',
        methods: ['GET'],
        hostPattern: 'example.com',
        pathPattern: '/api',
      );

      final json = rule.toJson();

      expect(json['id'], 'rule-1');
      expect(json['enabled'], true);
      expect(json['priority'], 10);
      expect(json['kind'], 'remote');
      expect(json['methods'], ['GET']);
      expect(json['hostPattern'], 'example.com');
      expect(json['pathPattern'], '/api');
      expect(json['targetURLTemplate'], '');
      expect(json['preserveHost'], false);
      expect(json.containsKey('filePath'), false);
      expect(json.containsKey('blobPath'), false);
      expect(json.containsKey('statusOverride'), false);
      expect(json.containsKey('contentTypeOverride'), false);
      expect(json.containsKey('updatedAt'), false);
    });

    test('toJson serializes local rule correctly', () {
      const rule = MappingRule(
        id: 'rule-2',
        enabled: true,
        priority: 10,
        kind: 'local',
        methods: ['GET'],
        hostPattern: 'example.com',
        pathPattern: '/api',
        filePath: '/tmp/file.json',
        statusOverride: 201,
        contentTypeOverride: 'application/json',
      );

      final json = rule.toJson();

      expect(json['id'], 'rule-2');
      expect(json['kind'], 'local');
      expect(json['filePath'], '/tmp/file.json');
      expect(json['statusOverride'], 201);
      expect(json['contentTypeOverride'], 'application/json');
      expect(json.containsKey('targetURLTemplate'), false);
      expect(json.containsKey('preserveHost'), false);
    });

    test('toJson includes updatedAt when present', () {
      final now = DateTime.utc(2025, 1, 15, 10, 30);
      final rule = MappingRule(
        id: 'rule-3',
        kind: 'local',
        updatedAt: now,
      );

      final json = rule.toJson();
      expect(json['updatedAt'], '2025-01-15T10:30:00.000Z');
    });

    test('copyWith works', () {
      const rule = MappingRule(
        id: 'rule-1',
        priority: 10,
        hostPattern: 'old.com',
      );

      final updated = rule.copyWith(
        priority: 99,
        hostPattern: 'new.com',
        statusOverride: 404,
      );

      expect(updated.id, 'rule-1');
      expect(updated.priority, 99);
      expect(updated.hostPattern, 'new.com');
      expect(updated.statusOverride, 404);
      // Original unchanged
      expect(rule.priority, 10);
      expect(rule.hostPattern, 'old.com');
      expect(rule.statusOverride, isNull);
    });

    test('equality works', () {
      const a = MappingRule(
        id: 'rule-1',
        enabled: true,
        priority: 10,
        kind: 'local',
        methods: ['GET'],
      );
      const b = MappingRule(
        id: 'rule-1',
        enabled: true,
        priority: 10,
        kind: 'local',
        methods: ['GET'],
      );

      expect(a, equals(b));
      expect(a.hashCode, b.hashCode);

      final c = a.copyWith(priority: 20);
      expect(a, isNot(equals(c)));
    });
  });
}

// ---------------------------------------------------------------------------
// Mock repository with configurable function fields
// ---------------------------------------------------------------------------

class _MockRepo implements MappingRepository {
  Future<MappingConfig> Function() getConfigFn =
      () async => const MappingConfig();

  Future<MappingConfig> Function(MappingConfig cfg) setConfigFn =
      (cfg) async => cfg;

  Future<List<MappingRule>> Function() listRulesFn = () async => [];

  Future<MappingRule> Function(MappingRule rule) upsertFn =
      (rule) async => rule;

  Future<void> Function(List<String> ids) reorderFn = (_) async {};

  Future<void> Function(String id) deleteFn = (_) async {};

  @override
  Future<MappingConfig> getConfig() => getConfigFn();

  @override
  Future<MappingConfig> setConfig(MappingConfig cfg) => setConfigFn(cfg);

  @override
  Future<List<MappingRule>> listRules() => listRulesFn();

  @override
  Future<MappingRule> upsert(MappingRule rule) => upsertFn(rule);

  @override
  Future<void> reorder(List<String> ids) => reorderFn(ids);

  @override
  Future<void> delete(String id) => deleteFn(id);

  @override
  Future<Map<String, dynamic>> uploadFile(
      String fileName, List<int> bytes) async {
    return {};
  }

  @override
  Future<Map<String, dynamic>> uploadStream({
    required String fileName,
    required Stream<List<int>> stream,
    required int length,
  }) async {
    return {};
  }
}

// ---------------------------------------------------------------------------
// Failing repository that throws on all methods
// ---------------------------------------------------------------------------

class _FailingRepo implements MappingRepository {
  @override
  Future<MappingConfig> getConfig() async {
    throw Exception('getConfig failed');
  }

  @override
  Future<MappingConfig> setConfig(MappingConfig cfg) async {
    throw Exception('setConfig failed');
  }

  @override
  Future<List<MappingRule>> listRules() async {
    throw Exception('listRules failed');
  }

  @override
  Future<MappingRule> upsert(MappingRule rule) async {
    throw Exception('upsert failed');
  }

  @override
  Future<void> reorder(List<String> ids) async {
    throw Exception('reorder failed');
  }

  @override
  Future<void> delete(String id) async {
    throw Exception('delete failed');
  }

  @override
  Future<Map<String, dynamic>> uploadFile(
      String fileName, List<int> bytes) async {
    throw Exception('uploadFile failed');
  }

  @override
  Future<Map<String, dynamic>> uploadStream({
    required String fileName,
    required Stream<List<int>> stream,
    required int length,
  }) async {
    throw Exception('uploadStream failed');
  }
}
