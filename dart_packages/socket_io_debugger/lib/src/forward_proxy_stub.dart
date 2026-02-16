import 'package:socket_io_debugger/socket_io_debugger.dart';

/// Stub for forward proxy in environments without IO support.
///
/// Returns config without proxy since forward mode is unavailable.
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
