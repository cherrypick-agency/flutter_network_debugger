import 'dart:io';
import 'reverse_http_client_io.dart';

/// Global forward-proxy configuration via HttpOverrides.
class HttpDebuggerConfig {
  /// Proxy server host and port in 'host:port' format (no scheme).
  /// Used in HttpClient.findProxy to specify the proxy.
  final String proxyHostPort;

  /// Allow self-signed or invalid SSL certificates.
  /// Useful when developing with local certificates.
  final bool allowBadCertificates;

  /// List of hosts for which proxy is disabled (requests go direct).
  /// Can use exact strings or regular expressions.
  final List<Pattern> bypassHosts;

  const HttpDebuggerConfig({
    required this.proxyHostPort,
    this.allowBadCertificates = false,
    this.bypassHosts = const [],
  });
}

/// Reverse-proxy configuration for global interception via HttpOverrides.
class HttpReverseProxyConfig {
  /// Target server base URL that requests will be proxied to.
  final String upstreamBaseUrl;

  /// Proxy server base URL. Scheme may be omitted, will be normalized automatically.
  final String proxyBaseUrl;

  /// Path on the proxy server for HTTP requests. Default '/httpproxy'.
  final String proxyHttpPath;

  /// Allow self-signed or invalid SSL certificates.
  final bool allowBadCertificates;

  /// Path patterns to skip (not proxy).
  final List<Pattern>? skipPaths;

  /// Host patterns to skip.
  final List<Pattern>? skipHosts;

  /// HTTP methods to skip.
  final List<String>? skipMethods;

  /// Path patterns allowed to proxy (takes precedence over skipPaths).
  final List<Pattern>? allowPaths;

  /// Host patterns allowed to proxy (takes precedence over skipHosts).
  final List<Pattern>? allowHosts;

  /// HTTP methods allowed to proxy (takes precedence over skipMethods).
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

/// Enables global forward-proxy for all HTTP traffic (dart:io).
/// All clients using the default HttpClient (including package:http, Dio by default, etc.)
/// will route requests through the specified proxy.
class HttpDebugger {
  HttpDebugger._();

  static HttpOverrides? _previous;

  /// Enables global forward-proxy for all HTTP traffic.
  /// Repeated calls overwrite previous settings.
  static void enableForwardProxy(HttpDebuggerConfig config) {
    _previous ??= HttpOverrides.current;
    HttpOverrides.global = _ForwardProxyOverrides(config);
  }

  /// Enables global reverse-proxy for all HTTP traffic.
  static void enableReverseProxy(HttpReverseProxyConfig config) {
    _previous ??= HttpOverrides.current;
    HttpOverrides.global = _ReverseProxyOverrides(config);
  }

  /// Disables global proxy and restores previous overrides (if any).
  static void disable() {
    HttpOverrides.global = _previous;
    _previous = null;
  }

  /// Generic way to enable proxy.
  ///
  /// Uses 'reverse' mode by default. If [upstreamBaseUrl] is not set,
  /// automatically switches to forward-proxy with local proxy.
  ///
  /// [mode] - operation mode: 'reverse' (default) or 'forward'.
  /// [upstreamBaseUrl] - target server base URL for reverse-proxy mode.
  /// [proxyBaseUrl] - proxy server base URL. If not set, default value is used.
  /// [proxyHttpPath] - path on the proxy server for HTTP requests.
  /// [allowBadCertificates] - allow invalid SSL certificates.
  /// [skipPaths] - path patterns to skip (reverse-proxy mode).
  /// [skipHosts] - host patterns to skip (reverse-proxy mode).
  /// [skipMethods] - HTTP methods to skip (reverse-proxy mode).
  /// [allowPaths] - allowed path patterns (reverse-proxy mode).
  /// [allowHosts] - allowed host patterns (reverse-proxy mode).
  /// [allowMethods] - allowed HTTP methods (reverse-proxy mode).
  /// [bypassHosts] - hosts to bypass proxy (forward-proxy mode).
  static void enable({
    String mode = 'reverse',
    String? upstreamBaseUrl,
    String? proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    bool allowBadCertificates = false,
    // reverse filters
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
    // forward options
    List<Pattern> bypassHosts = const [],
  }) {
    final m = mode.trim().toLowerCase();
    final defaultProxy =
        Platform.isAndroid ? 'http://10.0.2.2:9091' : 'http://localhost:9091';
    final proxy = (proxyBaseUrl == null || proxyBaseUrl.trim().isEmpty)
        ? defaultProxy
        : proxyBaseUrl;

    if (m == 'forward') {
      final hostPort = _normalizeProxyHostPort(proxy);
      return enableForwardProxy(
        HttpDebuggerConfig(
          proxyHostPort: hostPort,
          allowBadCertificates: allowBadCertificates,
          bypassHosts: bypassHosts,
        ),
      );
    }

    // reverse (default)
    final upstream = upstreamBaseUrl?.trim() ?? '';
    if (upstream.isEmpty) {
      // safe fallback to forward with default proxy
      final hostPort = _normalizeProxyHostPort(proxy);
      return enableForwardProxy(
        HttpDebuggerConfig(
          proxyHostPort: hostPort,
          allowBadCertificates: allowBadCertificates,
          bypassHosts: bypassHosts,
        ),
      );
    }

    return enableReverseProxy(
      HttpReverseProxyConfig(
        upstreamBaseUrl: upstream,
        proxyBaseUrl: proxy,
        proxyHttpPath: proxyHttpPath,
        allowBadCertificates: allowBadCertificates,
        skipPaths: skipPaths,
        skipHosts: skipHosts,
        skipMethods: skipMethods,
        allowPaths: allowPaths,
        allowHosts: allowHosts,
        allowMethods: allowMethods,
      ),
    );
  }

