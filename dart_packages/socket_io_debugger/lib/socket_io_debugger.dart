library socket_io_debugger;

import 'dart:io';

import 'package:socket_io_debugger/src/env/env_reader_stub.dart'
    if (dart.library.io) 'package:socket_io_debugger/src/env/env_reader_io.dart';
import 'package:socket_io_debugger/src/forward_proxy_stub.dart'
    if (dart.library.io) 'package:socket_io_debugger/src/forward_proxy_io.dart';
import 'package:socket_io_debugger/src/utils/url_tools.dart';

const String _kDefineSocketProxy = String.fromEnvironment('SOCKET_PROXY');
const String _kDefineSocketProxyPath =
    String.fromEnvironment('SOCKET_PROXY_PATH');
const String _kDefineSocketUpstream =
    String.fromEnvironment('SOCKET_UPSTREAM_URL');
const String _kDefineSocketUpstreamPath =
    String.fromEnvironment('SOCKET_UPSTREAM_PATH');
const String _kDefineSocketUpstreamTarget =
    String.fromEnvironment('SOCKET_UPSTREAM_TARGET');
const String _kDefineSocketProxyEnabled =
    String.fromEnvironment('SOCKET_PROXY_ENABLED');
const String _kDefineProxyModeGeneric =
    String.fromEnvironment('SOCKET_PROXY_MODE');

class SocketIoConfig {
  const SocketIoConfig({
    required this.effectiveBaseUrl,
    required this.effectivePath,
    required this.query,
    required this.useForwardOverrides,
    this.httpClientFactory,
  });

  final String effectiveBaseUrl;
  final String effectivePath;
  final Map<String, dynamic> query;
  final bool useForwardOverrides;
  final HttpClient Function()? httpClientFactory;
}

class SocketIoDebugger {
  SocketIoDebugger._();

  static SocketIoConfig attach({
    required String baseUrl,
    required String socketPath,
    String? proxyBaseUrl,
    String? proxyHttpPath,
    bool? enabled,
  }) {
    final enabledEffective = enabled ?? _computeEnabledFromEnv();
    if (!enabledEffective) {
      return SocketIoConfig(
        effectiveBaseUrl: baseUrl,
        effectivePath: socketPath,
        query: const {},
        useForwardOverrides: false,
      );
    }

    final mode = _computeMode(); // reverse | forward | none
    final proxy = proxyBaseUrl ??
        _firstNonEmpty([
          _kDefineSocketProxy,
          readEnvVar('SOCKET_PROXY'),
        ]) ??
        '';
    final path = (proxyHttpPath ??
        _firstNonEmpty([
          _kDefineSocketProxyPath,
          readEnvVar('SOCKET_PROXY_PATH'),
        ]) ??
        '/wsproxy');

    // Диагностика (dev): печатаем ключевые параметры
    // ignore: avoid_print
    print(
        '[SocketIoDebugger] mode=$mode baseUrl=$baseUrl socketPath=$socketPath proxy=$proxy path=$path');

    if (mode == 'forward') {
      if (proxy.isEmpty) {
        return SocketIoConfig(
          effectiveBaseUrl: baseUrl,
          effectivePath: socketPath,
          query: const {},
          useForwardOverrides: false,
        );
      }
      final allowBadCerts = _computeAllowBadCerts();
      return forwardProxyAttach(
        baseUrl: baseUrl,
        socketPath: socketPath,
        proxyHostPort: _normalizeProxy(proxy),
        allowBadCerts: allowBadCerts,
      );
    }

    if (mode == 'reverse') {
      if (proxy.isEmpty) {
        return SocketIoConfig(
          effectiveBaseUrl: baseUrl,
          effectivePath: socketPath,
          query: const {},
          useForwardOverrides: false,
        );
      }
      final proxyBase = ensureHttpScheme(proxy);
      final effectivePath = path; // используем ровно указанный путь прокси
      // Сохраняем namespace из исходного baseUrl (например, '/chat')
      final srcNsPath = Uri.tryParse(ensureHttpScheme(baseUrl))?.path ?? '';
      final effectiveBaseWithNs =
          proxyBase + (srcNsPath.isNotEmpty ? srcNsPath : '');
      // целевой engine.io URL (ws->http)
      // Если baseUrl указывает на сам proxy (host:port совпадает) — пытаемся взять реальный upstream из defines/ENV
      var upstream = baseUrl;
      final proxyHost = _hostPort(proxyBase);
      final upstreamHost = _hostPort(ensureHttpScheme(upstream));
      if (proxyHost.isNotEmpty && upstreamHost == proxyHost) {
        final envUpstream = _firstNonEmpty([
          _kDefineSocketUpstream,
          readEnvVar('SOCKET_UPSTREAM_URL'),
        ]);
        if (envUpstream != null && envUpstream.trim().isNotEmpty) {
          upstream = envUpstream;
        }
      }
      // Если переданный socketPath относится к прокси (содержит wsproxy), пытаемся взять апстрим path/target из ENV
      var upstreamSocketPath = socketPath;
      // Полный target из ENV имеет приоритет
      final explicitTarget = _firstNonEmpty([
        _kDefineSocketUpstreamTarget,
        readEnvVar('SOCKET_UPSTREAM_TARGET'),
      ]);
      if (explicitTarget != null && explicitTarget.trim().isNotEmpty) {
        final t = explicitTarget.trim();
        // ignore: avoid_print
        print('[SocketIoDebugger] reverse (explicit target): $t');
        return SocketIoConfig(
          effectiveBaseUrl: effectiveBaseWithNs,
          effectivePath: effectivePath,
          query: {'_target': t},
          useForwardOverrides: false,
        );
      }
      if (socketPath.contains('wsproxy')) {
        upstreamSocketPath = _firstNonEmpty([
              _kDefineSocketUpstreamPath,
              readEnvVar('SOCKET_UPSTREAM_PATH'),
            ]) ??
            '/socket.io';
      }
      final target = buildEngineIoTarget(upstream, upstreamSocketPath);
      // ignore: avoid_print
      print(
          '[SocketIoDebugger] reverse: proxyBase=$proxyBase effectiveBaseUrl=$effectiveBaseWithNs effectivePath=$effectivePath upstream=$upstream upstreamPath=$upstreamSocketPath target=$target');
      return SocketIoConfig(
        effectiveBaseUrl: effectiveBaseWithNs,
        effectivePath: effectivePath,
        query: {'_target': target},
        useForwardOverrides: false,
      );
    }

    // none
    return SocketIoConfig(
      effectiveBaseUrl: baseUrl,
      effectivePath: socketPath,
      query: const {},
      useForwardOverrides: false,
    );
  }

