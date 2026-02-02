import 'dart:io';

/// Wrapper over HttpClient that rewrites requests to
/// reverse-proxy endpoint of the form:
///   {proxyBaseUrl}{proxyHttpPath}?_target=<FULL_UPSTREAM_URL>
class ReverseProxyHttpClient implements HttpClient {
  ReverseProxyHttpClient({
    required String upstreamBaseUrl,
    required String proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    bool allowBadCertificates = false,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
    SecurityContext? context,
    HttpClient? innerClient,
  })  : _inner = innerClient ?? HttpClient(context: context),
        _upstreamBaseUrl = upstreamBaseUrl,
        _proxyBaseUrl = _ensureHttpScheme(proxyBaseUrl),
        _proxyHttpPath =
            proxyHttpPath.startsWith('/') ? proxyHttpPath : '/$proxyHttpPath',
        _skipPaths = skipPaths,
        _skipHosts = skipHosts,
        _skipMethods = _upper(skipMethods),
        _allowPaths = allowPaths,
        _allowHosts = allowHosts,
        _allowMethods = _upper(allowMethods) {
    if (allowBadCertificates) {
      _inner.badCertificateCallback = (cert, host, port) => true;
    }
  }

  final HttpClient _inner;

  final String _upstreamBaseUrl;
  final String _proxyBaseUrl;
  final String _proxyHttpPath;

  final List<Pattern>? _skipPaths;
  final List<Pattern>? _skipHosts;
  final List<String>? _skipMethods; // uppercase
  final List<Pattern>? _allowPaths;
  final List<Pattern>? _allowHosts;
  final List<String>? _allowMethods; // uppercase

  @override
  Future<HttpClientRequest> open(
    String method,
    String host,
    int port,
    String path,
  ) {
    final scheme = port == 443 ? 'https' : 'http';
    return openUrl(
      method,
      Uri(scheme: scheme, host: host, port: port, path: path),
    );
  }

  @override
  Future<HttpClientRequest> openUrl(String method, Uri url) {
    if (!_isHttpScheme(url.scheme)) {
      return _inner.openUrl(method, url);
    }
    if (_isAlreadyProxy(url)) {
      return _inner.openUrl(method, url);
    }
    if (_shouldBypass(method, url)) {
      return _inner.openUrl(method, url);
    }
    final Uri? target = _buildTarget(url);
    if (target == null) {
      return _inner.openUrl(method, url);
    }
    final Uri proxyUri = _buildProxyUri(target);
    return _inner.openUrl(method, proxyUri);
  }

  @override
  Future<HttpClientRequest> get(String host, int port, String path) =>
      open('GET', host, port, path);

  @override
  Future<HttpClientRequest> getUrl(Uri url) => openUrl('GET', url);

  @override
  Future<HttpClientRequest> post(String host, int port, String path) =>
      open('POST', host, port, path);

  @override
  Future<HttpClientRequest> postUrl(Uri url) => openUrl('POST', url);

  @override
  Future<HttpClientRequest> put(String host, int port, String path) =>
      open('PUT', host, port, path);

  @override
  Future<HttpClientRequest> putUrl(Uri url) => openUrl('PUT', url);

  @override
  Future<HttpClientRequest> delete(String host, int port, String path) =>
      open('DELETE', host, port, path);

  @override
  Future<HttpClientRequest> deleteUrl(Uri url) => openUrl('DELETE', url);

  @override
  Future<HttpClientRequest> patch(String host, int port, String path) =>
      open('PATCH', host, port, path);

  @override
  Future<HttpClientRequest> patchUrl(Uri url) => openUrl('PATCH', url);

  @override
  Future<HttpClientRequest> head(String host, int port, String path) =>
      open('HEAD', host, port, path);

  @override
  Future<HttpClientRequest> headUrl(Uri url) => openUrl('HEAD', url);

  // Delegate other properties/methods
  @override
  set authenticate(
    Future<bool> Function(Uri url, String scheme, String? realm)? f,
  ) {
    _inner.authenticate = f;
  }

  @override
  void addCredentials(
    Uri url,
    String realm,
    HttpClientCredentials credentials,
  ) =>
      _inner.addCredentials(url, realm, credentials);

  @override
  set connectionFactory(
    Future<ConnectionTask<Socket>> Function(
      Uri url,
      String? proxyHost,
      int? proxyPort,
    )? f,
  ) {
    _inner.connectionFactory = f;
  }

  @override
  set findProxy(String Function(Uri url)? f) {
    _inner.findProxy = f;
  }

  @override
  set authenticateProxy(
    Future<bool> Function(String host, int port, String scheme, String? realm)?
        f,
  ) {
    _inner.authenticateProxy = f;
  }

  @override
  void addProxyCredentials(
    String host,
    int port,
    String realm,
    HttpClientCredentials credentials,
  ) =>
      _inner.addProxyCredentials(host, port, realm, credentials);

  @override
  set badCertificateCallback(
    bool Function(X509Certificate cert, String host, int port)? callback,
  ) {
    _inner.badCertificateCallback = callback;
  }

  @override
  set keyLog(Function(String line)? callback) {
    _inner.keyLog = callback;
  }

  @override
  void close({bool force = false}) => _inner.close(force: force);

