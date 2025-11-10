import 'package:multi_editor_plugins/multi_editor_plugins.dart';
import 'package:flutter_monaco_crossplatform/flutter_monaco_crossplatform.dart';

/// Service providing access to the multi_editor (Monaco) instance for plugins
///
/// Implements [EditorService] to allow plugins to interact with the Monaco editor.
/// The controller is set when the Monaco editor is initialized via [setController].
class MultiEditorService implements EditorService {
  MonacoController? _controller;

  /// Sets the Monaco controller instance
  ///
  /// Should be called when the Monaco editor is ready (e.g., in onControllerReady callback)
  void setController(MonacoController? controller) {
    _controller = controller;
  }

  /// Returns true if the Monaco controller is available
  @override
  bool get isAvailable => _controller != null;

  /// Gets the current Monaco controller
  ///
  /// Returns null if the editor is not yet initialized
  MonacoController? get controller => _controller;

  @override
  void dispose() {
    _controller = null;
  }
}
