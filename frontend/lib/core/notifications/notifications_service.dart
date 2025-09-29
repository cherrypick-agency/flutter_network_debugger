import 'dart:async';

import 'notification.dart';
import '../network/error_utils.dart';
import 'package:app_http_client/application/server_error.dart';

class NotificationsService {
  final _controller = StreamController<NotificationMessage>.broadcast();
  final Map<String, DateTime> _recent = {};
  final Duration _dedupWindow = const Duration(seconds: 3);

  Stream<NotificationMessage> get stream => _controller.stream;

  void info(String title, String description) {
    _emit(NotificationLevel.info, title, description);
  }

  void warn(String title, String description) {
    _emit(NotificationLevel.warning, title, description);
  }

  void error(String title, String description) {
    _emit(NotificationLevel.error, title, description);
  }

  void errorFromResolved(ResolvedErrorMessage msg) {
    // Подавляем шум "SocketException: Connection refused" (Flutter runtime)
    if (_isConnectionRefused(msg)) return;
    final details = msg.details ?? const {};
    final title = '[${msg.code?.name.toUpperCase() ?? 'ERROR'}] ${msg.title}';
    var desc = msg.description + (details.isNotEmpty ? ' — ${details['method'] ?? ''} ${details['url'] ?? ''} ${details['statusCode'] ?? ''}' : '');
    // Санитизация: если кто-то передал объект вместо строки
    if (desc.contains("Instance of 'ResolvedErrorMessage'")) {
      desc = (msg.raw?.toString().trim().isNotEmpty ?? false) ? msg.raw!.toString() : msg.description;
    }
    // Если совсем нечего показать и код unknown — не шумим
    if ((desc.trim().isEmpty || desc.trim() == 'Unexpected error occurred.') && (msg.code == ServerErrorCode.unknown)) {
      return;
    }
    _controller.add(NotificationMessage(
      level: NotificationLevel.error,
      title: title,
      description: desc,
      raw: msg.raw,
      stack: msg.stack,
      details: msg.details,
    ));
  }

  void dispose() {
    _controller.close();
  }

  void _emit(NotificationLevel level, String title, String description) {
    // Глобальное подавление Connection refused вне зависимости от канала
    if (_isConnRefusedStr(title) || _isConnRefusedStr(description)) return;
    final key = '${level.name}|$title|$description';
    final now = DateTime.now();
    final last = _recent[key];
    if (last != null && now.difference(last) < _dedupWindow) {
      return; // debounced duplicate
    }
    _recent[key] = now;
    _controller.add(NotificationMessage(level: level, title: title, description: description));
  }

  bool _isConnectionRefused(ResolvedErrorMessage msg) {
    final raw = (msg.raw ?? '').toLowerCase();
    final desc = (msg.description).toLowerCase();
    // Признаки: SocketException + Connection refused (errno 61 и аналоги)
    if (raw.contains('socketexception') && raw.contains('connection refused')) return true;
    if (desc.contains('connection refused')) return true;
    return false;
  }

  bool _isConnRefusedStr(String s) {
    final t = s.toLowerCase();
    return t.contains('socketexception') && t.contains('connection refused');
  }
}


