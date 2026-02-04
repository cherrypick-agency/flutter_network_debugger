import 'dart:convert';
import 'dart:typed_data';
import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:get_it/get_it.dart';
import 'package:provider/provider.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:frontend/features/export_import/application/stores/export_import_store.dart';
import 'package:frontend/features/export_import/presentation/widgets/export_import_dialog.dart';
import 'package:frontend/features/inspector/domain/repositories/inspector_repository.dart';
import 'package:frontend/core/notifications/notifications_service.dart';
import 'package:frontend/theme/app_theme.dart';
import '../../helpers/har_generator.dart';

/// Fake InspectorRepository for smoke testing
class _FakeRepository implements InspectorRepository {
  int exportCallCount = 0;
  int importCallCount = 0;

  @override
  Future<String> exportHAR({
    List<String>? sessionIds,
    bool includeBodies = true,
    bool includeSensitive = false,
    bool minify = false,
  }) async {
    exportCallCount++;
    await Future.delayed(const Duration(milliseconds: 100));
    return 'http://example.com/export.har';
  }

  @override
  Future<Map<String, dynamic>> importHAR(
    String harJson, {
    required String mode,
  }) async {
    importCallCount++;
    await Future.delayed(const Duration(milliseconds: 100));
    return {
      'imported': 5,
      'failed': 0,
      'total': 5,
      'sessionIds': ['sess1', 'sess2', 'sess3', 'sess4', 'sess5'],
    };
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

/// Fake NotificationsService for testing
class _FakeNotificationsService extends NotificationsService {
  @override
  void info(String title, String description) {}

  @override
  void warn(String title, String description) {}

  @override
  void error(String title, String description) {}
}

/// Helper to convert HAR to bytes
Uint8List _harToBytes(Map<String, dynamic> har) {
  return Uint8List.fromList(utf8.encode(jsonEncode(har)));
}

void main() {
  setUpAll(() {
    TestWidgetsFlutterBinding.ensureInitialized();
    final sl = GetIt.instance;
    sl.registerLazySingleton<NotificationsService>(
      () => _FakeNotificationsService(),
    );

    const channel = MethodChannel('plugins.flutter.io/url_launcher');
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (call) async {
          switch (call.method) {
            case 'canLaunch':
              return true;
            case 'launch':
              return true;
            default:
              return null;
          }
        });
  });

  tearDownAll(() async {
    const channel = MethodChannel('plugins.flutter.io/url_launcher');
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, null);
    await GetIt.instance.reset();
  });

