import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../domain/entities/update_info.dart';

/// Shows a dialog with information about the available update
Future<UpdateDialogResult?> showUpdateDialog(
  BuildContext context,
  UpdateInfo updateInfo,
) async {
  return await showDialog<UpdateDialogResult>(
    context: context,
    barrierDismissible: false,
    builder: (ctx) => _UpdateDialog(updateInfo: updateInfo),
  );
}

enum UpdateDialogResult { download, skip, remindLater, viewAllReleases }

class _UpdateDialog extends StatelessWidget {
  final UpdateInfo updateInfo;

  const _UpdateDialog({required this.updateInfo});

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          const Icon(Icons.system_update, color: Colors.blue, size: 28),
          const SizedBox(width: 12),
          const Text('Update Available'),
        ],
      ),
      content: SizedBox(
        width: 500,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Version info
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.blue.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.blue.withValues(alpha: 0.3)),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'New Version',
                          style: TextStyle(fontSize: 12, color: Colors.grey),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          updateInfo.version,
                          style: const TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
                            color: Colors.blue,
                          ),
                        ),
                      ],
                    ),
                    if (updateInfo.sizeBytes > 0)
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          const Text(
                            'Size',
                            style: TextStyle(fontSize: 12, color: Colors.grey),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            updateInfo.formattedSize,
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // Changelog
              const Text(
                'What\'s New:',
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),

              Container(
                constraints: const BoxConstraints(maxHeight: 200),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.grey.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
                ),
                child: SingleChildScrollView(
                  child: Text(
                    updateInfo.changelog.isNotEmpty
                        ? updateInfo.changelog
                        : 'No changelog available',
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
              ),

              const SizedBox(height: 16),

              // Info message
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.orange.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: Colors.orange.withValues(alpha: 0.3),
                  ),
                ),
                child: const Row(
                  children: [
                    Icon(Icons.info_outline, color: Colors.orange, size: 20),
                    SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        'The update will be downloaded from GitHub. You\'ll need to install it manually.',
                        style: TextStyle(fontSize: 12),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton.icon(
          onPressed: () =>
              Navigator.of(context).pop(UpdateDialogResult.viewAllReleases),
          icon: const Icon(Icons.history, size: 18),
          label: const Text('View All Releases'),
        ),
        const Spacer(),
        TextButton(
          onPressed: () => Navigator.of(context).pop(UpdateDialogResult.skip),
          child: const Text('Skip This Version'),
        ),
        TextButton(
          onPressed: () =>
              Navigator.of(context).pop(UpdateDialogResult.remindLater),
          child: const Text('Remind Me Later'),
        ),
        ElevatedButton.icon(
          onPressed: () =>
              Navigator.of(context).pop(UpdateDialogResult.download),
          icon: const Icon(Icons.download),
          label: const Text('Download Update'),
        ),
      ],
    );
  }
}

/// Opens URL to download the update
Future<void> openUpdateDownloadUrl(String url) async {
  final uri = Uri.parse(url);
  if (await canLaunchUrl(uri)) {
    await launchUrl(uri, mode: LaunchMode.externalApplication);
  } else {
    throw Exception('Could not launch $url');
  }
}
