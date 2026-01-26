library web_socket_debugger;

import 'dart:developer' as developer;

import 'package:web_socket/web_socket.dart' as ws;

import 'src/env/env_reader_stub.dart'
    if (dart.library.io) 'src/env/env_reader_io.dart';
import 'src/forward_proxy_stub.dart'
    if (dart.library.io) 'src/forward_proxy_io.dart';
import 'src/utils/url_tools.dart';
import 'src/ws_connector_stub.dart'
    if (dart.library.io) 'src/ws_connector_io.dart';

const String _kDefineProxy = String.fromEnvironment('SOCKET_PROXY');
const String _kDefineProxyPath = String.fromEnvironment('SOCKET_PROXY_PATH');
const String _kDefineMode = String.fromEnvironment('SOCKET_PROXY_MODE');
const String _kDefineEnabled = String.fromEnvironment('SOCKET_PROXY_ENABLED');
const String _kDefineAllowBadCerts =
    String.fromEnvironment('SOCKET_PROXY_ALLOW_BAD_CERTS');
const String _kDefineUpstreamUrl =
    String.fromEnvironment('SOCKET_UPSTREAM_URL');
const String _kDefineUpstreamTarget =
    String.fromEnvironment('SOCKET_UPSTREAM_TARGET');

void _debugLog(String message) {
  assert(() {
    developer.log(message, name: 'web_socket_debugger');
    return true;
  }());
}

class WebSocketProxyConfig {
  const WebSocketProxyConfig({
    required this.connectUrl,
    required this.query,
    required this.useForwardOverrides,
    this.httpClientFactory,
  });

  final Uri connectUrl;
  final Map<String, dynamic> query;
  final bool useForwardOverrides;
  final Object Function()? httpClientFactory;
}

class WebSocketDebugger {
  WebSocketDebugger._();

  static WebSocketProxyConfig attach({
    required String baseUrl,
    String proxyBaseUrl = 'http://localhost:9091',
    String proxyPath = '/wsproxy',
    bool? enabled,
    String? mode,
  }) {
    final enabledEffective = enabled ?? _computeEnabledFromEnv();
    if (!enabledEffective) {
      return WebSocketProxyConfig(
        connectUrl: Uri.parse(baseUrl),
        query: const {},
        useForwardOverrides: false,
      );
    }

    final modeEffective = (mode ?? _computeMode());
    final proxy = proxyBaseUrl.isNotEmpty
        ? proxyBaseUrl
        : (_firstNonEmpty([_kDefineProxy, readEnvVar('SOCKET_PROXY')]) ?? '');
    final rawPath = proxyPath.isNotEmpty
        ? proxyPath
        : (_firstNonEmpty(
                [_kDefineProxyPath, readEnvVar('SOCKET_PROXY_PATH')]) ??
            '/wsproxy');
    final path = _normalizeProxyPath(rawPath);

    _debugLog(
      '[WebSocketDebugger] mode=$modeEffective base=$baseUrl proxy=$proxy path=$path',
    );

    if (modeEffective == 'forward') {
      if (proxy.isEmpty) {
        return WebSocketProxyConfig(
          connectUrl: Uri.parse(baseUrl),
          query: const {},
          useForwardOverrides: false,
        );
      }
      final allowBad = _computeAllowBadCerts();
      final proxyHostPort = normalizeProxyHostPort(proxy);
      return forwardProxyAttach(
        baseUrl: baseUrl,
        proxyHostPort: proxyHostPort,
        allowBadCerts: allowBad,
      );
    }

    if (modeEffective == 'reverse') {
      if (proxy.isEmpty) {
        return WebSocketProxyConfig(
          connectUrl: Uri.parse(baseUrl),
          query: const {},
          useForwardOverrides: false,
        );
      }
      final proxyHttp = ensureHttpScheme(proxy);

      // If baseUrl points to the proxy itself, try to get real upstream from ENV/define
      var upstream = baseUrl;
      if (hostPort(proxyHttp) == hostPort(ensureHttpScheme(upstream))) {
        final envUp = _firstNonEmpty(
            [_kDefineUpstreamUrl, readEnvVar('SOCKET_UPSTREAM_URL')]);
        if (envUp != null && envUp.trim().isNotEmpty) upstream = envUp;
      }

      final explicitTarget = _firstNonEmpty(
          [_kDefineUpstreamTarget, readEnvVar('SOCKET_UPSTREAM_TARGET')]);
      final target = explicitTarget?.trim().isNotEmpty == true
          ? explicitTarget!.trim()
          : buildWsTarget(upstream);

      final uri = Uri.parse(proxyHttp);
      final wsScheme = uri.scheme == 'https' ? 'wss' : 'ws';
      final effective = uri.replace(scheme: wsScheme, path: path);
      return WebSocketProxyConfig(
        connectUrl: effective,
        query: {'_target': target},
        useForwardOverrides: false,
      );
    }

    return WebSocketProxyConfig(
      connectUrl: Uri.parse(baseUrl),
      query: const {},
      useForwardOverrides: false,
    );
  }

  static Future<ws.WebSocket> connect({
    required WebSocketProxyConfig config,
    Map<String, dynamic>? headers,
  }) async {
    // Build query manually to preserve ws:// scheme and avoid adding '#'
    Uri uri;
    if (config.query.isEmpty) {
      uri = config.connectUrl;
    } else {
      final merged = <String, String>{
        ...config.connectUrl.queryParameters,
        ...config.query.map((k, v) => MapEntry(k, v.toString())),
      };
      final encodedQuery = Uri(queryParameters: merged).query;
      uri = config.connectUrl.replace(query: encodedQuery, fragment: null);
    }
    // In IO environment pass headers (Authorization/Cookie etc.),
    // on web headers are ignored at transport level
    return connectWS(uri, headers: headers);
  }

  static bool _computeEnabledFromEnv() {
    final v =
        _firstNonEmpty([_kDefineEnabled, readEnvVar('SOCKET_PROXY_ENABLED')]);
    if (v == null) return true;
    final sv = v.trim().toLowerCase();
    return sv == '1' || sv == 'true' || sv == 'yes' || sv == 'on';
  }

  static String _computeMode() {
    final v = _firstNonEmpty([_kDefineMode, readEnvVar('SOCKET_PROXY_MODE')])
        ?.trim()
        .toLowerCase();
    if (v == 'forward' || v == 'reverse' || v == 'none') return v!;
    return 'reverse';
  }

  static bool _computeAllowBadCerts() {
    final v = _firstNonEmpty(
            [_kDefineAllowBadCerts, readEnvVar('SOCKET_PROXY_ALLOW_BAD_CERTS')])
        ?.trim()
        .toLowerCase();
    if (v == null) return false;
    return v == '1' || v == 'true' || v == 'yes' || v == 'on';
  }

  static String? _firstNonEmpty(List<String?> values) {
    for (final v in values) {
      if (v != null && v.trim().isNotEmpty) return v;
    }
    return null;
  }

  // Protect against invalid proxy path values: strip scheme/host, trim everything after '?',
  // ensure leading '/'. Without this the final URL may contain sequences like '%3F_target'.
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
