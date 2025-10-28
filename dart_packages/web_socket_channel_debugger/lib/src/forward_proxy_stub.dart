import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

WscProxyConfig forwardProxyAttach({
  required String baseUrl,
  required String proxyHostPort,
  bool allowBadCerts = false,
}) {
  return WscProxyConfig(
    connectUrl: Uri.parse(baseUrl),
    query: const {},
    useForwardOverrides: false,
  );
}
