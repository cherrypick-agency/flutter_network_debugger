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
    useForwardOverrides: false,
  );
}
