import 'package:flutter/material.dart';
import 'notification.dart';

class NotificationSnackbar {
  static void show(BuildContext context, NotificationMessage n) {
    final scheme = Theme.of(context).colorScheme;
    Color color;
    Color textColor;
    switch (n.level) {
      case NotificationLevel.error:
        color = scheme.error;
        textColor = scheme.onErrorContainer;
        break;
      case NotificationLevel.warning:
        color = scheme.tertiary;
        textColor = scheme.onTertiaryContainer;
        break;
      case NotificationLevel.info:
        color = scheme.primary;
        textColor = scheme.onPrimaryContainer;
        break;
    }
    final controller = ScaffoldMessenger.of(context);
    controller.showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Expanded(
              child: Text(
                '${n.title}: ${n.description}',
                style: TextStyle(color: textColor),
              ),
            ),
            TextButton(
              onPressed: () {
                controller.hideCurrentSnackBar();
                showModalBottomSheet(
                  context: context,
                  isScrollControlled: true,
                  builder: (_) {
                    return Padding(
                      padding: const EdgeInsets.all(12),
                      child: SingleChildScrollView(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              n.title,
                              style: Theme.of(context).textTheme.titleMedium,
                            ),
                            const SizedBox(height: 8),
                            if (n.description.isNotEmpty)
                              SelectableText(n.description),
                            const SizedBox(height: 8),
                            if ((n.details ?? {}).isNotEmpty)
                              SelectableText('Details: ${n.details}'),
                            if ((n.raw ?? '').isNotEmpty) ...[
                              const SizedBox(height: 12),
                              Text(
                                'Raw',
                                style: Theme.of(context).textTheme.labelLarge,
                              ),
                              SelectableText(n.raw!),
                            ],
                            if ((n.stack ?? '').isNotEmpty) ...[
                              const SizedBox(height: 12),
                              Text(
                                'Stack trace',
                                style: Theme.of(context).textTheme.labelLarge,
                              ),
                              SelectableText(n.stack!),
                            ],
                          ],
                        ),
                      ),
                    );
                  },
                );
              },
              child: Text('Details', style: TextStyle(color: textColor)),
            ),
          ],
        ),
        backgroundColor: color.withValues(alpha: 0.85),
        duration: const Duration(seconds: 6),
      ),
    );
  }
}