  @override
  Duration get idleTimeout => _inner.idleTimeout;
  @override
  set idleTimeout(Duration v) => _inner.idleTimeout = v;

  @override
  Duration? get connectionTimeout => _inner.connectionTimeout;
  @override
  set connectionTimeout(Duration? v) => _inner.connectionTimeout = v;

  @override
  int? get maxConnectionsPerHost => _inner.maxConnectionsPerHost;
  @override
  set maxConnectionsPerHost(int? v) => _inner.maxConnectionsPerHost = v;

  @override
  bool get autoUncompress => _inner.autoUncompress;
  @override
  set autoUncompress(bool v) => _inner.autoUncompress = v;

  @override
  String? get userAgent => _inner.userAgent;
  @override
  set userAgent(String? v) => _inner.userAgent = v;

  bool _shouldBypass(String method, Uri url) {
    final m = method.toUpperCase();
    final host = url.host;
    final path = url.path;

    final hasAllow = (_allowPaths?.isNotEmpty ?? false) ||
        (_allowHosts?.isNotEmpty ?? false) ||
        (_allowMethods?.isNotEmpty ?? false);

    if (hasAllow) {
      final lp = _allowPaths;
      final lh = _allowHosts;
      final lm = _allowMethods;
      final okPath = lp == null ? true : _matchesAny(path, lp);
      final okHost = lh == null ? true : _matchesAny(host, lh);
      final okMethod = lm == null ? true : lm.contains(m);
      return !(okPath && okHost && okMethod);
    }

    final sp = _skipPaths;
    final sh = _skipHosts;
    final sm = _skipMethods;
    final skipByPath = sp == null ? false : _matchesAny(path, sp);
    final skipByHost = sh == null ? false : _matchesAny(host, sh);
    final skipByMethod = sm == null ? false : sm.contains(m);
    return skipByPath || skipByHost || skipByMethod;
  }

  bool _isAlreadyProxy(Uri url) {
    final proxy = Uri.parse(_proxyBaseUrl);
    if (url.scheme != proxy.scheme) return false;
    if (url.host != proxy.host) return false;
    final urlPort = url.hasPort ? url.port : (url.scheme == 'https' ? 443 : 80);
    final proxyPort =
        proxy.hasPort ? proxy.port : (proxy.scheme == 'https' ? 443 : 80);
    if (urlPort != proxyPort) return false;
    return url.path.startsWith(_proxyHttpPath);
  }

  Uri? _buildTarget(Uri original) {
    if (_isHttpScheme(original.scheme) && original.hasAuthority) {
      return original;
    }
    if (_upstreamBaseUrl.trim().isEmpty) {
      return null;
    }
    final upstream = Uri.parse(_upstreamBaseUrl);
    final mergedPath = _concatPaths(upstream.path, original.path);
    final mergedQuery = _mergeQuery(upstream, original);
    return upstream.replace(path: mergedPath, query: mergedQuery);
  }

  static String _mergeQuery(Uri upstream, Uri original) {
    final merged = <String, List<String>>{};

    for (final e in upstream.queryParametersAll.entries) {
      final k = e.key.startsWith('?') ? e.key.substring(1) : e.key;
      merged[k] = e.value;
    }
    for (final e in original.queryParametersAll.entries) {
      final k = e.key.startsWith('?') ? e.key.substring(1) : e.key;
      merged[k] = e.value;
    }

    final parts = <String>[];
    for (final e in merged.entries) {
      for (final v in e.value) {
        parts.add(
          '${Uri.encodeQueryComponent(e.key)}=${Uri.encodeQueryComponent(v)}',
        );
      }
    }
    return parts.join('&');
  }

  Uri _buildProxyUri(Uri target) {
    final base = Uri.parse(_proxyBaseUrl);
    return Uri(
      scheme: base.scheme,
      host: base.host,
      port: base.hasPort ? base.port : null,
      path: _proxyHttpPath,
      queryParameters: {'_target': target.toString()},
    );
  }

  static bool _isHttpScheme(String scheme) =>
      scheme == 'http' || scheme == 'https';

  static String _ensureHttpScheme(String value) {
    final v = value.trim();
    if (v.isEmpty) return v;
    if (v.startsWith('http://') || v.startsWith('https://')) return v;
    final portMatch = RegExp(r":(\d+)$").firstMatch(v);
    if (portMatch != null && portMatch.group(1) == '443') return 'https://$v';
    return 'http://$v';
  }

  static List<String>? _upper(List<String>? methods) {
    if (methods == null) return null;
    return methods.map((e) => e.toUpperCase()).toList(growable: false);
  }

  static bool _matchesAny(String value, List<Pattern> patterns) {
    for (final p in patterns) {
      if (p is RegExp) {
        if (p.hasMatch(value)) return true;
      } else {
        if (value.contains(p)) return true;
      }
    }
    return false;
  }

  static String _concatPaths(String a, String b) {
    final left = a.endsWith('/') ? a.substring(0, a.length - 1) : a;
    final right = b.startsWith('/') ? b.substring(1) : b;
    if (left.isEmpty) return '/$right';
    if (right.isEmpty) return left.isEmpty ? '/' : left;
    return '$left/$right';
  }
}
