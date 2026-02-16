import 'package:web_socket/web_socket.dart' as ws;

/// Establishes WebSocket connection in IO environment.
///
/// Creates connection using web_socket library. The [headers] parameter
/// is ignored since web_socket does not support passing headers.
///
/// [uri] - URI for connection.
/// [headers] - HTTP headers (ignored by web_socket library).
///
/// Returns the established WebSocket connection.
Future<ws.WebSocket> connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return ws.WebSocket.connect(uri);
}
