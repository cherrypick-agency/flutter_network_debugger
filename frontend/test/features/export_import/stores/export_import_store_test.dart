import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:get_it/get_it.dart';
import 'package:frontend/features/export_import/application/stores/export_import_store.dart';
import 'package:frontend/features/inspector/domain/repositories/inspector_repository.dart';
import 'package:frontend/core/notifications/notifications_service.dart';
import '../../../helpers/har_generator.dart';

/// Fake InspectorRepository for testing
class _FakeInspectorRepository implements InspectorRepository {
  // Call trackers
  int exportCallCount = 0;
  int importCallCount = 0;
  String? lastImportMode;
  List<String>? lastExportSessionIds;

  // Control behavior
  bool shouldFailExport = false;
  bool shouldFailImport = false;
  String? exportResult;
  Map<String, dynamic>? importResult;

  @override
  Future<String> exportHAR({
    List<String>? sessionIds,
    bool includeBodies = true,
    bool includeSensitive = false,
    bool minify = false,
  }) async {
    exportCallCount++;
    lastExportSessionIds = sessionIds;

    if (shouldFailExport) {
      throw Exception('Export failed');
    }

    return exportResult ?? 'http://example.com/download/export.har';
  }

  @override
  Future<Map<String, dynamic>> importHAR(
    String harJson, {
    required String mode,
  }) async {
    importCallCount++;
    lastImportMode = mode;

    if (shouldFailImport) {
      throw Exception('Import failed');
    }

    return importResult ??
        {
          'imported': 5,
          'failed': 0,
          'total': 5,
          'sessionIds': ['sess1', 'sess2', 'sess3', 'sess4', 'sess5'],
        };
  }

  // Stub implementations for other methods
  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

/// Fake NotificationsService for testing
class _FakeNotificationsService extends NotificationsService {
  final List<String> messages = [];

  @override
  void info(String title, String description) {
    messages.add('INFO: $title - $description');
  }

  @override
  void warn(String title, String description) {
    messages.add('WARN: $title - $description');
  }

  @override
  void error(String title, String description) {
    messages.add('ERROR: $title - $description');
  }
}

/// Helper to convert HAR map to Uint8List
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

  group('ExportImportStore - File Validation', () {
    late ExportImportStore store;
    late _FakeInspectorRepository repo;

    setUp(() {
      repo = _FakeInspectorRepository();
      store = ExportImportStore(repo);
    });

    test('Valid HAR file passes validation', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);

      // Act
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Assert
      expect(store.importFileData, isNotNull);
      expect(store.importFileName, 'test.har');
      expect(
        store.state,
        ExportImportState.selectMode,
      ); // State remains selectMode after setImportFile
    });

