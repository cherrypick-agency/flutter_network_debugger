import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

Future<void> main() async {
  const upstream = 'wss://echo.websocket.events';
  final cfg = WebSocketChannelDebugger.attach(
    baseUrl: upstream,
    proxyBaseUrl: 'http://localhost:9091',
    proxyPath: '/wsproxy',
  );

  final ch = WebSocketChannelDebugger.connect(config: cfg);
  ch.stream.listen((m) => print('> $m'));
  ch.sink.add('hello');
}
