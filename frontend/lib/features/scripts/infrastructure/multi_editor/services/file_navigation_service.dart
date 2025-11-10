import 'package:multi_editor_plugins/multi_editor_plugins.dart';
import 'package:multi_editor_ui/multi_editor_ui.dart';

/// Service for navigating between files in the multi_editor
///
/// Implements [FileNavigationService] to allow plugins to trigger file opening
/// by delegating to [EditorController].
class EditorFileNavigationService implements FileNavigationService {
  final EditorController _editorController;

  EditorFileNavigationService(this._editorController);

  /// Opens a file by its ID in the editor
  ///
  /// Delegates to [EditorController.loadFile] to handle the actual file loading
  @override
  Future<void> openFile(String fileId) async {
    await _editorController.loadFile(fileId);
  }

  /// Always returns true as the navigation service is always available
  @override
  bool get isAvailable => true;
}
