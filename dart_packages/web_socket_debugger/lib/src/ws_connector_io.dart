import 'package:web_socket/web_socket.dart' as ws;

// Библиотека web_socket не принимает headers как именованный параметр — игнорируем
Future<ws.WebSocket> connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return ws.WebSocket.connect(uri);
}
