/// Library for debugging HTTP requests through a proxy server.
///
/// Supports two modes:
/// - Forward proxy: routes all requests through the proxy server
/// - Reverse proxy: rewrites URLs to work through reverse-proxy
///
/// On platforms with dart:io support (iOS, Android, Desktop) works fully;
/// on Web and other platforms exports stubs without side effects.
library http_debugger;

// Export platform-specific implementation: on IO - working,
// on Web and others - stub without side effects.
export 'src/overrides_stub.dart' if (dart.library.io) 'src/overrides_io.dart';
export 'src/http_client_wrapper.dart';
