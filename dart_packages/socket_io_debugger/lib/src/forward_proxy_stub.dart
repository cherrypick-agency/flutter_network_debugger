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
    useForwardOverrides: false,
  );
}
