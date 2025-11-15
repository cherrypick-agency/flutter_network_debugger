import 'dart:convert';
import 'package:mobx/mobx.dart';
import 'package:flutter/foundation.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../platform/file_operations.dart' as html;
import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/network/error_utils.dart';
import '../../../inspector/domain/repositories/inspector_repository.dart';

part 'export_import_store.g.dart';

/// Export/Import mode
enum ExportImportMode { export, import }

/// Export/Import state machine
enum ExportImportState {
  selectMode, // Initial state: choose Export or Import
  exportSettings, // Configure export options
  importSettings, // Configure import options
  importing, // Import in progress
  exporting, // Export in progress
  success, // Operation completed successfully
  error, // Operation failed
}

/// Export scope
enum ExportScope {
  visible, // Export visible (filtered) sessions
  all, // Export all sessions
}

/// Import mode
enum ImportMode {
  merge, // Add to existing sessions (default)
  replaceAll, // Delete all sessions before import
  replaceImported, // Delete only previously imported sessions
}

class ExportImportStore = _ExportImportStore with _$ExportImportStore;

abstract class _ExportImportStore with Store {
  _ExportImportStore(this._repository);

  final InspectorRepository _repository;

  // ============================================================================
  // STATE
  // ============================================================================

  @observable
  ExportImportState state = ExportImportState.selectMode;

  @observable
  ExportImportMode? selectedMode;

  // Export settings
  @observable
  ExportScope exportScope = ExportScope.visible;

  @observable
  bool includeBodies = true;

  @observable
  bool includeSensitive = false;

  @observable
  bool minify = false;

  // Import settings
  @observable
  ImportMode importMode = ImportMode.merge;

  // Progress tracking
  @observable
  double progress = 0.0;

  @observable
  String? progressMessage;

  // Results
  @observable
  String? errorMessage;

  @observable
  String? successMessage;

  @observable
  int? exportedCount;

  @observable
  int? importedCount;

  @observable
  int? failedCount;

  // Import file
  @observable
  Uint8List? importFileData;

  @observable
  String? importFileName;

  @observable
  int? importFileSize;

  // Preview data
  @observable
  int? visibleSessionsCount;

  @observable
  int? totalSessionsCount;

  @observable
  int? harEntriesCount;

  // ============================================================================
  // COMPUTED
  // ============================================================================

  @computed
  bool get canExport =>
      state == ExportImportState.exportSettings &&
      (exportScope == ExportScope.all || (visibleSessionsCount ?? 0) > 0);

  @computed
  bool get canImport =>
      state == ExportImportState.selectMode && importFileData != null;

  @computed
  String get exportScopeDescription {
    if (exportScope == ExportScope.visible) {
      final count = visibleSessionsCount ?? 0;
      return 'Visible sessions ($count filtered)';
    } else {
      final count = totalSessionsCount ?? 0;
      return 'All sessions ($count total)';
    }
  }

  @computed
  String get visibleScopeDescription {
    final count = visibleSessionsCount ?? 0;
    return 'Visible sessions ($count filtered)';
  }

  @computed
  String get allScopeDescription {
    final count = totalSessionsCount ?? 0;
    return 'All sessions ($count total)';
  }

  @computed
  String? get estimatedFileSize {
    final count = exportScope == ExportScope.visible
        ? (visibleSessionsCount ?? 0)
        : (totalSessionsCount ?? 0);

    if (count == 0) return null;

    // Rough estimation: ~10KB per session with bodies, ~2KB without
    final avgSize = includeBodies ? 10 : 2;
    final estimatedKb = count * avgSize;

    if (estimatedKb < 1024) {
      return '~${estimatedKb}KB';
    } else {
      final mb = (estimatedKb / 1024).toStringAsFixed(1);
      return '~${mb}MB';
    }
  }

  // ============================================================================
  // ACTIONS - STATE TRANSITIONS
  // ============================================================================

  @action
  void selectMode(ExportImportMode mode) {
    selectedMode = mode;
    if (mode == ExportImportMode.export) {
      state = ExportImportState.exportSettings;
    }
    // For import, stay in selectMode until file is selected
  }

  @action
  void backToModeSelection() {
    state = ExportImportState.selectMode;
    selectedMode = null;
    _resetState();
  }

  @action
  void setExportScope(ExportScope scope) {
    exportScope = scope;
  }

