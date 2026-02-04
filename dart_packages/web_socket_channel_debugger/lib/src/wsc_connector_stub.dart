import 'package:web_socket_channel/web_socket_channel.dart' as wsc;

// In web/non-IO environment headers are not supported - we simply ignore them
wsc.WebSocketChannel connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return wsc.WebSocketChannel.connect(uri);
}
