import 'package:web_socket_debugger/web_socket_debugger.dart';

Future<void> main() async {
  const upstream = 'wss://ws.postman-echo.com/raw';
  final cfg = WebSocketDebugger.attach(
    baseUrl: upstream,
    proxyBaseUrl: 'http://localhost:9091',
    proxyPath: '/wsproxy',
  );

  final socket = await WebSocketDebugger.connect(config: cfg);
  socket.events.listen((e) => print(e));
  socket.sendText('hello');
}
