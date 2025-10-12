library dio_debugger;

import 'package:dio/dio.dart';

import 'package:dio_debugger/src/reverse_proxy_interceptor.dart';
export 'package:dio_debugger/src/reverse_proxy_interceptor.dart';

// Условный импорт: на Web используем заглушку, на IO — читаем OS ENV
import 'package:dio_debugger/src/env/env_reader_stub.dart'
    if (dart.library.io) 'package:dio_debugger/src/env/env_reader_io.dart';
import 'package:dio_debugger/src/forward_proxy_stub.dart'
    if (dart.library.io) 'package:dio_debugger/src/forward_proxy_io.dart';

// Значения compile-time из --dart-define, при сборке Flutter/Dart
const String _kDefineUpstream = String.fromEnvironment('UPSTREAM_BASE_URL');
const String _kDefineApiHost = String.fromEnvironment('API_HOST');
const String _kDefineProxy = String.fromEnvironment('PROXY_BASE_URL');
const String _kDefineHttpProxy = String.fromEnvironment('HTTP_PROXY');
const String _kDefineProxyPath = String.fromEnvironment('PROXY_HTTP_PATH');
const String _kDefineHttpProxyPath = String.fromEnvironment('HTTP_PROXY_PATH');

/// Простая утилита для привязки reverse‑proxy к существующему Dio экземпляру.
/// По умолчанию пытается взять настройки из ENV переменных (через Platform.environment):
///   - UPSTREAM_BASE_URL (пример: https://dev.api.padelme.app)
///   - PROXY_BASE_URL (пример: http://localhost:9091 или localhost:9091)
///   - PROXY_HTTP_PATH (пример: /httpproxy)
/// Можно явно переопределить через аргументы.
class DioDebugger {
  DioDebugger._();

  /// Подключает интерсептор reverse‑proxy к [dio]. Возвращает тот же [dio] для чейнинга.
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
  }) {
    final enabledEffective = enabled ?? _computeEnabledFromEnv();
    if (!enabledEffective) return dio;

    final upstream = upstreamBaseUrl ??
        _firstNonEmpty([
          _kDefineUpstream,
          _kDefineApiHost,
          readEnvVar('UPSTREAM_BASE_URL'),
          readEnvVar('API_HOST'),
        ]) ??
        '';

    final proxy = proxyBaseUrl ??
        _firstNonEmpty([
          _kDefineProxy,
          _kDefineHttpProxy,
          readEnvVar('PROXY_BASE_URL'),
          readEnvVar('HTTP_PROXY'),
        ]) ??
        '';

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
      final alreadyAttached =
          dio.interceptors.any((i) => i is ReverseProxyInterceptor);
      if (!alreadyAttached) {
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
        );
        if (insertFirst) {
          dio.interceptors.insert(0, interceptor);
        } else {
          dio.interceptors.add(interceptor);
        }
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
      String.fromEnvironment('DIO_DEBUGGER_ENABLED'),
      String.fromEnvironment('HTTP_PROXY_ENABLED'),
      readEnvVar('DIO_DEBUGGER_ENABLED'),
      readEnvVar('HTTP_PROXY_ENABLED'),
    ]);
    if (v == null) return true; // по умолчанию включено в dev
    final sv = v.trim().toLowerCase();
    return sv == '1' || sv == 'true' || sv == 'yes' || sv == 'on';
  }

  static List<String>? _upper(List<String>? methods) {
    if (methods == null) return null;
    return methods.map((m) => m.toUpperCase()).toList(growable: false);
  }

  static String _computeMode() {
    final v = _firstNonEmpty([
      String.fromEnvironment('HTTP_PROXY_MODE'),
      String.fromEnvironment('PROXY_MODE'),
      readEnvVar('HTTP_PROXY_MODE'),
      readEnvVar('PROXY_MODE'),
    ])?.trim().toLowerCase();
    if (v == 'forward' || v == 'reverse' || v == 'none') return v!;
    return 'reverse';
  }

  static bool _computeAllowBadCerts() {
    final v = _firstNonEmpty([
      String.fromEnvironment('HTTP_PROXY_ALLOW_BAD_CERTS'),
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
}
