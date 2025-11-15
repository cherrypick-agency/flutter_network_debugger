import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:get_it/get_it.dart';
import 'package:provider/provider.dart';
import 'package:frontend/features/export_import/application/stores/export_import_store.dart';
import 'package:frontend/features/export_import/presentation/widgets/export_import_dialog.dart';
import 'package:frontend/features/inspector/domain/repositories/inspector_repository.dart';
import 'package:frontend/core/notifications/notifications_service.dart';
import 'package:frontend/theme/app_theme.dart';
import '../../../helpers/har_generator.dart';

/// Fake InspectorRepository for widget testing
class _FakeRepository implements InspectorRepository {
  int importCallCount = 0;
  String? lastMode;
  bool shouldFail = false;

  @override
  Future<Map<String, dynamic>> importHAR(
    String harJson, {
    required String mode,
  }) async {
    importCallCount++;
    lastMode = mode;
    if (shouldFail) throw Exception('Import failed');
    return {'imported': 3, 'failed': 0, 'total': 3};
  }

  @override
  Future<String> exportHAR({
    List<String>? sessionIds,
    bool includeBodies = true,
    bool includeSensitive = false,
    bool minify = false,
  }) async {
    return 'http://example.com/export.har';
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

/// Helper to wrap widget with necessary providers and theme
Widget _wrap(ExportImportStore store, Widget child) {
  return Provider<ExportImportStore>.value(
    value: store,
    child: MaterialApp(
      theme: buildLightTheme(),
      home: Scaffold(body: child),
    ),
  );
}

/// Helper to convert HAR to bytes
Uint8List _harToBytes(Map<String, dynamic> har) {
  return Uint8List.fromList(utf8.encode(jsonEncode(har)));
}

void main() {
  setUpAll(() {
    final sl = GetIt.instance;
    sl.registerLazySingleton<NotificationsService>(
      () => _FakeNotificationsService(),
    );
  });

  tearDownAll(() async {
    await GetIt.instance.reset();
  });

  group('ImportSettingsScreen - File Preview Display', () {
    testWidgets('Shows filename correctly', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'my-test-file.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      expect(find.text('my-test-file.har'), findsOneWidget);
    });

    testWidgets('Shows file size in KB', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      const fileSize = 2048; // 2 KB
      store.setImportFile(bytes, 'test.har', fileSize);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      expect(find.textContaining('2.0 KB'), findsOneWidget);
    });

    testWidgets('Shows entries count', (tester) async {
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
      store.setImportFile(bytes, 'multi.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      expect(find.textContaining('Entries: 3'), findsOneWidget);
    });

    testWidgets('Shows file icon', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      expect(find.byIcon(Icons.file_present), findsOneWidget);
    });
  });

  group('ImportSettingsScreen - Import Mode Selection', () {
    testWidgets('All 3 radio options are rendered', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      expect(find.text('Add to existing'), findsOneWidget);
      expect(find.text('Replace all sessions'), findsOneWidget);
      expect(find.text('Replace imported only'), findsOneWidget);
    });

    testWidgets('Default mode is merge (Add to existing)', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      // Find checked radio button with merge mode
      final radioIcons = find.byIcon(Icons.radio_button_checked);
      expect(radioIcons, findsOneWidget);

      // Verify it's the merge option
      expect(store.importMode, ImportMode.merge);
    });

