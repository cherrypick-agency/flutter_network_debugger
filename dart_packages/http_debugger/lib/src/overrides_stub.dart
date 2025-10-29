/// Заглушка для платформ без dart:io (например, Web). Ничего не делает.
class HttpDebuggerConfig {
  final String proxyHostPort;
  final bool allowBadCertificates;
  final List<Object> bypassHosts;

  const HttpDebuggerConfig({
    required this.proxyHostPort,
    this.allowBadCertificates = false,
    this.bypassHosts = const [],
  });
}

class HttpReverseProxyConfig {
  final String upstreamBaseUrl;
  final String proxyBaseUrl;
  final String proxyHttpPath;
  final bool allowBadCertificates;
  final List<Pattern>? skipPaths;
  final List<Pattern>? skipHosts;
  final List<String>? skipMethods;
  final List<Pattern>? allowPaths;
  final List<Pattern>? allowHosts;
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

class HttpDebugger {
  HttpDebugger._();

  static void enableForwardProxy(HttpDebuggerConfig _) {}
  static void enableReverseProxy(HttpReverseProxyConfig _) {}
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
  static void disable() {}

  static T runZonedWithForwardProxy<T>(
    HttpDebuggerConfig _,
    T Function() action,
  ) => action();

  static T runZonedWithReverseProxy<T>(
    HttpReverseProxyConfig _,
    T Function() action,
  ) => action();

  static void enableAuto() {}
}


