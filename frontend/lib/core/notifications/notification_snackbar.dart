import 'package:flutter/material.dart';
import 'notification.dart';

class NotificationSnackbar {
  static void show(BuildContext context, NotificationMessage n) {
    Color color;
    switch (n.level) {
      case NotificationLevel.error:
        color = Theme.of(context).colorScheme.error;
        break;
      case NotificationLevel.warning:
        color = Theme.of(context).colorScheme.tertiary;
        break;
      case NotificationLevel.info:
        color = Theme.of(context).colorScheme.primary;
        break;
    }
    final controller = ScaffoldMessenger.of(context);
    controller.showSnackBar(
      SnackBar(
        content: Row(children: [
          Expanded(child: Text('${n.title}: ${n.description}')),
          TextButton(
            onPressed: () {
              controller.hideCurrentSnackBar();
              showModalBottomSheet(context: context, isScrollControlled: true, builder: (_) {
                return Padding(
                  padding: const EdgeInsets.all(12),
                  child: SingleChildScrollView(
                    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text('Error Details', style: Theme.of(context).textTheme.titleMedium),
                      const SizedBox(height: 8),
                      if (n.description.isNotEmpty) SelectableText(n.description),
                      const SizedBox(height: 8),
                      if ((n.details ?? {}).isNotEmpty) SelectableText('Details: '+(n.details!.toString())),
                      if ((n.raw ?? '').isNotEmpty) ...[
                        const SizedBox(height: 12),
                        Text('Raw', style: Theme.of(context).textTheme.labelLarge),
                        SelectableText(n.raw!),
                      ],
                      if ((n.stack ?? '').isNotEmpty) ...[
                        const SizedBox(height: 12),
                        Text('Stack trace', style: Theme.of(context).textTheme.labelLarge),
                        SelectableText(n.stack!),
                      ],
                    ]),
                  ),
                );
              });
            },
            child: const Text('Details'),
          ),
        ]),
        backgroundColor: color.withOpacity(0.1),
        duration: const Duration(seconds: 6),
      ),
    );
  }
}


