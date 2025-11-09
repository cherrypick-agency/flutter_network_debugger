import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';
import 'package:multi_editor_plugins/multi_editor_plugins.dart';
import 'package:multi_editor_plugin_base/multi_editor_plugin_base.dart';
import '../../domain/entities/file_statistics.dart';

/// Plugin that calculates and displays file statistics.
///
/// ## Architecture improvements:
/// - Uses PluginManifestBuilder for cleaner manifest
/// - Throttles statistics calculation (500ms) to reduce CPU usage
/// - Debounces UI updates (200ms) to avoid excessive rebuilds
/// - Simplified through updated BaseEditorPlugin
class FileStatsPlugin extends BaseEditorPlugin with StatefulPlugin {
  final Map<String, FileStatistics> _statistics = {};
  final Map<String, DateTime> _lastCalculationTime = {};
  final Map<String, Timer> _updateTimers = {};
  String? _currentFileId; // Track currently open file

  /// Minimum time between statistics calculations for same file
  static const _calculationThrottle = Duration(milliseconds: 500);

  /// Delay before updating UI state
  static const _uiUpdateDelay = Duration(milliseconds: 200);

  @override
  PluginManifest get manifest => PluginManifestBuilder()
      .withId('plugin.file-stats')
      .withName('File Statistics')
      .withVersion('0.1.0')
      .withDescription('Displays file metrics (lines, characters, words, size)')
      .withAuthor('Editor Team')
      .addActivationEvent('onFileOpen')
      .addActivationEvent('onFileContentChange')
      .build();

  @override
  Future<void> onInitialize(PluginContext context) async {
    setState('statistics', _statistics);

    // Register UI with PluginUIService
    final uiService = context.getService<PluginUIService>();
    final descriptor = getUIDescriptor();
    if (descriptor != null && uiService != null) {
      uiService.registerUI(descriptor);
    }
  }

  @override
  Future<void> onDispose() async {
    // Cancel all pending timers
    for (final timer in _updateTimers.values) {
      timer.cancel();
    }
    _updateTimers.clear();
    disposeStateful();
  }

  @override
  void onFileOpen(FileDocument file) {
    safeExecute('Calculate file statistics', () {
      // Track current file
      _currentFileId = file.id;

      // Calculate immediately on file open
      _calculateAndScheduleUpdate(file.id, file.content, immediate: true);
    });
  }

  @override
  void onFileContentChange(String fileId, String content) {
    safeExecute('Update file statistics', () {
      // Throttle calculations on content change
      _calculateAndScheduleUpdate(fileId, content);
    });
  }

  @override
  void onFileClose(String fileId) {
    safeExecute('Clear file statistics', () {
      // Clear current file if it's the one being closed
      if (_currentFileId == fileId) {
        _currentFileId = null;
      }

      _updateTimers[fileId]?.cancel();
      _updateTimers.remove(fileId);
      _statistics.remove(fileId);
      _lastCalculationTime.remove(fileId);
      setState('statistics', _statistics);

      // Update UI descriptor in PluginUIService
      if (isInitialized) {
        final uiService = context.getService<PluginUIService>();
        final descriptor = getUIDescriptor();
        if (descriptor != null) {
          uiService?.registerUI(descriptor);
        } else {
          // No files open, unregister UI
          uiService?.unregisterUI(manifest.id);
        }
      }
    });
  }

  /// Calculate statistics with optional throttling and schedule UI update
  void _calculateAndScheduleUpdate(
    String fileId,
    String content, {
    bool immediate = false,
  }) {
    // Check throttle unless immediate
    if (!immediate) {
      final lastCalc = _lastCalculationTime[fileId];
      if (lastCalc != null) {
        final timeSinceLastCalc = DateTime.now().difference(lastCalc);
        if (timeSinceLastCalc < _calculationThrottle) {
          // Skip calculation, too soon
          return;
        }
      }
    }

    // Calculate statistics
    final stats = FileStatistics.calculate(fileId, content);
    _statistics[fileId] = stats;
    _lastCalculationTime[fileId] = DateTime.now();

    if (immediate) {
      // Update UI immediately without debouncing
      setState('statistics', _statistics);

      // Update UI descriptor in PluginUIService
      if (isInitialized) {
        final uiService = context.getService<PluginUIService>();
        final descriptor = getUIDescriptor();
        if (descriptor != null && uiService != null) {
          uiService.registerUI(descriptor);
        }
      }
    } else {
      // Schedule UI update with debouncing
      _updateTimers[fileId]?.cancel();
      _updateTimers[fileId] = Timer(_uiUpdateDelay, () {
        setState('statistics', _statistics);

        // Update UI descriptor in PluginUIService
        if (isInitialized) {
          final uiService = context.getService<PluginUIService>();
          final descriptor = getUIDescriptor();
          if (descriptor != null && uiService != null) {
            uiService.registerUI(descriptor);
          }
        }
      });
    }
  }

  @override
  PluginUIDescriptor? getUIDescriptor() {
    // Only show stats for currently open file
    if (_currentFileId == null) {
      return null;
    }

    final currentStats = _statistics[_currentFileId];
    if (currentStats == null) {
      return null;
    }

    return PluginUIDescriptor(
      pluginId: manifest.id,
      iconCode: MaterialIconCodes.barChart,
      iconFamily: 'MaterialIcons',
      tooltip: 'File Statistics',
      priority: 20,
      uiData: {
        'type': 'list',
        'items': [
          {
            'id': _currentFileId,
            'title': 'Lines: ${currentStats.lines}',
            'subtitle':
                'Chars: ${currentStats.characters} • Words: ${currentStats.words}',
            'iconCode': 0xe873, // Icons.insert_chart
            'onTap': 'showStats',
          },
        ],
      },
    );
  }

  FileStatistics? getStatistics(String fileId) => _statistics[fileId];

  Map<String, FileStatistics> get allStatistics =>
      Map<String, FileStatistics>.from(_statistics);
}
