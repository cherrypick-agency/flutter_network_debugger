import 'dart:io';
import 'package:web_socket_debugger/web_socket_debugger.dart';

/// Configures forward proxy for WebSocket in IO environment.
///
/// Creates a config that uses system HTTP client with proxy settings.
/// All connections go through the specified proxy server.
///
/// [baseUrl] - original WebSocket connection URL.
/// [proxyHostPort] - proxy server address in "host:port" format.
/// [allowBadCerts] - allow invalid SSL certificates (for debugging).
///
/// Returns configuration with forward proxy mode enabled.
WebSocketProxyConfig forwardProxyAttach({
  required String baseUrl,
  required String proxyHostPort,
  bool allowBadCerts = false,
}) {
  return WebSocketProxyConfig(
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
