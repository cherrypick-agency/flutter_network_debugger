library web_socket_channel_debugger;

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
    // Нормализуем путь до proxy: убираем схему/хост, хвост после '?', добавляем ведущий '/'
    final rawPath = proxyPath.isNotEmpty
        ? proxyPath
        : (_firstNonEmpty(
                [_kDefineProxyPath, readEnvVar('SOCKET_PROXY_PATH')]) ??
            '/wsproxy');
    final path = _normalizeProxyPath(rawPath);

    // ignore: avoid_print
    print(
        '[WscDebugger] mode=$modeEffective base=$baseUrl proxy=$proxy path=$path');

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
      // Если baseUrl указывает на сам proxy — попробуем взять реальный upstream из ENV/define
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
      // Собираем конечный URL подключения к прокси без мусорных символов в пути
      final effective = uri.replace(scheme: wsScheme, path: path, queryParameters: {});
      // ignore: avoid_print
      print('[WscDebugger] effective URL (should be proxy): $effective');
      // ignore: avoid_print
      print('[WscDebugger] target (upstream): $target');
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
      // Строим URI вручную, чтобы сохранить ws:// схему
      final allParams = <String, String>{
        ...config.connectUrl.queryParameters,
        ...config.query.map((k, v) => MapEntry(k, v.toString())),
      };
      // Используем Uri конструктор напрямую
      uri = Uri(
        scheme: config.connectUrl.scheme,
        userInfo: config.connectUrl.hasAuthority ? config.connectUrl.userInfo : '',
        host: config.connectUrl.host,
        port: config.connectUrl.hasPort ? config.connectUrl.port : null,
        path: config.connectUrl.path,
        queryParameters: allParams.isNotEmpty ? allParams : null,
      );
    }
    // Пробрасываем заголовки (в IO), на web игнорируем — см. коннектор
    // ignore: avoid_print
    print('[WscDebugger] Final URI for connection: $uri');
    // ignore: avoid_print
    print('[WscDebugger] Headers: ${headers?.keys.join(", ")}');
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

  // Страхуемся от кривых значений proxyPath: удаляем схему/хост, обрезаем всё после '?',
  // приводим к виду с ведущим '/'. Это предотвращает появление последовательностей вида
  // "?%3F_target=..." в запросах.
  static String _normalizeProxyPath(String value) {
    var v = value.trim();
    if (v.isEmpty) return '/wsproxy';
    // Если случайно передали полный URL – берём только путь
    if (v.startsWith('http://') ||
        v.startsWith('https://') ||
        v.startsWith('ws://') ||
        v.startsWith('wss://')) {
      final u = Uri.parse(v);
      v = u.path;
    }
    // Отрезаем всё после '?' (никаких query в path быть не должно)
    final q = v.indexOf('?');
    if (q >= 0) v = v.substring(0, q);
    if (!v.startsWith('/')) v = '/$v';
    // Упростим двойные слэши
    while (v.contains('//')) {
      v = v.replaceAll('//', '/');
    }
    return v.isEmpty ? '/wsproxy' : v;
  }
}
