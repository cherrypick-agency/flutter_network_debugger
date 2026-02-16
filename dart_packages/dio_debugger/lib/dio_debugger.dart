/// Library for enabling HTTP request debugging through a proxy server.
///
/// Supports two modes of operation:
/// - Reverse proxy: intercepts requests and routes them through the proxy server
///   for viewing and debugging
/// - Forward proxy: uses the system proxy for request routing
///
/// Settings can be configured via environment variables or method parameters.
library dio_debugger;

import 'package:dio/dio.dart';

import 'package:dio_debugger/src/reverse_proxy_interceptor.dart';
export 'package:dio_debugger/src/reverse_proxy_interceptor.dart';

// Conditional import: on Web we use a stub, on IO - we read OS ENV
import 'package:dio_debugger/src/env/env_reader_stub.dart'
    if (dart.library.io) 'package:dio_debugger/src/env/env_reader_io.dart';
import 'package:dio_debugger/src/forward_proxy_stub.dart'
    if (dart.library.io) 'package:dio_debugger/src/forward_proxy_io.dart';
import 'package:dio_debugger/src/platform_stub.dart'
    if (dart.library.io) 'package:dio_debugger/src/platform_io.dart'
    as platform;

// Compile-time values from --dart-define, during Flutter/Dart build
const String _kDefineUpstream = String.fromEnvironment('UPSTREAM_BASE_URL');
const String _kDefineApiHost = String.fromEnvironment('API_HOST');
const String _kDefineProxy = String.fromEnvironment('PROXY_BASE_URL');
const String _kDefineHttpProxy = String.fromEnvironment('HTTP_PROXY');
const String _kDefineProxyPath = String.fromEnvironment('PROXY_HTTP_PATH');
const String _kDefineHttpProxyPath = String.fromEnvironment('HTTP_PROXY_PATH');
const String _kDefineDioDebuggerEnabled =
    String.fromEnvironment('DIO_DEBUGGER_ENABLED');
const String _kDefineHttpProxyEnabled =
    String.fromEnvironment('HTTP_PROXY_ENABLED');
const String _kDefineHttpProxyMode = String.fromEnvironment('HTTP_PROXY_MODE');
const String _kDefineProxyMode = String.fromEnvironment('PROXY_MODE');
const String _kDefineHttpProxyAllowBadCerts =
    String.fromEnvironment('HTTP_PROXY_ALLOW_BAD_CERTS');

/// Simple utility for attaching reverse-proxy to an existing Dio instance.
/// By default, it tries to get settings from ENV variables (via Platform.environment):
///   - UPSTREAM_BASE_URL (example: https://dev.api.padelme.app)
///   - PROXY_BASE_URL (example: http://localhost:9091 or localhost:9091)
///   - PROXY_HTTP_PATH (example: /httpproxy)
/// Can be explicitly overridden via arguments.
class DioDebugger {
  DioDebugger._();

