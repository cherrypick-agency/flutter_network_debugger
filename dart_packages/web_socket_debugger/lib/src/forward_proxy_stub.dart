import 'package:web_socket_debugger/web_socket_debugger.dart';

WebSocketProxyConfig forwardProxyAttach({
  required String baseUrl,
  required String proxyHostPort,
  bool allowBadCerts = false,
}) {
  return WebSocketProxyConfig(
    connectUrl: Uri.parse(baseUrl),
    query: const {},
    useForwardOverrides: false,
  );
}
