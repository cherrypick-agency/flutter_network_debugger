import 'dart:async';
import 'package:multi_editor_plugins/multi_editor_plugins.dart';

/// Implementation of [PluginUIService] for managing plugin UI registrations
///
/// Stores plugin UI descriptors and notifies listeners when registrations change.
/// This is a presentation layer implementation of the domain service interface.
class PluginUIRegistry implements PluginUIService {
  /// Storage for plugin UI descriptors by plugin ID
  final Map<String, PluginUIDescriptor> _descriptors = {};

  /// Stream controller for UI update events
  final StreamController<PluginUIUpdateEvent> _updateController =
      StreamController<PluginUIUpdateEvent>.broadcast();

  @override
  void registerUI(PluginUIDescriptor descriptor) {
    final isUpdate = _descriptors.containsKey(descriptor.pluginId);
    _descriptors[descriptor.pluginId] = descriptor;

    _updateController.add(
      PluginUIUpdateEvent(
        type: isUpdate
            ? PluginUIUpdateType.updated
            : PluginUIUpdateType.registered,
        descriptor: descriptor,
      ),
    );
  }

  @override
  void unregisterUI(String pluginId) {
    final descriptor = _descriptors.remove(pluginId);
    if (descriptor != null) {
      _updateController.add(
        PluginUIUpdateEvent(
          type: PluginUIUpdateType.unregistered,
          descriptor: descriptor,
        ),
      );
    }
  }

  @override
  List<PluginUIDescriptor> getRegisteredUIs() {
    final descriptors = _descriptors.values.toList();
    // Sort by priority (lower number = higher priority)
    descriptors.sort((a, b) => a.priority.compareTo(b.priority));
    return descriptors;
  }

  @override
  PluginUIDescriptor? getUI(String pluginId) {
    return _descriptors[pluginId];
  }

  @override
  Stream<PluginUIUpdateEvent> get updates => _updateController.stream;

  @override
  bool hasUI(String pluginId) {
    return _descriptors.containsKey(pluginId);
  }

  /// Dispose resources
  void dispose() {
    _updateController.close();
  }
}
