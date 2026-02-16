import 'package:web_socket_channel/web_socket_channel.dart' as wsc;
import 'package:web_socket_channel/io.dart' as io;

/// Connects to WebSocket server in IO environment.
///
/// Supports passing headers (Authorization, Cookie, etc.).
wsc.WebSocketChannel connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return io.IOWebSocketChannel.connect(uri, headers: headers);
}
