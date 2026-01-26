import 'dart:async';
import 'package:http/http.dart' as http;

/// Client wrapper for package:http that rewrites URL to reverse-proxy format:
///   {proxyBaseUrl}{proxyHttpPath}?_target=<FULL_UPSTREAM_URL>
/// Works on Web and IO since rewriting happens before sending the request.
class HttpReverseProxyClient extends http.BaseClient {
  HttpReverseProxyClient({
    required http.Client inner,
    required String proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    String? upstreamBaseUrl,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
  })  : _inner = inner,
        _proxyBase = _ensureHttpScheme(proxyBaseUrl),
        _proxyHttpPath =
            proxyHttpPath.startsWith('/') ? proxyHttpPath : '/$proxyHttpPath',
        _upstreamBase =
            (upstreamBaseUrl == null || upstreamBaseUrl.trim().isEmpty)
                ? null
                : Uri.parse(upstreamBaseUrl),
        _skipPaths = skipPaths,
        _skipHosts = skipHosts,
        _skipMethods =
            skipMethods?.map((m) => m.toUpperCase()).toList(growable: false),
        _allowPaths = allowPaths,
        _allowHosts = allowHosts,
        _allowMethods =
            allowMethods?.map((m) => m.toUpperCase()).toList(growable: false);

  final http.Client _inner;
  final String _proxyBase;
  final String _proxyHttpPath;
  final Uri? _upstreamBase;

  final List<Pattern>? _skipPaths;
  final List<Pattern>? _skipHosts;
  final List<String>? _skipMethods;
  final List<Pattern>? _allowPaths;
  final List<Pattern>? _allowHosts;
  final List<String>? _allowMethods;

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) async {
    final method = request.method.toUpperCase();
    final url = request.url;

    if (_isAlreadyProxy(url) || _shouldBypass(method, url)) {
      return _inner.send(request);
    }

    final Uri? target = _buildTarget(url);
    if (target == null) {
      return _inner.send(request);
    }

    final Uri proxy = _buildProxyUri(target);

    // Base case: construct new Request, copy body and headers
    final bodyBytes = await request.finalize().toBytes();
    final rq = http.Request(method, proxy)
      ..headers.addAll(request.headers)
      ..followRedirects = request.followRedirects
      ..maxRedirects = request.maxRedirects
      ..persistentConnection = request.persistentConnection
      ..bodyBytes = bodyBytes;
    return _inner.send(rq);
  }

  bool _shouldBypass(String method, Uri url) {
    final host = url.host;
    final path = url.path;

    final hasAllow = (_allowPaths?.isNotEmpty ?? false) ||
        (_allowHosts?.isNotEmpty ?? false) ||
        (_allowMethods?.isNotEmpty ?? false);
    if (hasAllow) {
      final lp = _allowPaths, lh = _allowHosts, lm = _allowMethods;
      final okPath = lp == null ? true : _matchesAny(path, lp);
      final okHost = lh == null ? true : _matchesAny(host, lh);
      final okMethod = lm == null ? true : lm.contains(method);
      return !(okPath && okHost && okMethod);
    }

    final sp = _skipPaths, sh = _skipHosts, sm = _skipMethods;
    final skipByPath = sp == null ? false : _matchesAny(path, sp);
    final skipByHost = sh == null ? false : _matchesAny(host, sh);
    final skipByMethod = sm == null ? false : sm.contains(method);
    return skipByPath || skipByHost || skipByMethod;
  }

  bool _isAlreadyProxy(Uri url) {
    final base = Uri.parse(_proxyBase);
    if (url.scheme != base.scheme || url.host != base.host) return false;
    final urlPort = url.hasPort ? url.port : (url.scheme == 'https' ? 443 : 80);
    final basePort =
        base.hasPort ? base.port : (base.scheme == 'https' ? 443 : 80);
    if (urlPort != basePort) return false;
    return url.path.startsWith(_proxyHttpPath);
  }

  Uri? _buildTarget(Uri original) {
    if (_isHttp(original)) return original;
    final base = _upstreamBase;
    if (base == null) return null;
    if (original.path.isEmpty && original.query.isEmpty) return base;
    final mergedPath = _concatPaths(base.path, original.path);
    return base.replace(path: mergedPath, query: original.query);
  }

  Uri _buildProxyUri(Uri target) {
    final base = Uri.parse(_proxyBase);
    return Uri(
      scheme: base.scheme,
      host: base.host,
      port: base.hasPort ? base.port : null,
      path: _proxyHttpPath,
      queryParameters: {'_target': target.toString()},
    );
  }

  static bool _isHttp(Uri u) => u.scheme == 'http' || u.scheme == 'https';

  static String _ensureHttpScheme(String v) {
    final s = v.trim();
    if (s.isEmpty) return s;
    if (s.startsWith('http://') || s.startsWith('https://')) return s;
    final m = RegExp(r":(\d+)$").firstMatch(s);
    if (m != null && m.group(1) == '443') return 'https://$s';
    return 'http://$s';
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

/// Utility for convenient client creation.
class HttpDebuggerClient {
  static http.Client wrap(
    http.Client inner, {
    required String proxyBaseUrl,
    String proxyHttpPath = '/httpproxy',
    String? upstreamBaseUrl,
    List<Pattern>? skipPaths,
    List<Pattern>? skipHosts,
    List<String>? skipMethods,
    List<Pattern>? allowPaths,
    List<Pattern>? allowHosts,
    List<String>? allowMethods,
  }) {
    return HttpReverseProxyClient(
      inner: inner,
      proxyBaseUrl: proxyBaseUrl,
      proxyHttpPath: proxyHttpPath,
      upstreamBaseUrl: upstreamBaseUrl,
      skipPaths: skipPaths,
      skipHosts: skipHosts,
      skipMethods: skipMethods,
      allowPaths: allowPaths,
      allowHosts: allowHosts,
      allowMethods: allowMethods,
    );
  }
}