  @action
  void toggleIncludeBodies() {
    includeBodies = !includeBodies;
  }

  @action
  void toggleIncludeSensitive() {
    includeSensitive = !includeSensitive;
  }

  @action
  void toggleMinify() {
    minify = !minify;
  }

  @action
  void setImportMode(ImportMode mode) {
    importMode = mode;
  }

  @action
  void setVisibleSessionsCount(int count) {
    visibleSessionsCount = count;
  }

  @action
  void setTotalSessionsCount(int count) {
    totalSessionsCount = count;
  }

  @action
  void setImportFile(Uint8List data, String fileName, int fileSize) {
    // Validate file size (max 100MB for safety)
    const maxFileSize = 100 * 1024 * 1024; // 100MB
    if (fileSize > maxFileSize) {
      errorMessage =
          'File too large. Maximum size is 100MB.\nYour file: ${(fileSize / 1024 / 1024).toStringAsFixed(1)}MB';
      state = ExportImportState.error;
      sl<NotificationsService>().error(
        'File Too Large',
        'Maximum file size is 100MB',
      );
      return;
    }

    // Validate file extension
    final lowerName = fileName.toLowerCase();
    if (!lowerName.endsWith('.har') && !lowerName.endsWith('.json')) {
      errorMessage = 'Invalid file type.\nPlease select a .har or .json file';
      state = ExportImportState.error;
      sl<NotificationsService>().error(
        'Invalid File Type',
        'Please select a .har or .json file',
      );
      return;
    }

    // Validate JSON format
    try {
      final jsonStr = String.fromCharCodes(data);
      final parsed = jsonDecode(jsonStr);

      // Validate HAR structure
      if (parsed is! Map<String, dynamic>) {
        errorMessage = 'Invalid HAR format.\nFile must contain a JSON object';
        state = ExportImportState.error;
        sl<NotificationsService>().error(
          'Invalid HAR Format',
          'File must contain a JSON object',
        );
        return;
      }

      if (!parsed.containsKey('log')) {
        errorMessage = 'Invalid HAR format.\nMissing required "log" field';
        state = ExportImportState.error;
        sl<NotificationsService>().error(
          'Invalid HAR Format',
          'Missing required "log" field',
        );
        return;
      }

      final log = parsed['log'];
      if (log is! Map<String, dynamic>) {
        errorMessage = 'Invalid HAR format.\n"log" field must be an object';
        state = ExportImportState.error;
        sl<NotificationsService>().error(
          'Invalid HAR Format',
          '"log" field must be an object',
        );
        return;
      }

      if (!log.containsKey('entries')) {
        errorMessage = 'Invalid HAR format.\nMissing "entries" field in log';
        state = ExportImportState.error;
        sl<NotificationsService>().error(
          'Invalid HAR Format',
          'Missing "entries" field',
        );
        return;
      }

      final entries = log['entries'];
      if (entries is! List) {
        errorMessage = 'Invalid HAR format.\n"entries" must be an array';
        state = ExportImportState.error;
        sl<NotificationsService>().error(
          'Invalid HAR Format',
          '"entries" must be an array',
        );
        return;
      }

      // All validations passed
      importFileData = data;
      importFileName = fileName;
      importFileSize = fileSize;

      // Parse HAR preview
      _parseHARPreview(data);
    } on FormatException catch (e) {
      errorMessage = 'Invalid JSON format.\n${e.message}';
      state = ExportImportState.error;
      sl<NotificationsService>().error(
        'Invalid JSON',
        'Cannot parse file: ${e.message}',
      );
      return;
    } catch (e) {
      errorMessage = 'Error reading file.\n${e.toString()}';
      state = ExportImportState.error;
      sl<NotificationsService>().error('Error Reading File', e.toString());
      return;
    }
  }

  @action
  void clearImportFile() {
    importFileData = null;
    importFileName = null;
    importFileSize = null;
    harEntriesCount = null;
  }

  // ============================================================================
  // ACTIONS - OPERATIONS
  // ============================================================================

