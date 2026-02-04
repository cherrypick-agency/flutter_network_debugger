import 'dart:async';
import 'package:multi_editor_core/multi_editor_core.dart';
import 'package:multi_editor_plugins/multi_editor_plugins.dart';
import 'package:multi_editor_plugin_base/multi_editor_plugin_base.dart';
import '../../domain/entities/recent_files_list.dart';
import '../../domain/value_objects/recent_file_entry.dart';

/// Plugin that tracks recently opened files.
///
/// ## Architecture improvements:
/// - Uses PluginManifestBuilder for cleaner manifest
/// - Debounces state updates (200ms) to avoid excessive UI rebuilds
/// - Simplified through updated BaseEditorPlugin
class RecentFilesPlugin extends BaseEditorPlugin with StatefulPlugin {
  static const String _pluginId = 'plugin.recent-files';
  RecentFilesList _recentFiles = RecentFilesList.create(maxEntries: 10);
  Timer? _stateUpdateDebounce;

  /// Delay before updating state (debouncing)
  static const _stateUpdateDelay = Duration(milliseconds: 200);

  @override
  PluginManifest get manifest => PluginManifestBuilder()
      .withId('plugin.recent-files')
      .withName('Recent Files')
      .withVersion('0.1.0')
      .withDescription('Tracks and displays recently opened files')
      .withAuthor('Editor Team')
      .addActivationEvent('onFileOpen')
      .build();

  @override
  Future<void> onInitialize(PluginContext context) async {
    setState('recentFiles', _recentFiles);

    // Register UI with PluginUIService
    final uiService = context.getService<PluginUIService>();
    final descriptor = getUIDescriptor();
    if (descriptor != null && uiService != null) {
      uiService.registerUI(descriptor);
    }
  }

  @override
  Future<void> onDispose() async {
    _stateUpdateDebounce?.cancel();
    _stateUpdateDebounce = null;
    disposeStateful();
  }

  /// Schedule state update with debouncing
  void _scheduleStateUpdate() {
    _stateUpdateDebounce?.cancel();
    _stateUpdateDebounce = Timer(_stateUpdateDelay, () {
      setState('recentFiles', _recentFiles);

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

  /// Update state immediately without debouncing
  void _updateStateImmediately() {
    setState('recentFiles', _recentFiles);

    // Update UI descriptor in PluginUIService
    if (isInitialized) {
      final uiService = context.getService<PluginUIService>();
      final descriptor = getUIDescriptor();
      if (descriptor != null && uiService != null) {
        uiService.registerUI(descriptor);
      }
    }
  }

  @override
  void onFileOpen(FileDocument file) {
    safeExecute('Add to recent files', () {
      // Try to carefully get data from FileDocument without tight coupling
      String fileId;
      String fileName;
      try {
        fileId = (file as dynamic).id as String;
      } catch (_) {
        fileId = DateTime.now().millisecondsSinceEpoch.toString();
      }
      try {
        fileName = (file as dynamic).name as String;
      } catch (_) {
        fileName = 'Untitled';
      }
      print('[RecentFiles] File opened: $fileName');
      final entry = RecentFileEntry.create(
        fileId: fileId,
        fileName: fileName,
        filePath: fileName,
      );

      _recentFiles = _recentFiles.addFile(entry);
      _updateStateImmediately();
      print('[RecentFiles] Total files: ${_recentFiles.count}');
    });
  }

  @override
  void onFileDelete(String fileId) {
    safeExecute('Remove from recent files', () {
      _recentFiles = _recentFiles.removeFile(fileId);
      _scheduleStateUpdate();
    });
  }

  @override
  PluginUIDescriptor? getUIDescriptor() {
    return PluginUIDescriptor(
      pluginId: _pluginId,
      iconCode: MaterialIconCodes.copyAll,
      iconFamily: 'MaterialIcons',
      tooltip: 'Recent Files',
      priority: 10,
      uiData: {
        'type': 'list',
        'items': _recentFiles.sortedByRecent.map((entry) {
          return {
            'id': entry.fileId,
            'title': entry.fileName,
            'subtitle': entry.formattedTime, // Show relative time
            'iconCode': 0xe24d, // Icons.insert_drive_file
            'onTap': 'openFile',
          };
        }).toList(),
      },
    );
  }

  void clearRecentFiles() {
    _recentFiles = _recentFiles.clear();
    setState('recentFiles', _recentFiles);
  }

  List<RecentFileEntry> get recentFiles => _recentFiles.sortedByRecent;
}
