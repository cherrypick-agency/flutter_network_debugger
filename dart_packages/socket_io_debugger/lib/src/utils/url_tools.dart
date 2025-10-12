String ensureHttpScheme(String value) {
  var v = value.trim();
  if (v.isEmpty) return v;
  final lower = v.toLowerCase();
  if (lower.startsWith('http://') || lower.startsWith('https://')) return v;
  final portMatch = RegExp(r":(\d+)(?:/|\?|$)").firstMatch(v);
  if (portMatch != null && portMatch.group(1) == '443') return 'https://$v';
  return 'http://$v';
}

String buildEngineIoTarget(String baseUrl, String socketPath) {
  // ws(s) -> http(s), плюс EIO=4&transport=websocket
  var base = baseUrl.trim();
  final lower = base.toLowerCase();
  if (lower.startsWith('wss://')) base = 'https://' + base.substring(6);
  if (lower.startsWith('ws://')) base = 'http://' + base.substring(5);
  final path = socketPath.startsWith('/') ? socketPath : '/$socketPath';
  final uri = Uri.parse(base);
  final scheme = (uri.scheme == 'https') ? 'https' : 'http';
  final authority = uri.hasPort ? '${uri.host}:${uri.port}' : uri.host;
  final p = '$path?EIO=4&transport=websocket';
  return '$scheme://$authority$p';
}

String maybeAppendEngineIoPath(String path) {
  // Ничего не дописываем: путь прокси может быть без '/socket.io'
  return path;
}
