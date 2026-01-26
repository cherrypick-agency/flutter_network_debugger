import 'dart:io';
import 'reverse_http_client_io.dart';

/// Global forward-proxy configuration via HttpOverrides.
class HttpDebuggerConfig {
  /// host:port - what HttpClient.findProxy expects (without scheme)
  final String proxyHostPort;

  /// Allow self-signed/bad certificates - useful in dev.
  final bool allowBadCertificates;

  /// List of hosts for which proxy is disabled (DIRECT).
  /// Can be specified as exact strings or RegExp.
  final List<Pattern> bypassHosts;

  const HttpDebuggerConfig({
    required this.proxyHostPort,
    this.allowBadCertificates = false,
    this.bypassHosts = const [],
  });
}

/// Reverse-proxy configuration for global interception via HttpOverrides.
class HttpReverseProxyConfig {
  final String upstreamBaseUrl;
  final String proxyBaseUrl; // can be without scheme; normalized
  final String proxyHttpPath; // default /httpproxy
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

/// Enables global forward-proxy for all HTTP traffic (dart:io).
/// All clients using standard HttpClient (including package:http, Dio by default, etc.)
/// will start routing through the specified proxy.
class HttpDebugger {
  HttpDebugger._();

  static HttpOverrides? _previous;

  /// Enable global proxy. Repeated calls will overwrite settings.
  static void enableForwardProxy(HttpDebuggerConfig config) {
    _previous ??= HttpOverrides.current;
    HttpOverrides.global = _ForwardProxyOverrides(config);
  }

  /// Enable global reverse-proxy.
  static void enableReverseProxy(HttpReverseProxyConfig config) {
    _previous ??= HttpOverrides.current;
    HttpOverrides.global = _ReverseProxyOverrides(config);
  }

  /// Disable global proxy and restore previous overrides (if any).
  static void disable() {
    HttpOverrides.global = _previous;
    _previous = null;
  }

  /// Universal way to enable proxy.
  ///
  /// By default mode = 'reverse'. If [upstreamBaseUrl] is not set,
  /// automatically uses forward-proxy with local proxy.
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

    // reverse (по умолчанию)
    final upstream = upstreamBaseUrl?.trim() ?? '';
    if (upstream.isEmpty) {
      // безопасный фоллбек в forward с дефолтным proxy
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

  /// Выполнить [action] в зоне с включённым прокси. Удобно для изолированного запуска.
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

  /// Выполнить [action] в зоне с включённым reverse‑proxy.
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
          // Важно: внутренний клиент создаём через super.createHttpClient,
          // иначе получаем бесконечную рекурсию из-за HttpOverrides.runZoned.
          innerClient: _RawHttpOverrides().createRawHttpClient(context),
        );
      },
    );
  }

  /// Автовыбор режима из ENV/--dart-define.
  ///
  /// HTTP_PROXY_MODE|PROXY_MODE = reverse|forward|none (по умолчанию reverse)
  /// DIO_DEBUGGER_ENABLED|HTTP_PROXY_ENABLED — вкл/выкл (по умолчанию true)
  /// Для reverse необходимы UPSTREAM_BASE_URL|API_HOST и PROXY_BASE_URL|HTTP_PROXY.
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
    // Проксируем всё, кроме bypass‑хостов.
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
      // Важно: внутренний HttpClient должен создаваться БЕЗ текущих overrides,
      // иначе получаем бесконечную рекурсию (stack overflow).
      innerClient: super.createHttpClient(context),
    );
  }
}

/// Техническая обёртка, чтобы из обычной функции можно было безопасно вызвать
/// `super.createHttpClient()` (это создаёт "сырой" HttpClient без применения
/// текущих overrides).
class _RawHttpOverrides extends HttpOverrides {
  HttpClient createRawHttpClient(SecurityContext? context) {
    return super.createHttpClient(context);
  }
}

// Вспомогательная логика для ENV/defines — держим приватно здесь.
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
