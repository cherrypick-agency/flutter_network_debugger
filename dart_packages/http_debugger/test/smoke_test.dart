import 'dart:io';

import 'package:http_debugger/http_debugger.dart';
import 'package:test/test.dart';

void main() {
  test('enableReverseProxy sets HttpOverrides.global and disable restores it',
      () {
    final prev = HttpOverrides.current;
    addTearDown(HttpDebugger.disable);

    HttpDebugger.enableReverseProxy(
      const HttpReverseProxyConfig(
        upstreamBaseUrl: 'https://api.example.test',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/httpproxy',
      ),
    );

    expect(HttpOverrides.current, isNot(prev));

    HttpDebugger.disable();
    expect(HttpOverrides.current, prev);
  });
}
