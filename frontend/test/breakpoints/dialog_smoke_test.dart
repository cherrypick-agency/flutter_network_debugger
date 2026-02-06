import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:get_it/get_it.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:frontend/features/breakpoints/presentation/widgets/breakpoints_dialog.dart';
import 'package:frontend/features/breakpoints/application/stores/breakpoints_store.dart';
import 'package:frontend/features/breakpoints/application/stores/intercept_queue_store.dart';
import 'package:frontend/features/breakpoints/application/stores/intercept_editor_store.dart';
import 'package:frontend/features/breakpoints/domain/repositories/breakpoints_repository.dart';
import 'package:frontend/features/breakpoints/domain/entities/intercept_config.dart';
import 'package:frontend/features/breakpoints/domain/entities/intercept_rule.dart';
import 'package:frontend/features/breakpoints/domain/entities/intercept_item.dart';
import 'package:frontend/core/hotkeys/hotkeys_service.dart';
import 'package:frontend/core/notifications/notifications_service.dart';
import 'package:frontend/features/breakpoints/domain/entities/decisions.dart';
import 'package:frontend/features/inspector/application/services/monitor_service.dart';
import 'package:frontend/theme/app_theme.dart';

void main() {
  testWidgets('BreakpointsDialog smoke', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final sl = GetIt.instance;
    await sl.reset();
    // Repo fake
    sl.registerLazySingleton<BreakpointsRepository>(() => _Repo());
    sl.registerLazySingleton<MonitorService>(() => _FakeMonitor());
    // Stores
    sl.registerLazySingleton<BreakpointsStore>(
      () => BreakpointsStore(sl<BreakpointsRepository>()),
    );
    sl.registerLazySingleton<InterceptQueueStore>(
      () => InterceptQueueStore(sl<BreakpointsRepository>()),
    );
    sl.registerLazySingleton<InterceptEditorStore>(
      () => InterceptEditorStore(sl<BreakpointsRepository>()),
    );
    // Notifications
    sl.registerLazySingleton<NotificationsService>(() => NotificationsService());
    // Hotkeys
    final hk = HotkeysService();
    await hk.init();
    sl.registerSingleton<HotkeysService>(hk);

    await tester.pumpWidget(
      MaterialApp(
        theme: buildLightTheme(),
        home: const Scaffold(body: BreakpointsDialog()),
      ),
    );
    // initial frame
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.text('Breakpoints'), findsOneWidget);
    expect(find.text('Queue'), findsOneWidget);
    expect(find.text('Editor'), findsOneWidget);
    expect(find.text('Rules'), findsOneWidget);
  });
}

class _Repo implements BreakpointsRepository {
  @override
  Future<void> cancel(String id) async {}
  @override
  Future<void> continueRequest(String id, RequestDecision decision) async {}
  @override
  Future<void> continueResponse(String id, ResponseDecision decision) async {}
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

class _FakeMonitor extends MonitorService {
  @override
  void addListener(MonitorListener l) {}
  @override
  void removeListener(MonitorListener l) {}
  @override
  Future<void> connect() async {}
}
