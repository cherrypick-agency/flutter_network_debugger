import 'package:dio/dio.dart';

/// Интерсептор, который переписывает исходные запросы на reverse‑proxy эндпоинт
/// вида: {proxyBaseUrl}{proxyHttpPath}?_target=<FULL_UPSTREAM_URL>
class ReverseProxyInterceptor extends Interceptor {
  ReverseProxyInterceptor({
    required String upstreamBaseUrl,
    required String proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    this.skipPaths,
    this.skipHosts,
    this.skipMethods,
    this.allowPaths,
    this.allowHosts,
    this.allowMethods,
  })  : _upstreamBaseUrl = upstreamBaseUrl,
        _proxyBaseUrl = _ensureHttpScheme(proxyBaseUrl),
        _proxyHttpPath =
            proxyHttpPath.startsWith('/') ? proxyHttpPath : '/$proxyHttpPath';

  final String _upstreamBaseUrl; // реальный upstream, куда хотим ходить
  final String _proxyBaseUrl; // адрес прокси (может быть без схемы в ENV)
  final String _proxyHttpPath; // путь на прокси, обычно /httpproxy

  // Фильтры: если allow* заданы, то проксируем только совпадающие, иначе —
  // если skip* заданы, то пропускаем совпадающие.
  final List<Pattern>? skipPaths;
  final List<Pattern>? skipHosts;
  final List<String>? skipMethods; // в верхнем регистре
  final List<Pattern>? allowPaths;
  final List<Pattern>? allowHosts;
  final List<String>? allowMethods; // в верхнем регистре

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    // Если уже идём через прокси endpoint — не трогаем, чтобы избежать двойной переписки
    if (options.path.startsWith(_proxyHttpPath) &&
        options.baseUrl.startsWith(_proxyBaseUrl)) {
      return handler.next(options);
    }

    // Определяем, следует ли обходить прокси по фильтрам
    if (_shouldBypassProxy(options)) {
      return handler.next(options);
    }

    // Целевой URL: если path абсолютный http(s), используем его как target;
    // иначе собираем из upstreamBaseUrl + path + query. Значения query нормализуем в строки.
    final Uri target = _buildTargetUri(options);

    options
      ..baseUrl = _proxyBaseUrl
      ..path = _proxyHttpPath
      ..queryParameters = {'_target': target.toString()};

    handler.next(options);
  }

  bool _shouldBypassProxy(RequestOptions options) {
    final method = options.method.toUpperCase();
    final path = options.path;
    // host может быть в options.uri, он уже комбинированный; если path абсолютный — берём из него
    final Uri effectiveUri =
        _isAbsoluteHttpUrl(path) ? Uri.parse(path) : options.uri;
    final host = effectiveUri.host;

    // Если указаны allow-листы — применяем их (должно совпасть хотя бы что-то)
    final hasAllow = (allowPaths?.isNotEmpty ?? false) ||
        (allowHosts?.isNotEmpty ?? false) ||
        (allowMethods?.isNotEmpty ?? false);
    if (hasAllow) {
      final okPath = allowPaths == null || _matchesAny(path, allowPaths!);
      final okHost = allowHosts == null || _matchesAny(host, allowHosts!);
      final okMethod = allowMethods == null || allowMethods!.contains(method);
      return !(okPath &&
          okHost &&
          okMethod); // если не прошёл allow — обойти прокси
    }

    // Иначе, если есть skip-листы — обходим прокси при совпадении любого
    final skipByPath = (skipPaths != null && _matchesAny(path, skipPaths!));
    final skipByHost = (skipHosts != null && _matchesAny(host, skipHosts!));
    final skipByMethod = (skipMethods != null && skipMethods!.contains(method));
    return skipByPath || skipByHost || skipByMethod;
  }

  Uri _buildTargetUri(RequestOptions options) {
    final path = options.path;
    if (_isAbsoluteHttpUrl(path)) {
      final original = Uri.parse(path);
      return _buildWithNormalizedQuery(
        base: original.replace(queryParameters: const {}),
        baseQuery: original.queryParameters,
        overrideQuery: options.queryParameters,
      );
    }

    final upstream = Uri.parse(_upstreamBaseUrl);
    final targetPath = _concatPaths(upstream.path, path);
    return _buildWithNormalizedQuery(
      base: upstream.replace(path: targetPath, queryParameters: const {}),
      baseQuery: upstream.queryParameters,
      overrideQuery: options.queryParameters,
    );
  }

  Uri _buildWithNormalizedQuery({
    required Uri base,
    required Map<String, String> baseQuery,
    required Map<String, dynamic> overrideQuery,
  }) {
    final qpAll = <String, List<String>>{};
    // из baseQuery (уже String -> String)
    baseQuery.forEach((k, v) => qpAll[k] = [v]);
    // из overrideQuery (dynamic)
    overrideQuery.forEach((k, v) {
      if (v == null) return;
      if (v is Iterable) {
        final list = <String>[];
        for (final item in v) {
          if (item == null) continue;
          list.add(item.toString());
        }
        if (list.isNotEmpty) qpAll[k] = list;
      } else {
        qpAll[k] = [v.toString()];
      }
    });

    // Сборка query строки вручную
    final qp = <String>[];
    qpAll.forEach((k, values) {
      for (final v in values) {
        qp.add('${Uri.encodeQueryComponent(k)}=${Uri.encodeQueryComponent(v)}');
      }
    });
    final query = qp.isEmpty ? '' : '?${qp.join('&')}';
    return Uri.parse('$base$query');
  }

  bool _isAbsoluteHttpUrl(String value) {
    final v = value.trim();
    if (v.startsWith('http://') || v.startsWith('https://')) return true;
    return false;
  }

  static String _concatPaths(String a, String b) {
    final left = a.endsWith('/') ? a.substring(0, a.length - 1) : a;
    final right = b.startsWith('/') ? b.substring(1) : b;
    if (left.isEmpty) return '/$right';
    if (right.isEmpty) return left.isEmpty ? '/' : left;
    return '$left/$right';
  }

  static String _ensureHttpScheme(String value) {
    final v = value.trim();
    if (v.isEmpty) return v;
    if (v.startsWith('http://') || v.startsWith('https://')) return v;
    final portMatch = RegExp(r":(\d+)$").firstMatch(v);
    if (portMatch != null && portMatch.group(1) == '443') {
      return 'https://$v';
    }
    return 'http://$v';
  }

  bool _matchesAny(String value, List<Pattern> patterns) {
    for (final p in patterns) {
      if (p is RegExp) {
        if (p.hasMatch(value)) return true;
      } else {
        // String/Pattern: простая проверка вхождения
        if (value.contains(p)) return true;
      }
    }
    return false;
  }
}
