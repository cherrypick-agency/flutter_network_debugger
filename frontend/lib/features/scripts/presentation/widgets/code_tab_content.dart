import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../application/stores/script_editor_store.dart';
import '../../application/stores/scripts_store.dart';
import '../../domain/entities/script.dart';
import 'compilation_status_banner.dart';
import 'validation_status_banner.dart';
import 'code_mode_selector.dart';
import 'extism_toolbar.dart';
import 'mode_content_widget.dart';

class CodeTabContent extends StatelessWidget {
  final ScriptEditorStore editorStore;
  final ScriptsStore scriptsStore;
  final bool isDownloadingProject;
  final Future<void> Function() onDownloadProject;
  final Widget Function() buildMultiFileEditor;
  final VoidCallback onImportZip;

  const CodeTabContent({
    super.key,
    required this.editorStore,
    required this.scriptsStore,
    required this.isDownloadingProject,
    required this.onDownloadProject,
    required this.buildMultiFileEditor,
    required this.onImportZip,
  });

  @override
  Widget build(BuildContext context) {
    return Observer(
      builder: (_) {
        return Column(
          children: [
            if (editorStore.runtime == ScriptRuntime.extism &&
                editorStore.isEditing)
              CompilationStatusBanner(
                editorStore: editorStore,
                scriptsStore: scriptsStore,
              ),
            ValidationStatusBanner(store: editorStore),
            if (editorStore.runtime == ScriptRuntime.extism)
              CodeModeSelector(
                selectedMode: editorStore.codeCreationMode,
                onModeChanged: editorStore.setCodeCreationMode,
              ),
            if (editorStore.runtime == ScriptRuntime.extism &&
                editorStore.codeCreationMode != CodeCreationMode.uploadWasm)
              ExtismToolbar(
                store: editorStore,
                isDownloadingProject: isDownloadingProject,
                onDownloadProject: onDownloadProject,
              ),
            Expanded(
              child: ModeContentWidget(
                editorStore: editorStore,
                buildMultiFileEditor: buildMultiFileEditor,
                onImportZip: onImportZip,
              ),
            ),
          ],
        );
      },
    );
  }
}