    testWidgets('Tapping radio changes selection to replaceAll', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Tap on "Replace all sessions" option
      await tester.tap(find.text('Replace all sessions'));
      await tester.pumpAndSettle();

      // Assert
      expect(store.importMode, ImportMode.replaceAll);
    });

    testWidgets('Tapping radio changes selection to replaceImported', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Tap on "Replace imported only" option
      await tester.tap(find.text('Replace imported only'));
      await tester.pumpAndSettle();

      // Assert
      expect(store.importMode, ImportMode.replaceImported);
    });

    testWidgets('Only one option is selected at a time', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Initially merge is selected
      expect(find.byIcon(Icons.radio_button_checked), findsOneWidget);
      expect(find.byIcon(Icons.radio_button_unchecked), findsNWidgets(2));

      // Tap replaceAll
      await tester.tap(find.text('Replace all sessions'));
      await tester.pumpAndSettle();

      // Now replaceAll should be selected
      expect(find.byIcon(Icons.radio_button_checked), findsOneWidget);
      expect(find.byIcon(Icons.radio_button_unchecked), findsNWidgets(2));
    });
  });

  group('ImportSettingsScreen - Visual Indicators', () {
    testWidgets('Merge mode shows green elements when selected', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert - merge is selected by default, should show green
      final radioChecked = find.byIcon(Icons.radio_button_checked);
      expect(radioChecked, findsOneWidget);

      // The radio icon should have green color (from theme)
      final icon = tester.widget<Icon>(radioChecked);
      expect(icon.color, Colors.green);
    });

    testWidgets('Replace all mode shows red elements when selected', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));
      await tester.tap(find.text('Replace all sessions'));
      await tester.pumpAndSettle();

      // Assert
      final radioChecked = find.byIcon(Icons.radio_button_checked);
      expect(radioChecked, findsOneWidget);

      // The radio icon should have red color (destructive)
      final icon = tester.widget<Icon>(radioChecked);
      expect(icon.color, Colors.red);
    });

    testWidgets('Replace all mode shows warning emoji', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert - Warning emoji should be present in description
      expect(find.textContaining('⚠️'), findsOneWidget);
      expect(
        find.text('Delete ALL sessions before import (⚠️ destructive)'),
        findsOneWidget,
      );
    });

    testWidgets('Descriptions are readable and informative', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert - Check all descriptions are present
      expect(
        find.text('Keep all current sessions and add imported ones (safe)'),
        findsOneWidget,
      );
      expect(
        find.text('Delete ALL sessions before import (⚠️ destructive)'),
        findsOneWidget,
      );
      expect(
        find.text('Delete previously imported sessions, keep proxy sessions'),
        findsOneWidget,
      );
    });
  });

  group('ImportSettingsScreen - Import Button', () {
    testWidgets('Import button is rendered', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));
      await tester.pump();

      // Assert
      expect(find.text('Import HAR'), findsOneWidget);
      expect(find.byIcon(Icons.download_rounded), findsOneWidget);
    });

    testWidgets('Import button shows correct icon', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Assert
      expect(find.byIcon(Icons.download_rounded), findsOneWidget);
    });

    testWidgets('Tapping import button calls store.performImport()', (
      tester,
    ) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));
      await tester.tap(find.text('Import HAR'));
      await tester.pumpAndSettle();

      // Wait for progress simulation timers (up to 1000ms)
      await tester.pump(const Duration(milliseconds: 1100));

      // Assert - import was called on repository
      expect(repo.importCallCount, 1);
    });

    testWidgets('Import button uses selected mode', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Change mode to replaceAll
      await tester.tap(find.text('Replace all sessions'));
      await tester.pumpAndSettle();

      // Tap import
      await tester.tap(find.text('Import HAR'));
      await tester.pumpAndSettle();

      // Wait for progress simulation timers (up to 1000ms)
      await tester.pump(const Duration(milliseconds: 1100));

      // Assert
      expect(repo.lastMode, 'replace_all');
    });
  });

  group('ImportSettingsScreen - Integration', () {
    testWidgets('Complete flow: select mode and import', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      final har = HARGenerator.createMultipleHTTP(
        requests: [
          {'url': 'http://example.com/1'},
          {'url': 'http://example.com/2'},
        ],
      );
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'integration-test.har', bytes.length);

      // Act
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Verify file preview
      expect(find.text('integration-test.har'), findsOneWidget);
      expect(find.textContaining('Entries: 2'), findsOneWidget);

      // Select "Replace imported only" mode
      await tester.tap(find.text('Replace imported only'));
      await tester.pumpAndSettle();

      // Verify mode changed
      expect(store.importMode, ImportMode.replaceImported);

      // Tap import button
      await tester.tap(find.text('Import HAR'));
      await tester.pumpAndSettle();

      // Wait for progress simulation timers (up to 1000ms)
      await tester.pump(const Duration(milliseconds: 1100));

      // Assert
      expect(repo.importCallCount, 1);
      expect(repo.lastMode, 'replace_imported');
      expect(store.state, ExportImportState.success);
    });

    testWidgets('Widget renders correctly without file data', (tester) async {
      // Arrange
      final repo = _FakeRepository();
      final store = ExportImportStore(repo);
      // NO file set

      // Act & Assert - should not crash
      await tester.pumpWidget(_wrap(store, const ImportSettingsScreen()));

      // Widget should render even with null file data
      // (though this shouldn't happen in real usage due to state machine)
      expect(find.byType(ImportSettingsScreen), findsOneWidget);
    });
  });
}
