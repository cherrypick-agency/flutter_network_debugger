import 'dart:io';
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

/// Configures forward proxy for WebSocket connection.
///
/// All connections will go through the specified proxy server.
WscProxyConfig forwardProxyAttach({
  required String baseUrl,
  required String proxyHostPort,
  bool allowBadCerts = false,
}) {
  return WscProxyConfig(
    connectUrl: Uri.parse(baseUrl),
    query: const {},
    useForwardOverrides: true,
    httpClientFactory: () {
      final client = HttpClient();
      client.findProxy = (uri) => 'PROXY $proxyHostPort';
      if (allowBadCerts) {
        client.badCertificateCallback = (cert, host, port) => true;
      }
      return client;
    },
  );
}
