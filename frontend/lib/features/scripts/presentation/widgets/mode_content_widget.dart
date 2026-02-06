import 'package:flutter/material.dart';
import '../../application/stores/script_editor_store.dart';
import '../../domain/entities/script.dart';
import 'upload_wasm_mode.dart';
import 'import_zip_mode.dart';

class ModeContentWidget extends StatelessWidget {
  final ScriptEditorStore editorStore;
  final Widget Function() buildMultiFileEditor;
  final VoidCallback onImportZip;

  const ModeContentWidget({
    super.key,
    required this.editorStore,
    required this.buildMultiFileEditor,
    required this.onImportZip,
  });

  @override
  Widget build(BuildContext context) {
    if (editorStore.runtime == ScriptRuntime.extism &&
        editorStore.codeCreationMode == CodeCreationMode.uploadWasm) {
      return UploadWasmMode(editorStore: editorStore);
    }

    if (editorStore.codeCreationMode == CodeCreationMode.importZip &&
        editorStore.sourceFiles.isEmpty) {
      return ImportZipMode(onImportZip: onImportZip);
    }

    return buildMultiFileEditor();
  }
}
