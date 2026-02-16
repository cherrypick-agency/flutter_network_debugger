/// Ensures the string starts with http:// or https://.
///
/// If scheme is missing, adds http:// (or https:// for port 443).
String ensureHttpScheme(String value) {
  var v = value.trim();
  if (v.isEmpty) return v;
  final lower = v.toLowerCase();
  if (lower.startsWith('http://') || lower.startsWith('https://')) return v;
  final portMatch = RegExp(r":(\d+)(?:/|\?|$)").firstMatch(v);
  if (portMatch != null && portMatch.group(1) == '443') return 'https://$v';
  return 'http://$v';
}

/// Extracts host:port from URL.
///
/// If port is not specified, uses default 80 for http and 443 for https.
String hostPort(String url) {
  var u = url.trim();
  if (!u.startsWith('http://') && !u.startsWith('https://')) u = 'http://$u';
  final parsed = Uri.parse(u);
  final port =
      parsed.hasPort ? parsed.port : (parsed.scheme == 'https' ? 443 : 80);
  return '${parsed.host}:$port';
}

/// Normalizes proxy address to host:port format.
///
/// Removes http:// or https:// scheme and trailing characters.
String normalizeProxyHostPort(String proxy) {
  var p = proxy.trim();
  if (p.isEmpty) return p;
  if (p.startsWith('http://')) p = p.substring('http://'.length);
  if (p.startsWith('https://')) p = p.substring('https://'.length);
  if (p.endsWith(';')) p = p.substring(0, p.length - 1);
  return p; // host:port
}

/// Builds target URL for Engine.IO proxy.
///
/// Converts ws(s) to http(s) and adds EIO=4&transport=websocket params.
String buildEngineIoTarget(String baseUrl, String socketPath) {
  // ws(s) -> http(s), plus EIO=4&transport=websocket as the minimum set for WS proxy.
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

/// Converts HTTP URL to WebSocket URL.
///
/// Replaces http:// with ws:// and https:// with wss://.
String buildWsTarget(String baseUrl) {
  var base = baseUrl.trim();
  if (base.isEmpty) return base;
  final lower = base.toLowerCase();
  if (lower.startsWith('http://')) return 'ws://' + base.substring(7);
  if (lower.startsWith('https://')) return 'wss://' + base.substring(8);
  return base; // assume ws:// or wss:// is already passed
}

/// Ensures the string starts with ws:// or wss://.
///
/// If scheme is missing, adds ws:// (or wss:// for port 443).
String ensureWsScheme(String value) {
  var v = value.trim();
  if (v.isEmpty) return v;
  final lower = v.toLowerCase();
  if (lower.startsWith('ws://') || lower.startsWith('wss://')) return v;
  if (lower.startsWith('http://')) return 'ws://' + v.substring(7);
  if (lower.startsWith('https://')) return 'wss://' + v.substring(8);
  final portMatch = RegExp(r":(\d+)(?:/|\?|$)").firstMatch(v);
  if (portMatch != null && portMatch.group(1) == '443') return 'wss://$v';
  return 'ws://$v';
}
