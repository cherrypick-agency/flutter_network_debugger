import 'dart:io';
import 'reverse_http_client_io.dart';

/// Конфигурация глобального форвард‑прокси через HttpOverrides.
class HttpDebuggerConfig {
  /// host:port — то, что ждёт HttpClient.findProxy (без схемы)
  final String proxyHostPort;

  /// Разрешать самоподписанные/битые сертификаты — удобно в dev.
  final bool allowBadCertificates;

  /// Список хостов, для которых прокси отключается (DIRECT).
  /// Можно задавать как точные строки, так и RegExp.
  final List<Pattern> bypassHosts;

  const HttpDebuggerConfig({
    required this.proxyHostPort,
    this.allowBadCertificates = false,
    this.bypassHosts = const [],
  });
}

/// Конфигурация reverse‑proxy для глобального перехвата через HttpOverrides.
class HttpReverseProxyConfig {
  final String upstreamBaseUrl;
  final String proxyBaseUrl; // может быть без схемы; нормализуем
  final String proxyHttpPath; // по умолчанию /httpproxy
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

/// Включает глобальный форвард‑прокси для всего HTTP‑трафика (dart:io).
/// Все клиенты, которые используют стандартный HttpClient (включая package:http, Dio по умолчанию и т.д.),
/// начнут ходить через заданный прокси.
class HttpDebugger {
  HttpDebugger._();

  static HttpOverrides? _previous;

  /// Включить глобальный прокси. Повторный вызов перезапишет настройки.
  static void enableForwardProxy(HttpDebuggerConfig config) {
    _previous ??= HttpOverrides.current;
    HttpOverrides.global = _ForwardProxyOverrides(config);
  }

  /// Включить глобальный reverse‑proxy.
  static void enableReverseProxy(HttpReverseProxyConfig config) {
    _previous ??= HttpOverrides.current;
    HttpOverrides.global = _ReverseProxyOverrides(config);
  }

  /// Отключить глобальный прокси и восстановить предыдущие overrides (если были).
  static void disable() {
    HttpOverrides.global = _previous;
    _previous = null;
  }

  /// Универсальный способ включить прокси.
  ///
  /// По умолчанию mode = 'reverse'. Если [upstreamBaseUrl] не задан —
  /// автоматически используем forward‑proxy с локальным proxy.
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
    final proxy =
        (proxyBaseUrl == null || proxyBaseUrl.trim().isEmpty)
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
        final client = HttpClient(context: context);
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
      final proxy =
          _firstNonEmpty([
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
      final upstream =
          _firstNonEmpty([
            const String.fromEnvironment('UPSTREAM_BASE_URL'),
            const String.fromEnvironment('API_HOST'),
            _readEnvVar('UPSTREAM_BASE_URL'),
            _readEnvVar('API_HOST'),
          ]) ??
          '';
      var proxyBase =
          _firstNonEmpty([
            const String.fromEnvironment('PROXY_BASE_URL'),
            const String.fromEnvironment('HTTP_PROXY'),
            _readEnvVar('PROXY_BASE_URL'),
            _readEnvVar('HTTP_PROXY'),
          ]) ??
          (Platform.isAndroid
              ? 'http://10.0.2.2:9091'
              : 'http://localhost:9091');
      final proxyPath =
          _firstNonEmpty([
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
    );
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
  final v =
      _firstNonEmpty([
        const String.fromEnvironment('HTTP_PROXY_MODE'),
        const String.fromEnvironment('PROXY_MODE'),
        _readEnvVar('HTTP_PROXY_MODE'),
        _readEnvVar('PROXY_MODE'),
      ])?.trim().toLowerCase();
  if (v == 'forward' || v == 'reverse' || v == 'none') return v!;
  return 'reverse';
}

bool _computeAllowBadCerts() {
  final v =
      _firstNonEmpty([
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
