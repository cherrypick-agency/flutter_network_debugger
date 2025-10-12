import 'dart:io';
import 'package:socket_io_debugger/socket_io_debugger.dart';

SocketIoConfig forwardProxyAttach({
  required String baseUrl,
  required String socketPath,
  required String proxyHostPort,
  bool allowBadCerts = false,
}) {
  return SocketIoConfig(
    effectiveBaseUrl: baseUrl,
    effectivePath: socketPath,
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