  /// Runs [action] in a zone with forward-proxy enabled.
  /// Convenient for isolated execution of code with proxy.
  static T runZonedWithForwardProxy<T>(
    HttpDebuggerConfig config,
    T Function() action,
  ) {
    return HttpOverrides.runZoned(
      action,
      createHttpClient: (SecurityContext? context) {
        final client = _RawHttpOverrides().createRawHttpClient(context);
        _configureClient(client, config);
        return client;
      },
    );
  }

  /// Runs [action] in a zone with reverse-proxy enabled.
  static T runZonedWithReverseProxy<T>(
    HttpReverseProxyConfig config,
    T Function() action,
  ) {
    return HttpOverrides.runZoned(
      action,
      createHttpClient: (SecurityContext? context) {
        return ReverseProxyHttpClient(
          upstreamBaseUrl: config.upstreamBaseUrl,
          proxyBaseUrl: config.proxyBaseUrl,
          proxyHttpPath: config.proxyHttpPath,
          allowBadCertificates: config.allowBadCertificates,
          skipPaths: config.skipPaths,
          skipHosts: config.skipHosts,
          skipMethods: config.skipMethods,
          allowPaths: config.allowPaths,
          allowHosts: config.allowHosts,
          allowMethods: config.allowMethods,
          context: context,
          // Important: create the inner client via super.createHttpClient,
          // otherwise we get infinite recursion due to HttpOverrides.runZoned.
          innerClient: _RawHttpOverrides().createRawHttpClient(context),
        );
      },
    );
  }

  /// Auto-select mode from ENV/--dart-define.
  ///
  /// HTTP_PROXY_MODE|PROXY_MODE = reverse|forward|none (default reverse)
  /// DIO_DEBUGGER_ENABLED|HTTP_PROXY_ENABLED - enable/disable (default true)
  /// For reverse mode, UPSTREAM_BASE_URL|API_HOST and PROXY_BASE_URL|HTTP_PROXY are required.
  static void enableAuto() {
    final enabled = _computeEnabledFromEnv();
    if (!enabled) return;

    final mode = _computeMode();
    if (mode == 'forward') {
      final proxy = _firstNonEmpty([
            const String.fromEnvironment('HTTP_PROXY'),
            const String.fromEnvironment('PROXY_BASE_URL'),
            _readEnvVar('HTTP_PROXY'),
            _readEnvVar('PROXY_BASE_URL'),
          ]) ??
          (Platform.isAndroid
              ? 'http://10.0.2.2:9091'
              : 'http://localhost:9091');
      final allowBad = _computeAllowBadCerts();
      final hostPort = _normalizeProxyHostPort(proxy);
      enableForwardProxy(
        HttpDebuggerConfig(
          proxyHostPort: hostPort,
          allowBadCertificates: allowBad,
        ),
      );
      return;
    }

    if (mode == 'reverse') {
      final upstream = _firstNonEmpty([
            const String.fromEnvironment('UPSTREAM_BASE_URL'),
            const String.fromEnvironment('API_HOST'),
            _readEnvVar('UPSTREAM_BASE_URL'),
            _readEnvVar('API_HOST'),
          ]) ??
          '';
      var proxyBase = _firstNonEmpty([
            const String.fromEnvironment('PROXY_BASE_URL'),
            const String.fromEnvironment('HTTP_PROXY'),
            _readEnvVar('PROXY_BASE_URL'),
            _readEnvVar('HTTP_PROXY'),
          ]) ??
          (Platform.isAndroid
              ? 'http://10.0.2.2:9091'
              : 'http://localhost:9091');
      final proxyPath = _firstNonEmpty([
            const String.fromEnvironment('PROXY_HTTP_PATH'),
            const String.fromEnvironment('HTTP_PROXY_PATH'),
            _readEnvVar('PROXY_HTTP_PATH'),
            _readEnvVar('HTTP_PROXY_PATH'),
          ]) ??
          '/httpproxy';
      if (upstream.trim().isEmpty) {
        // Fallback to forward‑proxy with default proxy when upstream is unknown
        final allowBad = _computeAllowBadCerts();
        final hostPort = _normalizeProxyHostPort(proxyBase);
        enableForwardProxy(
          HttpDebuggerConfig(
            proxyHostPort: hostPort,
            allowBadCertificates: allowBad,
          ),
        );
        return;
      }

      final allowBad = _computeAllowBadCerts();
      enableReverseProxy(
        HttpReverseProxyConfig(
          upstreamBaseUrl: upstream,
          proxyBaseUrl: proxyBase,
          proxyHttpPath: proxyPath,
          allowBadCertificates: allowBad,
        ),
      );
      return;
    }
  }