  static bool _computeEnabledFromEnv() {
    final v = _firstNonEmpty([
      _kDefineSocketProxyEnabled,
      String.fromEnvironment('DIO_DEBUGGER_ENABLED'),
      readEnvVar('SOCKET_PROXY_ENABLED'),
      readEnvVar('DIO_DEBUGGER_ENABLED'),
    ]);
    if (v == null) return true; // включено в dev по умолчанию
    final sv = v.trim().toLowerCase();
    return sv == '1' || sv == 'true' || sv == 'yes' || sv == 'on';
  }

  static String _computeMode() {
    final v = _firstNonEmpty([
      _kDefineProxyModeGeneric,
      readEnvVar('SOCKET_PROXY_MODE'),
    ])?.trim().toLowerCase();
    if (v == 'forward' || v == 'reverse' || v == 'none') return v!;
    return 'reverse';
  }

  static bool _computeAllowBadCerts() {
    final v = _firstNonEmpty([
      String.fromEnvironment('SOCKET_PROXY_ALLOW_BAD_CERTS'),
      readEnvVar('SOCKET_PROXY_ALLOW_BAD_CERTS'),
    ])?.trim().toLowerCase();
    if (v == null) return false;
    return v == '1' || v == 'true' || v == 'yes' || v == 'on';
  }

  static String? _firstNonEmpty(List<String?> values) {
    for (final v in values) {
      if (v != null && v.trim().isNotEmpty) return v;
    }
    return null;
  }

  static String _normalizeProxy(String proxy) {
    var p = proxy.trim();
    if (p.isEmpty) return p;
    if (p.startsWith('http://')) p = p.substring('http://'.length);
    if (p.startsWith('https://')) p = p.substring('https://'.length);
    if (p.endsWith(';')) p = p.substring(0, p.length - 1);
    return p; // host:port
  }

  static String _hostPort(String url) {
    var u = url.trim();
    if (!u.startsWith('http://') && !u.startsWith('https://')) u = 'http://$u';
    final parsed = Uri.parse(u);
    final port =
        parsed.hasPort ? parsed.port : (parsed.scheme == 'https' ? 443 : 80);
    return '${parsed.host}:$port';
  }
}
