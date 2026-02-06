import 'package:flutter/material.dart';
import '../../application/stores/script_editor_store.dart';
import 'wasm_upload_zone.dart';

class UploadWasmMode extends StatelessWidget {
  final ScriptEditorStore editorStore;

  const UploadWasmMode({
    super.key,
    required this.editorStore,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 600),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: WasmUploadZone(
            onWasmUploaded: (base64) {
              editorStore.setCode(base64);
            },
            onClear: () {
              editorStore.setCode('');
            },
          ),
        ),
      ),
    );
  }
}
