import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

/// Stub for forward proxy in environments without IO support.
///
/// Returns config without proxy since forward mode is unavailable.
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
