String buildWsUrl(String base, String path) {
  var b = base;
  if (b.startsWith('http://')) {
    b = 'ws://${b.substring(7)}';
  }
  if (b.startsWith('https://')) {
    b = 'wss://${b.substring(8)}';
  }
  if (!path.startsWith('/')) path = '/$path';
  if (b.endsWith('/')) {
    b = b.substring(0, b.length - 1);
  }
  return '$b$path';
}
