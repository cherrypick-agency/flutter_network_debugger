import 'dart:io';
import 'package:socket_io_debugger/socket_io_debugger.dart';

SocketIoConfig forwardProxyAttach({
  required String baseUrl,
  required String path,
  required String proxyHostPort,
  bool allowBadCerts = false,
}) {
  return SocketIoConfig(
    effectiveBaseUrl: baseUrl,
    effectivePath: path,
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
