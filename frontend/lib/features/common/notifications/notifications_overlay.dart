import 'package:flutter/material.dart';

import '../../../core/di/di.dart';
import '../../../core/notifications/notifications_service.dart';
import '../../../core/notifications/notification_snackbar.dart';
import '../../../core/notifications/notification.dart';

class NotificationsOverlay extends StatelessWidget {
  const NotificationsOverlay({super.key});

  @override
  Widget build(BuildContext context) {
    return Positioned.fill(
      child: IgnorePointer(
        child: StreamBuilder(
          stream: sl<NotificationsService>().stream,
          builder: (context, snapshot) {
            if (!snapshot.hasData) return const SizedBox.shrink();
            final n = snapshot.data as NotificationMessage;
            WidgetsBinding.instance.addPostFrameCallback((_) {
              NotificationSnackbar.show(context, n);
            });
            return const SizedBox.shrink();
          },
        ),
      ),
    );
  }
}
