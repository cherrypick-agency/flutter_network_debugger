import 'package:web_socket_channel/web_socket_channel.dart' as wsc;
import 'package:web_socket_channel/io.dart' as io;

// In IO environment we pass through headers (Authorization/Cookie/etc.)
wsc.WebSocketChannel connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return io.IOWebSocketChannel.connect(uri, headers: headers);
}
