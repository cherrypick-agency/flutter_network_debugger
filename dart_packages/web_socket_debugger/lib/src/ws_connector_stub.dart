import 'package:web_socket/web_socket.dart' as ws;

// В средах без dart:io заголовки не поддерживаются — игнорируем
Future<ws.WebSocket> connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return ws.WebSocket.connect(uri);
}