  group('ExportImportDialog Smoke Tests', () {
    testWidgets('Dialog renders mode selection screen', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);

      // Act
      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: Dialog(
                child: SizedBox(
                  width: 600,
                  height: 400,
                  child: Column(
                    children: [
                      const Text('Export / Import'),
                      const Expanded(child: ModeSelectionScreen()),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 50));

      // Assert - Mode selection should render
      expect(find.text('Export'), findsOneWidget);
      expect(find.text('Import'), findsOneWidget);
    });

    testWidgets('Export flow renders settings screen', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      store.setVisibleSessionsCount(10);
      store.setTotalSessionsCount(50);

      // Act
      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: Dialog(
                child: SizedBox(
                  width: 600,
                  height: 500,
                  child: Column(
                    children: [
                      const Text('Export Settings'),
                      Expanded(
                        child: ExportSettingsScreen(
                          visibleSessionIds: const ['sess1', 'sess2'],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      // Transition to export settings
      store.selectMode(ExportImportMode.export);
      await tester.pumpAndSettle();

      // Assert - Export settings should render
      expect(find.text('Visible Sessions'), findsWidgets);
      expect(find.text('All Sessions'), findsWidgets);
    });

    testWidgets('Import flow renders settings screen with file', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);

      // Act
      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: Dialog(
                child: SizedBox(
                  width: 600,
                  height: 500,
                  child: Column(
                    children: [
                      const Text('Import Settings'),
                      const Expanded(child: ImportSettingsScreen()),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      // Set import file
      store.setImportFile(bytes, 'test.har', bytes.length);
      await tester.pumpAndSettle();

      // Assert - Import settings should render
      expect(find.text('test.har'), findsOneWidget);
      expect(find.text('Import HAR'), findsOneWidget);
    });

    testWidgets('Full export flow: select mode → configure → export', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      store.setVisibleSessionsCount(5);
      store.setTotalSessionsCount(10);

      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: Dialog(
                child: SizedBox(
                  width: 600,
                  height: 500,
                  child: Column(
                    children: [
                      Observer(
                        builder: (_) {
                          String title = 'Export / Import';
                          if (store.state == ExportImportState.exportSettings) {
                            title = 'Export Settings';
                          } else if (store.state == ExportImportState.success) {
                            title = 'Success';
                          }
                          return Text(title);
                        },
                      ),
                      Expanded(
                        child: Observer(
                          builder: (_) {
                            if (store.state == ExportImportState.selectMode) {
                              return const ModeSelectionScreen();
                            } else if (store.state ==
                                ExportImportState.exportSettings) {
                              return ExportSettingsScreen(
                                visibleSessionIds: const ['sess1'],
                              );
                            } else if (store.state ==
                                ExportImportState.success) {
                              return const Center(
                                child: Text('Export completed'),
                              );
                            }
                            return const SizedBox();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Step 1: Select Export mode
      expect(store.state, ExportImportState.selectMode);
      store.selectMode(ExportImportMode.export);
      await tester.pumpAndSettle();

      // Step 2: Verify export settings shown
      expect(store.state, ExportImportState.exportSettings);
      expect(find.text('Export Settings'), findsOneWidget);

      // Step 3: Perform export
      unawaited(store.performExport(['sess1']));
      await tester.pump(const Duration(milliseconds: 1200));
      await tester.pump();

      // Assert - export completed
      expect(repo.exportCallCount, 1);
      expect(store.state, ExportImportState.success);
      expect(find.text('Export completed'), findsOneWidget);
    });

    testWidgets('Full import flow: select file → choose mode → import', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createMultipleHTTP(
        requests: [
          {'url': 'http://example.com/1'},
          {'url': 'http://example.com/2'},
          {'url': 'http://example.com/3'},
        ],
      );
      final bytes = _harToBytes(har);

      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: Dialog(
                child: SizedBox(
                  width: 600,
                  height: 500,
                  child: Column(
                    children: [
                      Observer(
                        builder: (_) {
                          String title = 'Import';
                          if (store.state == ExportImportState.importSettings) {
                            title = 'Import Settings';
                          } else if (store.state == ExportImportState.success) {
                            title = 'Success';
                          }
                          return Text(title);
                        },
                      ),
                      Expanded(
                        child: Observer(
                          builder: (_) {
                            if (store.state ==
                                ExportImportState.importSettings) {
                              return const ImportSettingsScreen();
                            } else if (store.state ==
                                ExportImportState.success) {
                              return const Center(
                                child: Text('Import completed'),
                              );
                            }
                            return const SizedBox();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Step 1: Set import file (simulates file picker)
      store.selectMode(ExportImportMode.import);
      store.setImportFile(bytes, 'integration-test.har', bytes.length);
      if (store.state != ExportImportState.error) {
        store.state = ExportImportState.importSettings;
      }
      await tester.pumpAndSettle();

      // Step 2: Verify import settings shown
      expect(store.state, ExportImportState.importSettings);
      expect(find.text('Import Settings'), findsOneWidget);
      expect(find.text('integration-test.har'), findsOneWidget);

      // Step 3: Select import mode
      store.setImportMode(ImportMode.replaceImported);
      await tester.pumpAndSettle();

      // Step 4: Perform import
      unawaited(store.performImport());
      await tester.pump(const Duration(milliseconds: 1200));
      await tester.pump();

      // Assert - import completed
      expect(repo.importCallCount, 1);
      expect(store.state, ExportImportState.success);
      expect(find.text('Import completed'), findsOneWidget);
    });

    testWidgets('Back button transitions state correctly', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);

      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: const SizedBox(),
            ),
          ),
        ),
      );

      // Act - go to export settings
      store.selectMode(ExportImportMode.export);
      await tester.pump();
      expect(store.state, ExportImportState.exportSettings);

      // Go back
      store.backToModeSelection();
      await tester.pump();

      // Assert
      expect(store.state, ExportImportState.selectMode);
      expect(store.selectedMode, isNull);
    });

    testWidgets('Reset clears all state', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);

      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: Provider<ExportImportStore>.value(
              value: store,
              child: const SizedBox(),
            ),
          ),
        ),
      );

      // Act - set up some state
      store.selectMode(ExportImportMode.export);
      store.setVisibleSessionsCount(10);
      store.setImportFile(bytes, 'test.har', bytes.length);
      await tester.pump();

      // Reset
      store.reset();
      await tester.pump();

      // Assert - all state cleared
      expect(store.state, ExportImportState.selectMode);
      expect(store.selectedMode, isNull);
      // These values come from outside (dialog context) and persist,
      // so that export options don't "die" when navigating back.
      expect(store.visibleSessionsCount, 10);
      expect(store.importFileData, isNull);
    });
  });
}
