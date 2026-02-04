import 'package:flutter/material.dart';
import 'package:multi_editor_ui/multi_editor_ui.dart';
import '../../infrastructure/editor_di.dart';

/// Wrapper around EditorScaffold from multi_file_code_editor
/// Simplified widget for integration with ScriptEditorDialog
class MultiFileEditorWidget extends StatelessWidget {
  final EditorDI editorDI;

  const MultiFileEditorWidget({super.key, required this.editorDI});

  @override
  Widget build(BuildContext context) {
    return EditorScaffold(
      fileTreeController: editorDI.fileTreeController,
      editorController: editorDI.editorController,
      pluginManager: editorDI.pluginManager,
      pluginUIService: editorDI.pluginUIRegistry,
      editorConfig: const EditorConfig(
        fontSize: 14.0,
        fontFamily: 'Consolas, Monaco, monospace',
        tabSize: 2,
        showMinimap: false,
        wordWrap: true,
        showLineNumbers: true,
        bracketPairColorization: true,
        autoSave: true,
        autoSaveDelay: 2,
      ),
      onControllerReady: (controller) {
        // Register Monaco controller in MultiEditorService for plugin access
        editorDI.multiEditorService.setController(controller);
      },
    );
  }
}