  static void _configureClient(HttpClient client, HttpDebuggerConfig config) {
    // Proxy everything except bypass hosts.
    client.findProxy = (Uri uri) {
      final host = uri.host;
      for (final pattern in config.bypassHosts) {
        if (pattern is RegExp) {
          if (pattern.hasMatch(host)) return 'DIRECT';
        } else {
          if (host == pattern.toString()) return 'DIRECT';
        }
      }
      return 'PROXY ${config.proxyHostPort}';
    };

    if (config.allowBadCertificates) {
      client.badCertificateCallback = (cert, host, port) => true;
    }
  }
}

class _ForwardProxyOverrides extends HttpOverrides {
  final HttpDebuggerConfig _config;

  _ForwardProxyOverrides(this._config);

  @override
  HttpClient createHttpClient(SecurityContext? context) {
    final client = super.createHttpClient(context);
    HttpDebugger._configureClient(client, _config);
    return client;
  }
}

class _ReverseProxyOverrides extends HttpOverrides {
  final HttpReverseProxyConfig _config;

  _ReverseProxyOverrides(this._config);

  @override
  HttpClient createHttpClient(SecurityContext? context) {
    return ReverseProxyHttpClient(
      upstreamBaseUrl: _config.upstreamBaseUrl,
      proxyBaseUrl: _config.proxyBaseUrl,
      proxyHttpPath: _config.proxyHttpPath,
      allowBadCertificates: _config.allowBadCertificates,
      skipPaths: _config.skipPaths,
      skipHosts: _config.skipHosts,
      skipMethods: _config.skipMethods,
      allowPaths: _config.allowPaths,
      allowHosts: _config.allowHosts,
      allowMethods: _config.allowMethods,
      context: context,
      // Important: the inner HttpClient must be created WITHOUT current overrides,
      // otherwise we get infinite recursion (stack overflow).
      innerClient: super.createHttpClient(context),
    );
  }
}

/// Technical wrapper to safely call `super.createHttpClient()` from a regular function
/// (this creates a "raw" HttpClient without applying current overrides).
class _RawHttpOverrides extends HttpOverrides {
  HttpClient createRawHttpClient(SecurityContext? context) {
    return super.createHttpClient(context);
  }
}

// Helper logic for ENV/defines - kept private here.
String? _readEnvVar(String name) {
  try {
    return Platform.environment[name];
  } catch (_) {
    return null;
  }
}

String? _firstNonEmpty(List<String?> values) {
  for (final v in values) {
    if (v != null && v.trim().isNotEmpty) return v;
  }
  return null;
}

bool _computeEnabledFromEnv() {
  final v = _firstNonEmpty([
    const String.fromEnvironment('DIO_DEBUGGER_ENABLED'),
    const String.fromEnvironment('HTTP_PROXY_ENABLED'),
    _readEnvVar('DIO_DEBUGGER_ENABLED'),
    _readEnvVar('HTTP_PROXY_ENABLED'),
  ]);
  if (v == null) return true;
  final sv = v.trim().toLowerCase();
  return sv == '1' || sv == 'true' || sv == 'yes' || sv == 'on';
}

String _computeMode() {
  final v = _firstNonEmpty([
    const String.fromEnvironment('HTTP_PROXY_MODE'),
    const String.fromEnvironment('PROXY_MODE'),
    _readEnvVar('HTTP_PROXY_MODE'),
    _readEnvVar('PROXY_MODE'),
  ])?.trim().toLowerCase();
  if (v == 'forward' || v == 'reverse' || v == 'none') return v!;
  return 'reverse';
}

bool _computeAllowBadCerts() {
  final v = _firstNonEmpty([
    const String.fromEnvironment('HTTP_PROXY_ALLOW_BAD_CERTS'),
    _readEnvVar('HTTP_PROXY_ALLOW_BAD_CERTS'),
  ])?.trim().toLowerCase();
  if (v == null) return false;
  return v == '1' || v == 'true' || v == 'yes' || v == 'on';
}

String _normalizeProxyHostPort(String proxy) {
  var p = proxy.trim();
  if (p.isEmpty) return p;
  if (p.startsWith('http://')) p = p.substring('http://'.length);
  if (p.startsWith('https://')) p = p.substring('https://'.length);
  if (p.endsWith(';')) p = p.substring(0, p.length - 1);
  return p;
}
