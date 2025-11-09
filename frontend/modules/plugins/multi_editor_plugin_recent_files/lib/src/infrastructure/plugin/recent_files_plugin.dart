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
      print('[RecentFiles] File opened: ${file.name}');
      final entry = RecentFileEntry.create(
        fileId: file.id,
        fileName: file.name,
        filePath: file.name,
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
      pluginId: manifest.id,
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
            'subtitle': entry.formattedTime, // Показываем относительное время
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
