library socket_io_debugger;

import 'dart:developer' as developer;

import 'package:socket_io_debugger/src/env/env_reader_stub.dart'
    if (dart.library.io) 'package:socket_io_debugger/src/env/env_reader_io.dart';
import 'package:socket_io_debugger/src/forward_proxy_stub.dart'
    if (dart.library.io) 'package:socket_io_debugger/src/forward_proxy_io.dart';
import 'package:socket_io_debugger/src/io_types_io.dart'
    if (dart.library.html) 'package:socket_io_debugger/src/io_types_stub.dart';
import 'package:socket_io_debugger/src/platform_io.dart'
    if (dart.library.html) 'package:socket_io_debugger/src/platform_stub.dart'
    as platform;
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
const String _kDefineDioDebuggerEnabled =
    String.fromEnvironment('DIO_DEBUGGER_ENABLED');
const String _kDefineSocketProxyAllowBadCerts =
    String.fromEnvironment('SOCKET_PROXY_ALLOW_BAD_CERTS');

void _debugLog(String message) {
  assert(() {
    developer.log(message, name: 'socket_io_debugger');
    return true;
  }());
}

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
    String path = '/socket.io/',
    String? proxyBaseUrl,
    String? proxyHttpPath,
    bool? enabled,
  }) {
    final enabledEffective = enabled ?? _computeEnabledFromEnv();
    if (!enabledEffective) {
      return SocketIoConfig(
        effectiveBaseUrl: baseUrl,
        effectivePath: path,
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
        (platform.isAndroid ? 'http://10.0.2.2:9091' : 'http://localhost:9091');
    final proxyPath = (proxyHttpPath ??
        _firstNonEmpty([
          _kDefineSocketProxyPath,
          readEnvVar('SOCKET_PROXY_PATH'),
        ]) ??
        '/wsproxy');
    final effectiveProxyPath = _normalizeProxyPath(proxyPath);

    _debugLog(
      '[SocketIoDebugger] mode=$mode baseUrl=$baseUrl path=$path proxy=$proxy proxyPath=$effectiveProxyPath',
    );

    if (mode == 'forward') {
      if (proxy.isEmpty) {
        return SocketIoConfig(
          effectiveBaseUrl: baseUrl,
          effectivePath: path,
          query: const {},
          useForwardOverrides: false,
        );
      }
      final allowBadCerts = _computeAllowBadCerts();
      return forwardProxyAttach(
        baseUrl: baseUrl,
        path: path,
        proxyHostPort: _normalizeProxy(proxy),
        allowBadCerts: allowBadCerts,
      );
    }

    if (mode == 'reverse') {
      if (proxy.isEmpty) {
        return SocketIoConfig(
          effectiveBaseUrl: baseUrl,
          effectivePath: path,
          query: const {},
          useForwardOverrides: false,
        );
      }
      final proxyBase = ensureHttpScheme(proxy);
      final effectivePath = effectiveProxyPath;
      // Preserve namespace from original baseUrl (e.g. '/chat')
      final srcNsPath = Uri.tryParse(ensureHttpScheme(baseUrl))?.path ?? '';
      final effectiveBaseWithNs =
          proxyBase + (srcNsPath.isNotEmpty ? srcNsPath : '');
      // Target engine.io URL (ws->http)
      // If baseUrl points to the proxy itself (host:port match), try to get real upstream from defines/ENV
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
      // If passed path belongs to proxy (contains wsproxy), try to get upstream path/target from ENV
      var upstreamPath = path;
      // Full target from ENV takes priority
      final explicitTarget = _firstNonEmpty([
        _kDefineSocketUpstreamTarget,
        readEnvVar('SOCKET_UPSTREAM_TARGET'),
      ]);
      if (explicitTarget != null && explicitTarget.trim().isNotEmpty) {
        final t = explicitTarget.trim();
        _debugLog('[SocketIoDebugger] reverse (explicit target): $t');
        return SocketIoConfig(
          effectiveBaseUrl: effectiveBaseWithNs,
          effectivePath: effectivePath,
          query: {'_target': t},
          useForwardOverrides: false,
        );
      }
      if (path.contains('wsproxy')) {
        upstreamPath = _firstNonEmpty([
              _kDefineSocketUpstreamPath,
              readEnvVar('SOCKET_UPSTREAM_PATH'),
            ]) ??
            '/socket.io/';
      }
      final target = buildEngineIoTarget(upstream, upstreamPath);
      _debugLog(
        '[SocketIoDebugger] reverse: proxyBase=$proxyBase effectiveBaseUrl=$effectiveBaseWithNs effectivePath=$effectivePath upstream=$upstream upstreamPath=$upstreamPath target=$target',
      );
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
      effectivePath: path,
      query: const {},
      useForwardOverrides: false,
    );
  }

  static bool _computeEnabledFromEnv() {
    final v = _firstNonEmpty([
      _kDefineSocketProxyEnabled,
      _kDefineDioDebuggerEnabled,
      readEnvVar('SOCKET_PROXY_ENABLED'),
      readEnvVar('DIO_DEBUGGER_ENABLED'),
    ]);
    if (v == null) return true; // enabled by default in dev
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
      _kDefineSocketProxyAllowBadCerts,
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

  // Protect against invalid proxyPath values: strip scheme/host, trim after '?',
  // ensure leading '/'. This prevents sequences like "?%3F_target=..." in requests.
  static String _normalizeProxyPath(String value) {
    var v = value.trim();
    if (v.isEmpty) return '/wsproxy';
    if (v.startsWith('http://') ||
        v.startsWith('https://') ||
        v.startsWith('ws://') ||
        v.startsWith('wss://')) {
      final u = Uri.parse(v);
      v = u.path;
    }
    final q = v.indexOf('?');
    if (q >= 0) v = v.substring(0, q);
    if (!v.startsWith('/')) v = '/$v';
    while (v.contains('//')) {
      v = v.replaceAll('//', '/');
    }
    return v.isEmpty ? '/wsproxy' : v;
  }
}
