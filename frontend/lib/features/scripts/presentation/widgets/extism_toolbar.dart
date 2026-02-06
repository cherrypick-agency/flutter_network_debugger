import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import '../../application/stores/script_editor_store.dart';

class ExtismToolbar extends StatelessWidget {
  final ScriptEditorStore store;
  final bool isDownloadingProject;
  final VoidCallback onDownloadProject;

  const ExtismToolbar({
    super.key,
    required this.store,
    required this.isDownloadingProject,
    required this.onDownloadProject,
  });

  @override
  Widget build(BuildContext context) {
    return Observer(
      builder: (_) {
        final hasSourceFiles = store.sourceFiles.isNotEmpty;
        final isImportZipWithFiles =
            store.codeCreationMode == CodeCreationMode.importZip &&
            hasSourceFiles;
        final isEditing = store.isEditing;
        final isValidating = store.isValidating;

        final leftButtons = <Widget>[];

        if (hasSourceFiles) {
          leftButtons.add(
            OutlinedButton.icon(
              onPressed: isValidating ? null : store.validateSyntax,
              icon: isValidating
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.fact_check_outlined, size: 18),
              label: Text(isValidating ? 'Validating...' : 'Validate'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
              ),
            ),
          );
        }

        if (isImportZipWithFiles) {
          if (leftButtons.isNotEmpty) leftButtons.add(const SizedBox(width: 8));
          leftButtons.add(
            OutlinedButton.icon(
              onPressed: () {
                store.clearSourceFiles();
              },
              icon: const Icon(Icons.refresh, size: 18),
              label: const Text('Re-import ZIP'),
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 8,
                ),
              ),
            ),
          );
        }

        Widget? rightButton;
        if (isEditing) {
          rightButton = OutlinedButton.icon(
            onPressed: isDownloadingProject ? null : onDownloadProject,
            icon: isDownloadingProject
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.download, size: 18),
            label: Text(
              isDownloadingProject ? 'Downloading...' : 'Download Project',
            ),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 8,
              ),
            ),
          );
        }

        if (leftButtons.isEmpty && rightButton == null) {
          return const SizedBox.shrink();
        }

        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: Theme.of(
              context,
            ).colorScheme.surfaceContainerHighest.withOpacity(0.2),
            border: Border(
              bottom: BorderSide(
                color: Theme.of(context).dividerColor,
                width: 1,
              ),
            ),
          ),
          child: Row(
            children: [
              ...leftButtons,
              const Spacer(),
              if (rightButton != null) rightButton,
            ],
          ),
        );
      },
    );
  }
}
