import 'package:web_socket_channel/web_socket_channel.dart' as wsc;

/// Connects to WebSocket server in web environment.
///
/// Headers are not supported and are ignored.
wsc.WebSocketChannel connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return wsc.WebSocketChannel.connect(uri);
}
