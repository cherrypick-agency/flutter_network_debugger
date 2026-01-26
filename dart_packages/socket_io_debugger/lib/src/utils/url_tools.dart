String ensureHttpScheme(String value) {
  var v = value.trim();
  if (v.isEmpty) return v;
  final lower = v.toLowerCase();
  if (lower.startsWith('http://') || lower.startsWith('https://')) return v;
  final portMatch = RegExp(r":(\d+)(?:/|\?|$)").firstMatch(v);
  if (portMatch != null && portMatch.group(1) == '443') return 'https://$v';
  return 'http://$v';
}

String buildEngineIoTarget(String baseUrl, String path) {
  // ws(s) -> http(s), plus EIO=4&transport=websocket
  var base = baseUrl.trim();
  final lower = base.toLowerCase();
  if (lower.startsWith('wss://')) base = 'https://' + base.substring(6);
  if (lower.startsWith('ws://')) base = 'http://' + base.substring(5);
  final pathWithSlash = path.startsWith('/') ? path : '/$path';
  final uri = Uri.parse(base);
  final scheme = (uri.scheme == 'https') ? 'https' : 'http';
  // socket_io_client uses uri.port even if port is not specified (returns 0 in that case),
  // so it's important to always explicitly set 80/443.
  final port = uri.hasPort ? uri.port : (scheme == 'https' ? 443 : 80);
  final authority = '${uri.host}:$port';
  final p = '$pathWithSlash?EIO=4&transport=websocket';
  return '$scheme://$authority$p';
}

String maybeAppendEngineIoPath(String path) {
  // Don't append anything: proxy path may not include '/socket.io'
  return path;
}
