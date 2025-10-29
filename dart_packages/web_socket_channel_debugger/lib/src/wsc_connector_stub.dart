import 'package:web_socket_channel/web_socket_channel.dart' as wsc;

// В вебе/не-IO окружении headers не поддерживаются — просто игнорируем их
wsc.WebSocketChannel connectWS(Uri uri, {Map<String, dynamic>? headers}) {
  return wsc.WebSocketChannel.connect(uri);
}


