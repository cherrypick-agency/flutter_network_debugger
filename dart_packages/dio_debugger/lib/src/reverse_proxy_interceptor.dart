import 'package:dio/dio.dart';

/// Interceptor that rewrites original requests to a reverse-proxy endpoint
/// of the form: {proxyBaseUrl}{proxyHttpPath}?_target=<FULL_UPSTREAM_URL>
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
    this.resetCaptureOnFirstRequest = false,
  })  : _upstreamBaseUrl = upstreamBaseUrl,
        _proxyBaseUrl = _ensureHttpScheme(proxyBaseUrl),
        _proxyHttpPath =
            proxyHttpPath.startsWith('/') ? proxyHttpPath : '/$proxyHttpPath';

  final String
      _upstreamBaseUrl; // actual upstream where we want to send requests
  final String _proxyBaseUrl; // proxy address (may be without scheme in ENV)
  final String _proxyHttpPath; // path on proxy, usually /httpproxy

  // Filters: if allow* are set, we proxy only matching ones, otherwise -
  // if skip* are set, we skip matching ones.
  final List<Pattern>? skipPaths;
  final List<Pattern>? skipHosts;
  final List<String>? skipMethods; // in upper case
  final List<Pattern>? allowPaths;
  final List<Pattern>? allowHosts;
  final List<String>? allowMethods; // in upper case

  // Reset capture: if true, adds _resetCapture=true to the first request
  final bool resetCaptureOnFirstRequest;
  bool _isFirstRequest = true;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    // If already going through proxy endpoint - don't touch to avoid double rewriting
    if (options.path.startsWith(_proxyHttpPath) &&
        options.baseUrl.startsWith(_proxyBaseUrl)) {
      return handler.next(options);
    }

    // Determine whether to bypass proxy based on filters
    if (_shouldBypassProxy(options)) {
      return handler.next(options);
    }

    // Target URL: if path is absolute http(s), use it as target;
    // otherwise build from upstreamBaseUrl + path + query. Normalize query values to strings.
    final Uri target = _buildTargetUri(options);

    // Build query parameters for proxy
    final proxyQueryParams = <String, String>{'_target': target.toString()};

    // Add _resetCapture=true for the first request if required
    if (resetCaptureOnFirstRequest && _isFirstRequest) {
      proxyQueryParams['_resetCapture'] = 'true';
      _isFirstRequest = false;
    }

    options
      ..baseUrl = _proxyBaseUrl
      ..path = _proxyHttpPath
      ..queryParameters = proxyQueryParams;

    handler.next(options);
  }

  bool _shouldBypassProxy(RequestOptions options) {
    final method = options.method.toUpperCase();
    final path = options.path;
    // host can be in options.uri, it's already combined; if path is absolute - take from it
    final Uri effectiveUri =
        _isAbsoluteHttpUrl(path) ? Uri.parse(path) : options.uri;
    final host = effectiveUri.host;

    // If allow-lists are specified - apply them (must match at least something)
    final hasAllow = (allowPaths?.isNotEmpty ?? false) ||
        (allowHosts?.isNotEmpty ?? false) ||
        (allowMethods?.isNotEmpty ?? false);
    if (hasAllow) {
      final okPath = allowPaths == null || _matchesAny(path, allowPaths!);
      final okHost = allowHosts == null || _matchesAny(host, allowHosts!);
      final okMethod = allowMethods == null || allowMethods!.contains(method);
      return !(okPath &&
          okHost &&
          okMethod); // if didn't pass allow - bypass proxy
    }

    // Otherwise, if there are skip-lists - bypass proxy when any matches
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
    // from baseQuery (already String -> String)
    baseQuery.forEach((k, v) {
      final key = k.startsWith('?') ? k.substring(1) : k;
      qpAll[key] = [v];
    });
    // from overrideQuery (dynamic)
    overrideQuery.forEach((k, v) {
      if (v == null) return;
      final kk = k.toString();
      final key = kk.startsWith('?') ? kk.substring(1) : kk;
      if (v is Iterable) {
        final list = <String>[];
        for (final item in v) {
          if (item == null) continue;
          list.add(item.toString());
        }
        if (list.isNotEmpty) qpAll[key] = list;
      } else {
        qpAll[key] = [v.toString()];
      }
    });

    // Build query string manually, then safely substitute via Uri.replace
    final parts = <String>[];
    qpAll.forEach((k, values) {
      for (final v in values) {
        parts.add(
            '${Uri.encodeQueryComponent(k)}=${Uri.encodeQueryComponent(v)}');
      }
    });
    final q = parts.join('&');
    // Remove possible existing query from base and substitute new one
    final cleanBase = base.replace(query: null, queryParameters: null);
    return cleanBase.replace(query: q);
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
        // String/Pattern: simple containment check
        if (value.contains(p)) return true;
      }
    }
    return false;
  }
}