  @action
  Future<void> performExport(List<String>? sessionIds) async {
    state = ExportImportState.exporting;
    progress = 0.0;
    progressMessage = 'Preparing export...';

    try {
      // Simulate progress (real progress would come from backend streaming)
      _simulateProgress();

      final downloadUrl = await _repository.exportHAR(
        sessionIds: sessionIds,
        includeBodies: includeBodies,
        includeSensitive: includeSensitive,
        minify: minify,
      );

      // Open download URL
      await _openDownloadUrl(downloadUrl);

      exportedCount = sessionIds?.length ?? totalSessionsCount ?? 0;
      successMessage = 'Successfully exported $exportedCount sessions to HAR';
      state = ExportImportState.success;

      sl<NotificationsService>().info(
        'Export Complete',
        'Exported $exportedCount sessions',
      );
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      errorMessage = '${msg.title}: ${msg.description}';
      state = ExportImportState.error;
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      progress = 1.0;
    }
  }

  @action
  Future<void> performImport() async {
    if (importFileData == null) return;

    state = ExportImportState.importing;
    progress = 0.0;
    progressMessage = 'Uploading HAR file...';

    try {
      _simulateProgress();

      // Convert Uint8List to String
      final harJson = String.fromCharCodes(importFileData!);

      // Convert ImportMode to string for backend
      final modeStr = importMode == ImportMode.merge
          ? 'merge'
          : importMode == ImportMode.replaceAll
          ? 'replace_all'
          : 'replace_imported';

      final result = await _repository.importHAR(harJson, mode: modeStr);

      importedCount = result['imported'] as int? ?? 0;
      failedCount = result['failed'] as int? ?? 0;
      final total = result['total'] as int? ?? 0;

      successMessage = 'Imported $importedCount/$total entries';
      if ((failedCount ?? 0) > 0) {
        successMessage = '$successMessage ($failedCount failed)';
      }

      state = ExportImportState.success;
      sl<NotificationsService>().info(
        'Import Complete',
        'Imported $importedCount entries',
      );
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      errorMessage = '${msg.title}: ${msg.description}';
      state = ExportImportState.error;
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      // Clear file data after operation completes (success or failure)
      clearImportFile();
      progress = 1.0;
    }
  }

  @action
  void reset() {
    state = ExportImportState.selectMode;
    _resetState();
  }

  // ============================================================================
  // PRIVATE HELPERS
  // ============================================================================

  void _resetState() {
    selectedMode = null;
    exportScope = ExportScope.visible;
    includeBodies = true;
    includeSensitive = false;
    minify = false;
    importMode = ImportMode.merge;
    progress = 0.0;
    progressMessage = null;
    errorMessage = null;
    successMessage = null;
    exportedCount = null;
    importedCount = null;
    failedCount = null;
    clearImportFile();
  }

  void _simulateProgress() {
    // In real implementation, this would be updated from backend streaming
    Future.delayed(const Duration(milliseconds: 100), () {
      if (progress < 0.3) {
        progress = 0.3;
        progressMessage = 'Processing...';
      }
    });
    Future.delayed(const Duration(milliseconds: 500), () {
      if (progress < 0.6) {
        progress = 0.6;
        progressMessage = 'Generating HAR...';
      }
    });
    Future.delayed(const Duration(milliseconds: 1000), () {
      if (progress < 0.9) {
        progress = 0.9;
        progressMessage = 'Finalizing...';
      }
    });
  }

  void _parseHARPreview(Uint8List data) {
    try {
      // Parse JSON to get entries count
      final jsonStr = String.fromCharCodes(data);
      final parsed = _quickParseHAR(jsonStr);
      harEntriesCount = parsed;
    } catch (e) {
      harEntriesCount = null;
    }
  }

  int? _quickParseHAR(String jsonStr) {
    // Quick regex to find entries array length without full JSON parsing
    // Pattern: "entries":\s*\[
    final entriesMatch = RegExp(r'"entries"\s*:\s*\[').firstMatch(jsonStr);
    if (entriesMatch == null) return null;

    // Count objects in entries array (rough estimation)
    final afterEntries = jsonStr.substring(entriesMatch.end);
    final objectCount = '"startedDateTime"'.allMatches(afterEntries).length;
    return objectCount;
  }

  Future<void> _openDownloadUrl(String url) async {
    // Open URL in new window to trigger download
    // Backend sets Content-Disposition header for file download
    if (kIsWeb) {
      // Web: use window.open
      html.window.open(url, '_blank');
    } else {
      // Desktop/Mobile: use url_launcher to open in system browser
      final uri = Uri.parse(url);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        throw Exception('Could not launch $url');
      }
    }
  }
}
