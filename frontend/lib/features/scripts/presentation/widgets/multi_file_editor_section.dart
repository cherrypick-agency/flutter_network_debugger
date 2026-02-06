import 'package:flutter/material.dart';
import '../../infrastructure/editor_di.dart';
import 'multi_file_editor_widget.dart';

class MultiFileEditorSection extends StatelessWidget {
  final Future<void>? initEditorFuture;
  final EditorDI? editorDI;

  const MultiFileEditorSection({
    super.key,
    required this.initEditorFuture,
    required this.editorDI,
  });

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<void>(
      future: initEditorFuture,
      builder: (context, snapshot) {
        if (editorDI == null) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError && editorDI == null) {
          return Center(
            child: Text('Failed to initialize editor: ${snapshot.error}'),
          );
        }

        return MultiFileEditorWidget(editorDI: editorDI!);
      },
    );
  }
}
