import 'package:socket_io_client/socket_io_client.dart' as io;
import 'package:app_http_client/application/app_http_client.dart'
    as http_client;
import '../di/di.dart';
import 'package:socket_io_debugger/socket_io_debugger.dart';

typedef SocketHandler = void Function(dynamic data);

class SocketIoService {
  io.Socket? _socket;

  Future<void> connect() async {
    if (_socket != null) {
      print('[SocketIO] connect() called but already connected');
      return;
    }
    print('[SocketIO] connect() starting...');
    final base = sl<http_client.AppHttpClient>().defaultHost;
    // Socket.IO конфигурация:
    // baseUrl - только хост (http://localhost:9092)
    // path - полный HTTP путь к Socket.IO endpoint (/_api/v1/monitor/io/socket.io/)
    final cfg = SocketIoDebugger.attach(
      enabled: false,
      baseUrl: base, // Только хост
      path:
          '/_api/v1/monitor/io/socket.io/', // Полный HTTP путь к Socket.IO endpoint
      proxyBaseUrl: 'http://localhost:9091',
    );
    print(
      '[SocketIO] Config: baseUrl=${cfg.effectiveBaseUrl}, path=${cfg.effectivePath}',
    );
    final socket = io.io(
      cfg.effectiveBaseUrl,
      io.OptionBuilder()
          .setTransports(['websocket'])
          .setPath(cfg.effectivePath)
          .setQuery(cfg.query)
          .enableReconnection()
          .setReconnectionAttempts(1 << 20)
          .setReconnectionDelay(500)
          .setReconnectionDelayMax(5000)
          .build(),
    );

    // Универсальный обработчик ВСЕХ событий для debug
    socket.onAny((event, data) {
      print('[SocketIO] ⬇️ Received event: "$event" with data: $data');
    });

    socket.on('connect', (_) {
      print('[SocketIO] ✅ Connected! Socket ID: ${socket.id}');
    });

    socket.on('disconnect', (reason) {
      print('[SocketIO] ❌ Disconnected. Reason: $reason');
    });

    socket.on('connect_error', (error) {
      print('[SocketIO] 🔴 Connection error: $error');
    });

    _socket = socket;
    print('[SocketIO] Socket instance created and stored');
  }

  void on(String event, SocketHandler handler) {
    print('[SocketIO] 📝 Registering handler for event: "$event"');
    _socket?.on(event, handler);
  }

  void off(String event, [SocketHandler? handler]) {
    print('[SocketIO] 🗑️ Removing handler for event: "$event"');
    if (handler != null) {
      _socket?.off(event, handler);
    } else {
      _socket?.off(event);
    }
  }

  void emit(String event, dynamic data) {
    print('[SocketIO] ⬆️ Emitting event: "$event" with data: $data');
    _socket?.emit(event, data);
  }

  void dispose() {
    try {
      _socket?.dispose();
    } catch (_) {}
    _socket = null;
  }
}
