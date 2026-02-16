import 'package:web_socket/web_socket.dart' as ws;

/// Establishes WebSocket connection in environments without dart:io.
///
/// Creates connection using web_socket library. On web, headers are not
/// supported at transport level, so [headers] is ignored.
///
/// [uri] - URI for connection.
/// [headers] - HTTP headers (not supported on web).
///
/// Returns the established WebSocket connection.
Future<ws.WebSocket> connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return ws.WebSocket.connect(uri);
}