  /// Attaches the reverse-proxy interceptor to [dio]. Returns the same [dio] for chaining.
  ///
  /// [resetCaptureOnHotRestart] - if true, sends a request to the proxy to clear
  /// previous sessions and create a new capture ID. Useful for separating
  /// hot restarts during development.
  static Dio attach(
    Dio dio, {
    String? upstreamBaseUrl,
    String? proxyBaseUrl,
    String? proxyHttpPath,
    bool? enabled,
    bool insertFirst = true,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
    bool resetCaptureOnHotRestart = false,
  }) {
    final enabledEffective = enabled ?? _computeEnabledFromEnv();
    if (!enabledEffective) return dio;

    // Upstream source in priority order:
    // 1) explicit argument upstreamBaseUrl
    // 2) dio.options.baseUrl (if not empty)
    // 3) --dart-define / ENV
    String upstream = upstreamBaseUrl ?? '';
    if (upstream.trim().isEmpty) {
      final fromDio = dio.options.baseUrl;
      if (fromDio.trim().isNotEmpty) {
        upstream = fromDio;
      }
    }
    if (upstream.trim().isEmpty) {
      upstream = _firstNonEmpty([
            _kDefineUpstream,
            _kDefineApiHost,
            readEnvVar('UPSTREAM_BASE_URL'),
            readEnvVar('API_HOST'),
          ]) ??
          '';
    }

    final proxy = proxyBaseUrl ??
        _firstNonEmpty([
          _kDefineProxy,
          _kDefineHttpProxy,
          readEnvVar('PROXY_BASE_URL'),
          readEnvVar('HTTP_PROXY'),
        ]) ??
        _getDefaultProxyUrl();

    final path = (proxyHttpPath ??
        _firstNonEmpty([
          _kDefineProxyPath,
          _kDefineHttpProxyPath,
          readEnvVar('PROXY_HTTP_PATH'),
          readEnvVar('HTTP_PROXY_PATH'),
        ]) ??
        '/httpproxy');

    final mode = _computeMode(); // none | reverse | forward
    if (mode == 'forward') {
      final p = proxy.isEmpty
          ? _firstNonEmpty([_kDefineHttpProxy, readEnvVar('HTTP_PROXY')]) ?? ''
          : proxy;
      if (p.isEmpty) return dio;
      final allowBadCerts = _computeAllowBadCerts();
      final normalized = _normalizeProxy(p);
      return forwardProxyAttach(dio,
          proxyHostPort: normalized, allowBadCerts: allowBadCerts);
    }

    if (mode == 'reverse') {
      if (upstream.isEmpty || proxy.isEmpty) return dio;

      // Remove old interceptor if exists (for hot restart support)
      dio.interceptors.removeWhere((i) => i is ReverseProxyInterceptor);

      // Create new interceptor
      final interceptor = ReverseProxyInterceptor(
        upstreamBaseUrl: upstream,
        proxyBaseUrl: proxy,
        proxyHttpPath: path,
        skipPaths: skipPaths,
        skipHosts: skipHosts,
        skipMethods: _upper(skipMethods),
        allowPaths: allowPaths,
        allowHosts: allowHosts,
        allowMethods: _upper(allowMethods),
        resetCaptureOnFirstRequest: resetCaptureOnHotRestart,
      );

      if (insertFirst) {
        dio.interceptors.insert(0, interceptor);
      } else {
        dio.interceptors.add(interceptor);
      }

      return dio;
    }

    return dio;
  }

  static String? _firstNonEmpty(List<String?> values) {
    for (final v in values) {
      if (v != null && v.trim().isNotEmpty) return v;
    }
    return null;
  }

  static bool _computeEnabledFromEnv() {
    final v = _firstNonEmpty([
      _kDefineDioDebuggerEnabled,
      _kDefineHttpProxyEnabled,
      readEnvVar('DIO_DEBUGGER_ENABLED'),
      readEnvVar('HTTP_PROXY_ENABLED'),
    ]);
    if (v == null) return true; // enabled by default in dev
    final sv = v.trim().toLowerCase();
    return sv == '1' || sv == 'true' || sv == 'yes' || sv == 'on';
  }

  static List<String>? _upper(List<String>? methods) {
    if (methods == null) return null;
    return methods.map((m) => m.toUpperCase()).toList(growable: false);
  }

  static String _computeMode() {
    final v = _firstNonEmpty([
      _kDefineHttpProxyMode,
      _kDefineProxyMode,
      readEnvVar('HTTP_PROXY_MODE'),
      readEnvVar('PROXY_MODE'),
    ])?.trim().toLowerCase();
    if (v == 'forward' || v == 'reverse' || v == 'none') return v!;
    return 'reverse';
  }

  static bool _computeAllowBadCerts() {
    final v = _firstNonEmpty([
      _kDefineHttpProxyAllowBadCerts,
      readEnvVar('HTTP_PROXY_ALLOW_BAD_CERTS'),
    ])?.trim().toLowerCase();
    if (v == null) return false;
    return v == '1' || v == 'true' || v == 'yes' || v == 'on';
  }

  static String _normalizeProxy(String proxy) {
    var p = proxy.trim();
    if (p.isEmpty) return p;
    if (p.startsWith('http://')) p = p.substring('http://'.length);
    if (p.startsWith('https://')) p = p.substring('https://'.length);
    if (p.endsWith(';')) p = p.substring(0, p.length - 1);
    return p;
  }

  /// Returns default proxy URL based on platform.
  /// Android emulator: 10.0.2.2 (special IP for host machine)
  /// Other platforms: localhost
  static String _getDefaultProxyUrl() {
    return platform.isAndroid
        ? 'http://10.0.2.2:9091'
        : 'http://localhost:9091';
  }
}
