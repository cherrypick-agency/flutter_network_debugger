/// Stub for platforms without dart:io (e.g. Web). Performs no operations.
class HttpDebuggerConfig {
  /// Proxy server host and port (not used in stub).
  final String proxyHostPort;

  /// Allow invalid SSL certificates (not used in stub).
  final bool allowBadCertificates;

  /// List of hosts to bypass proxy (not used in stub).
  final List<Object> bypassHosts;

  const HttpDebuggerConfig({
    required this.proxyHostPort,
    this.allowBadCertificates = false,
    this.bypassHosts = const [],
  });
}

/// Stub for reverse-proxy configuration on platforms without dart:io.
class HttpReverseProxyConfig {
  /// Target server base URL (not used in stub).
  final String upstreamBaseUrl;

  /// Proxy server base URL (not used in stub).
  final String proxyBaseUrl;

  /// Path on the proxy server (not used in stub).
  final String proxyHttpPath;

  /// Allow invalid SSL certificates (not used in stub).
  final bool allowBadCertificates;

  /// Path patterns to skip (not used in stub).
  final List<Pattern>? skipPaths;

  /// Host patterns to skip (not used in stub).
  final List<Pattern>? skipHosts;

  /// HTTP methods to skip (not used in stub).
  final List<String>? skipMethods;

  /// Allowed path patterns (not used in stub).
  final List<Pattern>? allowPaths;

  /// Allowed host patterns (not used in stub).
  final List<Pattern>? allowHosts;

  /// Allowed HTTP methods (not used in stub).
  final List<String>? allowMethods;

  const HttpReverseProxyConfig({
    required this.upstreamBaseUrl,
    required this.proxyBaseUrl,
    this.proxyHttpPath = '/httpproxy',
    this.allowBadCertificates = false,
    this.skipPaths,
    this.skipHosts,
    this.skipMethods,
    this.allowPaths,
    this.allowHosts,
    this.allowMethods,
  });
}

/// Stub class for proxy management on platforms without dart:io.
class HttpDebugger {
  HttpDebugger._();

  /// Enables forward-proxy (stub, no-op).
  static void enableForwardProxy(HttpDebuggerConfig _) {}

  /// Enables reverse-proxy (stub, no-op).
  static void enableReverseProxy(HttpReverseProxyConfig _) {}

  /// Generic way to enable proxy (stub, no-op).
  static void enable({
    String mode = 'reverse',
    String? upstreamBaseUrl,
    String? proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    bool allowBadCertificates = false,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
    List<Pattern> bypassHosts = const [],
  }) {}

  /// Disables proxy (stub, no-op).
  static void disable() {}

  /// Runs action in a zone with forward-proxy (stub, just runs the action).
  static T runZonedWithForwardProxy<T>(
    HttpDebuggerConfig _,
    T Function() action,
  ) =>
      action();

  /// Runs action in a zone with reverse-proxy (stub, just runs the action).
  static T runZonedWithReverseProxy<T>(
    HttpReverseProxyConfig _,
    T Function() action,
  ) =>
      action();

  /// Automatically enables proxy from environment variables (stub, no-op).
  static void enableAuto() {}
}
