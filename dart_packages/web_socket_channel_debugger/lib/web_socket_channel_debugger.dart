library web_socket_channel_debugger;

import 'dart:developer' as developer;

import 'package:web_socket_channel/web_socket_channel.dart' as wsc;
import 'package:web_socket_channel_debugger/src/wsc_connector_stub.dart'
    if (dart.library.io) 'package:web_socket_channel_debugger/src/wsc_connector_io.dart';

import 'src/env/env_reader_stub.dart'
    if (dart.library.io) 'src/env/env_reader_io.dart';
import 'src/forward_proxy_stub.dart'
    if (dart.library.io) 'src/forward_proxy_io.dart';
import 'src/utils/url_tools.dart';

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
    developer.log(message, name: 'web_socket_channel_debugger');
    return true;
  }());
}

class WscProxyConfig {
  const WscProxyConfig({
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

class WebSocketChannelDebugger {
  WebSocketChannelDebugger._();

  static WscProxyConfig attach({
    required String baseUrl,
    String proxyBaseUrl = 'http://localhost:9091',
    String proxyPath = '/wsproxy',
    bool? enabled,
    String? mode,
  }) {
    final enabledEffective = enabled ?? _computeEnabledFromEnv();
    if (!enabledEffective) {
      return WscProxyConfig(
        connectUrl: Uri.parse(baseUrl),
        query: const {},
        useForwardOverrides: false,
      );
    }

    final modeEffective = (mode ?? _computeMode());
    final proxy = proxyBaseUrl.isNotEmpty
        ? proxyBaseUrl
        : (_firstNonEmpty([_kDefineProxy, readEnvVar('SOCKET_PROXY')]) ?? '');
    // Normalize proxy path: strip scheme/host, trailing query, ensure leading '/'
    final rawPath = proxyPath.isNotEmpty
        ? proxyPath
        : (_firstNonEmpty(
                [_kDefineProxyPath, readEnvVar('SOCKET_PROXY_PATH')]) ??
            '/wsproxy');
    final path = _normalizeProxyPath(rawPath);

    _debugLog(
      '[WscDebugger] mode=$modeEffective base=$baseUrl proxy=$proxy path=$path',
    );

    if (modeEffective == 'forward') {
      if (proxy.isEmpty) {
        return WscProxyConfig(
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
        return WscProxyConfig(
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
          ? ensureWsScheme(explicitTarget!.trim())
          : ensureWsScheme(upstream);

      final uri = Uri.parse(proxyHttp);
      final wsScheme = uri.scheme == 'https' ? 'wss' : 'ws';
      // Build final proxy connection URL without junk characters in path
      final effective =
          uri.replace(scheme: wsScheme, path: path, queryParameters: null);
      _debugLog('[WscDebugger] effective URL (proxy): $effective');
      _debugLog('[WscDebugger] target (upstream): $target');
      return WscProxyConfig(
        connectUrl: effective,
        query: {'_target': target},
        useForwardOverrides: false,
      );
    }

    return WscProxyConfig(
      connectUrl: Uri.parse(baseUrl),
      query: const {},
      useForwardOverrides: false,
    );
  }

  static wsc.WebSocketChannel connect({
    required WscProxyConfig config,
    Map<String, dynamic>? headers,
  }) {
    Uri uri;
    if (config.query.isEmpty) {
      uri = config.connectUrl;
    } else {
      // Build URI manually to preserve ws:// scheme
      final allParams = <String, String>{
        ...config.connectUrl.queryParameters,
        ...config.query.map((k, v) => MapEntry(k, v.toString())),
      };
      // Use Uri constructor directly
      uri = Uri(
        scheme: config.connectUrl.scheme,
        userInfo:
            config.connectUrl.hasAuthority ? config.connectUrl.userInfo : '',
        host: config.connectUrl.host,
        port: config.connectUrl.hasPort ? config.connectUrl.port : null,
        path: config.connectUrl.path,
        queryParameters: allParams.isNotEmpty ? allParams : null,
      );
    }
    // Pass headers (IO only), ignored on web - see connector
    _debugLog('[WscDebugger] Final URI for connection: $uri');
    _debugLog('[WscDebugger] Headers: ${headers?.keys.join(", ")}');
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

  // Protect against invalid proxyPath values: strip scheme/host, trim after '?',
  // ensure leading '/'. This prevents sequences like "?%3F_target=..." in requests.
  static String _normalizeProxyPath(String value) {
    var v = value.trim();
    if (v.isEmpty) return '/wsproxy';
    // If accidentally passed full URL, take only the path
    if (v.startsWith('http://') ||
        v.startsWith('https://') ||
        v.startsWith('ws://') ||
        v.startsWith('wss://')) {
      final u = Uri.parse(v);
      v = u.path;
    }
    // Trim everything after '?' (no query params should be in path)
    final q = v.indexOf('?');
    if (q >= 0) v = v.substring(0, q);
    if (!v.startsWith('/')) v = '/$v';
    // Simplify double slashes
    while (v.contains('//')) {
      v = v.replaceAll('//', '/');
    }
    return v.isEmpty ? '/wsproxy' : v;
  }
}
