import 'package:flutter_test/flutter_test.dart';

String wsURL(String base, String path) {
  var b = base;
  if (b.startsWith('http://')) { b = 'ws://${b.substring(7)}'; }
  if (b.startsWith('https://')) { b = 'wss://${b.substring(8)}'; }
  if (!path.startsWith('/')) path = '/$path';
  if (b.endsWith('/')) { b = b.substring(0, b.length - 1); }
  return '$b$path';
}

void main() {
  test('wsURL converts http to ws and appends path correctly', () {
    expect(wsURL('http://localhost:8080', '/api/monitor/ws'), 'ws://localhost:8080/api/monitor/ws');
    expect(wsURL('http://localhost:8080/', 'wsproxy'), 'ws://localhost:8080/wsproxy');
    expect(wsURL('https://example.com', '/x'), 'wss://example.com/x');
  });
}


