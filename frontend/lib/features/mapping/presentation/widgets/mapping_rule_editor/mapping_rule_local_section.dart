import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../../../theme/context_ext.dart';

class MappingRuleLocalSection extends StatelessWidget {
  const MappingRuleLocalSection({
    super.key,
    required this.filePathController,
    required this.statusOverrideController,
    required this.contentTypeOverrideController,
    required this.saving,
    required this.blobPath,
    required this.onPickAndUpload,
    required this.onClearUploaded,
    required this.onFilePathChanged,
    this.filePathErrorText,
    this.blobPathErrorText,
    this.statusOverrideErrorText,
    this.contentTypeOverrideErrorText,
    this.onStatusOverrideChanged,
    this.onContentTypeOverrideChanged,
  });

  final TextEditingController filePathController;
  final TextEditingController statusOverrideController;
  final TextEditingController contentTypeOverrideController;

  final bool saving;
  final String? blobPath;

  final VoidCallback onPickAndUpload;
  final VoidCallback onClearUploaded;
  final ValueChanged<String> onFilePathChanged;

  final String? filePathErrorText;
  final String? blobPathErrorText;
  final String? statusOverrideErrorText;
  final String? contentTypeOverrideErrorText;

  final ValueChanged<String>? onStatusOverrideChanged;
  final ValueChanged<String>? onContentTypeOverrideChanged;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: filePathController,
                enabled: !saving,
                decoration: InputDecoration(
                  labelText: 'File path (optional)',
                  hintText: '/path/to/file',
                  errorText: filePathErrorText,
                ),
                onChanged: onFilePathChanged,
              ),
            ),
            const SizedBox(width: 12),
            SizedBox(
              width: 200,
              child: TextButton.icon(
                onPressed: saving ? null : onPickAndUpload,
                icon: const Icon(Icons.file_upload),
                label: Text(blobPath == null ? 'Upload file' : 'Re-upload'),
              ),
            ),
          ],
        ),
        if (blobPath != null) ...[
          const SizedBox(height: 6),
          Row(
            children: [
              Expanded(
                child: Text(
                  'Uploaded: $blobPath',
                  style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                tooltip: 'Clear uploaded',
                onPressed: saving ? null : onClearUploaded,
                icon: const Icon(Icons.clear),
              ),
            ],
          ),
        ],
        if (blobPathErrorText != null) ...[
          const SizedBox(height: 6),
          Text(
            blobPathErrorText!,
            style: context.appText.body.copyWith(
              color: Theme.of(context).colorScheme.error,
            ),
          ),
        ],
        const SizedBox(height: 8),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: statusOverrideController,
                enabled: !saving,
                keyboardType: TextInputType.number,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: InputDecoration(
                  labelText: 'HTTP status override (optional)',
                  errorText: statusOverrideErrorText,
                ),
                onChanged: onStatusOverrideChanged,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: TextField(
                controller: contentTypeOverrideController,
                enabled: !saving,
                decoration: InputDecoration(
                  labelText: 'Content-Type override (optional)',
                  errorText: contentTypeOverrideErrorText,
                ),
                onChanged: onContentTypeOverrideChanged,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        Text(
          'Если указан file path — blob не используется.',
          style: context.appText.body.copyWith(
            color: context.appColors.textSecondary,
          ),
        ),
      ],
    );
  }
}
