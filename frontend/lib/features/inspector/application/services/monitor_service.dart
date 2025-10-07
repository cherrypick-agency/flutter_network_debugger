import 'dart:async';
import 'dart:convert';
import 'dart:io' as io;
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart' as ws_io;
import 'package:app_http_client/application/app_http_client.dart'
    as http_client;

import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/network/ws_url.dart';

typedef MonitorListener = void Function(Map<String, dynamic> event);

class MonitorService {
  WebSocketChannel? _channel;
  final List<MonitorListener> _listeners = [];
  StreamSubscription? _sub;

  void addListener(MonitorListener l) {
    _listeners.add(l);
  }

  void removeListener(MonitorListener l) {
    _listeners.remove(l);
  }

  Future<void> connect() async {
    final base = sl<http_client.AppHttpClient>().defaultHost;
    final url = buildWsUrl(base, '/_api/v1/monitor/ws');
    try {
      final sock = await io.WebSocket.connect(
        url,
      ).timeout(const Duration(seconds: 3));
      final ch = ws_io.IOWebSocketChannel(sock);
      _channel = ch;
      _sub?.cancel();
      _sub = _channel!.stream.listen(
        (msg) {
          try {
            final s = msg.toString();
            final Map<String, dynamic> ev = jsonDecode(s);
            for (final l in _listeners) {
              l(ev);
            }
          } catch (_) {}
        },
        onError: (_) {},
        onDone: () {
          // reconnect silently later
        },
      );
    } catch (e) {
      // swallow; visual indicators handle connectivity
      sl<NotificationsService>();
    }
  }

  void dispose() {
    _sub?.cancel();
    _channel?.sink.close();
    _sub = null;
    _channel = null;
  }
}
