import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/features/breakpoints/application/stores/breakpoints_store.dart';
import 'package:frontend/features/breakpoints/application/stores/intercept_queue_store.dart';
import 'package:frontend/features/breakpoints/application/stores/intercept_editor_store.dart';
import 'package:frontend/features/breakpoints/domain/entities/decisions.dart';
import 'package:frontend/features/breakpoints/domain/entities/intercept_config.dart';
import 'package:frontend/features/breakpoints/domain/entities/intercept_item.dart';
import 'package:frontend/features/breakpoints/domain/entities/intercept_rule.dart';
import 'package:frontend/features/breakpoints/domain/repositories/breakpoints_repository.dart';

void main() {
  group('Stores', () {
    late _Repo repo;

    setUp(() {
      repo = _Repo();
    });

    test('BreakpointsStore.load fills config and rules', () async {
      final s = BreakpointsStore(repo);
      await s.load();
      expect(s.config?.enabled, true);
      expect(s.rules.length, 1);
    });

    test('InterceptQueueStore refresh/select/quick actions', () async {
      final q = InterceptQueueStore(repo);
      await q.refresh();
      expect(q.items.length, 1);
      q.select('i1');
      expect(q.selected?.id, 'i1');
      await q.quickContinue('i1');
      expect(repo.continued, true);
      await q.quickCancel('i1');
      expect(repo.canceled, true);
    });

    test('InterceptEditorStore continue/cancel/validation', () async {
      final e = InterceptEditorStore(repo);
      e.setItem(_item());
      e.validateJson('{');
      expect(e.bodyJsonError.isNotEmpty, true);
      await e.continueRequest(method: 'POST');
      await e.continueResponse(status: 201);
      await e.cancel();
      expect(repo.requestMethod, 'POST');
      expect(repo.responseStatus, 201);
    });
  });
}

class _Repo implements BreakpointsRepository {
  bool continued = false;
  bool canceled = false;
  String? requestMethod;
  int? responseStatus;

  @override
  Future<void> cancel(String id) async {
    canceled = true;
  }

  @override
  Future<void> continueRequest(String id, RequestDecision decision) async {
    continued = true;
    requestMethod = decision.method;
  }

  @override
  Future<void> continueResponse(String id, ResponseDecision decision) async {
    continued = true;
    responseStatus = decision.status;
  }

  @override
  Future<InterceptConfig> getConfig() async => const InterceptConfig(
    enabled: true,
    requests: true,
    responses: true,
    timeoutMs: 1000,
    queueMax: 10,
    bodyMaxBytes: 1024,
    reencode: true,
    overflow: 'auto-continue-oldest',
  );

  @override
  Future<InterceptItem> getItem(String id) async => _item();

  @override
  Future<List<InterceptItem>> listPending({int? limit}) async => [_item()];

  @override
  Future<List<InterceptRule>> listRules() async => [
    InterceptRule(
      id: 'r1',
      enabled: true,
      priority: 0,
      action: 'both',
      once: false,
      stopProcessing: true,
      when: const InterceptWhen(),
    ),
  ];

  @override
  Future<void> replaceRules(List<InterceptRule> rules) async {}

  @override
  Future<void> setConfig(InterceptConfig cfg) async {}
}

InterceptItem _item() => InterceptItem(
  id: 'i1',
  createdAt: DateTime.now().toUtc(),
  deadline: DateTime.now().toUtc().add(const Duration(seconds: 1)),
  direction: 'request',
  sessionId: 's',
  state: 'PENDING',
  req: const HTTPRequestSnapshot(method: 'GET', url: 'https://e', headers: {}),
);
