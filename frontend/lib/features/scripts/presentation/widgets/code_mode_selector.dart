import 'package:flutter/material.dart';
import '../../application/stores/script_editor_store.dart';

/// Widget for selecting code creation mode
class CodeModeSelector extends StatelessWidget {
  final CodeCreationMode selectedMode;
  final ValueChanged<CodeCreationMode> onModeChanged;

  const CodeModeSelector({
    super.key,
    required this.selectedMode,
    required this.onModeChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(
          context,
        ).colorScheme.surfaceContainerHighest.withOpacity(0.3),
        border: Border(
          bottom: BorderSide(color: Theme.of(context).dividerColor, width: 1),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Code Creation Mode',
            style: Theme.of(
              context,
            ).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),
          SegmentedButton<CodeCreationMode>(
            segments: const [
              ButtonSegment(
                value: CodeCreationMode.uploadWasm,
                label: Text('Upload WASM'),
                icon: Icon(Icons.upload_file, size: 18),
              ),
              ButtonSegment(
                value: CodeCreationMode.writeSource,
                label: Text('Write Source'),
                icon: Icon(Icons.code, size: 18),
              ),
              ButtonSegment(
                value: CodeCreationMode.importZip,
                label: Text('Import ZIP'),
                icon: Icon(Icons.folder_zip, size: 18),
              ),
            ],
            selected: {selectedMode},
            onSelectionChanged: (Set<CodeCreationMode> newSelection) {
              onModeChanged(newSelection.first);
            },
          ),
          const SizedBox(height: 8),
          _buildModeDescription(context),
        ],
      ),
    );
  }

  Widget _buildModeDescription(BuildContext context) {
    final description = switch (selectedMode) {
      CodeCreationMode.uploadWasm =>
        'Upload a pre-compiled WASM file (.wasm or .wat)',
      CodeCreationMode.writeSource =>
        'Write source code in Rust, Go, AssemblyScript, or other languages and compile to WASM',
      CodeCreationMode.importZip =>
        'Import a multi-file project from a ZIP archive',
    };

    return Text(
      description,
      style: Theme.of(
        context,
      ).textTheme.bodySmall?.copyWith(color: Colors.grey[600]),
    );
  }
}