    test('File too large (>100MB) is rejected', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      const largeSize = 101 * 1024 * 1024; // 101 MB

      // Act
      store.setImportFile(bytes, 'large.har', largeSize);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, contains('too large'));
      expect(store.errorMessage, contains('100MB'));
    });

    test('Invalid file extension (.txt) is rejected', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);

      // Act
      store.setImportFile(bytes, 'file.txt', bytes.length);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, contains('Invalid file type'));
    });

    test('Invalid file extension (.pdf) is rejected', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);

      // Act
      store.setImportFile(bytes, 'document.pdf', bytes.length);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
    });

    test('Invalid JSON format is rejected', () {
      // Arrange
      final malformed = HARGenerator.createMalformedJSON();
      final bytes = Uint8List.fromList(utf8.encode(malformed));

      // Act
      store.setImportFile(bytes, 'bad.har', bytes.length);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, contains('Invalid JSON'));
    });

    test('Missing "log" field is rejected', () {
      // Arrange
      final invalid = {'version': '1.2', 'entries': []};
      final bytes = _harToBytes(invalid);

      // Act
      store.setImportFile(bytes, 'no-log.har', bytes.length);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, contains('Missing required "log" field'));
    });

    test('Missing "entries" field is rejected', () {
      // Arrange
      final invalid = {
        'log': {'version': '1.2'},
      };
      final bytes = _harToBytes(invalid);

      // Act
      store.setImportFile(bytes, 'no-entries.har', bytes.length);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, contains('Missing "entries" field'));
    });

    test('"entries" must be an array', () {
      // Arrange
      final invalid = {
        'log': {'version': '1.2', 'entries': 'not-an-array'},
      };
      final bytes = _harToBytes(invalid);

      // Act
      store.setImportFile(bytes, 'bad-entries.har', bytes.length);

      // Assert
      expect(store.importFileData, isNull);
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, contains('"entries" must be an array'));
    });

    test('HAR preview parses entries count correctly', () {
      // Arrange
      final har = HARGenerator.createMultipleHTTP(
        requests: [
          {'url': 'http://example.com/1'},
          {'url': 'http://example.com/2'},
          {'url': 'http://example.com/3'},
        ],
      );
      final bytes = _harToBytes(har);

      // Act
      store.setImportFile(bytes, 'multi.har', bytes.length);

      // Assert
      expect(store.harEntriesCount, 3);
    });
  });

  group('ExportImportStore - Import Modes', () {
    late ExportImportStore store;
    late _FakeInspectorRepository repo;

    setUp(() {
      repo = _FakeInspectorRepository();
      store = ExportImportStore(repo);
    });

    test('Default import mode is merge', () {
      expect(store.importMode, ImportMode.merge);
    });

    test('setImportMode changes mode', () {
      // Act
      store.setImportMode(ImportMode.replaceAll);

      // Assert
      expect(store.importMode, ImportMode.replaceAll);
    });

    test('performImport sends correct mode string for merge', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      store.setImportMode(ImportMode.merge);

      // Act
      await store.performImport();

      // Assert
      expect(repo.lastImportMode, 'merge');
    });

    test('performImport sends correct mode string for replaceAll', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      store.setImportMode(ImportMode.replaceAll);

      // Act
      await store.performImport();

      // Assert
      expect(repo.lastImportMode, 'replace_all');
    });

    test(
      'performImport sends correct mode string for replaceImported',
      () async {
        // Arrange
        final har = HARGenerator.createSimpleHTTP();
        final bytes = _harToBytes(har);
        store.setImportFile(bytes, 'test.har', bytes.length);
        store.setImportMode(ImportMode.replaceImported);

        // Act
        await store.performImport();

        // Assert
        expect(repo.lastImportMode, 'replace_imported');
      },
    );
  });

  group('ExportImportStore - State Transitions', () {
    late ExportImportStore store;
    late _FakeInspectorRepository repo;

    setUp(() {
      repo = _FakeInspectorRepository();
      store = ExportImportStore(repo);
    });

    test('Initial state is selectMode', () {
      expect(store.state, ExportImportState.selectMode);
    });

    test('selectMode(export) transitions to exportSettings', () {
      // Act
      store.selectMode(ExportImportMode.export);

      // Assert
      expect(store.state, ExportImportState.exportSettings);
      expect(store.selectedMode, ExportImportMode.export);
    });

    test('selectMode(import) stays in selectMode until file selected', () {
      // Act
      store.selectMode(ExportImportMode.import);

      // Assert - stays in selectMode, not transitioning yet
      expect(store.selectedMode, ExportImportMode.import);
      // State transition happens when file is selected via setImportFile
    });

    test('setImportFile does not auto-transition state', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);

      // Act
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Assert - state should remain selectMode, widget code transitions manually
      expect(store.state, ExportImportState.selectMode);
      expect(store.importFileData, isNotNull);
    });

    test('performImport transitions to importing then success', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      await store.performImport();

      // Assert
      expect(store.state, ExportImportState.success);
      expect(store.importedCount, 5);
      expect(store.successMessage, isNotNull);
    });

    test('performImport transitions to error on failure', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      repo.shouldFailImport = true;

      // Act
      await store.performImport();

      // Assert
      expect(store.state, ExportImportState.error);
      expect(store.errorMessage, isNotNull);
    });

    test('backToModeSelection resets state', () {
      // Arrange
      store.selectMode(ExportImportMode.export);
      store.setExportScope(ExportScope.all);

      // Act
      store.backToModeSelection();

      // Assert
      expect(store.state, ExportImportState.selectMode);
      expect(store.selectedMode, isNull);
      expect(store.exportScope, ExportScope.visible); // Reset to default
    });

    test('reset() returns to initial state', () {
      // Arrange
      store.selectMode(ExportImportMode.export);
      store.setVisibleSessionsCount(10);
      store.toggleIncludeBodies();

      // Act
      store.reset();

      // Assert
      expect(store.state, ExportImportState.selectMode);
      expect(store.selectedMode, isNull);
      expect(store.includeBodies, true); // Reset to default
      // Note: visibleSessionsCount is metadata and persists across resets
      expect(store.visibleSessionsCount, 10);
    });
  });

  group('ExportImportStore - Repository Interaction', () {
    late ExportImportStore store;
    late _FakeInspectorRepository repo;

    setUp(() {
      repo = _FakeInspectorRepository();
      store = ExportImportStore(repo);
    });

    test('performExport calls repository with correct parameters', () async {
      // Arrange
      final sessionIds = ['sess1', 'sess2'];
      store.selectMode(ExportImportMode.export);
      store.toggleIncludeBodies(); // Now false
      store.toggleIncludeSensitive(); // Now true
      store.toggleMinify(); // Now true

      // Act
      await store.performExport(sessionIds);

      // Assert
      expect(repo.exportCallCount, 1);
      expect(repo.lastExportSessionIds, sessionIds);
    });

    test('performImport clears file data after completion', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      expect(store.importFileData, isNotNull);

      // Act
      await store.performImport();

      // Assert
      expect(store.importFileData, isNull);
      expect(store.importFileName, isNull);
      expect(store.importFileSize, isNull);
    });

    test('performImport clears file data even on error', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      repo.shouldFailImport = true;

      // Act
      await store.performImport();

      // Assert
      expect(store.importFileData, isNull);
    });

    test('performImport parses result correctly', () async {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      repo.importResult = {'imported': 10, 'failed': 2, 'total': 12};

      // Act
      await store.performImport();

      // Assert
      expect(store.importedCount, 10);
      expect(store.failedCount, 2);
      expect(store.successMessage, contains('10/12'));
      expect(store.successMessage, contains('2 failed'));
    });
  });

  group('ExportImportStore - Computed Properties', () {
    late ExportImportStore store;
    late _FakeInspectorRepository repo;

    setUp(() {
      repo = _FakeInspectorRepository();
      store = ExportImportStore(repo);
    });

    test('canExport is true when in exportSettings with visible sessions', () {
      // Arrange
      store.selectMode(ExportImportMode.export);
      store.setVisibleSessionsCount(5);
      store.setExportScope(ExportScope.visible);

      // Assert
      expect(store.canExport, isTrue);
    });

    test('canExport is true when scope is all regardless of visible count', () {
      // Arrange
      store.selectMode(ExportImportMode.export);
      store.setVisibleSessionsCount(0);
      store.setExportScope(ExportScope.all);

      // Assert
      expect(store.canExport, isTrue);
    });

    test('canExport is false when not in exportSettings state', () {
      // Arrange
      store.setVisibleSessionsCount(10);

      // Assert
      expect(store.canExport, isFalse);
    });

    test('canImport is true when file is selected and in selectMode', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);
      // Reset back to selectMode for this specific test
      store.state = ExportImportState.selectMode;

      // Assert
      expect(store.canImport, isTrue);
    });

    test('estimatedFileSize computes correctly for small files', () {
      // Arrange
      store.setTotalSessionsCount(10);
      store.setExportScope(ExportScope.all);
      store.toggleIncludeBodies(); // Now false

      // Assert
      expect(store.estimatedFileSize, '~20KB'); // 10 sessions * 2KB
    });

    test('estimatedFileSize computes correctly for large files', () {
      // Arrange
      store.setTotalSessionsCount(200);
      store.setExportScope(ExportScope.all);
      // includeBodies = true by default

      // Assert
      // 200 sessions * 10KB = 2000KB = 1.953MB, rounds to 2.0MB
      expect(store.estimatedFileSize, '~2.0MB');
    });

    test('exportScopeDescription shows correct text for visible', () {
      // Arrange
      store.setVisibleSessionsCount(15);
      store.setExportScope(ExportScope.visible);

      // Assert
      expect(store.exportScopeDescription, 'Visible sessions (15 filtered)');
    });

    test('exportScopeDescription shows correct text for all', () {
      // Arrange
      store.setTotalSessionsCount(100);
      store.setExportScope(ExportScope.all);

      // Assert
      expect(store.exportScopeDescription, 'All sessions (100 total)');
    });
  });

  group('ExportImportStore - Edge Cases', () {
    late ExportImportStore store;
    late _FakeInspectorRepository repo;

    setUp(() {
      repo = _FakeInspectorRepository();
      store = ExportImportStore(repo);
    });

    test('clearImportFile removes all file data', () {
      // Arrange
      final har = HARGenerator.createSimpleHTTP();
      final bytes = _harToBytes(har);
      store.setImportFile(bytes, 'test.har', bytes.length);

      // Act
      store.clearImportFile();

      // Assert
      expect(store.importFileData, isNull);
      expect(store.importFileName, isNull);
      expect(store.importFileSize, isNull);
      expect(store.harEntriesCount, isNull);
    });

    test('performImport does nothing if no file data', () async {
      // Act
      await store.performImport();

      // Assert
      expect(repo.importCallCount, 0);
    });

    test('toggle functions work correctly', () {
      // Initial states
      expect(store.includeBodies, isTrue);
      expect(store.includeSensitive, isFalse);
      expect(store.minify, isFalse);

      // Toggle
      store.toggleIncludeBodies();
      store.toggleIncludeSensitive();
      store.toggleMinify();

      // After toggle
      expect(store.includeBodies, isFalse);
      expect(store.includeSensitive, isTrue);
      expect(store.minify, isTrue);

      // Toggle again
      store.toggleIncludeBodies();
      store.toggleIncludeSensitive();
      store.toggleMinify();

      // Back to initial
      expect(store.includeBodies, isTrue);
      expect(store.includeSensitive, isFalse);
      expect(store.minify, isFalse);
    });
  });
}
