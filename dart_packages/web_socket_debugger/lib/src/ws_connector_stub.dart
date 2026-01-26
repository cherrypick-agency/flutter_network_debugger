import 'package:web_socket/web_socket.dart' as ws;

// In environments without dart:io headers are not supported - ignored
Future<ws.WebSocket> connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return ws.WebSocket.connect(uri);
}
