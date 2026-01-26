import 'package:web_socket/web_socket.dart' as ws;

// web_socket library doesn't accept headers as named parameter - ignored
Future<ws.WebSocket> connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return ws.WebSocket.connect(uri);
}
