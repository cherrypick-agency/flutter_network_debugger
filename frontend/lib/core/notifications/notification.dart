enum NotificationLevel { info, warning, error }

class NotificationMessage {
  final NotificationLevel level;
  final String title;
  final String description;
  final DateTime ts;
  final String? raw;
  final String? stack;
  final Map<String, dynamic>? details;

  NotificationMessage({
    required this.level,
    required this.title,
    required this.description,
    DateTime? ts,
    this.raw,
    this.stack,
    this.details,
  }) : ts = ts ?? DateTime.now();
}


